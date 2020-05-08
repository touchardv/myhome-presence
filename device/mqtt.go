package device

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/touchardv/myhome-presence/config"

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
		return "unknown"
	}
	return hostname
}

func (r *Registry) connect() {
	if token := r.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}
}

func (r *Registry) disconnect() {
	if r.mqttClient.IsConnected() {
		r.mqttClient.Disconnect(500)
	}
}

func (r *Registry) publishPresence(d *config.Device) {
	bytes, err := json.Marshal(d)
	if err == nil {
		r.mqttClient.Publish(r.mqttTopic, 0, false, bytes)
	} else {
		log.Error(err)
	}
}
