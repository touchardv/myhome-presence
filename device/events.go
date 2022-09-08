package device

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/model"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

func newMQTTClient(c config.MQTT) MQTT.Client {
	server := fmt.Sprintf("tcp://%s:%d", c.Hostname, c.Port)
	opts := MQTT.NewClientOptions().AddBroker(server)
	opts.SetAutoReconnect(true)
	opts.SetClientID(mqttClientID())
	return MQTT.NewClient(opts)
}

func mqttClientID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s-%d", hostname, os.Getpid())
}

func (r *Registry) connect() {
	if r.mqttClient == nil {
		return
	}
	log.Debug("Connecting to MQTT")
	if token := r.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}
}

func (r *Registry) disconnect() {
	if r.mqttClient == nil {
		return
	}
	if r.mqttClient.IsConnected() {
		log.Debug("Disconnecting from MQTT")
		r.mqttClient.Disconnect(500)
		log.Debug("Disconnected from MQTT")
	}
}

func (r *Registry) onAdded(d *model.Device) {
	if d.Status != model.StatusTracked {
		return
	}

	r.publish(model.EventTypeAdded, model.DeviceAdded{
		Description: d.Description,
		Identifier:  d.Identifier,
		Present:     d.Present,
		LastSeenAt:  d.LastSeenAt,
	})
}

func (r *Registry) onPresenceUpdated(d *model.Device) {
	if d.Status != model.StatusTracked {
		return
	}

	if d.Present {
		log.Info("Device '", d.Description, "' is present")
	} else {
		log.Info("Device '", d.Description, "' is not present")
	}
	r.publish(model.EventTypePresenceUpdated, model.DevicePresenceUpdated{
		Identifier: d.Identifier,
		Present:    d.Present,
		LastSeenAt: d.LastSeenAt,
	})
}

func (r *Registry) onUpdated(d *model.Device, previousStatus model.Status, previousUpdatedAt time.Time) {
	switch previousStatus {
	case model.StatusDiscovered, model.StatusIgnored:
		if d.Status == model.StatusTracked {
			r.publish(model.EventTypeAdded, model.DeviceAdded{
				Description: d.Description,
				Identifier:  d.Identifier,
				Present:     d.Present,
				LastSeenAt:  d.LastSeenAt,
			})
		}

	case model.StatusTracked:
		if d.Status == model.StatusIgnored {
			r.publish(model.EventTypeRemoved, model.DeviceRemoved{
				Identifier: d.Identifier,
			})
		} else if d.Status == model.StatusTracked {
			// limit the number of update events
			elapsed := time.Since(previousUpdatedAt)
			if elapsed.Minutes() > 1 {
				r.publish(model.EventTypeUpdated, model.DeviceUpdated{
					Identifier:  d.Identifier,
					Description: d.Description,
					Present:     d.Present,
					LastSeenAt:  d.LastSeenAt,
				})
			}
		}
	}
}

func (r *Registry) onRemoved(d *model.Device) {
	if d.Status != model.StatusTracked {
		return
	}

	r.publish(model.EventTypeRemoved, model.DeviceRemoved{
		Identifier: d.Identifier,
	})
}

func (r *Registry) publish(t model.EventType, itf interface{}) {
	if r.mqttClient == nil {
		log.Debugf("Event: %s - %s", t.String(), itf)
		return
	}
	data, err := json.Marshal(itf)
	if err == nil {
		bytes, err := json.Marshal(model.Event{Type: t, Data: data})
		if err == nil {
			r.mqttClient.Publish(r.mqttTopic, 0, false, bytes)
			return
		}
	}
	log.Error(err)
}
