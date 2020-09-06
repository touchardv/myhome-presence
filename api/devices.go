package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/touchardv/myhome-presence/device"

	"github.com/gorilla/mux"
	"github.com/touchardv/myhome-presence/config"
)

// swagger:parameters registerDevice updateDevice
type deviceBodyParameter struct {
	// A device
	//
	// in: body
	// required: true
	Device deviceParameter
}

type deviceParameter struct {
	// example: My phone
	Description string `json:"description"`
	// example: my-phone
	// required: true
	Identifier string `json:"identifier"`
	// example: AA:BB:CC:DD:EE
	BLEAddress string `json:"ble_address"`
	// example: AA:BB:CC:DD:EE
	BTAddress string `json:"bt_address"`
	// example: { "wifi": { "ip_address": "10.10.10.124", "mac_address": "AB:CD:EF:01:02:03" } }
	IPInterfaces map[string]ipInterfaceParameter `json:"ip_interfaces"`
}

type ipInterfaceParameter struct {
	// required: true
	IPAddress string `json:"ip_address"`
	// required: true
	MACAddress string `json:"mac_address"`
}

// swagger:route POST /devices devices registerDevice
//
// Register a new device.
//
// responses:
//   201: description: Success
//   400: description: Invalid parameters
func (c *apiContext) registerDevice(w http.ResponseWriter, r *http.Request) {
	param := deviceParameter{}
	err := json.NewDecoder(r.Body).Decode(&param)
	if err == nil {
		d := convert(param)
		err = c.registry.AddDevice(d)
		if err == nil {
			w.WriteHeader(http.StatusCreated)
			return
		}
	}

	w.WriteHeader(http.StatusBadRequest)
}

func convert(p deviceParameter) config.Device {
	d := config.Device{
		Description:  p.Description,
		Identifier:   p.Identifier,
		BLEAddress:   p.BLEAddress,
		BTAddress:    p.BTAddress,
		IPInterfaces: make(map[string]config.IPInterface, len(p.IPInterfaces)),
	}
	for name, itf := range p.IPInterfaces {
		d.IPInterfaces[name] = config.IPInterface{
			IPAddress:  itf.IPAddress,
			MACAddress: itf.MACAddress,
		}
	}
	return d
}

// swagger:parameters findDevice unregisterDevice updateDevice
type deviceID struct {
	// The ID of the device
	//
	// in: path
	// required: true
	ID string `json:"id"`
}

// swagger:route DELETE /devices/{id} devices unregisterDevice
//
// Unregister a device given its identifier.
//
// responses:
// 	 204: description: Success
//   404: description: Not found
func (c *apiContext) unregisterDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := c.registry.RemoveDevice(vars["id"])
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.NotFound(w, r)
	}
}

// swagger:route PUT /devices/{id} devices updateDevice
//
// Update a device given its identifier.
//
// responses:
//   200: Device
//   400: description: Invalid parameters
//   404: description: Not found
func (c *apiContext) updateDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	param := deviceParameter{}
	err := json.NewDecoder(r.Body).Decode(&param)
	if err == nil {
		d := convert(param)
		d, err = c.registry.UpdateDevice(vars["id"], d)
		if err == nil {
			w.Header().Add("Content-Type", "application/json")
			json.NewEncoder(w).Encode(d)
			return
		}
	}
	if errors.Is(err, device.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// swagger:route GET /devices/{id} devices findDevice
//
// Find a device given its identifier.
//
// responses:
//   200: Device
//   404: description: Not found
func (c *apiContext) findDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	d, err := c.registry.FindDevice(vars["id"])
	if err == nil {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d)
	} else {
		http.NotFound(w, r)
	}
}

// swagger:response deviceArray
type deviceArray []config.Device

// swagger:route GET /devices devices listDevices
//
// List all known device(s).
//
// responses:
//   200: deviceArray
func (c *apiContext) listDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c.registry.GetDevices())
}
