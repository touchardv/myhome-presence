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
		for _, itf := range d.Interfaces {
			if itf.Type == model.InterfaceBluetooth {
				log.Debug("Try to ping: ", itf.Address)
				if respondToPing(itf.Address) {
					presence <- d.Identifier
					delete(devices, d.Identifier)
					break
				}
			}
		}
	}
	t.mux.Unlock()
}
