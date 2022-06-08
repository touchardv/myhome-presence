package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/touchardv/myhome-presence/device"
	"github.com/touchardv/myhome-presence/model"

	"github.com/gorilla/mux"
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

func (c *apiContext) contactDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := c.registry.ContactDevice(vars["id"])
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

func (c *apiContext) listDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c.registry.GetDevices())
}
