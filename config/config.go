package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/model"
)

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

type Settings map[string]string

// Config contains the list of all devices to be tracked.
type Config struct {
	Devices    map[string]*model.Device `yaml:"-"`
	MQTTServer MQTT                     `yaml:"mqtt_server"`
	Server     Server                   `yaml:"server"`
	Trackers   map[string]Settings      `yaml:"trackers"`
	location   string                   `yaml:"-"`
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
	cfg.Devices = make(map[string]*model.Device)
	return cfg
}

func (cfg *Config) load(location string, name string) {
	devices, err := load(location, name)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range devices {
		if _, ok := cfg.Devices[d.Identifier]; !ok {
			cfg.Devices[d.Identifier] = &model.Device{}
		}
		*cfg.Devices[d.Identifier] = d
	}
}

// Save persists the device list to disk.
func (cfg *Config) Save(devices []model.Device) {
	save(devices, cfg.location, devicesFilename)
}
