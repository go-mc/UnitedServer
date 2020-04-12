package protocol

import (
	"fmt"
	pk "github.com/Tnze/go-mc/net/packet"
	log "github.com/sirupsen/logrus"
)

type proto15 struct {
	unchanged
	generalDimRecorder
}

func NewProto15() Protocol {
	return proto15{
		unchanged: unchanged{
			versionID:  578,
			disconnect: 0x1b,
			chatClient: 0x0f,
			cmdInject:  0x03,
		},
		generalDimRecorder: [2]byte{0x26, 0x3B},
	}
}

func (proto15) JoinGame2Respawn(packet pk.Packet, dim int32) ([]pk.Packet, int32, error) {
	if packet.ID != 0x26 {
		return nil, 0, fmt.Errorf("packet id 0x%02x is not JoinGame", packet.ID)
	}
	var (
		EID           pk.Int
		Gamemode      pk.UnsignedByte
		Dimension     pk.Int
		HashSeed      pk.Long
		MaxPlayers    pk.UnsignedByte
		LevelType     pk.String
		ViewDistance  pk.VarInt
		DebugInfo     pk.Boolean
		RespawnScreen pk.Boolean
	)
	if err := packet.Scan(&EID, &Gamemode, &Dimension, &HashSeed, &MaxPlayers, &LevelType, &ViewDistance,
		&DebugInfo, &RespawnScreen); err != nil {
		log.WithError(err).Error("Scan JoinGame packet error")
	}

	respawn := pk.Marshal(0x3b, Dimension, HashSeed, Gamemode, LevelType)
	if int32(Dimension) != dim {
		return []pk.Packet{respawn}, int32(Dimension), nil
	}
	// client programs cannot re-spawn to the same dimension they are already in.
	// so we send a extra Respawn packet to respawn them to another dimension first.
	otherDim := pk.Int(0)
	if otherDim == Dimension {
		otherDim = 1
	}
	extra := pk.Marshal(0x3b, Dimension, HashSeed, Gamemode, LevelType)

	return []pk.Packet{extra, respawn}, int32(Dimension), nil
}
