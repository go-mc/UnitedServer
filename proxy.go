package main

import (
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/chat"
	mcnet "github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
)

func handleConn(c *mcnet.Conn) {
	loge := log.WithField("addr", c.Socket.RemoteAddr())
	defer c.Close()
	// handshake
	_, _, intention, err := recvHandshake(c)
	if err != nil {
		loge.WithError(err).Error("Handshake error")
	}
	switch intention {
	case 0x01:
		if err := Status(c); err != nil {
			loge.WithError(err).Error("Send status packet error")
		}
	case 0x02:
		// client login
		p, err := Login(c)
		if err != nil {
			loge.WithError(err).Error("Player login fail")
		}
		s, err := p.Connect("localhost:25565")
		if err != nil {
			loge.WithError(err).Error("Connect server error")
		}
		stop := make(chan struct{})
		stoped := p.JoinServer(stop, s)
		<-stoped
	default:
		loge.WithField("intention", intention).Error("Unknown intention in handshake")
		_ = c.WritePacket(pk.Marshal(0x00, chat.Message{Text: fmt.Sprintf("unknown intention 0x%x in handshake", intention)}))
	}
}

func recvHandshake(c *mcnet.Conn) (address pk.String, port pk.UnsignedShort, intention pk.Byte, err error) {
	var p pk.Packet
	if p, err = c.ReadPacket(); err != nil {
		return
	}
	if p.ID != 0x00 {
		err = errors.New("not a handshake packet")
		return
	}
	var version pk.VarInt
	if err = p.Scan(&version, &address, &port, &intention); err != nil {
		return
	}
	// check protocol version
	if version < ProtocolVersion {
		err = c.WritePacket(pk.Marshal(0x00, chat.Message{Translate: "multiplayer.disconnect.outdated_client"}))
	} else if version > ProtocolVersion {
		err = c.WritePacket(pk.Marshal(0x00, chat.Message{Translate: "multiplayer.disconnect.outdated_server"}))
	} else {
		return // all right
	}
	// version different
	if err != nil {
		err = fmt.Errorf("sending disconnect packet error: %w", err)
		return
	}
	err = errors.New("different protocol version")
	return
}

type Player struct {
	*mcnet.Conn
	Name string
}

type Server struct {
	sonn *mcnet.Conn
}

func (p *Player) Connect(serverAddr string) (*Server, error) {
	addr, portStr, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("look up port for %s error: %w", serverAddr, err)
	}
	port, err := strconv.ParseInt(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("port %s isn't a intiger: %w", portStr, err)
	}
	conn, err := mcnet.DialMC(serverAddr)
	if err != nil {
		return nil, err
	}
	//Handshake
	err = conn.WritePacket(
		//Handshake Packet
		pk.Marshal(
			0x00,                       //Handshake packet ID
			pk.VarInt(ProtocolVersion), //Protocol version
			pk.String(addr),            //Server's address
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
		//Recive Packet
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
			return &Server{sonn: conn}, nil //switches the connection state to PLAY.
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

// connect a player and server
// to stop this, close "stop chan"
// after completely stop, the returned chan will be closed.
func (p *Player) JoinServer(stop <-chan struct{}, s *Server) <-chan struct{} {
	ret := make(chan struct{})
	s1 := make(chan struct{})
	go func() {
		for {
			select {
			default:
				packet, err := s.sonn.ReadPacket()
				if err != nil {
					log.WithError(err).Error("recv target server packet error")
					return
				}
				if err := p.WritePacket(packet); err != nil {
					log.WithError(err).Error("send packet to client error")
				}
			case <-stop:
				s1 <- struct{}{}
				return
			}
		}
	}()
	go func() {
		for {
			select {
			default:
				packet, err := p.ReadPacket()
				if err != nil {
					log.WithError(err).Error("recv client packet error")
					return
				}
				if err := s.sonn.WritePacket(packet); err != nil {
					log.WithError(err).Error("send packet to target server error")
				}
			case <-stop:
				<-s1
				close(ret)
				return
			}
		}
	}()
	return ret
}
