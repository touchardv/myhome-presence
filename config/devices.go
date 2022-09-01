package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/model"
	"gopkg.in/yaml.v2"
)

func load(location string, name string) ([]model.Device, error) {
	filename := filepath.Join(location, name)
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return []model.Device{}, nil
	}
	log.Debug("Loading devices from: ", filename)
	content, err := ioutil.ReadFile(filename)
	devices := make([]model.Device, 10)
	if err == nil {
		err = yaml.Unmarshal(content, &devices)
	}
	return devices, err
}

func save(devices []model.Device, location string, name string) error {
	trackedDevices := make([]model.Device, 0, len(devices))
	for _, d := range devices {
		if d.Status == model.StatusTracked ||
			d.Status == model.StatusIgnored {
			trackedDevices = append(trackedDevices, d)
		}
	}
	bytes, err := yaml.Marshal(trackedDevices)
	if err == nil {
		filename := filepath.Join(location, name)
		log.Debug("Saving devices to: ", filename)
		err = ioutil.WriteFile(filename, bytes, 0644)
	}
	return err
}
