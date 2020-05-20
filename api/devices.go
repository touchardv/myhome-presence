package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/touchardv/myhome-presence/config"
)

// swagger:parameters findDevice
type deviceID struct {
	// The ID of the device
	//
	// in: path
	// required: true
	ID string `json:"id"`
}

// swagger:route GET /devices/{id} devices findDevice
// Find a device given its identifier.
// responses:
//   200: Device
func (c *apiContext) findDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	d, found := c.registry.FindDevice(vars["id"])
	if found {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(d)
	} else {
		http.NotFound(w, r)
	}
}

// swagger:response deviceArray
type deviceArray []config.Device

// swagger:route GET /devices devices listDevices
// List all known device(s).
// responses:
//   200: deviceArray
func (c *apiContext) listDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c.registry.GetDevices())
}
