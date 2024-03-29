package model

import (
	"bytes"
	"encoding/json"
	"time"
)

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

var statusToString = map[Status]string{
	StatusUndefined:  "undefined",
	StatusDiscovered: "discovered",
	StatusIgnored:    "ignored",
	StatusTracked:    "tracked",
}

var stringToStatus = map[string]Status{
	"undefined":  StatusUndefined,
	"discovered": StatusDiscovered,
	"ignored":    StatusIgnored,
	"tracked":    StatusTracked,
}

func StatusOf(s string) Status {
	if t, ok := stringToStatus[s]; ok {
		return t
	} else {
		return StatusUndefined
	}
}

func (s Status) String() string {
	return statusToString[s]
}

// MarshalJSON marshals the enum as a quoted json string
func (s Status) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(statusToString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshals a quoted json string to the enum value
func (s *Status) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	if t, ok := stringToStatus[j]; ok {
		*s = t
	} else {
		return ErrInvalidDeviceStatus
	}
	return nil
}

// MarshalYAML marshals the enum as yaml string
func (s Status) MarshalYAML() (interface{}, error) {
	return statusToString[s], nil
}

// UnmarshalYAML unmarshals a yaml string to the enum value
func (s *Status) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var t string
	if err := unmarshal(&t); err != nil {
		return err
	}
	if t, ok := stringToStatus[t]; ok {
		*s = t
	} else {
		*s = StatusUndefined
	}
	return nil
}

// Device represents a single device that can be tracked.
type Device struct {
	Description string            `json:"description"`
	Identifier  string            `json:"identifier"`
	Interfaces  []Interface       `json:"interfaces" yaml:"interfaces"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	FirstSeenAt time.Time         `json:"first_seen_at" yaml:"first_seen_at"`
	LastSeenAt  time.Time         `json:"last_seen_at" yaml:"last_seen_at"`
	Present     bool              `json:"present" yaml:"present"`
	Properties  map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`
	Status      Status            `json:"status" yaml:"status"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
}
