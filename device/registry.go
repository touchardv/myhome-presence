package device

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/model"
)

var (
	ErrNotFound       = errors.New("device not found")
	ErrInvalidID      = errors.New("invalid device identifier")
	ErrIDAlreadyTaken = errors.New("device identifier already taken")
)

// Registry maintains the status of all tracked devices
// together with their presence status.
type Registry struct {
	devices    map[string]*model.Device
	mutex      *sync.RWMutex
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
		mutex:      &sync.RWMutex{},
		mqttClient: mqttClient,
		mqttTopic:  cfg.MQTTServer.Topic,
		watchdog:   newWatchDog(cfg),
	}
}

// AddDevice adds a new device to the registry.
func (r *Registry) AddDevice(d model.Device) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(strings.TrimSpace(d.Identifier)) == 0 {
		return ErrInvalidID
	}
	if _, found := r.devices[d.Identifier]; found {
		return ErrIDAlreadyTaken
	}
	d.CreatedAt = time.Now()
	r.devices[d.Identifier] = &d
	r.onAdded(&d)
	log.Info("Device added: ", d.Identifier)
	return nil
}

// ContactDevice attempts to contact a device given its identifier.
func (r *Registry) ContactDevice(id string) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if d, found := r.devices[id]; found {
		r.watchdog.ping(d)
		return nil
	}
	return ErrNotFound
}

// FindDevice lookups a device given its identifier.
func (r *Registry) FindDevice(id string) (model.Device, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if d, found := r.devices[id]; found {
		return *d, nil
	}
	return model.Device{}, ErrNotFound
}

// GetDevices returns all known devices.
func (r *Registry) GetDevices(status model.Status) []model.Device {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	devices := make([]model.Device, 0)
	for _, d := range r.devices {
		if status == model.StatusUndefined || status == d.Status {
			devices = append(devices, *d)
		}
	}
	return devices
}

func (r *Registry) newDevice(itf model.Interface, optData map[string]string) *model.Device {
	now := time.Now()
	id := identifier(optData)
	if _, found := r.devices[id]; found {
		id = fmt.Sprintf("%s-%s", id, now.Format(time.RFC3339))
	}
	return &model.Device{
		Description: description(optData, now),
		Identifier:  id,
		Interfaces:  []model.Interface{itf},
		Present:     true,
		Properties:  optData,
		CreatedAt:   now,
		FirstSeenAt: now,
		LastSeenAt:  now,
		Status:      model.StatusDiscovered,
	}
}

func description(optData map[string]string, ts time.Time) string {
	if optData != nil {
		if description, ok := optData[ReportDataSuggestedDescription]; ok {
			return description
		}
	}
	return fmt.Sprintf("Unidentified device seen at %s", ts.Format(time.RFC822))
}

func identifier(optData map[string]string) string {
	if optData != nil {
		if id, ok := optData[ReportDataSuggestedIdentifier]; ok {
			return strings.ToLower(id)
		}
	}
	return "unidentified-device"
}

func (r *Registry) lookupDevice(itf model.Interface) *model.Device {
	for _, d := range r.devices {
		for _, di := range d.Interfaces {
			match := true
			if itf.Type != model.InterfaceUnknown {
				match = match && (itf.Type == di.Type)
			}
			if len(itf.MACAddress) > 0 {
				match = match && (itf.MACAddress == di.MACAddress)
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
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, found := r.devices[id]; found {
		delete(r.devices, id)
		r.onRemoved(id)
		log.Info("Device removed: ", id)
		return nil
	}
	return ErrNotFound
}

func (r *Registry) reportPresence(itf model.Interface, optData map[string]string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	itf = sanitized(itf)
	d := r.lookupDevice(itf)
	if d == nil {
		d = r.newDevice(itf, optData)
		r.devices[d.Identifier] = d
		log.Info("Discovered a new device: ", d.Identifier)
	} else {
		// Merge device properties
		if optData != nil {
			if d.Properties == nil {
				d.Properties = optData
			} else {
				for k, v := range optData {
					d.Properties[k] = v
				}
			}
		}
		wasNotPresent := !d.Present
		lastReportAt := d.LastSeenAt

		now := time.Now()
		if wasNotPresent {
			d.FirstSeenAt = now
			d.Present = true
		}
		d.LastSeenAt = now

		if d.Status == model.StatusTracked {
			if wasNotPresent {
				log.Info("Device '", d.Description, "' is present")
			}
			elapsedMinutes := now.Sub(lastReportAt).Minutes()
			if elapsedMinutes > 1 {
				r.onPresenceUpdated(d)
			}
		}
	}
}

func sanitized(in model.Interface) model.Interface {
	return model.Interface{
		Type:        in.Type,
		MACAddress:  strings.ToLower(in.MACAddress),
		IPv4Address: in.IPv4Address,
	}
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
	r.mutex.Lock()
	defer r.mutex.Unlock()

	d, found := r.devices[id]
	if !found {
		return model.Device{}, ErrNotFound
	}
	if id != ud.Identifier {
		return model.Device{}, ErrInvalidID
	}

	// identifier, creation date are left untouched
	d.Description = ud.Description
	d.Interfaces = ud.Interfaces
	d.Properties = ud.Properties
	d.Status = ud.Status
	return *d, nil
}

func (r *Registry) UpdateDevicesPresence(t time.Time) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	removedIDs := make([]string, 0)
	for _, d := range r.devices {
		elapsedMinutes := t.Sub(d.LastSeenAt).Minutes()

		switch d.Status {
		case model.StatusDiscovered:
			if elapsedMinutes > 60 {
				removedIDs = append(removedIDs, d.Identifier)
			} else if elapsedMinutes >= 10 {
				d.Present = false
			}

		case model.StatusTracked:
			if elapsedMinutes >= 10 {
				if d.Present {
					d.Present = false
					log.Info("Device '", d.Description, "' is not present")
					r.onPresenceUpdated(d)
				}
			}
		}
	}

	for _, id := range removedIDs {
		delete(r.devices, id)
		log.Debug("Discovered device automatically removed: ", id)
	}
}
