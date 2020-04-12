package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"sync/atomic"
)

func Status(conn *net.Conn, version int) error {
	for i := 0; i < 2; i++ { //要么ping，要么list，只允许查询两次
		p, err := conn.ReadPacket()
		if err != nil {
			return err
		}

		switch p.ID {
		case 0x00: //List
			var list struct {
				Version struct {
					Name     string `json:"name"`
					Protocol int    `json:"protocol"`
				} `json:"version"`
				Players struct {
					Max    int        `json:"max"`
					Online int        `json:"online"`
					Sample []struct{} `json:"sample"` //must init with
				} `json:"players"`
				Description chat.Message `json:"description"`
				FavIcon     string       `json:"favicon,omitempty"`
			}

			list.Version.Name = ServerName
			list.Version.Protocol = version
			list.Players.Max = viper.GetInt("MaxPlayers")
			list.Players.Online = int(atomic.LoadInt64(&countOnline))
			list.Players.Sample = []struct{}{}
			list.Description = chat.Text(viper.GetString("Motd"))

			data, err := json.Marshal(list)
			if err != nil {
				return errors.New("marshal JSON for status checking fail")
			}
			err = conn.WritePacket(pk.Marshal(0x00, pk.String(data)))
		case 0x01: //Ping
			err = conn.WritePacket(p)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func Login(conn *net.Conn) (string, uuid.UUID, error) {
	var name pk.String
	// LoginStart
	p, err := conn.ReadPacket()
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("recv LoginStart pk error: %w", err)
	}
	if err := p.Scan(&name); err != nil {
		return "", uuid.Nil, fmt.Errorf("scan LoginStart pk error: %w", err)
	}
	// Check max players
	if !counterInc() {
		err := conn.WritePacket(pk.Marshal(0x00, // Disconnect
			chat.Message{Translate: "multiplayer.disconnect.server_full"}))
		if err != nil {
			return "", uuid.Nil, err
		}
		return "", uuid.Nil, errors.New("server full")
	}
	// TODO: player whitelist and blacklist
	id := OfflineUUID(string(name))
	err = conn.WritePacket(pk.Marshal(0x02, // LoginSuccess
		pk.String(id.String()), name))
	if err != nil {
		return "", uuid.Nil, err
	}
	return string(name), id, nil
}

var countOnline int64

func counterInc() bool {
	max := int64(viper.GetInt("MaxPlayers"))
	for {
		online := atomic.LoadInt64(&countOnline)
		if online >= max {
			return false // server is full
		}
		if atomic.CompareAndSwapInt64(&countOnline, online, online+1) {
			return true // successfully join
		}
	}
}

func counterDec() {
	atomic.AddInt64(&countOnline, -1)
}
