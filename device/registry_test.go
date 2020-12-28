package device

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/model"
)

var device = model.Device{
	Description: "dummy",
	Identifier:  "foo",
	Interfaces: []model.Interface{
		{Type: model.InterfaceBluetoothLowEnergy, Address: "BLE"},
		{Type: model.InterfaceBluetooth, Address: "BT"},
		{Type: model.InterfaceWifi, IPv4Address: "1.2.3.4"},
	},
}

var cfg = config.Config{
	Devices: map[string]*model.Device{"foo": &device},
}

type dummyTracker struct {
	pingCount int
	scanCount int
}

var tracker dummyTracker

func newDummyTracker() Tracker {
	return &tracker
}

func (t *dummyTracker) Scan(existence chan model.Interface, stopping chan struct{}) {
	t.scanCount++
}

func (t *dummyTracker) Ping(devices map[string]model.Device, presence chan string) {
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

	existence := make(chan model.Interface)
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
	registry := NewRegistry(config.Config{Devices: map[string]*model.Device{}})

	existence := make(chan model.Interface)
	presence := make(chan string)
	done := make(chan struct{})
	go func() {
		registry.handle(existence, presence)
		close(done)
	}()
	existence <- model.Interface{Type: model.InterfaceBluetoothLowEnergy, Address: "12:34:56:78:90"}

	close(registry.stopping)
	<-done

	devices := registry.GetDevices()
	assert.Equal(t, 1, len(devices))

	assert.True(t, devices[0].Present)
	assert.False(t, devices[0].LastSeenAt.IsZero())
	assert.Equal(t, model.InterfaceBluetoothLowEnergy, devices[0].Interfaces[0].Type)
	assert.Equal(t, "12:34:56:78:90", devices[0].Interfaces[0].Address)
	assert.Equal(t, model.StatusDiscovered, devices[0].Status)
}

func TestNewDevice(t *testing.T) {
	registry := NewRegistry(cfg)
	d := registry.newDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, Address: "one"})

	devices := registry.GetDevices()
	assert.Equal(t, 2, len(devices))
	assert.NotEmpty(t, d.Identifier, d.Description)
	assert.Equal(t, model.InterfaceBluetoothLowEnergy, d.Interfaces[0].Type)
	assert.Equal(t, "one", d.Interfaces[0].Address)
	assert.Equal(t, model.StatusDiscovered, d.Status)

	d = registry.newDevice(model.Interface{Type: model.InterfaceBluetooth, Address: "two"})
	assert.Equal(t, model.InterfaceBluetooth, d.Interfaces[0].Type)
	assert.Equal(t, "two", d.Interfaces[0].Address)

	d = registry.newDevice(model.Interface{Type: model.InterfaceWifi, IPv4Address: "three"})
	assert.Equal(t, 1, len(d.Interfaces))
	assert.Equal(t, model.InterfaceWifi, d.Interfaces[0].Type)
	assert.Equal(t, "three", d.Interfaces[0].IPv4Address)
}

func TestLookupDevice(t *testing.T) {
	registry := NewRegistry(cfg)

	d := registry.lookupDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, Address: "BLE"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(model.Interface{Type: model.InterfaceBluetooth, Address: "BT"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(model.Interface{Type: model.InterfaceWifi, IPv4Address: "1.2.3.4"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, Address: "foobar"})
	assert.Nil(t, d)
}

func TestPingMissingDevices(t *testing.T) {
	device := model.Device{Identifier: "foo", Status: model.StatusUndefined}
	registry := NewRegistry(config.Config{
		Devices:  map[string]*model.Device{"foo": &device},
		Trackers: []string{"dummy"},
	})
	presence := make(chan string)

	// No ping: device is not tracked
	registry.pingMissingDevices(presence)
	assert.Equal(t, 0, tracker.pingCount)

	// No ping: device is present and seen less than 5 minutes ago
	device.Status = model.StatusTracked
	device.LastSeenAt = time.Now().Add(-3 * time.Minute)
	device.Present = true
	registry.pingMissingDevices(presence)
	assert.Equal(t, 0, tracker.pingCount)
	assert.True(t, device.Present)

	// Ping: device is present and seen more than 5 minutes ago
	device.LastSeenAt = time.Now().Add(-7 * time.Minute)
	registry.pingMissingDevices(presence)
	assert.Equal(t, 1, tracker.pingCount)
	assert.True(t, device.Present)

	// Ping: device is absent and seen more than 10 minutes ago
	device.LastSeenAt = time.Now().Add(-15 * time.Minute)
	device.Present = false
	registry.pingMissingDevices(presence)
	assert.Equal(t, 2, tracker.pingCount)
	assert.False(t, device.Present)
}

func TestUpdateDevice(t *testing.T) {
	registry := NewRegistry(cfg)
	registry.devices["foo"].Present = false
	registry.devices["foo"].LastSeenAt = time.Time{}

	ud := model.Device{
		Description: "Foobar",
		Identifier:  "foo",
		LastSeenAt:  time.Now(), // should not be updated
		Present:     true,       // should not be updated
		Status:      model.StatusIgnored,
	}
	d, err := registry.UpdateDevice("foo", ud)
	assert.Nil(t, err)
	assert.Equal(t, "Foobar", d.Description)
	assert.True(t, d.LastSeenAt.IsZero())
	assert.False(t, d.Present)
	assert.Equal(t, model.StatusIgnored, d.Status)

	ud.Identifier = "bar"
	_, err = registry.UpdateDevice("foo", ud)
	assert.Equal(t, ErrInvalidID, err)

	_, err = registry.UpdateDevice("miss", ud)
	assert.Equal(t, ErrNotFound, err)
}
