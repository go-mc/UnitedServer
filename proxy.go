package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/chat"
	mcnet "github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/go-mc/UnitedServer/protocol"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

func handleConn(ctx context.Context, c *mcnet.Conn) {
	loge := log.WithField("addr", c.Socket.RemoteAddr())
	defer c.Close()
	// handshake
	version, _, _, intention, err := recvHandshake(c)
	if err != nil {
		loge.WithError(err).Error("Handshake error")
	}
	proto := protocol.GetProtocol(int(version))
	switch intention {
	case 0x01: // ping & list
		if err := Status(c, proto.VersionID()); err != nil {
			loge.WithError(err).Error("Send status packet error")
		}
	case 0x02: // client login
		// Check protocol
		if !proto.Support() {
			if err := c.WritePacket(pk.Marshal(0x00, chat.Text("Unsupported version!"))); err != nil {
				loge.WithError(err).Error("Write disconnect packet error")
			}
			return
		}
		// Login
		name, id, err := Login(c)
		if err != nil {
			loge.WithError(err).Error("Player login fail")
			return
		}
		defer counterDec() // decrease counter when player leave
		// StartGame
		loge = loge.WithField("player", name)
		defer loge.Info("Player left the game")
		(&Player{
			Conn: c,
			Name: name,
			UUID: id,
		}).Start(ctx, proto, loge)
	default:
		loge.WithField("intention", intention).Error("Unknown intention in handshake")
		_ = c.WritePacket(pk.Marshal(0x00, chat.Message{Text: fmt.Sprintf("unknown intention 0x%x in handshake", intention)}))
	}
}

func recvHandshake(c *mcnet.Conn) (version pk.VarInt, address pk.String, port pk.UnsignedShort, intention pk.Byte, err error) {
	var p pk.Packet
	if p, err = c.ReadPacket(); err != nil {
		return
	}
	if p.ID != 0x00 {
		err = errors.New("not a handshake packet")
		return
	}
	err = p.Scan(&version, &address, &port, &intention)
	return
}

type Player struct {
	*mcnet.Conn
	Name      string
	UUID      uuid.UUID
	Dimension int32
}

type Server struct {
	*mcnet.Conn
}

func (p *Player) Start(ctx context.Context, proto protocol.Protocol, loge *log.Entry) {
	server, err := p.connect(viper.GetString("LobbyServer"), proto)
	if err != nil {
		loge.WithError(err).Error("Connect server error")
		return
	}
	for {
		errChan := make(chan [2]error, 1)
		cmdChan := make(chan string)
		subCtx, cancel := context.WithCancel(ctx)
		go func(server *Server, err chan [2]error) {
			loge = loge.WithField("server", server.Socket.RemoteAddr())
			loge.Info("Player join server")
			errChan <- p.JoinServer(subCtx, server, proto.CmdInjector(ConnectCmdHandler(cmdChan)), proto.DimRecorder(&p.Dimension))()
			_ = server.Close()
			loge.Debug("Disconnect server")
		}(server, errChan)
	CmdLoop:
		for {
			select {
			case addr := <-cmdChan:
				_ = p.WritePacket(proto.SysChat(chat.Message{
					Text:  "[UnitedServer] Connecting " + addr,
					Color: "blue",
				}))
				secServer, err := p.connect(addr, proto)
				if err != nil {
					loge.WithField("server", addr).WithError(err).Error("Connect server error")
					_ = p.WritePacket(proto.SysChat(chat.Message{
						Text:  "[UnitedServer] Connect server error: " + err.Error(),
						Color: "red",
					}))
					break
				}
				cancel()
				<-errChan
				server = secServer
				p.SwitchTo(server, proto)
				break CmdLoop

			case errs := <-errChan:
				loge.WithField("errs", errs).Error("Transmit packets error")
				return
			case <-ctx.Done():
				_ = p.WritePacket(proto.Disconnect(chat.Message{Translate: "multiplayer.disconnect.server_shutdown"}))
				return
			}
		}
	}
}

func ConnectCmdHandler(cmdChan chan string) func(cmd string) (bool, error) {
	return func(cmd string) (bool, error) {
		if strings.HasPrefix(cmd, "/connect ") {
			cmdChan <- strings.TrimPrefix(cmd, "/connect ")
			return false, nil
		}
		return true, nil
	}
}

// transmit continued read packet from src, then write to dst.
// The middle func will be called for each packet before send to dst.
// The packet will be transmit only if middle func return pass==true.
func transmit(ctx context.Context, dst mcnet.Writer, src mcnet.Reader, middle middleFunc) error {
	for {
		select {
		default:
			packet, err := src.ReadPacket()
			if err != nil {
				return fmt.Errorf("recv packet error: %w", err)
			}
			if middle != nil {
				pass, err := middle(packet)
				if err != nil {
					return fmt.Errorf("middle func error: %w", err)
				} else if !pass {
					break // ignore this packet
				}
			}
			if err := dst.WritePacket(packet); err != nil {
				return fmt.Errorf("send packet error: %w", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (p *Player) connect(serverAddr string, proto protocol.Protocol) (*Server, error) {
	addr, portStr, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("look up port for %s error: %w", serverAddr, err)
	}
	port, err := strconv.ParseInt(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("port %s isn't a intiger: %w", portStr, err)
	}
	conn, err := mcnet.DialMC(serverAddr) // TODO: server whitelist and blacklist
	if err != nil {
		return nil, err
	}
	//Handshake
	err = conn.WritePacket(
		//Handshake Packet
		pk.Marshal(
			0x00,                         //Handshake packet ID
			pk.VarInt(proto.VersionID()), //Protocol version
			pk.String(addr),              //Server's address
			pk.UnsignedShort(port),
			pk.Byte(2),
		))
	if err != nil {
		return nil, fmt.Errorf("send handshake packect fail: %w", err)
	}
	//Login
	err = conn.WritePacket(
		//LoginStart Packet
		pk.Marshal(0, pk.String(p.Name)))
	if err != nil {
		return nil, fmt.Errorf("send login start packect fail: %w", err)
	}
	for {
		//Receive Packet
		var pack pk.Packet
		pack, err = conn.ReadPacket()
		if err != nil {
			return nil, fmt.Errorf("recv packet for Login fail: %w", err)
		}

		//Handle Packet
		switch pack.ID {
		case 0x00: //Disconnect
			var reason pk.String
			err = pack.Scan(&reason)
			if err != nil {
				err = fmt.Errorf("connect disconnected by server: %w",
					fmt.Errorf("read Disconnect message fail: %w", err))
			} else {
				err = fmt.Errorf("connect disconnected by server: %w", errors.New(string(reason)))
			}
			return nil, err
		case 0x01: //Encryption Request
			return nil, errors.New("this demo don't support encryption")
			//if err := handleEncryptionRequest(c, pack); err != nil {
			//	return fmt.Errorf("bot: encryption fail: %v", err)
			//}
		case 0x02: //Login Success
			// uuid, l := pk.UnpackString(pack.Data)
			// name, _ := unpackString(pack.Data[l:])
			return &Server{Conn: conn}, nil //switches the connection state to PLAY.
		case 0x03: //Set Compression
			var threshold pk.VarInt
			if err := pack.Scan(&threshold); err != nil {
				return nil, fmt.Errorf("set compression fail: %w", err)
			}
			conn.SetThreshold(int(threshold))
		case 0x04: //Login Plugin Request
			return nil, errors.New("this demo don't support login plugin request")
			//if err := handlePluginPacket(c, pack); err != nil {
			//	return fmt.Errorf("bot: handle plugin packet fail: %v", err)
			//}
		}
	}
}

type middleFunc func(packet pk.Packet) (pass bool, err error)

// connect a player and server
// to stop this, close "stop chan"
// after completely stop, the returned chan will be closed.
// TODO: online-mode support
func (p *Player) JoinServer(ctx context.Context, s *Server, middle1, middle2 middleFunc) (wait func() [2]error) {
	var wg sync.WaitGroup
	var errs [2]error
	wg.Add(2)
	go func() {
		defer wg.Done()
		errs[0] = transmit(ctx, s, p, middle1)
	}()
	go func() {
		defer wg.Done()
		errs[1] = transmit(ctx, p, s, middle2)
	}()
	return func() [2]error {
		wg.Wait()
		return errs
	}
}

func (p *Player) SwitchTo(s *Server, proto protocol.Protocol) {
	packet, err := s.ReadPacket()
	if err != nil {
		log.WithError(err).Error("Read JoinGame packet error")
		return
	}
	// returned respawn may be more then one
	respawn, dim, err := proto.ToRespawn(packet, atomic.LoadInt32(&p.Dimension))
	for _, v := range respawn {
		if err := p.WritePacket(v); err != nil {
			log.WithError(err).Error("Write Respawn packet error")
			return
		}
	}
	atomic.StoreInt32(&p.Dimension, dim)
}
