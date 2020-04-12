package protocol

import (
	"fmt"
	pk "github.com/Tnze/go-mc/net/packet"
	log "github.com/sirupsen/logrus"
)

type proto8 struct {
	unchanged
	generalDimRecorder
}

func NewProto8() Protocol {
	return proto8{
		unchanged: unchanged{
			versionID:  47,
			disconnect: 0x40,
			chatClient: 0x02,
			cmdInject:  0x01,
		},
		generalDimRecorder: [2]byte{0x01, 0x07},
	}
}

func (proto8) JoinGame2Respawn(packet pk.Packet, dim int32) ([]pk.Packet, int32, error) {
	if packet.ID != 0x01 {
		return nil, 0, fmt.Errorf("packet id 0x%02x is not JoinGame", packet.ID)
	}
	// parse JoinGame packet
	var (
		EID        pk.Int
		Gamemode   pk.UnsignedByte
		Dimension  pk.Byte
		Difficulty pk.UnsignedByte
		MaxPlayers pk.UnsignedByte
		LevelType  pk.String
		DebugInfo  pk.Boolean
	)
	if err := packet.Scan(
		&EID, &Gamemode, &Dimension, &Difficulty,
		&MaxPlayers, &LevelType, &DebugInfo); err != nil {
		log.WithError(err).Error("Scan JoinGame packet error")
	}

	respawn := pk.Marshal(0x07, pk.Int(Dimension), Difficulty, Gamemode, LevelType)
	if int32(Dimension) != dim {
		return []pk.Packet{respawn}, int32(Dimension), nil
	}
	// client programs cannot re-spawn to the same dimension they are already in.
	// so we send a extra Respawn packet to respawn them to another dimension first.
	otherDim := pk.Int(0)
	if otherDim == pk.Int(Dimension) {
		otherDim = 1
	}
	extra := pk.Marshal(0x07, otherDim, Difficulty, Gamemode, LevelType)

	return []pk.Packet{extra, respawn}, int32(Dimension), nil
}
