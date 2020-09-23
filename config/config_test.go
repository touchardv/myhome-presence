package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/touchardv/myhome-presence/model"
)

func TestRetrieving(t *testing.T) {
	cwd, _ := os.Getwd()
	cfg := retrieve(cwd, "config.yaml.example")
	assert.Equal(t, 3, len(cfg.Devices))

	device := cfg.Devices["my-laptop"]
	assert.Equal(t, "My laptop", device.Description)
	assert.Equal(t, 2, len(device.IPInterfaces))
	assert.Equal(t, model.StatusTracked, device.Status)

	device = cfg.Devices["my-smartwatch"]
	assert.Equal(t, "My smartwatch", device.Description)
	assert.Equal(t, "9d329f8ba3c24ae0a494b195dda27d41", device.BLEAddress)
	assert.Equal(t, model.StatusTracked, device.Status)

	device = cfg.Devices["my-phone"]
	assert.Equal(t, "My phone", device.Description)
	assert.Equal(t, "AA:BB:CC:DD:EE", device.BTAddress)
	assert.Equal(t, model.StatusTracked, device.Status)

	assert.Equal(t, 2, len(cfg.Trackers))
	assert.Equal(t, "ipv4", cfg.Trackers[0])
	assert.Equal(t, "bluetooth", cfg.Trackers[1])
}

func TestLoadingDevicesState(t *testing.T) {
	cwd, _ := os.Getwd()
	cfg := retrieve(cwd, "config.yaml.example")
	cfg.load(cwd, "devices.yaml")
	assert.Equal(t, 3, len(cfg.Devices))
	device := cfg.Devices["my-smartwatch"]
	assert.True(t, device.LastSeenAt.IsZero())
	assert.False(t, device.Present)

	cfg.load(cwd, "devices.yaml.example")
	assert.Equal(t, 4, len(cfg.Devices))

	device = cfg.Devices["my-smartwatch"]
	assert.False(t, device.LastSeenAt.IsZero())
	assert.True(t, device.Present)
	assert.Equal(t, model.StatusTracked, device.Status)

	device = cfg.Devices["my-ip-camera"]
	assert.False(t, device.LastSeenAt.IsZero())
	assert.True(t, device.Present)
	assert.Equal(t, model.StatusUndefined, device.Status)
}

func TestSavingDevicesState(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "config_test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	devices := []model.Device{model.Device{Identifier: "foobar", Present: true}}
	err = save(devices, tempDir, "test-devices.yaml")
	assert.Nil(t, err)
	assert.FileExists(t, filepath.Join(tempDir, "test-devices.yaml"))
}
