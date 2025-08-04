package device

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/internal/config"
	"github.com/touchardv/myhome-presence/pkg/model"
)

// ReportPresenceFunc is a function that can be used by a Tracker to report the presence of one or more detected device interfaces.
// Optionally some data related to the interface/device may be provided.
type ReportPresenceFunc func([]model.DetectedInterface)

const (
	ReportDataSuggestedIdentifier  = "Identifier"
	ReportDataSuggestedDescription = "Description"
)

// Tracker tracks the presence of devices.
type Tracker interface {
	Loop(deviceReport ReportPresenceFunc, ctx context.Context, wg *sync.WaitGroup) error

	Ping([]model.Device)
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
