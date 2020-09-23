package bluetooth

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/model"
)

func (t *btTracker) Ping(devices map[string]model.Device, presence chan string) {
	t.mux.Lock()
	if t.scanning {
		t.stopScanning()
		// Wait a little before doing the ping
		time.Sleep(2 * time.Second)
	}
	for _, d := range devices {
		if len(d.BTAddress) == 0 {
			continue
		}

		log.Debug("Try to ping: ", d.BTAddress)
		if respondToPing(d.BTAddress) {
			presence <- d.Identifier
			delete(devices, d.Identifier)
		}
	}
	t.mux.Unlock()
}
