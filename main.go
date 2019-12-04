package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"
	"github.com/touchardv/myhome-presence/config"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/api"
	"github.com/touchardv/myhome-presence/device"
)

func main() {
	daemonized := pflag.Bool("daemon", false, "Start as daemon")
	logLevel := pflag.String("log-level", log.InfoLevel.String(), "The logging level (trace, debug, info...)")
	configLocation := pflag.String("config-location", config.DefaultLocation, "The path to the directory where the configuration file is stored.")
	pflag.Parse()

	close := config.SetupLogging(*logLevel, *daemonized)
	defer close()

	log.Info("Starting...")
	config := config.Retrieve(*configLocation)
	registry := device.NewRegistry(config)
	server := api.NewServer(registry)
	registry.Start()
	server.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	server.Stop()
	registry.Stop()
	log.Info("...Stopped")
	log.Exit(0)
}
