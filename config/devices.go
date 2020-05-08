package config

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

// Status represents the various states a device can be in.
type Status int

const (
	// Undefined is the default initialised device status.
	Undefined Status = iota

	// Discovered is the initial state given to a device that was found
	// by a tracker scan operation for the first time.
	Discovered

	// Ignored is the state given to a device that should not be tracked.
	Ignored

	// Tracked is the state given to a device that should be watched.
	Tracked
)

func (s Status) String() string {
	status := [...]string{"Discovered", "Ignored", "Tracked"}
	if s < Undefined || s > Tracked {
		return "Unknown"
	}
	return status[s]
}

// Device represents a single device that can be tracked.
type Device struct {
	Description  string
	Identifier   string
	BLEAddress   string                 `yaml:"ble_address"`
	BTAddress    string                 `yaml:"bt_address"`
	IPInterfaces map[string]IPInterface `yaml:"ip_interfaces"`
	Status       Status                 `yaml:"status"`
	Present      bool                   `yaml:"present"`
	LastSeenAt   time.Time              `yaml:"last_seen_at"`
}

func save(devices []Device, location string) error {
	bytes, err := yaml.Marshal(devices)
	if err == nil {
		err = ioutil.WriteFile(filepath.Join(location, "devices.yaml"), bytes, 0644)
	}
	return err
}
