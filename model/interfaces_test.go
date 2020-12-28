package model

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

func TestJSONSerialization(t *testing.T) {
	i := InterfaceWifi
	data, _ := json.Marshal(i)
	assert.Equal(t, `"wifi"`, string(data))
}

func TestJSONDeSerialization(t *testing.T) {
	data := `"ethernet"`
	var i InterfaceType
	err := json.Unmarshal([]byte(data), &i)
	assert.Nil(t, err)
	assert.Equal(t, InterfaceEthernet, i)

	data = `"foobar"`
	err = json.Unmarshal([]byte(data), &i)
	assert.Nil(t, err)
	assert.Equal(t, InterfaceUnknown, i)
}

func TestYAMLSerialization(t *testing.T) {
	i := InterfaceBluetoothLowEnergy
	data, _ := yaml.Marshal(i)
	assert.Equal(t, "ble\n", string(data))
}

func TestYAMLDeserialization(t *testing.T) {
	data := "bluetooth"
	var i InterfaceType
	err := yaml.Unmarshal([]byte(data), &i)
	assert.Nil(t, err)
	assert.Equal(t, InterfaceBluetooth, i)

	data = "foobar"
	err = yaml.Unmarshal([]byte(data), &i)
	assert.Nil(t, err)
	assert.Equal(t, InterfaceUnknown, i)
}
