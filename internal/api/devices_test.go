package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/touchardv/myhome-presence/internal/config"
	"github.com/touchardv/myhome-presence/internal/device"
	"github.com/touchardv/myhome-presence/pkg/model"
)

func TestDeviceRegistration(t *testing.T) {
	registry := device.NewRegistry(config.Config{Devices: map[string]*model.Device{}})
	server := NewServer(config.Server{}, registry)

	response := performRequest(server, registerRequest(`{}`))
	assert.Equal(t, http.StatusBadRequest, response.Code)

	response = performRequest(server, registerRequest(`{"identifier": "foo"}`))
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assertEqualBody(t, "missing device status", response)

	response = performRequest(server, registerRequest(`{"identifier": "foo", "status": "bad"}`))
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assertEqualBody(t, "invalid device status", response)

	response = performRequest(server, registerRequest(`{"identifier": "foo", "status": "tracked"}`))
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, 1, len(registry.GetDevices(model.StatusTracked)))

	response = performRequest(server, registerRequest(`{"identifier": "bar", "properties": { "key": "value" }, "status": "ignored"}`))
	assert.Equal(t, http.StatusCreated, response.Code)
	devices := registry.GetDevices(model.StatusIgnored)
	assert.Equal(t, 1, len(devices))
	assert.Equal(t, "value", devices[0].Properties["key"])
}

func TestFindDevice(t *testing.T) {
	devices := make(map[string]*model.Device, 0)
	devices["foo"] = &model.Device{Identifier: "foo"}
	registry := device.NewRegistry(config.Config{Devices: devices})
	server := NewServer(config.Server{}, registry)

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
	devices := make(map[string]*model.Device, 0)

	registry := device.NewRegistry(config.Config{Devices: devices})
	server := NewServer(config.Server{}, registry)

	req, _ := http.NewRequest("GET", "/api/devices", nil)
	response := performRequest(server, req)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "application/json", response.Header().Get("Content-Type"))
	bytes, _ := ioutil.ReadAll(response.Body)
	body := string(bytes)
	assert.Contains(t, body, "[]")
}

func TestUnregisterDevice(t *testing.T) {
	devices := make(map[string]*model.Device, 0)
	devices["foo"] = &model.Device{Identifier: "foo"}
	registry := device.NewRegistry(config.Config{Devices: devices})
	server := NewServer(config.Server{}, registry)

	req, _ := http.NewRequest("DELETE", "/api/devices/bar", nil)
	response := performRequest(server, req)
	assert.Equal(t, http.StatusNotFound, response.Code)

	req, _ = http.NewRequest("DELETE", "/api/devices/foo", nil)
	response = performRequest(server, req)
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Equal(t, 0, len(registry.GetDevices(model.StatusUndefined)))
}

func TestUpdateDevice(t *testing.T) {
	devices := make(map[string]*model.Device, 0)
	devices["foo"] = &model.Device{Identifier: "foo", Description: "old foo", Status: model.StatusIgnored}
	registry := device.NewRegistry(config.Config{Devices: devices})
	server := NewServer(config.Server{}, registry)

	jsonStr := []byte(`{"identifier": "foo", "description": "new foo", "status": "tracked"}`)
	req, _ := http.NewRequest("PUT", "/api/devices/bar", bytes.NewBuffer(jsonStr))
	response := performRequest(server, req)
	assert.Equal(t, http.StatusNotFound, response.Code)

	req, _ = http.NewRequest("PUT", "/api/devices/foo", bytes.NewBuffer(jsonStr))
	response = performRequest(server, req)
	assert.Equal(t, http.StatusOK, response.Code)

	d, err := registry.FindDevice("foo")
	assert.Nil(t, err)
	assert.Equal(t, "new foo", d.Description)
	assert.Equal(t, "tracked", d.Status.String())
}

func performRequest(server *Server, req *http.Request) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	server.router.ServeHTTP(response, req)
	return response
}

func registerRequest(b string) *http.Request {
	req, _ := http.NewRequest("POST", "/api/devices", bytes.NewBuffer([]byte(b)))
	return req
}

func assertEqualBody(t *testing.T, expected string, r *httptest.ResponseRecorder) {
	bytes, _ := ioutil.ReadAll(r.Body)
	actual := string(bytes)
	assert.Equal(t, expected, actual)
}
