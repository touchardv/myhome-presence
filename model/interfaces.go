package model

import (
	"bytes"
	"encoding/json"
)

// InterfaceType defines the type of physical/software interface
type InterfaceType uint

// Interface defines a physical/software interface that can be uniquely addressed
type Interface struct {
	Type        InterfaceType
	MACAddress  string
	IPv4Address string
}

const (
	// InterfaceUnknown corresponds to an unsupported/unknown interface
	InterfaceUnknown InterfaceType = iota

	// InterfaceEthernet corresponds to an Ethernet interface
	InterfaceEthernet

	// InterfaceWifi corresponds to a WiFi interface
	InterfaceWifi

	// InterfaceBluetooth corresponds to a Bluetooth interface
	InterfaceBluetooth

	// InterfaceBluetoothLowEnergy corresponds to a Bluetooth Low Energy interface
	InterfaceBluetoothLowEnergy
)

var interfaceTypeToString = map[InterfaceType]string{
	InterfaceUnknown:            "unknown",
	InterfaceEthernet:           "ethernet",
	InterfaceWifi:               "wifi",
	InterfaceBluetooth:          "bluetooth",
	InterfaceBluetoothLowEnergy: "ble",
}

var stringToInterfaceType = map[string]InterfaceType{
	"unknown":   InterfaceUnknown,
	"ethernet":  InterfaceEthernet,
	"wifi":      InterfaceWifi,
	"bluetooth": InterfaceBluetooth,
	"ble":       InterfaceBluetoothLowEnergy,
}

func (i InterfaceType) String() string {
	return interfaceTypeToString[i]
}

// MarshalJSON marshals the enum as a quoted json string
func (i InterfaceType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(interfaceTypeToString[i])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshals a quoted json string to the enum value
func (i *InterfaceType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	if t, ok := stringToInterfaceType[j]; ok {
		*i = t
	} else {
		*i = InterfaceUnknown
	}
	return nil
}

// MarshalYAML marshals the enum as yaml string
func (i InterfaceType) MarshalYAML() (interface{}, error) {
	return interfaceTypeToString[i], nil
}

// UnmarshalYAML unmarshals a yaml string to the enum value
func (i *InterfaceType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var t string
	if err := unmarshal(&t); err != nil {
		return err
	}
	if t, ok := stringToInterfaceType[t]; ok {
		*i = t
	} else {
		*i = InterfaceUnknown
	}
	return nil
}
