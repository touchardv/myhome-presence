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
		{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "BLE"},
		{Type: model.InterfaceBluetooth, MACAddress: "BT"},
		{Type: model.InterfaceWifi, IPv4Address: "1.2.3.4"},
	},
}

var cfg = config.Config{
	Devices:  map[string]*model.Device{"foo": &device},
	Trackers: map[string]config.Settings{"dummy": {}},
}

func TestAddDevice(t *testing.T) {
	registry := NewRegistry(cfg)
	registry.AddDevice(model.Device{Identifier: "bar"})

	devices := registry.GetDevices()
	assert.Equal(t, 2, len(devices))
}

func TestFindAnExistingDevice(t *testing.T) {
	registry := NewRegistry(cfg)

	d, err := registry.FindDevice("foo")
	assert.Nil(t, err)
	assert.Equal(t, d.Identifier, "foo")
}

func TestFindAnUnknownDevice(t *testing.T) {
	registry := NewRegistry(cfg)

	d, err := registry.FindDevice("unknown")
	assert.NotNil(t, err)
	assert.Equal(t, d.Identifier, "")
}

func TestGetDevices(t *testing.T) {
	registry := NewRegistry(cfg)
	devices := registry.GetDevices()

	assert.Equal(t, 1, len(devices))
	assert.Equal(t, "foo", devices[0].Identifier)
}

func TestReportPresenceOfAExistingDevice(t *testing.T) {
	registry := NewRegistry(cfg)
	devices := registry.GetDevices()

	assert.Equal(t, 1, len(devices))
	assert.False(t, devices[0].Present)
	assert.True(t, devices[0].LastSeenAt.IsZero())

	// matching the interface type
	registry.reportPresence(model.Interface{Type: model.InterfaceWifi, IPv4Address: "1.2.3.4"})

	devices = registry.GetDevices()
	assert.Equal(t, 1, len(devices))

	assert.True(t, devices[0].Present)
	assert.Equal(t, "foo", devices[0].Identifier)
	assert.False(t, devices[0].LastSeenAt.IsZero())

	// with an unknown interface type
	registry.reportPresence(model.Interface{Type: model.InterfaceUnknown, IPv4Address: "1.2.3.4"})

	devices = registry.GetDevices()
	assert.Equal(t, 1, len(devices))

	assert.True(t, devices[0].Present)
	assert.Equal(t, "foo", devices[0].Identifier)
	assert.False(t, devices[0].LastSeenAt.IsZero())
}

func TestReportPresenceOfANewDevice(t *testing.T) {
	registry := NewRegistry(config.Config{Devices: map[string]*model.Device{}})

	registry.reportPresence(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "12:34:56:78:90"})

	devices := registry.GetDevices()
	assert.Equal(t, 1, len(devices))

	assert.True(t, devices[0].Present)
	assert.False(t, devices[0].LastSeenAt.IsZero())
	assert.Equal(t, model.InterfaceBluetoothLowEnergy, devices[0].Interfaces[0].Type)
	assert.Equal(t, "12:34:56:78:90", devices[0].Interfaces[0].MACAddress)
	assert.Equal(t, model.StatusDiscovered, devices[0].Status)
}

func TestNewDevice(t *testing.T) {
	registry := NewRegistry(cfg)
	d := registry.newDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "one"})

	devices := registry.GetDevices()
	assert.Equal(t, 2, len(devices))
	assert.NotEmpty(t, d.Identifier, d.Description)
	assert.Equal(t, model.InterfaceBluetoothLowEnergy, d.Interfaces[0].Type)
	assert.Equal(t, "one", d.Interfaces[0].MACAddress)
	assert.Equal(t, model.StatusDiscovered, d.Status)

	d = registry.newDevice(model.Interface{Type: model.InterfaceBluetooth, MACAddress: "two"})
	assert.Equal(t, model.InterfaceBluetooth, d.Interfaces[0].Type)
	assert.Equal(t, "two", d.Interfaces[0].MACAddress)

	d = registry.newDevice(model.Interface{Type: model.InterfaceWifi, IPv4Address: "three"})
	assert.Equal(t, 1, len(d.Interfaces))
	assert.Equal(t, model.InterfaceWifi, d.Interfaces[0].Type)
	assert.Equal(t, "three", d.Interfaces[0].IPv4Address)
}

func TestLookupDevice(t *testing.T) {
	registry := NewRegistry(cfg)

	d := registry.lookupDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "BLE"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(model.Interface{Type: model.InterfaceBluetooth, MACAddress: "BT"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(model.Interface{Type: model.InterfaceWifi, IPv4Address: "1.2.3.4"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "foobar"})
	assert.Nil(t, d)
}

func TestRemoveDevicee(t *testing.T) {
	registry := NewRegistry(cfg)

	registry.RemoveDevice("foo")
	devices := registry.GetDevices()
	assert.Equal(t, 0, len(devices))
}

func TestRegistryStartStop(t *testing.T) {
	registry := NewRegistry(cfg)
	registry.Start()
	registry.Stop()
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

func TestUpdateDevicesPresence(t *testing.T) {
	registry := NewRegistry(cfg)
	device.LastSeenAt = time.Now()
	device.Present = true
	device.Status = model.StatusTracked

	registry.updateDevicesPresence(time.Now().Add(5 * time.Minute))
	assert.True(t, device.Present)

	registry.updateDevicesPresence(time.Now().Add(11 * time.Minute))
	assert.False(t, device.Present)
}
