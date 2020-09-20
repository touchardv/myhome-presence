package device

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/touchardv/myhome-presence/config"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type deviceAdded struct {
	Description string    `json:"description"`
	Identifier  string    `json:"identifier"`
	Present     bool      `json:"present"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}

type devicePresenceUpdated struct {
	Identifier string    `json:"identifier"`
	Present    bool      `json:"present"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type deviceRemoved struct {
	Identifier string `json:"identifier"`
}

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
		return "unknown"
	}
	return hostname
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

func (r *Registry) onAdded(d *config.Device) {
	r.publish(deviceAdded{
		Description: d.Description,
		Identifier:  d.Identifier,
		Present:     d.Present,
		LastSeenAt:  d.LastSeenAt,
	})
}

func (r *Registry) onPresenceUpdated(d *config.Device) {
	r.publish(devicePresenceUpdated{
		Identifier: d.Identifier,
		Present:    d.Present,
		LastSeenAt: d.LastSeenAt,
	})
}

func (r *Registry) onRemoved(id string) {
	r.publish(deviceRemoved{
		Identifier: id,
	})
}

func (r *Registry) publish(data interface{}) {
	if r.mqttClient == nil {
		return
	}
	bytes, err := json.Marshal(data)
	if err == nil {
		r.mqttClient.Publish(r.mqttTopic, 0, false, bytes)
	} else {
		log.Error(err)
	}
}
