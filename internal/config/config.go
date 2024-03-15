package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/pkg/model"
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
	Devices      map[string]*model.Device `yaml:"-"`
	MQTTServer   MQTT                     `yaml:"mqtt_server"`
	Server       Server                   `yaml:"server"`
	Trackers     map[string]Settings      `yaml:"trackers"`
	cfgLocation  string                   `yaml:"-"`
	dataLocation string                   `yaml:"-"`
}

// DefaultCfgLocation corresponds to the default path to the directory where
// the configuration file is stored.
const DefaultCfgLocation = "/etc/myhome"

const cfgFilename = "config.yaml"

// DefaultDataLocation corresponds to the default path to the directory where
// the data is stored.
const DefaultDataLocation = "/var/lib/myhome"

const devicesFilename = "devices.yaml"

// Retrieve reads and parses the configuration file.
func Retrieve(cfgLocation string, dataLocation string) Config {
	cfg := Config{
		cfgLocation:  cfgLocation,
		dataLocation: dataLocation,
	}
	cfg.loadConfig(cfgLocation, cfgFilename)
	cfg.loadDevicesData(dataLocation, devicesFilename)
	return cfg
}

func (cfg *Config) loadConfig(location string, name string) {
	filename := filepath.Join(location, name)
	log.Debug("Loading config from: ", filename)
	content, err := os.ReadFile(filename)
	if err == nil {
		err = yaml.Unmarshal(content, &cfg)
	}
	if err != nil {
		log.Fatal(err)
	}
	cfg.Devices = make(map[string]*model.Device)
}

func (cfg *Config) loadDevicesData(location string, name string) {
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
	save(devices, cfg.dataLocation, devicesFilename)
}
