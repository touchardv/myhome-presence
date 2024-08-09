package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/touchardv/myhome-presence/internal/api"
	"github.com/touchardv/myhome-presence/internal/config"
	"github.com/touchardv/myhome-presence/internal/device"
	"github.com/touchardv/myhome-presence/internal/trackers/bluetooth"
	"github.com/touchardv/myhome-presence/internal/trackers/ipv4"
	"github.com/touchardv/myhome-presence/internal/trackers/linksys"
	"github.com/touchardv/myhome-presence/internal/trackers/tplink"
	"github.com/touchardv/myhome-presence/pkg/model"
)

var (
	buildDate     = "undefined"
	gitCommitHash = "undefined"
	gitVersionTag = "undefined"
)

func main() {
	fmt.Println("myhome-presence - version:", gitVersionTag)
	fmt.Println("built date:", buildDate)
	fmt.Println("git commit:", gitCommitHash)

	daemonized := pflag.Bool("daemon", false, "Start as daemon")
	logLevel := pflag.String("log-level", log.InfoLevel.String(), "The logging level (trace, debug, info...)")
	configLocation := pflag.String("config-location", config.DefaultCfgLocation, "The path to the directory where the configuration file is stored.")
	dataLocation := pflag.String("data-location", config.DefaultDataLocation, "The path to the directory where the data file is stored.")
	pflag.Parse()

	close := config.SetupLogging(*logLevel, *daemonized)
	defer close()

	log.Info("Starting...")
	config := config.Retrieve(*configLocation, *dataLocation)
	bluetooth.EnableTracker()
	ipv4.EnableTracker()
	linksys.EnableTracker()
	tplink.EnableTrackers()
	registry := device.NewRegistry(config)
	server := api.NewServer(config.Server, registry)

	ctx, stopFunc := context.WithCancel(context.Background())
	registry.Start(ctx)
	server.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	server.Stop()
	stopFunc()
	registry.Stop()
	config.Save(registry.GetDevices(model.StatusUndefined))
	log.Info("...Stopped")
	log.Exit(0)
}
