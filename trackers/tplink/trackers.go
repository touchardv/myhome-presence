package tplink

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/device"
	"github.com/touchardv/myhome-presence/model"
)

// EnableTrackers registers the "tplink" trackers.
func EnableTrackers() {
	device.Register("tplink-c2600", newArcherC2600Tracker)
	device.Register("tplink-re450", newRE450Tracker)
}

type loginFunc func(baseUrl string, username string, password string) (credentials, error)

type statusFunc func(baseUrl string, c credentials) (statusResponse, error)

type tplinkTracker struct {
	name     string
	baseURL  string
	username string
	password string
	login    loginFunc
	status   statusFunc
}

type credentials struct {
	Token string
	Nonce string
}

type loginResponse struct {
	Data      map[string]string `json:"data"`
	ErrorCode string            `json:"errorcode"`
	Success   bool              `json:"success"`
}

type statusResponse struct {
	Success   bool       `json:"success"`
	ErrorCode string     `json:"errorcode"`
	Data      statusData `json:"data"`
}

type statusDataDevice struct {
	MACAddress string `json:"macaddr"`
	IPAddress  string `json:"ipaddr"`
	Hostname   string `json:"hostname"`
}

func (t *tplinkTracker) Loop(deviceReport device.ReportPresenceFunc, ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	log.Infof("Starting: %s tracker", t.name)
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Infof("Stopped: %s tracker", t.name)
			return nil

		case <-ticker.C:
			ticker.Reset(5 * time.Minute)
			c, err := t.login(t.baseURL, t.username, t.password)
			if err != nil {
				log.Errorf("[%s] login failed: %s", t.name, err)
				break
			}
			r, err := t.status(t.baseURL, c)
			if err != nil {
				log.Errorf("[%s] status failed: %s", t.name, err)
				break
			}
			log.Debugf("[%s] detected %d wired device(s)", t.name, len(r.Data.WiredDevices))
			for _, device := range r.Data.WiredDevices {
				itf := model.Interface{Type: model.InterfaceEthernet, IPv4Address: device.IPAddress}
				deviceReport(itf, nil)
			}
			log.Debugf("[%s] detected %d wireless device(s)", t.name, len(r.Data.WirelessDevices))
			for _, device := range r.Data.WirelessDevices {
				itf := model.Interface{Type: model.InterfaceWifi, IPv4Address: device.IPAddress}
				deviceReport(itf, nil)
			}
		}
	}
}

func (t *tplinkTracker) Ping(model.Device) {
	// Nothing to be done here. The tracker is purely asynchronous.
}

func ensureSetting(key string, cfg config.Settings, name string) string {
	if v, found := cfg[key]; found {
		return v
	}
	log.Fatalf("[%s] Missing device '%s' configuration setting", name, key)
	return ""
}
