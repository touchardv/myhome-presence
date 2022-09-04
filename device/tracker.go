package device

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/model"
)

// ReportPresenceFunc is a function that can be used by a Tracker to report the presence of a given device interface.
// Optionally some data related to the inteface/device may be provided.
type ReportPresenceFunc func(model.Interface, map[string]string)

const (
	ReportDataSuggestedIdentifier  = "Identifier"
	ReportDataSuggestedDescription = "Description"
)

// Tracker tracks the presence of devices.
type Tracker interface {
	Loop(deviceReport ReportPresenceFunc, ctx context.Context, wg *sync.WaitGroup) error

	Ping(model.Device)
}

// NewTrackerFunc is a factory function for instantiating a new Tracker.
type NewTrackerFunc func(config.Settings) Tracker

var factories map[string]NewTrackerFunc = make(map[string]NewTrackerFunc)

// Register records a Tracker factory function by name.
func Register(name string, f NewTrackerFunc) {
	factories[name] = f
}

func newTracker(name string, settings config.Settings) Tracker {
	if f, ok := factories[name]; ok {
		return f(settings)
	}
	log.Fatal("No suck tracker: ", name)
	return nil
}
