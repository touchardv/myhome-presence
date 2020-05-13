package device

import (
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
)

// Registry maintains the status of all tracked devices
// together with their presence status.
type Registry struct {
	devices    map[string]*config.Device
	trackers   []Tracker
	mqttClient MQTT.Client
	mqttTopic  string
	stopping   chan struct{}
	waitGroup  sync.WaitGroup
}

// NewRegistry builds a new device registry.
func NewRegistry(cfg config.Config) *Registry {
	devices := make(map[string]*config.Device, 0)
	for _, d := range cfg.Devices {
		device := config.Device{}
		device = d
		devices[device.Identifier] = &device
	}
	trackers := make([]Tracker, 0)
	for _, name := range cfg.Trackers {
		trackers = append(trackers, newTracker(name))
	}
	return &Registry{
		devices:    devices,
		trackers:   trackers,
		mqttClient: newMQTTClient(cfg.MQTTServer),
		mqttTopic:  cfg.MQTTServer.Topic,
		stopping:   make(chan struct{}),
	}
}

// GetDevices returns all tracked devices.
func (r *Registry) GetDevices() []config.Device {
	devices := make([]config.Device, 0)
	for _, d := range r.devices {
		devices = append(devices, *d)
	}
	return devices
}

func (r *Registry) handle(presence chan string) {
	log.Info("Starting: presence handler")
	for {
		select {
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

func (r *Registry) pingMissingDevices(presence chan string) {
	log.Info("Starting: device watchdog")
	check := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-r.stopping:
			log.Info("Stopped: device watchdog")
			return

		case <-check.C:
			devices := make(map[string]config.Device)
			now := time.Now()
			for _, d := range r.devices {
				elapsedMinutes := now.Sub(d.LastSeenAt).Minutes()
				if d.Present && elapsedMinutes > 5 {
					log.Info("Device '", d.Description, "' is not present")
					d.Present = false
					r.publishPresence(d)
				}

				if d.Present == false || elapsedMinutes > 3 {
					devices[d.Identifier] = *d
				}
			}

			if len(devices) > 0 {
				for _, t := range r.trackers {
					t.Ping(devices, presence)
				}
			}
			check.Reset(1 * time.Minute)
		}
	}
}

// Start activates the tracking of devices.
func (r *Registry) Start() {
	log.Info("Starting: registry")
	r.connect()
	presence := make(chan string, 10)
	go func() {
		r.handle(presence)
		r.waitGroup.Done()
	}()
	go func() {
		r.pingMissingDevices(presence)
		r.waitGroup.Done()
	}()
	r.waitGroup.Add(2)

	for _, t := range r.trackers {
		r.waitGroup.Add(1)
		go func(t Tracker) {
			t.Scan(presence, r.stopping)
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
