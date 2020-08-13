package api

import (
	"bytes"
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
	c := apiContext{registry}
	handler := http.HandlerFunc(c.registerDevice)

	var jsonStr = []byte(`{}`)
	req, _ := http.NewRequest("POST", "/api/devices", bytes.NewBuffer(jsonStr))
	response := performRequest(handler, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	jsonStr = []byte(`{"identifier": "foo"}`)
	req, _ = http.NewRequest("POST", "/api/devices", bytes.NewBuffer(jsonStr))
	response = performRequest(handler, req)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, 1, len(devices))
}

func performRequest(handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}
