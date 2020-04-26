package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/spf13/viper"
	"sync/atomic"
)

func Status(conn *net.Conn) error {
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

			list.Version.Name = UnitedServerVersion
			list.Version.Protocol = ProtocolVersion
			list.Players.Max = viper.GetInt("MaxPlayers")
			list.Players.Online = int(atomic.LoadInt64(&countOnline))
			list.Players.Sample = []struct{}{}
			list.Description = chat.Message{Text: "demo"}

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

func Login(conn *net.Conn) (*Player, error) {
	var name pk.String
	// LoginStart
	p, err := conn.ReadPacket()
	if err != nil {
		return nil, fmt.Errorf("recv LoginStart pk error: %w", err)
	}
	if err := p.Scan(&name); err != nil {
		return nil, fmt.Errorf("scan LoginStart pk error: %w", err)
	}
	// Check max players
	if !counterInc() {
		err := conn.WritePacket(pk.Marshal(0x00,
			chat.Message{Translate: "multiplayer.disconnect.server_full"}))
		if err != nil {
			return nil, err
		}
		return nil, errors.New("server full")
	}
	// LoginSuccess // TODO: player whitelist and blacklist
	err = conn.WritePacket(pk.Marshal(0x02,
		pk.String(bot.OfflineUUID(string(name)).String()), name))
	if err != nil {
		return nil, err
	}

	return &Player{
		Conn: conn,
		Name: string(name),
	}, nil
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
