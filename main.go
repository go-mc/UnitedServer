package main

import (
	"github.com/Tnze/go-mc/net"
	log "github.com/sirupsen/logrus"
)

const ProtocolVersion = 578

func main() {
	l, err := net.ListenMC(":25566")
	if err != nil {
		log.WithError(err).Panic("Listen fail")
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.WithError(err).Panic("Accept connection error")
		}
		go handleConn(&conn)
	}
}
