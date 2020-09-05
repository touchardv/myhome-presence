package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/device"
)

func TestDeviceRegistration(t *testing.T) {
	devices := make(map[string]*config.Device, 0)
	registry := device.NewRegistry(config.Config{Devices: devices})
	server := NewServer(registry)

	var jsonStr = []byte(`{}`)
	req, _ := http.NewRequest("POST", "/api/devices", bytes.NewBuffer(jsonStr))
	response := performRequest(server, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	jsonStr = []byte(`{"identifier": "foo"}`)
	req, _ = http.NewRequest("POST", "/api/devices", bytes.NewBuffer(jsonStr))
	response = performRequest(server, req)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, 1, len(devices))
}

func TestFindDevice(t *testing.T) {
	devices := make(map[string]*config.Device, 0)
	devices["foo"] = &config.Device{Identifier: "foo"}
	registry := device.NewRegistry(config.Config{Devices: devices})
	server := NewServer(registry)

	req, _ := http.NewRequest("GET", "/api/devices/bar", nil)
	response := performRequest(server, req)
	assert.Equal(t, http.StatusNotFound, response.Code)

	req, _ = http.NewRequest("GET", "/api/devices/foo", nil)
	response = performRequest(server, req)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "application/json", response.Header().Get("Content-Type"))
	bytes, _ := ioutil.ReadAll(response.Body)
	body := string(bytes)
	assert.Contains(t, body, "\"identifier\":\"foo\"")
}

func TestListDevices(t *testing.T) {
	devices := make(map[string]*config.Device, 0)

	registry := device.NewRegistry(config.Config{Devices: devices})
	server := NewServer(registry)

	req, _ := http.NewRequest("GET", "/api/devices", nil)
	response := performRequest(server, req)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "application/json", response.Header().Get("Content-Type"))
	bytes, _ := ioutil.ReadAll(response.Body)
	body := string(bytes)
	assert.Contains(t, body, "[]")
}

func TestUnregisterDevice(t *testing.T) {
	devices := make(map[string]*config.Device, 0)
	devices["foo"] = &config.Device{Identifier: "foo"}
	registry := device.NewRegistry(config.Config{Devices: devices})
	server := NewServer(registry)

	req, _ := http.NewRequest("DELETE", "/api/devices/bar", nil)
	response := performRequest(server, req)
	assert.Equal(t, http.StatusNotFound, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/devices/foo", nil)
	response = performRequest(server, req)
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Equal(t, 0, len(devices))
}

func TestUpdateDevice(t *testing.T) {
	devices := make(map[string]*config.Device, 0)
	devices["foo"] = &config.Device{Identifier: "foo", Description: "old foo"}
	registry := device.NewRegistry(config.Config{Devices: devices})
	server := NewServer(registry)

	jsonStr := []byte(`{"identifier": "foo", "description": "new foo"}`)
	req, _ := http.NewRequest("PUT", "/api/devices/bar", bytes.NewBuffer(jsonStr))
	response := performRequest(server, req)
	assert.Equal(t, http.StatusNotFound, response.Code)

	req, _ = http.NewRequest("PUT", "/api/devices/foo", bytes.NewBuffer(jsonStr))
	response = performRequest(server, req)
	assert.Equal(t, http.StatusOK, response.Code)

	d, err := registry.FindDevice("foo")
	assert.Nil(t, err)
	assert.Equal(t, "new foo", d.Description)
}

func performRequest(server *Server, req *http.Request) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	server.router.ServeHTTP(response, req)
	return response
}
