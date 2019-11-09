package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/api"
)

func main() {
	log.Info("Starting...")
	server := api.NewServer()
	server.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	server.Stop()
	log.Info("...Stopped")
	log.Exit(0)
}
