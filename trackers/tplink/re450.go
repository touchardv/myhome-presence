package tplink

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/device"
)

const re450 = "tplink-re450"

func newRE450Tracker(cfg config.Settings) device.Tracker {
	return &tplinkTracker{
		name:     re450,
		baseURL:  ensureSetting("url", cfg, re450),
		password: ensureSetting("password", cfg, re450),
		login:    re450Login,
		status:   re450Status,
	}
}

const re450SessionCookie = "COOKIE"

func re450Login(baseUrl string, username string, password string) (credentials, error) {
	cred := credentials{}

	req, _ := http.NewRequest("GET", baseUrl, nil)
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return cred, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return cred, errors.New("unexpected http response " + res.Status)
	}
	for _, c := range res.Cookies() {
		if c.Name == re450SessionCookie {
			cred.Nonce = c.Value
			break
		}
	}
	if len(cred.Nonce) == 0 {
		return cred, errors.New("missing session cookie")
	}

	v := url.Values{}
	v.Set("operation", "login")
	v.Set("encoded", re450Token(password, cred.Nonce))
	v.Set("nonce", cred.Nonce)
	data := v.Encode()

	url := fmt.Sprintf("%s/data/login.json", baseUrl)
	req, _ = http.NewRequest("POST", url, strings.NewReader(data))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Referer", baseUrl)
	req.AddCookie(&http.Cookie{Name: re450SessionCookie, Value: cred.Nonce})

	res, err = client.Do(req)
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
		if c.Name == re450SessionCookie {
			cred.Nonce = c.Value
			break
		}
	}
	return cred, nil
}

func re450Token(password string, nonce string) string {
	return strings.ToUpper(md5Func(strings.ToUpper(password) + ":" + nonce))
}

func md5Func(in string) string {
	h := md5.New()
	io.WriteString(h, in)
	return fmt.Sprintf("%x", h.Sum(nil))
}

type re450StatusResponse struct {
	Data      []re450Device `json:"data"`
	ErrorCode string        `json:"errorcode"`
	Success   bool          `json:"success"`
	Timeout   bool          `json:"timeout"`
}

type re450Device struct {
	MACAddress string `json:"mac"`
	IPAddress  string `json:"ipaddr"`
	Hostname   string `json:"name"`
}

func re450Status(baseUrl string, a credentials) (statusResponse, error) {
	status := statusResponse{}

	v := url.Values{}
	v.Set("operation", "read")
	data := v.Encode()

	url := fmt.Sprintf("%s/data/device.all.json", baseUrl)
	req, _ := http.NewRequest("POST", url, strings.NewReader(data))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Referer", baseUrl)
	req.AddCookie(&http.Cookie{Name: re450SessionCookie, Value: a.Nonce})

	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return status, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return status, errors.New("unexpected http response " + res.Status)
	}
	r := re450StatusResponse{}
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		return status, err
	}
	if !r.Success {
		return status, errors.New(r.ErrorCode)
	}
	status.Data = statusData{WirelessDevices: make([]statusDataDevice, len(r.Data))}
	for i, device := range r.Data {
		status.Data.WirelessDevices[i] = statusDataDevice(device)
	}
	status.ErrorCode = r.ErrorCode
	status.Success = r.Success
	return status, nil
}
