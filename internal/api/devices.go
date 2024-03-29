package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/touchardv/myhome-presence/internal/device"
	"github.com/touchardv/myhome-presence/pkg/model"
)

func (c *apiContext) registerDevice(w http.ResponseWriter, r *http.Request) {
	d := model.Device{}
	err := json.NewDecoder(r.Body).Decode(&d)
	if err == nil {
		err = c.registry.AddDevice(d)
		if err == nil {
			w.WriteHeader(http.StatusCreated)
			return
		}
	}

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}

func (c *apiContext) unregisterDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := c.registry.RemoveDevice(vars["id"])
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.NotFound(w, r)
	}
}

func (c *apiContext) executeDeviceAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	q := r.URL.Query()
	err := c.registry.ExecuteDeviceAction(vars["id"], q.Get("action"))
	if err == nil {
		w.WriteHeader(http.StatusAccepted)
	} else {
		http.NotFound(w, r)
	}
}

func (c *apiContext) updateDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	d := model.Device{}
	err := json.NewDecoder(r.Body).Decode(&d)
	if err == nil {
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
		w.Write([]byte(err.Error()))
	}
}

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

func (c *apiContext) queryDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	q := r.URL.Query()
	status := model.StatusOf(q.Get("status"))
	json.NewEncoder(w).Encode(c.registry.GetDevices(status))
}
