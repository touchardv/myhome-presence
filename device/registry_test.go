package device

import (
	"strings"
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
		{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "bb:77:33:00:00:00"},
		{Type: model.InterfaceBluetooth, MACAddress: "bb:11:00:00:00:00"},
		{Type: model.InterfaceWifi, IPv4Address: "1.2.3.4"},
	},
	Status: model.StatusTracked,
}

var cfg = config.Config{
	Devices:  map[string]*model.Device{"foo": &device},
	Trackers: map[string]config.Settings{"dummy": {}},
}

func TestAddDevice(t *testing.T) {
	registry := NewRegistry(cfg)
	registry.AddDevice(model.Device{Identifier: "bar"})

	devices := registry.GetDevices(model.StatusUndefined)
	assert.Equal(t, 2, len(devices))
	assert.NotZero(t, devices[1].CreatedAt)
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
	registry.AddDevice(model.Device{Identifier: "discobar", Status: model.StatusDiscovered})
	registry.AddDevice(model.Device{Identifier: "bad", Status: model.StatusIgnored})

	devices := registry.GetDevices(model.StatusUndefined)
	assert.Equal(t, 3, len(devices))

	devices = registry.GetDevices(model.StatusTracked)
	assert.Equal(t, 1, len(devices))
	assert.Equal(t, "foo", devices[0].Identifier)

	devices = registry.GetDevices(model.StatusDiscovered)
	assert.Equal(t, 1, len(devices))
	assert.Equal(t, "discobar", devices[0].Identifier)

	devices = registry.GetDevices(model.StatusIgnored)
	assert.Equal(t, 1, len(devices))
	assert.Equal(t, "bad", devices[0].Identifier)
}

func TestReportPresenceOfAExistingDevice(t *testing.T) {
	registry := NewRegistry(cfg)
	devices := registry.GetDevices(model.StatusUndefined)

	assert.Equal(t, 1, len(devices))
	assert.False(t, devices[0].Present)
	assert.True(t, devices[0].LastSeenAt.IsZero())

	// matching the interface type (via IP address)
	registry.reportPresence(model.Interface{Type: model.InterfaceWifi, IPv4Address: "1.2.3.4"}, nil)

	devices = registry.GetDevices(model.StatusUndefined)
	assert.Equal(t, 1, len(devices))

	assert.True(t, devices[0].Present)
	assert.Equal(t, "foo", devices[0].Identifier)
	assert.False(t, devices[0].LastSeenAt.IsZero())

	// matching the interface type (via uppercased MAC address)
	registry.reportPresence(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "BB:77:33:00:00:00"}, nil)
	devices = registry.GetDevices(model.StatusUndefined)
	assert.Equal(t, 1, len(devices))

	// with an unknown interface type
	registry.reportPresence(model.Interface{Type: model.InterfaceUnknown, IPv4Address: "1.2.3.4"}, nil)

	devices = registry.GetDevices(model.StatusUndefined)
	assert.Equal(t, 1, len(devices))

	assert.True(t, devices[0].Present)
	assert.Equal(t, "foo", devices[0].Identifier)
	assert.False(t, devices[0].LastSeenAt.IsZero())
}

func TestReportPresenceOfANewDevice(t *testing.T) {
	registry := NewRegistry(config.Config{Devices: map[string]*model.Device{}})

	registry.reportPresence(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "12:34:56:78:9A"}, nil)

	devices := registry.GetDevices(model.StatusUndefined)
	assert.Equal(t, 1, len(devices))

	assert.True(t, devices[0].Present)
	assert.NotZero(t, devices[0].CreatedAt)
	assert.NotZero(t, devices[0].FirstSeenAt)
	assert.NotZero(t, devices[0].LastSeenAt)
	assert.Equal(t, model.InterfaceBluetoothLowEnergy, devices[0].Interfaces[0].Type)
	assert.Equal(t, "12:34:56:78:9a", devices[0].Interfaces[0].MACAddress)
	assert.Equal(t, model.StatusDiscovered, devices[0].Status)
}

func TestNewDevice(t *testing.T) {
	registry := NewRegistry(cfg)
	d := registry.newDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "one"}, nil)

	assert.NotEmpty(t, d.Identifier)
	assert.NotEmpty(t, d.Description)
	assert.Equal(t, model.StatusDiscovered, d.Status)
	assert.True(t, d.Present)
	assert.Equal(t, 1, len(d.Interfaces))
	assert.Equal(t, "one", d.Interfaces[0].MACAddress)
	assert.Equal(t, model.InterfaceBluetoothLowEnergy, d.Interfaces[0].Type)

	d = registry.newDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "two"}, map[string]string{
		ReportDataSuggestedIdentifier: "baz",
	})
	assert.Equal(t, "baz", d.Identifier) // no ID conflict

	d = registry.newDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "three"}, map[string]string{
		ReportDataSuggestedIdentifier: "foo",
	})
	assert.NotEqual(t, "foo", d.Identifier)
	assert.True(t, strings.HasPrefix(d.Identifier, "foo-")) // ID exists => suffix is appended
}

func TestLookupDevice(t *testing.T) {
	registry := NewRegistry(cfg)

	d := registry.lookupDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "bb:77:33:00:00:00"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(model.Interface{Type: model.InterfaceBluetooth, MACAddress: "bb:11:00:00:00:00"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(model.Interface{Type: model.InterfaceWifi, IPv4Address: "1.2.3.4"})
	assert.NotNil(t, d)
	assert.Equal(t, "foo", d.Identifier)

	d = registry.lookupDevice(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: "foobar"})
	assert.Nil(t, d)
}

func TestRemoveDevice(t *testing.T) {
	registry := NewRegistry(cfg)

	registry.RemoveDevice("foo")
	devices := registry.GetDevices(model.StatusUndefined)
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

func TestUpdateDiscoveredDevicePresence(t *testing.T) {
	registry := NewRegistry(cfg)
	device.LastSeenAt = time.Now()
	device.Present = true
	device.Status = model.StatusDiscovered

	registry.UpdateDevicesPresence(time.Now().Add(5 * time.Minute))
	assert.True(t, device.Present)

	registry.UpdateDevicesPresence(time.Now().Add(11 * time.Minute))
	assert.False(t, device.Present)

	registry.UpdateDevicesPresence(time.Now().Add(61 * time.Minute))
	assert.Equal(t, 0, len(registry.devices))
}

func TestUpdateTrackedDevicePresence(t *testing.T) {
	registry := NewRegistry(cfg)
	device.LastSeenAt = time.Now()
	device.Present = true
	device.Status = model.StatusTracked

	registry.UpdateDevicesPresence(time.Now().Add(5 * time.Minute))
	assert.True(t, device.Present)

	registry.UpdateDevicesPresence(time.Now().Add(11 * time.Minute))
	assert.False(t, device.Present)
}
