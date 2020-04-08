package main

import (
	"errors"
	"fmt"
	mcnet "github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
)

func handleConn(c *mcnet.Conn) {
	p := Player{
		Conn: c,
		Name: "Tnze",
	}
	_, err := p.Connect("localhost:25565")
	if err != nil {
		log.WithError(err).Error("Connect server error")
	}
	//select{}
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
