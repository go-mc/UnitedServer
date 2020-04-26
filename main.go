package main

import (
	"context"
	"github.com/Tnze/go-mc/net"
	"github.com/shiena/ansicolor"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"sync"
)

const UnitedServerVersion = "UnitedServer Alpha"
const ProtocolVersion = 578

func main() {
	parseConf()
	if viper.GetBool("DebugMode") {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
		log.Warn("Starting in debug mode")
	}

	l, err := net.ListenMC(viper.GetString("ListenAddr"))
	if err != nil {
		log.WithError(err).Fatal("Start listening fail")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go gracefulShutdown(func() {
		cancel()
		if err := l.Close(); err != nil {
			log.WithError(err).Error("Close listener error")
		}
	})
	var wg sync.WaitGroup
	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() == nil {
				log.WithError(err).Error("Accept connection error")
			}
			break
		}
		wg.Add(1)
		go func() {
			handleConn(ctx, &conn)
			wg.Done()
		}()
	}
	wg.Wait() // wait for all connection close
}

func gracefulShutdown(shutdown func()) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	for i := 0; i < 4; i++ {
		<-sig
		if i == 0 {
			log.Warn("Server is stopping, please wait.")
			shutdown()
		} else {
			log.Warnf("Press %d times more to force stop this server", 4-i)
		}
	}
	os.Exit(2)
}
