package model

import "time"

// Status represents the various states a device can be in.
type Status uint

const (
	// StatusUndefined is the default initialised device status.
	StatusUndefined Status = iota

	// StatusDiscovered is the initial state given to a device that was found
	// by a tracker scan operation for the first time.
	StatusDiscovered

	// StatusIgnored is the state given to a device that should not be tracked.
	StatusIgnored

	// StatusTracked is the state given to a device that should be watched.
	StatusTracked
)

func (s Status) String() string {
	status := [...]string{"Undefined", "Discovered", "Ignored", "Tracked"}
	if s < StatusUndefined || s > StatusTracked {
		return "Unknown"
	}
	return status[s]
}

// Device represents a single device that can be tracked.
type Device struct {
	// example: My phone
	Description string `json:"description"`

	// example: my-phone
	// required: true
	Identifier string `json:"identifier"`

	Interfaces []Interface `json:"interfaces" yaml:"interfaces"`
	Status     Status      `json:"status" yaml:"status"`
	Present    bool        `json:"present" yaml:"present"`
	LastSeenAt time.Time   `json:"last_seen_at" yaml:"last_seen_at"`
}
