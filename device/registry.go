package device

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
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
	mqttClient MQTT.Client
	mqttTopic  string
	watchdog   *watchdog
}

// NewRegistry builds a new device registry.
func NewRegistry(cfg config.Config) *Registry {
	devices := make(map[string]*model.Device)
	for identifier, d := range cfg.Devices {
		devices[identifier] = d
	}
	var mqttClient MQTT.Client
	if cfg.MQTTServer.Enabled {
		mqttClient = newMQTTClient(cfg.MQTTServer)
	}
	return &Registry{
		devices:    devices,
		mqttClient: mqttClient,
		mqttTopic:  cfg.MQTTServer.Topic,
		watchdog:   newWatchDog(cfg),
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
			match := true
			if itf.Type != model.InterfaceUnknown {
				match = match && (itf.Type == di.Type)
			}
			if len(itf.Address) > 0 {
				match = match && (itf.Address == di.Address)
			}
			if len(itf.IPv4Address) > 0 {
				match = match && (itf.IPv4Address == di.IPv4Address)
			}
			if match {
				return d
			}
		}
	}
	return nil
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

func (r *Registry) reportPresence(itf model.Interface) {
	d := r.lookupDevice(itf)
	if d == nil {
		d = r.newDevice(itf)
	}
	d.Present = true
	d.LastSeenAt = time.Now()
	r.onPresenceUpdated(d)
	log.Info("Device '", d.Description, "' is present")
}

// Start activates the tracking of devices.
func (r *Registry) Start() {
	log.Info("Starting: registry")
	r.connect()
	go r.watchdog.loop(r)
}

// Stop de-activates the tracking of devices.
func (r *Registry) Stop() {
	log.Info("Stopping: registry")
	r.watchdog.stop()
	r.disconnect()
	log.Info("Stopped: registry")
}

// UpdateDevice updates an existing device.
func (r *Registry) UpdateDevice(id string, ud model.Device) (model.Device, error) {
	d, found := r.devices[id]
	if !found {
		return model.Device{}, ErrNotFound
	}
	if id != ud.Identifier {
		return model.Device{}, ErrInvalidID
	}

	d.Description = ud.Description
	d.Identifier = ud.Identifier
	d.Interfaces = ud.Interfaces
	d.Status = ud.Status
	return *d, nil
}

func (r *Registry) updateDevicesPresence(t time.Time) {
	for _, d := range r.devices {
		if d.Status == model.StatusTracked {
			elapsedMinutes := t.Sub(d.LastSeenAt).Minutes()
			if elapsedMinutes >= 10 {
				d.Present = false
				r.onPresenceUpdated(d)
				log.Info("Device '", d.Description, "' is not present")
			}
		}
	}
}
