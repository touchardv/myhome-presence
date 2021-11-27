package bluetooth

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/touchardv/myhome-presence/model"

	"github.com/bettercap/gatt"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/device"
)

// EnableTracker registers the "bluetooth" tracker so that it can be used.
func EnableTracker() {
	device.Register("bluetooth", newBTTracker)
}

type btTracker struct {
	device       gatt.Device
	scanning     bool
	scan         chan model.Interface
	mux          sync.Mutex
	deviceReport device.ReportPresenceFunc
}

func newBTTracker() device.Tracker {
	return &btTracker{
		scanning: false,
	}
}

func (t *btTracker) Loop(deviceReport device.ReportPresenceFunc, ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	log.Info("Starting: Bluetooth tracker")
	t.deviceReport = deviceReport
	t.startScanning()
	timer := time.NewTimer(30 * time.Second)
	for {
		select {
		case <-timer.C:
			t.mux.Lock()
			if t.scanning {
				t.stopScanning()
			} else {
				t.startScanning()
			}
			timer.Reset(randomDuration(t.scanning))
			t.mux.Unlock()
		case <-ctx.Done():
			timer.Stop()
			if t.scanning {
				t.stopScanning()
			}
			log.Info("Stopped: Bluetooth tracker")
			return nil
		}
	}
}

const minScanDuration = 20
const minDurationBetweenScans = 15

func randomDuration(scanning bool) time.Duration {
	var n int
	if scanning {
		n = minScanDuration
		n += rand.Intn(10)
	} else {
		n = minDurationBetweenScans
		n += rand.Intn(30)
	}
	return time.Duration(n) * time.Second
}
