package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

// Device represents a single device being tracked.
type Device struct {
	Description string
	Address     string
	Identifier  string
}

// MQTT contains the MQTT server connection information.
type MQTT struct {
	Hostname string
	Port     uint
	Topic    string
}

// Config contains the list of all devices to be tracked.
type Config struct {
	BluetoothDevices []Device `yaml:"bluetooth_devices"`
	IPDevices        []Device `yaml:"ip_devices"`
	MQTTServer       MQTT     `yaml:"mqtt_server"`
}

// DefaultLocation corresponds to the default path to the directory where
// the configuration file is stored.
const DefaultLocation = "/etc/myhome"

const defaultFilename = "config.yaml"

// Retrieve reads and parses the configuration file.
func Retrieve(location string) Config {
	config := Config{}
	filename := filepath.Join(location, defaultFilename)
	content, err := ioutil.ReadFile(filename)
	if err == nil {
		err = yaml.Unmarshal(content, &config)
	}
	if err != nil {
		log.Fatal(err)
	}
	return config
}
