package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func parseConf() {
	pflag.Bool("DebugMode", false, "Start in debug mode")
	pflag.Int("MaxPlayers", 20, "Max connections `number`")
	pflag.String("ListenAddr", ":25565", "`ip`:port")
	pflag.String("LobbyServer", "localhost:25566", "The first server `ip` player joined")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatal("Fatal error parse arg: %s \n", err)
	}
}
