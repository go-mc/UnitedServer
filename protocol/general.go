package protocol

import (
	"errors"
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
	JoinGame2Respawn(JoinGame pk.Packet, dim int32) (p []pk.Packet, newDim int32, err error)
	CmdInjector(func(cmd string) (bool, error)) func(packet pk.Packet) (pass bool, err error)
	DimRecorder(*int32) func(packet pk.Packet) (pass bool, err error)
}

func GetProtocol(ver int) Protocol {
	switch {
	case ver == 578: // 1.15.2
		return NewProto15()
	case ver == 47: // 1.8.9 to 1.8
		return NewProto8()
	default:
		return unsupported{
			unchanged: unchanged{versionID: Latest},
		}
	}
}

type unsupported struct {
	unchanged
	generalDimRecorder
}

func (unsupported) Support() bool                                                 { return false } // override
func (unsupported) JoinGame2Respawn(pk.Packet, int32) ([]pk.Packet, int32, error) { return nil, 0, nil }

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

// [2]byte{JoinGame, Respawn}
type generalDimRecorder [2]byte

func (g generalDimRecorder) DimRecorder(dim *int32) func(packet pk.Packet) (pass bool, err error) {
	return func(packet pk.Packet) (pass bool, err error) {
		var dimension pk.Byte
		if packet.ID == g[0] {
			if err := packet.Scan(new(pk.Int), new(pk.UnsignedByte), &dimension); err != nil {
				return false, errors.New("handle JoinGame packet error")
			}
			atomic.StoreInt32(dim, int32(dimension))
		} else if packet.ID == g[1] {
			if err := packet.Scan(&dimension); err != nil {
				return false, errors.New("handle Respawn packet error")
			}
			atomic.StoreInt32(dim, int32(dimension))
		}
		return true, nil
	}
}
