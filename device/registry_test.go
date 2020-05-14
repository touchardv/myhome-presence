package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/touchardv/myhome-presence/config"
)

var device = config.Device{
	Description: "dummy",
	IPInterfaces: map[string]config.IPInterface{
		"ethernet": {
			IPAddress: "1.2.3.4",
		},
	},
	Identifier: "foo",
}

var cfg = config.Config{
	Devices: []config.Device{device},
}

type dummyTracker struct{}

func newDummyTracker() Tracker {
	return &dummyTracker{}
}

func (t *dummyTracker) Scan(presence chan string, stopping chan struct{}) {
	// noop
}

func (t *dummyTracker) Ping(devices map[string]config.Device, presence chan string) {
	// noop
}

func init() {
	Register("bluetooth", newDummyTracker)
	Register("ipv4", newDummyTracker)
}

func TestGetDevices(t *testing.T) {
	registry := NewRegistry(cfg)
	devices := registry.GetDevices()

	assert.Equal(t, 1, len(devices))
	assert.Equal(t, "foo", devices[0].Identifier)
}

func TestHandle(t *testing.T) {
	registry := NewRegistry(cfg)
	devices := registry.GetDevices()

	assert.False(t, devices[0].Present)
	assert.True(t, devices[0].LastSeenAt.IsZero())

	presence := make(chan string)
	d := config.Device{}
	d.Identifier = "foo"
	go func() {
		registry.handle(presence)
	}()
	presence <- "foo"

	devices = registry.GetDevices()
	assert.True(t, devices[0].Present)
	assert.False(t, devices[0].LastSeenAt.IsZero())
	close(registry.stopping)
}
