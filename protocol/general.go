package protocol

import (
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/chat"
	pk "github.com/Tnze/go-mc/net/packet"
	"sync/atomic"
)

const Latest = 578

type Protocol interface {
	VersionID() int
	Support() bool
	SysChat(msg chat.Message) pk.Packet
	Disconnect(reason chat.Message) pk.Packet
	CmdInjector(func(cmd string) (bool, error)) func(packet pk.Packet) (pass bool, err error)
	DimRecorder(*int32) func(packet pk.Packet) (pass bool, err error)
	ToRespawn(packet pk.Packet, dim int32) ([]pk.Packet, int32, error)
}

func GetProtocol(ver int) Protocol {
	// Get packet IDs for current version
	joinGame := joinGameID[ver]
	disconnect := disconnectID[ver]
	chatClient := chatMessageClientboundID[ver]
	chatServer := chatMessageServerboundID[ver]
	respawn := respawnID[ver]
	// Generate protocol
	unc := unchanged{
		versionID:  ver,
		disconnect: disconnect,
		chatClient: chatClient,
		cmdInject:  chatServer,
	}
	switch {
	case ver >= 578: // 1.15.2
		return supported{
			unchanged: unc,
			dimRecorder: dimRecorder{joinGame: joinGame, respawn: respawn,
				JoinGamePacket: joinGame19w36a{joinGameID: joinGame, respawnID: respawn},
				RespawnPacket:  respawn19w36a{packetID: respawn},
			},
		}
	case ver >= 47: // 1.8.9 to 1.8
		return supported{
			unchanged: unc,
			dimRecorder: dimRecorder{joinGame: joinGame, respawn: respawn,
				JoinGamePacket: joinGame14w29a{joinGameID: joinGame, respawnID: respawn},
				RespawnPacket:  respawn13w42a{packetID: respawn},
			},
		}
	default:
		return unsupported{
			unchanged: unchanged{versionID: Latest},
		}
	}
}

type supported struct {
	unchanged
	dimRecorder
}

type unsupported struct {
	unchanged
	dimRecorder
}

func (unsupported) Support() bool { return false } // override

// chat.TranslateMsg("multiplayer.disconnect.outdated_client", chat.Text(ServerName))

type unchanged struct {
	versionID  int
	disconnect byte
	chatClient byte
	cmdInject  byte
}

func (u unchanged) Support() bool  { return true }
func (u unchanged) VersionID() int { return u.versionID }
func (u unchanged) SysChat(msg chat.Message) pk.Packet {
	return pk.Marshal(u.chatClient, msg, pk.Byte(1))
}
func (u unchanged) Disconnect(reason chat.Message) pk.Packet {
	return pk.Marshal(u.disconnect, reason)
}
func (u unchanged) CmdInjector(cmdHandler func(cmd string) (bool, error)) func(packet pk.Packet) (pass bool, err error) {
	return func(packet pk.Packet) (pass bool, err error) {
		if packet.ID == u.cmdInject {
			var msg pk.String
			if err := packet.Scan(&msg); err != nil {
				return false, errors.New("handle chat message error")
			}
			return cmdHandler(string(msg))
		}
		return true, nil
	}
}

type dimRecorder struct {
	joinGame byte
	respawn  byte
	JoinGamePacket
	RespawnPacket
}

type JoinGamePacket interface {
	// convert JoinGame packet into Respawn packet
	// client programs cannot re-spawn to the same dimension they are already in.
	// so we send a extra Respawn packet to respawn them to another dimension first.
	ToRespawn(packet pk.Packet, dim int32) ([]pk.Packet, int32, error)
	Dimension(packet pk.Packet) (int32, error)
}

type RespawnPacket interface {
	Dimension(packet pk.Packet) (int32, error)
}

func (d dimRecorder) DimRecorder(dim *int32) func(packet pk.Packet) (pass bool, err error) {
	return func(packet pk.Packet) (pass bool, err error) {
		if packet.ID == d.joinGame {
			dimension, err := d.JoinGamePacket.Dimension(packet)
			if err != nil {
				return false, fmt.Errorf("parse JoinGame packet error: %w", err)
			}
			atomic.StoreInt32(dim, dimension)
		} else if packet.ID == d.respawn {
			dimension, err := d.RespawnPacket.Dimension(packet)
			if err != nil {
				return false, fmt.Errorf("parse Respawn packet error: %w", err)
			}
			atomic.StoreInt32(dim, dimension)
		}
		return true, nil
	}
}
