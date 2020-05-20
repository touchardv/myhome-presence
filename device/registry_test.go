package device

import (
	"testing"
	"time"

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
	Devices: map[string]*config.Device{"foo": &device},
}

type dummyTracker struct {
	pingCount int
	scanCount int
}

var tracker dummyTracker

func newDummyTracker() Tracker {
	return &tracker
}

func (t *dummyTracker) Scan(existence chan ScanResult, stopping chan struct{}) {
	t.scanCount++
}

func (t *dummyTracker) Ping(devices map[string]config.Device, presence chan string) {
	t.pingCount++
}

func init() {
	Register("dummy", newDummyTracker)
}

func TestGetDevices(t *testing.T) {
	registry := NewRegistry(cfg)
	devices := registry.GetDevices()

	assert.Equal(t, 1, len(devices))
	assert.Equal(t, "foo", devices[0].Identifier)
}

func TestHandleDevicePresence(t *testing.T) {
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

	close(registry.stopping)
	<-done

	devices = registry.GetDevices()
	assert.Equal(t, 1, len(devices))

	assert.True(t, devices[0].Present)
	assert.Equal(t, "foo", devices[0].Identifier)
	assert.False(t, devices[0].LastSeenAt.IsZero())
}

func TestHandleNewDevice(t *testing.T) {
	registry := NewRegistry(config.Config{Devices: map[string]*config.Device{}})

	existence := make(chan ScanResult)
	presence := make(chan string)
	done := make(chan struct{})
	go func() {
		registry.handle(existence, presence)
		close(done)
	}()
	existence <- ScanResult{ID: BLEAddress, Value: "12:34:56:78:90"}

	close(registry.stopping)
	<-done

	devices := registry.GetDevices()
	assert.Equal(t, 1, len(devices))

	assert.True(t, devices[0].Present)
	assert.False(t, devices[0].LastSeenAt.IsZero())
	assert.Equal(t, "12:34:56:78:90", devices[0].BLEAddress)
	assert.Equal(t, config.Discovered, devices[0].Status)
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

func TestPingMissingDevices(t *testing.T) {
	device := config.Device{Identifier: "foo", Status: config.Undefined}
	registry := NewRegistry(config.Config{
		Devices:  map[string]*config.Device{"foo": &device},
		Trackers: []string{"dummy"},
	})
	presence := make(chan string)

	// No ping: device is not tracked
	registry.pingMissingDevices(presence)
	assert.Equal(t, 0, tracker.pingCount)

	// No ping: device is present and seen recently
	device.Status = config.Tracked
	device.LastSeenAt = time.Now()
	device.Present = true
	registry.pingMissingDevices(presence)
	assert.Equal(t, 0, tracker.pingCount)

	// Ping: device is absent and seen a while ago
	device.LastSeenAt = time.Now().Add(-24 * time.Hour)
	device.Present = false
	registry.pingMissingDevices(presence)
	assert.Equal(t, 1, tracker.pingCount)
}
