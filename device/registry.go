package device

import (
	"time"

	"github.com/touchardv/myhome-presence/config"
)

// Device represents a tracked device and its presence status.
type Device struct {
	config.Device
	Present    bool
	LastSeenAt time.Time
}

// Registry maintains the status of all tracked devices
// together with their presence status.
type Registry struct {
	devices []Device
}

// NewRegistry builds a new device registry.
func NewRegistry(config config.Config) *Registry {
	devices := make([]Device, 0)
	for _, d := range config.IPDevices {
		device := Device{Device: d, Present: false}
		devices = append(devices, device)
	}
	return &Registry{devices}
}

// GetDevices returns all tracked devices.
func (r *Registry) GetDevices() []Device {
	return r.devices
}
