package device

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/model"
)

var (
	ErrNotFound       = errors.New("Device not found")
	ErrInvalidID      = errors.New("Invalid device identifier")
	ErrIDAlreadyTaken = errors.New("Device identifier already taken")
)

// Registry maintains the status of all tracked devices
// together with their presence status.
type Registry struct {
	devices    map[string]*model.Device
	trackers   []Tracker
	mqttClient MQTT.Client
	mqttTopic  string
	stopping   chan struct{}
	waitGroup  sync.WaitGroup
}

// NewRegistry builds a new device registry.
func NewRegistry(cfg config.Config) *Registry {
	devices := cfg.Devices
	trackers := make([]Tracker, 0)
	for _, name := range cfg.Trackers {
		trackers = append(trackers, newTracker(name))
	}
	var mqttClient MQTT.Client
	if cfg.MQTTServer.Enabled {
		mqttClient = newMQTTClient(cfg.MQTTServer)
	}
	return &Registry{
		devices:    devices,
		trackers:   trackers,
		mqttClient: mqttClient,
		mqttTopic:  cfg.MQTTServer.Topic,
		stopping:   make(chan struct{}),
	}
}

// AddDevice adds a new device to the registry.
func (r *Registry) AddDevice(d model.Device) error {
	if len(strings.TrimSpace(d.Identifier)) == 0 {
		return ErrInvalidID
	}
	if _, found := r.devices[d.Identifier]; found {
		return ErrIDAlreadyTaken
	}
	r.devices[d.Identifier] = &d
	r.onAdded(&d)
	log.Info("Device added: ", d.Identifier)
	return nil
}

// FindDevice lookups a device given its identifier.
func (r *Registry) FindDevice(id string) (model.Device, error) {
	if d, found := r.devices[id]; found {
		return *d, nil
	}
	return model.Device{}, ErrNotFound
}

// GetDevices returns all known devices.
func (r *Registry) GetDevices() []model.Device {
	devices := make([]model.Device, 0)
	for _, d := range r.devices {
		devices = append(devices, *d)
	}
	return devices
}

func (r *Registry) handle(scan chan model.Interface, presence chan string) {
	log.Info("Starting: scan/presence handler")
	for {
		select {
		case <-r.stopping:
			log.Info("Stopped: scan/presence handler")
			return

		case s := <-scan:
			d := r.lookupDevice(s)
			if d == nil {
				d = r.newDevice(s)
			}
			if d.Present == false {
				log.Info("Device '", d.Description, "' is present")
				d.Present = true
			}
			d.LastSeenAt = time.Now()
			r.onPresenceUpdated(d)
			break

		case identifier := <-presence:
			if d, ok := r.devices[identifier]; ok {
				if d.Present == false {
					log.Info("Device '", d.Description, "' is present")
					d.Present = true
				}
				d.LastSeenAt = time.Now()
				r.onPresenceUpdated(d)
			} else {
				log.Warn("Unknown device: ", identifier)
			}
		}
	}
}

func (r *Registry) newDevice(itf model.Interface) *model.Device {
	now := time.Now()
	d := model.Device{
		Description: fmt.Sprintf("Discovered device at %s", now.Format(time.RFC822)),
		Identifier:  fmt.Sprintf("device-%d-%d", now.Unix(), rand.Intn(1000)),
		Interfaces:  make([]model.Interface, 1),
		Present:     false,
		Status:      model.StatusDiscovered,
	}
	d.Interfaces[0] = itf
	r.devices[d.Identifier] = &d
	r.onAdded(&d)
	log.Info("Discovered a new device: ", d.Identifier)
	return &d
}

func (r *Registry) lookupDevice(itf model.Interface) *model.Device {
	for _, d := range r.devices {
		for _, di := range d.Interfaces {
			if di.Type == itf.Type && di.Address == itf.Address {
				return d
			}
		}
	}
	return nil
}

func (r *Registry) pingLoop(presence chan string) {
	log.Info("Starting: device watchdog")
	check := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-r.stopping:
			log.Info("Stopped: device watchdog")
			return

		case <-check.C:
			r.pingMissingDevices(presence)
			check.Reset(1 * time.Minute)
		}
	}
}

func (r *Registry) pingMissingDevices(presence chan string) {
	missing := make(map[string]model.Device)
	now := time.Now()
	for _, d := range r.devices {
		if d.Status != model.StatusTracked {
			continue
		}
		elapsedMinutes := now.Sub(d.LastSeenAt).Minutes()
		if d.Present == false || elapsedMinutes >= 5 {
			missing[d.Identifier] = *d
		}
	}

	if len(missing) > 0 {
		for _, t := range r.trackers {
			t.Ping(missing, presence)
		}

		if len(missing) > 0 {
			for _, m := range missing {
				d := r.devices[m.Identifier]
				if d == nil {
					continue
				}
				elapsedMinutes := now.Sub(d.LastSeenAt).Minutes()
				if d.Present && elapsedMinutes > 10 {
					d.Present = false
					r.onPresenceUpdated(d)
					log.Info("Device '", d.Description, "' is not present")
				}
			}
		}
	}
}

// RemoveDevice removes a device.
func (r *Registry) RemoveDevice(id string) error {
	if _, found := r.devices[id]; found {
		delete(r.devices, id)
		r.onRemoved(id)
		log.Info("Device removed: ", id)
		return nil
	}
	return ErrNotFound
}

// Start activates the tracking of devices.
func (r *Registry) Start() {
	log.Info("Starting: registry")
	r.connect()
	existence := make(chan model.Interface, 10)
	presence := make(chan string, 10)
	go func() {
		r.handle(existence, presence)
		r.waitGroup.Done()
	}()
	go func() {
		r.pingLoop(presence)
		r.waitGroup.Done()
	}()
	r.waitGroup.Add(2)

	for _, t := range r.trackers {
		r.waitGroup.Add(1)
		go func(t Tracker) {
			t.Scan(existence, r.stopping)
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

// UpdateDevice updates an existing device.
func (r *Registry) UpdateDevice(id string, ud model.Device) (model.Device, error) {
	d, found := r.devices[id]
	if !found {
		return model.Device{}, ErrNotFound
	}
	if len(strings.TrimSpace(ud.Identifier)) == 0 {
		return model.Device{}, ErrInvalidID
	}
	if id != ud.Identifier {
		if _, found := r.devices[ud.Identifier]; found {
			return model.Device{}, ErrIDAlreadyTaken
		}
		r.devices[ud.Identifier] = d
		delete(r.devices, id)
		log.Infof("Device '%s' renamed to '%s'", id, ud.Identifier)
	}

	*d = ud
	return *d, nil
}
