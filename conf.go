package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var conf struct {
	DebugMode   bool
	MaxPlayers  int
	ListenAddr  string
	LobbyServer string
}

func parseConf() {
	pflag.BoolVar(&conf.DebugMode, "DebugMode", false, "Start in debug mode")
	pflag.IntVar(&conf.MaxPlayers, "MaxPlayers", 20, "Max connections `number`")
	pflag.StringVar(&conf.ListenAddr, "ListenAddr", ":25565", "`ip`:port")
	pflag.StringVar(&conf.LobbyServer, "LobbyServer", "localhost:25566", "The first server `ip` player joined")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatal("Fatal error parse arg: %s \n", err)
	}
}
