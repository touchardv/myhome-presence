package device

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/internal/config"
	"github.com/touchardv/myhome-presence/pkg/model"
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

func (r *Registry) connect(ctx context.Context) {
	if r.mqttClient == nil {
		return
	}
	log.Info("Connecting to MQTT")
	retry := time.NewTicker(5 * time.Second)

connectLoop:
	for {
		if token := r.mqttClient.Connect(); token.Wait() && token.Error() != nil {
			log.Error("Failed to connect to MQTT: ", token.Error())
		} else {
			retry.Stop()
			log.Info("Connected to MQTT")
			break connectLoop
		}

		select {
		case <-retry.C:
			continue

		case <-ctx.Done():
			retry.Stop()
			break connectLoop
		}
	}
}

func (r *Registry) disconnect() {
	if r.mqttClient == nil {
		return
	}
	if r.mqttClient.IsConnected() {
		log.Info("Disconnecting from MQTT")
		r.mqttClient.Disconnect(500)
		log.Info("Disconnected from MQTT")
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
		Properties:  d.Properties,
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
			log.Info("Device '", d.Description, "' is now tracked")
			r.publish(model.EventTypeAdded, model.DeviceAdded{
				Description: d.Description,
				Identifier:  d.Identifier,
				Present:     d.Present,
				Properties:  d.Properties,
				LastSeenAt:  d.LastSeenAt,
			})
		}

	case model.StatusTracked:
		if d.Status == model.StatusIgnored {
			log.Info("Device '", d.Description, "' is now ignored")
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
					Properties:  d.Properties,
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
