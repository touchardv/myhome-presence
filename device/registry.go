package device

import (
	"sync"
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
	stopping  chan struct{}
	waitGroup sync.WaitGroup
}

// NewRegistry builds a new device registry.
func NewRegistry(config config.Config) *Registry {
	devices := make(map[string]*Device, 0)
	for _, d := range config.IPDevices {
		device := Device{Device: d, Present: false}
		devices[device.Identifier] = &device
	}
	return &Registry{
		devices:   devices,
		ipTracker: newIPTracker(),
		stopping:  make(chan struct{}),
	}
}

// GetDevices returns all tracked devices.
func (r *Registry) GetDevices() []Device {
	devices := make([]Device, 0)
	for _, d := range r.devices {
		devices = append(devices, *d)
	}
	return devices
}

func (r *Registry) handle(presence chan string) {
	log.Info("Starting: presence handler")
	for {
		check := time.NewTimer(1 * time.Minute)
		select {
		case <-check.C:
			now := time.Now()
			for _, d := range r.devices {
				if d.Present && now.Sub(d.LastSeenAt).Minutes() > 5 {
					log.Info("Device '", d.Description, "' is not present")
					d.Present = false
				}
			}
		case <-r.stopping:
			log.Info("Stopped: presence handler")
			return
		case identifier := <-presence:
			if d, ok := r.devices[identifier]; ok {
				if d.Present == false {
					log.Info("Device '", d.Description, "' is present")
					d.Present = true
				}
				d.LastSeenAt = time.Now()
			} else {
				log.Warn("Unknown device: ", identifier)
			}
		}
	}
}

// Start activates the tracking of devices.
func (r *Registry) Start() {
	log.Info("Starting: registry")
	presence := make(chan string, 10)
	r.waitGroup.Add(1)
	go func() {
		r.handle(presence)
		r.waitGroup.Done()
	}()

	r.waitGroup.Add(1)
	go func() {
		devices := make([]config.Device, 0)
		for _, d := range r.devices {
			devices = append(devices, d.Device)
		}
		r.ipTracker.track(devices, presence, r.stopping)
		r.waitGroup.Done()
	}()
}

// Stop de-activates the tracking of devices.
func (r *Registry) Stop() {
	log.Info("Stopping: registry")
	close(r.stopping)
	r.waitGroup.Wait()
	log.Info("Stopped: registry")
}
