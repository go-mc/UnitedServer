package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var config struct {
	DebugMode   bool
	MaxPlayers  int
	ListenAddr  string
	LobbyServer string
}

func parseConf() {
	pflag.BoolVar(&config.DebugMode, "DebugMode", false, "Start in debug mode")
	pflag.IntVar(&config.MaxPlayers, "MaxPlayers", 20, "Max connections `number`")
	pflag.StringVar(&config.ListenAddr, "ListenAddr", ":25565", "`ip`:port")
	pflag.StringVar(&config.LobbyServer, "LobbyServer", "localhost:25566", "The first server `ip` player joined")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatal()
	}
}
