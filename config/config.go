package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

// IPInterface represents one IP device interface.
type IPInterface struct {
	IPAddress  string `json:"ip_address" yaml:"ip_address"`
	MACAddress string `json:"mac_address" yaml:"mac_address"`
}

// MQTT contains the MQTT server connection information.
type MQTT struct {
	Enabled  bool
	Hostname string
	Port     uint
	Topic    string
}

// Server contains the local web server configuration.
type Server struct {
	Address      string `yaml:"address"`
	Port         uint   `yaml:"port"`
	SwaggerUIURL string `yaml:"swagger_ui_url"`
}

// Config contains the list of all devices to be tracked.
type Config struct {
	Devices    map[string]*Device `yaml:"devices"`
	MQTTServer MQTT               `yaml:"mqtt_server"`
	Server     Server             `yaml:"server"`
	Trackers   []string           `yaml:"trackers"`
	location   string             `yaml:"-"`
}

// DefaultLocation corresponds to the default path to the directory where
// the configuration file is stored.
const DefaultLocation = "/etc/myhome"

const defaultFilename = "config.yaml"
const devicesFilename = "devices.yaml"

// Retrieve reads and parses the configuration file.
func Retrieve(location string) Config {
	cfg := retrieve(location, defaultFilename)
	cfg.load(location, devicesFilename)
	return cfg
}

func retrieve(location string, name string) Config {
	cfg := Config{location: location}
	filename := filepath.Join(location, name)
	content, err := ioutil.ReadFile(filename)
	if err == nil {
		err = yaml.Unmarshal(content, &cfg)
	}
	if err != nil {
		log.Fatal(err)
	}
	for id, d := range cfg.Devices {
		if len(d.Identifier) > 0 && d.Identifier != id {
			log.Fatal("Invalid configuration, the device identifier does not match: ", d.Identifier)
		}
		d.Identifier = id
		d.Status = Tracked
	}
	return cfg
}

func (cfg *Config) load(location string, name string) {
	devices, err := load(location, name)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range devices {
		if device, ok := cfg.Devices[d.Identifier]; ok {
			device.Present = d.Present
			device.LastSeenAt = d.LastSeenAt
		} else {
			cfg.Devices[d.Identifier] = &d
		}
	}
}

// Save persists the device list to disk.
func (cfg *Config) Save(devices []Device) {
	save(devices, cfg.location, devicesFilename)
}
