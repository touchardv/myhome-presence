package bluetooth

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/device"
)

// EnableTracker registers the "bluetooth" tracker so that it can be used.
func EnableTracker() {
	device.Register("bluetooth", newBTTracker)
}

type btTracker struct{}

func newBTTracker(config.Settings) device.Tracker {
	return &btTracker{}
}

func (t *btTracker) Loop(deviceReport device.ReportPresenceFunc, ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	log.Info("Starting: bluetooth tracker")
	mgr := newBtManager()
	err := mgr.scan(deviceReport, ctx)
	if err != nil {
		log.Warn("Scan failed: ", err)
	} else {
		<-ctx.Done()
		mgr.stopScan()
	}

	log.Info("Stopped: bluetooth tracker")
	return nil
}

type btManager interface {
	scan(device.ReportPresenceFunc, context.Context) error
	stopScan()
}
