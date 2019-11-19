package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

// Device represents a single device being tracked.
type Device struct {
	Description string
	Address     string
	Identifier  string
}

// Config contains the list of all devices to be tracked.
type Config struct {
	BluetoothDevices []Device `yaml:"bluetooth_devices"`
	IPDevices        []Device `yaml:"ip_devices"`
}

const defaultFilename = "config.yaml"

// Retrieve reads and parses the configuration file.
func Retrieve() Config {
	config := Config{}
	content, err := ioutil.ReadFile(defaultFilename)
	if err == nil {
		err = yaml.Unmarshal(content, &config)
	}
	if err != nil {
		log.Error(err)
	}
	return config
}
