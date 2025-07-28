package device

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/internal/config"
	"github.com/touchardv/myhome-presence/pkg/model"
)

type watchdog struct {
	stopped  chan bool
	stopping chan interface{}
	trackers []Tracker
}

func newWatchDog(cfg config.Config) *watchdog {
	trackers := make([]Tracker, 0)
	for name, settings := range cfg.Trackers {
		trackers = append(trackers, newTracker(name, settings))
	}
	return &watchdog{
		stopped:  make(chan bool),
		stopping: make(chan interface{}),
		trackers: trackers,
	}
}

func (w *watchdog) loop(r *Registry, ctx context.Context) {
	log.Info("Starting: device watchdog")
	var trackersWg sync.WaitGroup

	trackersWg.Add(len(w.trackers))
	for _, t := range w.trackers {
		go t.Loop(r.reportPresence, ctx, &trackersWg)
	}

	needUpdate := false
	check := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-w.stopping:
			log.Info("Stopping: trackers...")
			trackersWg.Wait()
			log.Info("Stopped: trackers")
			w.stopped <- true
			return

		case <-check.C:
			now := time.Now()
			if needUpdate {
				r.UpdateDevicesPresence(now)
			} else {
				w.pingMissingDevices(r, now)
			}
			needUpdate = !needUpdate
			check.Reset(30 * time.Second)
		}
	}
}

func (w *watchdog) stop() {
	log.Info("Stopping: device watchdog...")
	close(w.stopping)
	<-w.stopped
	log.Info("Stopped: device watchdog")
}

func (w *watchdog) pingMissingDevices(r *Registry, now time.Time) {
	devices := r.GetDevices(model.StatusTracked)
	missingDevices := []model.Device{}
	for _, d := range devices {
		elapsedMinutes := now.Sub(d.LastSeenAt).Minutes()
		if elapsedMinutes >= 5 {
			missingDevices = append(missingDevices, d)
		}
	}
	if len(missingDevices) > 0 {
		w.ping(missingDevices)
	} else {
		log.Debug("No missing devices to ping")
	}
}

func (w *watchdog) ping(devices []model.Device) {
	for _, t := range w.trackers {
		t.Ping(devices)
	}
}
