package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

// IPInterface represents one IP device interface.
type IPInterface struct {
	IPAddress  string `yaml:"ip_address"`
	MACAddress string `yaml:"mac_address"`
}

// Device represents a single device being tracked.
type Device struct {
	Description  string
	Identifier   string
	BLEAddress   string                 `yaml:"ble_address"`
	IPInterfaces map[string]IPInterface `yaml:"ip_interfaces"`
}

// MQTT contains the MQTT server connection information.
type MQTT struct {
	Hostname string
	Port     uint
	Topic    string
}

// Config contains the list of all devices to be tracked.
type Config struct {
	Devices    []Device `yaml:"devices"`
	MQTTServer MQTT     `yaml:"mqtt_server"`
}

// DefaultLocation corresponds to the default path to the directory where
// the configuration file is stored.
const DefaultLocation = "/etc/myhome"

const defaultFilename = "config.yaml"

// Retrieve reads and parses the configuration file.
func Retrieve(location string) Config {
	return retrieve(location, defaultFilename)
}

func retrieve(location string, name string) Config {
	config := Config{}
	filename := filepath.Join(location, name)
	content, err := ioutil.ReadFile(filename)
	if err == nil {
		err = yaml.Unmarshal(content, &config)
	}
	if err != nil {
		log.Fatal(err)
	}
	return config
}
