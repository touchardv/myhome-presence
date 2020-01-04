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
	devices   map[string]*Device
	ipTracker ipTracker
}

// NewRegistry builds a new device registry.
func NewRegistry(config config.Config) *Registry {
	devices := make(map[string]*Device, 0)
	for _, d := range config.IPDevices {
		device := Device{Device: d, Present: false}
		devices[device.Identifier] = &device
	}
	return &Registry{devices, newIPTracker()}
}

// GetDevices returns all tracked devices.
func (r *Registry) GetDevices() []Device {
	devices := make([]Device, 0)
	for _, d := range r.devices {
		devices = append(devices, *d)
	}
	return devices
}

func (r *Registry) notifyPresent(device Device) {
	if d, ok := r.devices[device.Identifier]; ok {
		if d.Present == false {
			log.Info("Device '", device.Description, "' is present")
			d.LastSeenAt = time.Now()
			d.Present = true
		}
	} else {
		log.Warn("Unknown device: ", device.Identifier)
		return
	}
}

// Start activates the tracking of devices.
func (r *Registry) Start() {
	log.Info("Starting: registry")
	r.ipTracker.track(r.GetDevices(), r)
}

// Stop de-activates the tracking of devices.
func (r *Registry) Stop() {
	log.Info("Stopping: registry")
	r.ipTracker.stop()
	log.Info("Stopped: registry")
}
