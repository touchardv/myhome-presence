package device

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/model"
)

type watchdog struct {
	stopped  chan bool
	stopping chan bool
	trackers []Tracker
}

func newWatchDog(cfg config.Config) *watchdog {
	trackers := make([]Tracker, 0)
	for name, settings := range cfg.Trackers {
		trackers = append(trackers, newTracker(name, settings))
	}
	return &watchdog{
		stopped:  make(chan bool),
		stopping: make(chan bool),
		trackers: trackers,
	}
}

func (w *watchdog) loop(r *Registry) {
	log.Info("Starting: device watchdog")
	var trackersWg sync.WaitGroup

	ctx, trackersStopFunc := context.WithCancel(context.Background())
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
			trackersStopFunc()
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
	w.stopping <- true
	<-w.stopped
	log.Info("Stopped: device watchdog")
}

func (w *watchdog) pingMissingDevices(r *Registry, now time.Time) {
	devices := r.GetDevices()
	for _, d := range devices {
		if d.Status == model.StatusTracked {
			elapsedMinutes := now.Sub(d.LastSeenAt).Minutes()
			if elapsedMinutes >= 5 {
				w.ping(&d)
			}
		}
	}
}

func (w *watchdog) ping(d *model.Device) {
	for _, t := range w.trackers {
		t.Ping(*d)
	}
}
