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
	d := Device{}
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
