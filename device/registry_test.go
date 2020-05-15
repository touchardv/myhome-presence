package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/touchardv/myhome-presence/config"
)

var device = config.Device{
	Description: "dummy",
	BLEAddress:  "BLE",
	BTAddress:   "BT",
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

func (t *dummyTracker) Scan(existence chan ScanResult, stopping chan struct{}) {
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

	assert.Equal(t, 1, len(devices))
	assert.False(t, devices[0].Present)
	assert.True(t, devices[0].LastSeenAt.IsZero())

	existence := make(chan ScanResult)
	presence := make(chan string)
	done := make(chan struct{})
	go func() {
		registry.handle(existence, presence)
		close(done)
	}()
	presence <- "foo"
	existence <- ScanResult{ID: BLEAddress, Value: "12:34:56:78:90"}

	close(registry.stopping)
	<-done

	devices = registry.GetDevices()
	assert.Equal(t, 2, len(devices))

	assert.True(t, devices[0].Present)
	assert.Equal(t, "foo", devices[0].Identifier)
	assert.False(t, devices[0].LastSeenAt.IsZero())

	assert.True(t, devices[1].Present)
	assert.False(t, devices[1].LastSeenAt.IsZero())
	assert.Equal(t, "12:34:56:78:90", devices[1].BLEAddress)
}

func TestNewDevice(t *testing.T) {
	registry := NewRegistry(cfg)
	d := registry.newDevice(ScanResult{ID: BLEAddress, Value: "one"})

	devices := registry.GetDevices()
	assert.Equal(t, 2, len(devices))
	assert.NotEmpty(t, d.Identifier, d.Description)
	assert.Equal(t, d.BLEAddress, "one")
	assert.Equal(t, config.Discovered, d.Status)

	d = registry.newDevice(ScanResult{ID: BTAddress, Value: "two"})
	assert.Equal(t, d.BTAddress, "two")

	d = registry.newDevice(ScanResult{ID: IPAddress, Value: "three"})
	assert.Equal(t, 1, len(d.IPInterfaces))
	assert.Equal(t, d.IPInterfaces["unknown"].IPAddress, "three")
}

func TestLookupDevice(t *testing.T) {
	registry := NewRegistry(cfg)

	d := registry.lookupDevice(ScanResult{ID: BLEAddress, Value: "BLE"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(ScanResult{ID: BTAddress, Value: "BT"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(ScanResult{ID: IPAddress, Value: "1.2.3.4"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(ScanResult{ID: BLEAddress, Value: "foobar"})
	assert.Nil(t, d)
}
