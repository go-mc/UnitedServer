package main

import "flag"

var conf struct {
	DebugMode   bool
	MaxPlayers  int
	ListenAddr  string
	LobbyServer string
}

func parseConf() {
	flag.BoolVar(&conf.DebugMode, "DebugMode", false, "Start in debug mode")
	flag.IntVar(&conf.MaxPlayers, "MaxPlayers", 20, "Max connections number")
	flag.StringVar(&conf.ListenAddr, "ListenAddr", ":25565", "ip:port")
	flag.StringVar(&conf.LobbyServer, "LobbyServer", "localhost:25566", "The first server player joined")
	flag.Parse()
}
