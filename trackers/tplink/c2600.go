package tplink

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/device"
)

const c2600 = "tplink-c2600"

func newArcherC2600Tracker(cfg config.Settings) device.Tracker {
	return &tplinkTracker{
		name:     c2600,
		baseURL:  ensureSetting("url", cfg, c2600),
		username: ensureSetting("username", cfg, c2600),
		password: ensureSetting("password", cfg, c2600),
		login:    c2600Login,
		status:   c2600Status,
	}
}

const c2600SessionCookie = "sysauth"

func c2600Login(baseUrl string, username string, password string) (credentials, error) {
	data := url.Values{}
	data.Set("operation", "login")
	data.Set("username", username)
	data.Set("password", password)

	url := fmt.Sprintf("%s/cgi-bin/luci/;stok=/login?form=login", baseUrl)
	req, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	cred := credentials{}
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return cred, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return cred, errors.New("unexpected http response " + res.Status)
	}
	response := loginResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return cred, err
	}
	if !response.Success {
		return cred, errors.New(response.ErrorCode)
	}
	for _, c := range res.Cookies() {
		if c.Name == c2600SessionCookie {
			cred.Token = response.Data["stok"]
			cred.Nonce = c.Value
			return cred, nil
		}
	}
	return cred, errors.New("missing session cookie")

}

type statusData struct {
	WiredDevices    []statusDataDevice `json:"access_devices_wired"`
	WirelessDevices []statusDataDevice `json:"access_devices_wireless_host"`
}

func c2600Status(baseUrl string, a credentials) (statusResponse, error) {
	data := url.Values{}
	data.Set("operation", "read")

	url := fmt.Sprintf("%s/cgi-bin/luci/;stok=%s/admin/status?form=all", baseUrl, a.Token)
	req, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.AddCookie(&http.Cookie{Name: c2600SessionCookie, Value: a.Nonce})

	response := statusResponse{}
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return response, err
	}
	if res.StatusCode != http.StatusOK {
		return response, errors.New("unexpected http response " + res.Status)
	}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return response, err
	}
	if !response.Success {
		return response, errors.New(response.ErrorCode)
	}
	return response, err
}
