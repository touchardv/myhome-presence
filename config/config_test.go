package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetrieving(t *testing.T) {
	cwd, _ := os.Getwd()
	cfg := retrieve(cwd, "config.yaml.example")
	assert.Equal(t, 3, len(cfg.Devices))

	device := cfg.Devices[0]
	assert.Equal(t, "My laptop", device.Description)
	assert.Equal(t, 2, len(device.IPInterfaces))

	device = cfg.Devices[1]
	assert.Equal(t, "My smartwatch", device.Description)
	assert.Equal(t, "9d329f8ba3c24ae0a494b195dda27d41", device.BLEAddress)

	device = cfg.Devices[2]
	assert.Equal(t, "My phone", device.Description)
	assert.Equal(t, "AA:BB:CC:DD:EE", device.BTAddress)

	assert.Equal(t, 2, len(cfg.Trackers))
	assert.Equal(t, "ipv4", cfg.Trackers[0])
	assert.Equal(t, "bluetooth", cfg.Trackers[1])
}
