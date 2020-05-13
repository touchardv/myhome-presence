package bluetooth

import (
	"math/rand"
	"time"

	"github.com/bettercap/gatt"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/device"
)

// EnableTracker registers the "bluetooth" tracker so that it can be used.
func EnableTracker() {
	device.Register("bluetooth", newBTTracker)
}

type btTracker struct {
	device   gatt.Device
	scanning bool
	presence chan string
}

func newBTTracker() device.Tracker {
	return &btTracker{
		scanning: false,
	}
}

func (t *btTracker) Scan(presence chan string, stopping chan struct{}) {
	log.Info("Starting: Bluetooth tracker")
	t.startScanning()
	timer := time.NewTimer(30 * time.Second)
	for {
		select {
		case <-timer.C:
			if t.scanning {
				t.stopScanning()
				// Wait a little before doing the ping
				time.Sleep(2 * time.Second)
			} else {
				t.startScanning()
			}
			timer.Reset(randomDuration(t.scanning))
		case <-stopping:
			timer.Stop()
			if t.scanning {
				t.stopScanning()
			}
			log.Info("Stopped: Bluetooth tracker")
			return
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
