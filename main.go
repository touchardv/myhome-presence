package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/touchardv/myhome-presence/config"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/api"
	"github.com/touchardv/myhome-presence/device"
)

func main() {
	log.Info("Starting...")
	config := config.Retrieve()
	registry := device.NewRegistry(config.Devices)
	server := api.NewServer(registry)
	server.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	server.Stop()
	log.Info("...Stopped")
	log.Exit(0)
}
