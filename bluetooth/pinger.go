package bluetooth

import log "github.com/sirupsen/logrus"

func (t *btTracker) ping() {
	log.Debug("Start pinging Bluetooth devices...")
	for _, d := range t.devices {
		if len(d.BTAddress) == 0 {
			continue
		}
		log.Debug("Try to ping: ", d.BTAddress)
		if respondToPing(d.BTAddress) {
			t.presence <- d.Identifier
		}
	}
}
