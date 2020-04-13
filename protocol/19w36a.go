package protocol

import (
	"fmt"
	pk "github.com/Tnze/go-mc/net/packet"
)

type joinGame19w36a struct {
	joinGameID, respawnID byte
	eid                   pk.Int
	gameMode              pk.UnsignedByte
	dimension             pk.Int
	hashSeed              pk.Long
	maxPlayers            pk.UnsignedByte
	levelType             pk.String
	viewDistance          pk.VarInt
	debugInfo             pk.Boolean
	respawnScreen         pk.Boolean
}

func (j *joinGame19w36a) scan(packet pk.Packet) error {
	if packet.ID != j.joinGameID {
		return fmt.Errorf("packet id 0x%02x is not JoinGame", packet.ID)
	}
	return packet.Scan(
		&j.eid, &j.gameMode, &j.dimension, &j.hashSeed,
		&j.maxPlayers, &j.levelType, &j.viewDistance,
		&j.debugInfo, &j.respawnScreen)
}

func (j joinGame19w36a) ToRespawn(packet pk.Packet, dim int32) ([]pk.Packet, int32, error) {
	if err := j.scan(packet); err != nil {
		return nil, dim, err
	}
	respawn := pk.Marshal(j.respawnID, j.dimension, j.hashSeed, j.gameMode, j.levelType)
	if int32(j.dimension) != dim {
		return []pk.Packet{respawn}, int32(j.dimension), nil
	}
	otherDim := pk.Int(0)
	if otherDim == j.dimension {
		otherDim = 1
	}
	extra := pk.Marshal(j.respawnID, otherDim, j.hashSeed, j.gameMode, j.levelType)
	return []pk.Packet{extra, respawn}, int32(j.dimension), nil
}

func (j joinGame19w36a) Dimension(packet pk.Packet) (int32, error) {
	if err := j.scan(packet); err != nil {
		return 0, err
	}
	return int32(j.dimension), nil
}

type respawn19w36a struct {
	packetID  byte
	dimension pk.Int
	hashSeed  pk.Long
	gameMode  pk.UnsignedByte
	levelType pk.String
}

func (r respawn19w36a) Dimension(packet pk.Packet) (int32, error) {
	err := packet.Scan(&r.dimension, &r.hashSeed, &r.gameMode, &r.levelType)
	return int32(r.dimension), err
}
