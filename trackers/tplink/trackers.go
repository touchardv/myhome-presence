package tplink

import (
	"context"
	"sync"

	"github.com/touchardv/myhome-presence/device"
	"github.com/touchardv/myhome-presence/model"
)

// EnableTrackers registers the "tplink" trackers.
func EnableTrackers() {
	device.Register("tplink-c2600", newArcherC2600Tracker)
	device.Register("tplink-re450", newRE450Tracker)
}

func newArcherC2600Tracker() device.Tracker {
	return &tplinkTracker{}
}

func newRE450Tracker() device.Tracker {
	return &tplinkTracker{}
}

type tplinkTracker struct {
}

func (t *tplinkTracker) Loop(deviceReport device.ReportPresenceFunc, ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	<-ctx.Done()

	return nil
}

func (t *tplinkTracker) Ping(model.Device) {

}
