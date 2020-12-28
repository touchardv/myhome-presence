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
	assert.Equal(t, 0, len(cfg.Devices))
}

func TestLoadingDevicesState(t *testing.T) {
	cwd, _ := os.Getwd()
	cfg := retrieve(cwd, "config.yaml.example")
	cfg.load(cwd, "devices.yaml.example")
	assert.Equal(t, 2, len(cfg.Devices))

	device := cfg.Devices["my-smartwatch"]
	assert.False(t, device.LastSeenAt.IsZero())
	assert.Equal(t, 1, len(device.Interfaces))
	assert.Equal(t, model.InterfaceBluetoothLowEnergy, device.Interfaces[0].Type)
	assert.True(t, device.Present)
	assert.Equal(t, model.StatusIgnored, device.Status)

	device = cfg.Devices["my-ip-camera"]
	assert.False(t, device.LastSeenAt.IsZero())
	assert.Equal(t, 2, len(device.Interfaces))
	assert.Equal(t, model.InterfaceWifi, device.Interfaces[0].Type)
	assert.Equal(t, "10.1.2.3", device.Interfaces[0].IPv4Address)
	assert.Equal(t, model.InterfaceEthernet, device.Interfaces[1].Type)
	assert.Equal(t, "10.2.3.4", device.Interfaces[1].IPv4Address)
	assert.True(t, device.Present)
	assert.Equal(t, model.StatusDiscovered, device.Status)
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
