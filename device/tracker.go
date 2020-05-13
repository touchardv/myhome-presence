package device

import (
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
)

// Tracker tracks the presence of devices.
type Tracker interface {
	Scan(presence chan string, stopping chan struct{})

	Ping(devices map[string]config.Device, presence chan string)
}

// NewTrackerFunc is a factory function for instantiating a new Tracker.
type NewTrackerFunc func() Tracker

var factories map[string]NewTrackerFunc = make(map[string]NewTrackerFunc)

// Register records a Tracker factory function by name.
func Register(name string, f NewTrackerFunc) {
	factories[name] = f
}

func newTracker(name string) Tracker {
	if f, ok := factories[name]; ok {
		return f()
	}
	log.Fatal("No suck tracker: ", name)
	return nil
}
