package ipv4

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/device"
	"golang.org/x/net/icmp"
)

func (t *ipTracker) Loop(deviceReport device.ReportPresenceFunc, ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	log.Info("Starting: ipv4 tracker")
	socket, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		log.Error("Failed to create udp4/icmp socket: ", err)
		return err
	}
	t.socket = socket

	stopped := make(chan bool)
	go func() {
		t.receiveLoop(deviceReport)
		stopped <- true
	}()

	<-ctx.Done()
	t.stopReceiving = true
	t.socket.Close()
	<-stopped
	log.Info("Stopped: ipv4 tracker")

	return nil
}
