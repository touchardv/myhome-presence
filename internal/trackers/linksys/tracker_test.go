package linksys

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/touchardv/myhome-presence/internal/config"
	"github.com/touchardv/myhome-presence/pkg/model"
)

func TestToJSON(t *testing.T) {
	assert.Equal(t, `{}`, toJSON(-1))
	assert.Equal(t, `{"sinceRevision": 123}`, toJSON(123))
}

func TestNew(t *testing.T) {
	cfg := config.Settings{
		"auth":                  "XZY",
		"base_url":              "http://foo",
		"sync_interval_minutes": "60",
	}
	tracker := newLinksysTracker(cfg)
	linksysTracker := tracker.(*linksysTracker)
	assert.Equal(t, "XZY", linksysTracker.auth)
	assert.Equal(t, "http://foo", linksysTracker.baseURL)
	assert.Equal(t, noRevision, linksysTracker.lastChangeRevision)
	assert.Equal(t, 60, linksysTracker.syncIntervalMinutes)
}

func TestLoop(t *testing.T) {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	tracker := linksysTracker{}

	go tracker.Loop(nil, ctx, wg)

	cancel()
	wg.Wait()
}

func TestInvalidHTTPResponse(t *testing.T) {
	for _, json := range []string{"", "bar", "{}"} {
		server := mockHTTPServerReturningResponse(t, json)
		defer server.Close()

		m := new(reportMock)
		tracker := linksysTracker{
			baseURL: server.URL,
		}

		tracker.fetchAndReportDevices(m.report, nil)

		m.AssertNotCalled(t, "report")
	}
}

func TestValidHTTPResponse(t *testing.T) {
	json := `{
	  "output": {
		"deletedDeviceIDs": [],
		"devices": [],
		"revision": 1234
	  },
	  "result": "OK"
	}`
	server := mockHTTPServerReturningResponse(t, json)
	defer server.Close()

	m := new(reportMock)
	tracker := linksysTracker{
		baseURL: server.URL,
	}

	tracker.fetchAndReportDevices(m.report, nil)

	m.AssertNotCalled(t, "report")
	assert.Equal(t, 1234, tracker.lastChangeRevision)
}

func TestValidHTTPResponseWithDevice(t *testing.T) {
	json := `{
		"output": {
			"deletedDeviceIDs": [],
			"devices": [
				{
					"connections": [
						{
							"ipAddress": "192.168.1.2",
							"macAddress": "AB:CD:EF:12:34:56",
							"parentDeviceID": "c24b3766-1355-4501-b670-ecc0a9411603"
						}
					],
					"deviceID": "b161ba98-ae85-4f76-b567-140df5bfafc8",
					"friendlyName": "My Device",
					"isAuthority": false,
					"knownInterfaces": [
						{
							"interfaceType": "Wired",
							"macAddress": "AB:CD:EF:12:34:56"
						}
					],
					"lastChangeRevision": 8320,
					"maxAllowedProperties": 16,
					"model": {
						"deviceType": ""
					},
					"properties": [
						{
							"name": "userDeviceName",
							"value": "mydevice"
						},
						{
							"name": "userDeviceType",
							"value": "digital-media-player"
						}
					],
					"unit": {}
				}
			],
			"revision": 1234
		},
		"result": "OK"
	}`
	server := mockHTTPServerReturningResponse(t, json)
	defer server.Close()

	var noProps map[string]string
	m := new(reportMock)
	m.On("report", model.Interface{
		Type:        model.InterfaceEthernet,
		IPv4Address: "192.168.1.2",
		MACAddress:  "AB:CD:EF:12:34:56",
	}, noProps)
	tracker := linksysTracker{
		baseURL: server.URL,
	}

	tracker.fetchAndReportDevices(m.report, nil)

	m.AssertNotCalled(t, "report")
	assert.Equal(t, 1234, tracker.lastChangeRevision)
}

func mockHTTPServerReturningResponse(t *testing.T, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.Method, "POST")
		assert.Equal(t, req.URL.String(), "/JNAP/")

		rw.Write([]byte(body))
	}))
}

type reportMock struct {
	mock.Mock
}

func (m *reportMock) report(itf model.Interface, props map[string]string) {
	m.Called(itf, props)
}
