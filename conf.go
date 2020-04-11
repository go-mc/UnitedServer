package main

import "flag"

var conf struct {
	DebugMode   bool
	HelpInf     bool
	MaxPlayers  int
	ListenAddr  string
	LobbyServer string
}

func parseConf() {
	flag.BoolVar(&conf.DebugMode, "DebugMode", false, "Start in debug mode")
	flag.BoolVar(&conf.HelpInf, "h", false, "Help")
	flag.IntVar(&conf.MaxPlayers, "MaxPlayers", 20, "Max connections `number`")
	flag.StringVar(&conf.ListenAddr, "ListenAddr", ":25565", "`ip`:port")
	flag.StringVar(&conf.LobbyServer, "LobbyServer", "localhost:25566", "The first server `ip` player joined")
	flag.Parse()
}
