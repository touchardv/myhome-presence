package ipv4

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func (t *ipTracker) Scan(presence chan string, stopping chan struct{}) {
	log.Info("Starting: ip scanner")
	ticker := time.NewTicker(1 * time.Minute)
	select {
	case <-stopping:
		ticker.Stop()
		log.Info("Stopped: ip scanner")
		return

	case <-ticker.C:
		// TODO implement a background IP scanner (e.g. ping a configured IP range)
	}
}
