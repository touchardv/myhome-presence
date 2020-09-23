package model

import "time"

// Status represents the various states a device can be in.
type Status int

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

// IPInterface represents one IP device interface.
type IPInterface struct {
	IPAddress  string `json:"ip_address" yaml:"ip_address"`
	MACAddress string `json:"mac_address" yaml:"mac_address"`
}

// Device represents a single device that can be tracked.
type Device struct {
	Description  string                 `json:"description"`
	Identifier   string                 `json:"identifier"`
	BLEAddress   string                 `json:"ble_address" yaml:"ble_address"`
	BTAddress    string                 `json:"bt_address" yaml:"bt_address"`
	IPInterfaces map[string]IPInterface `json:"ip_interfaces" yaml:"ip_interfaces"`
	Status       Status                 `json:"status" yaml:"status"`
	Present      bool                   `json:"present" yaml:"present"`
	LastSeenAt   time.Time              `json:"last_seen_at" yaml:"last_seen_at"`
}
