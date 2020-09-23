package device

import (
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/model"
)

// ID represents the unique ID that can be used to track a device.
type ID int

const (
	// Undefined is the default initialised device ID.
	Undefined ID = iota

	// BLEAddress is used when a new device is discovered via BLE scanning.
	BLEAddress

	// BTAddress is used when a new device is discovered via BT scanning.
	BTAddress

	//IPAddress is used when a new device is discoverd via IP scanning.
	IPAddress
)

// ScanResult contains the information on a newly discovered device.
type ScanResult struct {
	ID    ID
	Value string
}

// Tracker tracks the presence of devices.
type Tracker interface {
	Scan(existence chan ScanResult, stopping chan struct{})

	Ping(devices map[string]model.Device, presence chan string)
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
