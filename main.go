package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/touchardv/myhome-presence/api"
	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/device"
	"github.com/touchardv/myhome-presence/trackers/bluetooth"
	"github.com/touchardv/myhome-presence/trackers/ipv4"
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
	bluetooth.EnableTracker()
	ipv4.EnableTracker()
	registry := device.NewRegistry(config)
	server := api.NewServer(config.Server, registry)
	registry.Start()
	server.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	server.Stop()
	registry.Stop()
	config.Save(registry.GetDevices())
	log.Info("...Stopped")
	log.Exit(0)
}
