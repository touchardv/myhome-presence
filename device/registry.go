package device

import (
	"time"

	log "github.com/sirupsen/logrus"

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
	devices   []Device
	ipTracker ipTracker
}

// NewRegistry builds a new device registry.
func NewRegistry(config config.Config) *Registry {
	devices := make([]Device, 0)
	for _, d := range config.IPDevices {
		device := Device{Device: d, Present: false}
		devices = append(devices, device)
	}
	return &Registry{devices, newIPTracker()}
}

// GetDevices returns all tracked devices.
func (r *Registry) GetDevices() []Device {
	return r.devices
}

func (r *Registry) notify(device Device, present bool) {
	log.Info("Device ", device.Identifier, " presence=", present)
}

// Start activates the tracking of devices.
func (r *Registry) Start() {
	log.Info("Starting: registry")
	r.ipTracker.track(r.devices, r)
}

// Stop de-activates the tracking of devices.
func (r *Registry) Stop() {
	log.Info("Stopping: registry")
	r.ipTracker.stop()
	log.Info("Stopped: registry")
}
