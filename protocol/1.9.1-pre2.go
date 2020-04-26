package protocol

import (
	"fmt"
	pk "github.com/Tnze/go-mc/net/packet"
)

type joinGame108 struct {
	joinGameID, respawnID byte

	EID        pk.Int
	Gamemode   pk.UnsignedByte
	Dim        pk.Int
	Difficulty pk.UnsignedByte
	MaxPlayers pk.UnsignedByte
	LevelType  pk.String
	DebugInfo  pk.Boolean
}

func (j *joinGame108) scan(packet pk.Packet) error {
	if packet.ID != j.joinGameID {
		return fmt.Errorf("packet id 0x%02x is not JoinGame", packet.ID)
	}
	return packet.Scan(
		&j.EID, &j.Gamemode, &j.Dim, &j.Difficulty,
		&j.MaxPlayers, &j.LevelType, &j.DebugInfo)
}

func (j joinGame108) ToRespawn(packet pk.Packet, dim int32) ([]pk.Packet, int32, error) {
	if err := j.scan(packet); err != nil {
		return nil, dim, err
	}
	respawn := pk.Marshal(j.respawnID, pk.Int(j.Dim), j.Difficulty, j.Gamemode, j.LevelType)
	if int32(j.Dim) != dim {
		return []pk.Packet{respawn}, int32(j.Dim), nil
	}
	otherDim := pk.Int(0)
	if otherDim == pk.Int(j.Dim) {
		otherDim = 1
	}
	extra := pk.Marshal(j.respawnID, otherDim, j.Difficulty, j.Gamemode, j.LevelType)
	return []pk.Packet{extra, respawn}, int32(j.Dim), nil
}

func (j joinGame108) Dimension(packet pk.Packet) (int32, error) {
	if err := j.scan(packet); err != nil {
		return 0, err
	}
	return int32(j.Dim), nil
}
