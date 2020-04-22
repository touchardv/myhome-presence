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
}
