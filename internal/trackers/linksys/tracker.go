package linksys

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/internal/config"
	"github.com/touchardv/myhome-presence/internal/device"
	"github.com/touchardv/myhome-presence/pkg/model"
)

// EnableTracker registers the "linksys" tracker so that it can be used.
func EnableTracker() {
	device.Register("linksys", newLinksysTracker)
}

type linksysTracker struct {
	auth                string
	baseURL             string
	lastChangeRevision  int
	syncIntervalMinutes int
}

func newLinksysTracker(cfg config.Settings) device.Tracker {
	syncIntervalMinutes, err := strconv.Atoi(cfg["sync_interval_minutes"])
	if err != nil {
		log.Fatal("Invalid sync_interval_minutes setting value: ", err)
	}
	return &linksysTracker{
		auth:                cfg["auth"],
		baseURL:             cfg["base_url"],
		lastChangeRevision:  noRevision,
		syncIntervalMinutes: syncIntervalMinutes,
	}
}

const noRevision = -1

type jnapDeviceConnection struct {
	IPAddress  string `json:"ipAddress"`
	MACAddress string `json:"macAddress"`
}

type jnapDeviceInterface struct {
	InterfaceType string `json:"interfaceType"`
	MACAddress    string `json:"macAddress"`
}

type jnapDevice3 struct {
	Connections        []jnapDeviceConnection `json:"connections"`
	DeviceID           string                 `json:"deviceID"`
	KnownInterfaces    []jnapDeviceInterface  `json:"knownInterfaces"`
	LastChangeRevision int                    `json:"lastChangeRevision"`
}

type jnapDevices3Response struct {
	Result string `json:"result"`
	Output struct {
		DeletedDeviceIDs []string      `json:"deletedDeviceIDs"`
		Devices          []jnapDevice3 `json:"devices"`
		Revision         int           `json:"revision"`
	} `json:"output"`
}

func (t *linksysTracker) Loop(deviceReport device.ReportPresenceFunc, ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	log.Infof("Starting: linksys tracker")
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			log.Infof("Stopped: linksys tracker")
			return nil

		case <-ticker.C:
			ticker.Reset(time.Duration(t.syncIntervalMinutes) * time.Minute)
			t.fetchAndReportDevices(deviceReport, ctx)
		}
	}
}

func (t *linksysTracker) fetchAndReportDevices(deviceReport device.ReportPresenceFunc, _ context.Context) {
	url := fmt.Sprintf("%s/JNAP/", t.baseURL)
	req, _ := http.NewRequest("POST", url, strings.NewReader("{}"))
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("X-Jnap-Action", "http://linksys.com/jnap/devicelist/GetDevices3")
	req.Header.Add("X-Jnap-Authorization", t.auth)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 5 * time.Second, Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Performing http request failed: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Unexpected http request response status code: ", resp.Status)
		return
	}

	response := jnapDevices3Response{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Error("Decoding http body failed: ", err)
		return
	}

	if response.Result != "OK" {
		log.Error("Unexpected response result: ", response.Result)
		return
	}
	log.Debugf("Reporting %d device(s)", len(response.Output.Devices))
	itfs := []model.DetectedInterface{}
	for i, d := range response.Output.Devices {
		log.Tracef("response.Output.Devices[%d]=%+v", i, response.Output.Devices[i])
		if len(d.Connections) == 0 {
			continue
		}
		if len(d.Connections) > 1 {
			log.Warnf("Found %d connections for %s: ", len(d.Connections), d.DeviceID)
		}
		itfs = append(itfs, model.DetectedInterface{Interface: toInterface(d.Connections[0], d.KnownInterfaces)})
	}

	if len(itfs) > 0 {
		deviceReport(itfs)
	}
	t.lastChangeRevision = response.Output.Revision
}

func toInterface(conn jnapDeviceConnection, itfs []jnapDeviceInterface) model.Interface {
	out := model.Interface{
		Type:        toInterfaceType(conn.MACAddress, itfs),
		IPv4Address: conn.IPAddress,
		MACAddress:  conn.MACAddress,
	}
	return out
}

func toInterfaceType(macAddress string, itfs []jnapDeviceInterface) model.InterfaceType {
	for _, itf := range itfs {
		if itf.MACAddress == macAddress {
			switch itf.InterfaceType {
			case "Wired":
				return model.InterfaceEthernet
			case "Wireless":
				return model.InterfaceWifi
			default:
				return model.InterfaceUnknown
			}
		}
	}
	return model.InterfaceUnknown
}

func (t *linksysTracker) Ping([]model.Device) {
	// Nothing to be done here. The tracker is purely asynchronous.
}
