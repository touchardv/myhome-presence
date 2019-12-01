package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/touchardv/myhome-presence/config"
)

var device = config.Device{
	Description: "dummy",
	Address:     "1.2.3.4",
	Identifier:  "foo",
}

var cfg = config.Config{
	IPDevices: []config.Device{device},
}

func TestGetDevices(t *testing.T) {
	registry := NewRegistry(cfg)
	devices := registry.GetDevices()

	assert.Equal(t, 1, len(devices))
	assert.Equal(t, "foo", devices[0].Identifier)
}

func TestNotify(t *testing.T) {
	registry := NewRegistry(cfg)
	devices := registry.GetDevices()

	assert.False(t, devices[0].Present)
	assert.True(t, devices[0].LastSeenAt.IsZero())

	d := Device{}
	d.Identifier = "foo"
	registry.notify(d, true)

	devices = registry.GetDevices()
	assert.True(t, devices[0].Present)
	assert.False(t, devices[0].LastSeenAt.IsZero())
}
