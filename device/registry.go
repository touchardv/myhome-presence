package device

import (
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
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
	devices    map[string]*Device
	trackers   []Tracker
	mqttClient MQTT.Client
	mqttTopic  string
	stopping   chan struct{}
	waitGroup  sync.WaitGroup
}

// NewRegistry builds a new device registry.
func NewRegistry(config config.Config) *Registry {
	devices := make(map[string]*Device, 0)
	for _, d := range config.Devices {
		device := Device{Device: d, Present: false}
		devices[device.Identifier] = &device
	}
	return &Registry{
		devices: devices,
		trackers: []Tracker{
			newTracker("bluetooth"),
			newTracker("ipv4"),
		},
		mqttClient: newMQTTClient(config.MQTTServer),
		mqttTopic:  config.MQTTServer.Topic,
		stopping:   make(chan struct{}),
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
	check := time.NewTimer(1 * time.Minute)
	for {
		select {
		case <-check.C:
			now := time.Now()
			for _, d := range r.devices {
				if d.Present && now.Sub(d.LastSeenAt).Minutes() > 5 {
					log.Info("Device '", d.Description, "' is not present")
					d.Present = false
					r.publishPresence(d)
				}
			}
			check.Reset(1 * time.Minute)
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
				r.publishPresence(d)
			} else {
				log.Warn("Unknown device: ", identifier)
			}
		}
	}
}

// Start activates the tracking of devices.
func (r *Registry) Start() {
	log.Info("Starting: registry")
	r.connect()
	presence := make(chan string, 10)
	r.waitGroup.Add(1)
	go func() {
		r.handle(presence)
		r.waitGroup.Done()
	}()

	devices := make([]config.Device, 0)
	for _, d := range r.devices {
		devices = append(devices, d.Device)
	}
	for _, t := range r.trackers {
		r.waitGroup.Add(1)
		go func(t Tracker) {
			t.Track(devices, presence, r.stopping)
			r.waitGroup.Done()
		}(t)
	}
}

// Stop de-activates the tracking of devices.
func (r *Registry) Stop() {
	log.Info("Stopping: registry")
	close(r.stopping)
	r.waitGroup.Wait()
	r.disconnect()
	log.Info("Stopped: registry")
}
