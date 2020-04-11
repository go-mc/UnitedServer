package main

import (
	"flag"
	"github.com/shiena/ansicolor"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

var config struct {
	DebugMode   bool
	HelpInf     bool
	MaxPlayers  int
	ListenAddr  string
	LobbyServer string
}

func initConf() {
	flag.BoolVar(&config.DebugMode, "DebugMode", false, "Start in debug mode")
	flag.IntVar(&config.MaxPlayers, "MaxPlayers", 20, "Max connections `number`")
	flag.StringVar(&config.ListenAddr, "ListenAddr", ":25565", "`ip`:port")
	flag.StringVar(&config.LobbyServer, "LobbyServer", "localhost:25566", "The first server `ip` player joined")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if config.DebugMode {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
		log.Warn("Starting in debug mode")
	}
}
