package device

import (
	"math/rand"
	"time"

	"github.com/bettercap/gatt"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
)

type bleTracker struct {
	device   gatt.Device
	devices  map[string]config.Device
	scanning bool
}

func newBLETracker() Tracker {
	d, err := gatt.NewDevice(defaultClientOptions...)
	if err != nil {
		log.Fatal("Failed to create BLE device: ", err)
	}
	return &bleTracker{
		device:   d,
		devices:  make(map[string]config.Device, 10),
		scanning: false,
	}
}

func (t *bleTracker) init(devices []config.Device, f func(p gatt.Peripheral, a *gatt.Advertisement, rssi int)) error {
	for _, device := range devices {
		if len(device.BLEAddress) == 0 {
			continue
		}
		t.devices[device.BLEAddress] = device
	}
	t.device.Handle(gatt.PeripheralDiscovered(f))
	return t.device.Init(t.onDeviceStateChanged)
}

func (t *bleTracker) onDeviceStateChanged(d gatt.Device, s gatt.State) {
	log.Debug("BLE device state:", s)
	if s == gatt.StatePoweredOn {
		t.startScanning()
	}
}

func (t *bleTracker) startScanning() {
	log.Debug("Start BLE device scanning...")
	t.device.Scan([]gatt.UUID{}, false)
	t.scanning = true
}

func (t *bleTracker) stopScanning() {
	log.Debug("Stop BLE device scanning...")
	t.device.StopScanning()
	t.scanning = false
}

func (t *bleTracker) Track(devices []config.Device, presence chan string, stopping chan struct{}) {
	log.Info("Starting: ble tracker")
	err := t.init(devices, func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
		log.Debugf("Discovered %s %s %s", p.ID(), p.Name(), a.LocalName)
		d, ok := t.devices[p.ID()]
		if ok {
			presence <- d.Identifier
		}
	})
	if err != nil {
		log.Error("Failed to init BLE device: ", err)
	}
	timer := time.NewTimer(30 * time.Second)
	for {
		select {
		case <-timer.C:
			if t.scanning {
				t.stopScanning()
			} else {
				t.startScanning()
			}
			timer.Reset(randomDuration(t.scanning))
		case <-stopping:
			timer.Stop()
			if t.scanning {
				t.stopScanning()
			}
			log.Info("Stopped: ble tracker")
			return
		}
	}
}

const minScanDuration = 20
const minDurationBetweenScans = 15

func randomDuration(scanning bool) time.Duration {
	var n int
	if scanning {
		n = minScanDuration
		n += rand.Intn(10)
	} else {
		n = minDurationBetweenScans
		n += rand.Intn(30)
	}
	return time.Duration(n) * time.Second
}
