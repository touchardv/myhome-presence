package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/touchardv/myhome-presence/model"
	"gopkg.in/yaml.v2"
)

func load(location string, name string) ([]model.Device, error) {
	filename := filepath.Join(location, name)
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return []model.Device{}, nil
	}
	content, err := ioutil.ReadFile(filename)
	devices := make([]model.Device, 10)
	if err == nil {
		err = yaml.Unmarshal(content, &devices)
	}
	return devices, err
}

func save(devices []model.Device, location string, name string) error {
	bytes, err := yaml.Marshal(devices)
	if err == nil {
		err = ioutil.WriteFile(filepath.Join(location, name), bytes, 0644)
	}
	return err
}
