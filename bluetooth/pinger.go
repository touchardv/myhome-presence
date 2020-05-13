package bluetooth

import (
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
)

func (t *btTracker) Ping(devices map[string]config.Device, presence chan string) {
	for _, d := range devices {
		if len(d.BTAddress) == 0 {
			continue
		}
		log.Debug("Try to ping: ", d.BTAddress)
		if respondToPing(d.BTAddress) {
			t.presence <- d.Identifier
			delete(devices, d.Identifier)
		}
	}
}
