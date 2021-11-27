package bluetooth

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/model"
)

func (t *btTracker) Ping(d model.Device) {
	t.mux.Lock()
	if t.scanning {
		t.stopScanning()
		// Wait a little before doing the ping
		time.Sleep(2 * time.Second)
	}
	for _, itf := range d.Interfaces {
		if itf.Type == model.InterfaceBluetooth {
			log.Debug("Try to ping: ", itf.Address)
			if respondToPing(itf.Address) {
				t.deviceReport(itf)
				break
			}
		}
	}
	t.mux.Unlock()
}
