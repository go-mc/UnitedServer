package main

import (
	"encoding/json"
	"errors"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
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

			list.Version.Name = "MCServerSwitch"
			list.Version.Protocol = ProtocolVersion
			list.Players.Max = 20
			list.Players.Online = -1
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

func Login(conn *net.Conn) (Player, error) {

	return Player{
		Conn: conn,
		Name: "",
	}, nil
}
