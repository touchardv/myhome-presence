package device

import (
	"encoding/json"
	"fmt"
	"os"

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
	if token := r.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}
}

func (r *Registry) disconnect() {
	if r.mqttClient == nil {
		return
	}
	if r.mqttClient.IsConnected() {
		r.mqttClient.Disconnect(500)
	}
}

func (r *Registry) onAdded(d *model.Device) {
	r.publish(model.EventTypeAdded, model.DeviceAdded{
		Description: d.Description,
		Identifier:  d.Identifier,
		Present:     d.Present,
		LastSeenAt:  d.LastSeenAt,
	})
}

func (r *Registry) onPresenceUpdated(d *model.Device) {
	r.publish(model.EventTypePresenceUpdated, model.DevicePresenceUpdated{
		Identifier: d.Identifier,
		Present:    d.Present,
		LastSeenAt: d.LastSeenAt,
	})
}

func (r *Registry) onRemoved(id string) {
	r.publish(model.EventTypeRemoved, model.DeviceRemoved{
		Identifier: id,
	})
}

func (r *Registry) publish(t model.EventType, itf interface{}) {
	if r.mqttClient == nil {
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
