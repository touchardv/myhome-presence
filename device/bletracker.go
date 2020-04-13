package device

import (
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
)

type bleTracker struct {
}

func newBLETracker() Tracker {
	return &bleTracker{}
}

func (t *bleTracker) Track(devices []config.Device, presence chan string, stopping chan struct{}) {
	log.Info("Starting: ble tracker")
	for {
		select {
		case <-stopping:
			log.Info("Stopped: ble tracker")
			return
		}
	}
}
