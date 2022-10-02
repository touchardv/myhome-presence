package model

import (
	"bytes"
	"encoding/json"
	"time"
)

type EventType uint

const (
	EventTypeUndefined EventType = iota
	EventTypeAdded
	EventTypePresenceUpdated
	EventTypeUpdated
	EventTypeRemoved
)

var eventTypeToString = map[EventType]string{
	EventTypeUndefined:       "undefined",
	EventTypeAdded:           "added",
	EventTypePresenceUpdated: "presenceupdated",
	EventTypeUpdated:         "updated",
	EventTypeRemoved:         "removed",
}

var stringToEventType = map[string]EventType{
	"undefined":       EventTypeUndefined,
	"added":           EventTypeAdded,
	"presenceupdated": EventTypePresenceUpdated,
	"updated":         EventTypeUpdated,
	"removed":         EventTypeRemoved,
}

// MarshalJSON marshals the enum as a quoted json string
func (e EventType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(eventTypeToString[e])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshals a quoted json string to the enum value
func (e *EventType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	if t, ok := stringToEventType[s]; ok {
		*e = t
	} else {
		*e = EventTypeUndefined
	}
	return nil
}

func (e *EventType) String() string {
	return eventTypeToString[*e]
}

type Event struct {
	Type EventType       `json:"type"`
	Data json.RawMessage `json:"data"`
}

type DeviceAdded struct {
	Description string            `json:"description"`
	Identifier  string            `json:"identifier"`
	Present     bool              `json:"present"`
	Properties  map[string]string `json:"properties"`
	LastSeenAt  time.Time         `json:"last_seen_at"`
}

type DevicePresenceUpdated struct {
	Identifier string    `json:"identifier"`
	Present    bool      `json:"present"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type DeviceUpdated struct {
	Description string            `json:"description"`
	Identifier  string            `json:"identifier"`
	Present     bool              `json:"present"`
	Properties  map[string]string `json:"properties"`
	LastSeenAt  time.Time         `json:"last_seen_at"`
}

type DeviceRemoved struct {
	Identifier string `json:"identifier"`
}
