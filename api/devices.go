package api

import (
	"encoding/json"
	"net/http"

	"github.com/touchardv/myhome-presence/config"
)

// swagger:response deviceArray
type deviceArray []config.Device

// swagger:route GET /devices devices listDevices
// responses:
//   200: deviceArray
func (c *apiContext) listDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c.registry.GetDevices())
}
