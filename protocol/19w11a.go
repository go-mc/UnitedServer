package protocol

import (
	"fmt"
	pk "github.com/Tnze/go-mc/net/packet"
)

type joinGame19w11a struct {
	joinGameID, respawnID byte

	EID        pk.Int
	Gamemode   pk.UnsignedByte
	Dim        pk.Int
	MaxPlayers pk.UnsignedByte
	LevelType  pk.String
	DebugInfo  pk.Boolean
}

func (j *joinGame19w11a) scan(packet pk.Packet) error {
	if packet.ID != j.joinGameID {
		return fmt.Errorf("packet id 0x%02x is not JoinGame", packet.ID)
	}
	return packet.Scan(
		&j.EID, &j.Gamemode, &j.Dim,
		&j.MaxPlayers, &j.LevelType, &j.DebugInfo)
}

func (j joinGame19w11a) ToRespawn(packet pk.Packet, dim int32) ([]pk.Packet, int32, error) {
	if err := j.scan(packet); err != nil {
		return nil, dim, err
	}
	respawn := pk.Marshal(j.respawnID, pk.Int(j.Dim), j.Gamemode, j.LevelType)
	if int32(j.Dim) != dim {
		return []pk.Packet{respawn}, int32(j.Dim), nil
	}
	otherDim := pk.Int(0)
	if otherDim == pk.Int(j.Dim) {
		otherDim = 1
	}
	extra := pk.Marshal(j.respawnID, otherDim, j.Gamemode, j.LevelType)
	return []pk.Packet{extra, respawn}, int32(j.Dim), nil
}

func (j joinGame19w11a) Dimension(packet pk.Packet) (int32, error) {
	if err := j.scan(packet); err != nil {
		return 0, err
	}
	return int32(j.Dim), nil
}

type respawn19w11a struct {
	packetID  byte
	Dim       pk.Int
	Gamemode  pk.UnsignedByte
	LevelType pk.String
}

func (r respawn19w11a) Dimension(packet pk.Packet) (int32, error) {
	err := packet.Scan(&r.Dim, &r.Gamemode, &r.LevelType)
	return int32(r.Dim), err
}
