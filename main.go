package main

import (
	"flag"
	"github.com/Tnze/go-mc/net"
	"github.com/shiena/ansicolor"
	log "github.com/sirupsen/logrus"
	"os"
)

const ProtocolVersion = 578

var DebugMode = flag.Bool("debug", false, "Start in debug mode")

func main() {
	flag.Parse()
	if *DebugMode {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
		log.Warn("Starting in debug mode")
	}

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
