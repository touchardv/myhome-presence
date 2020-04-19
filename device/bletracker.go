package device

import (
	"github.com/bettercap/gatt"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
)

type bleTracker struct {
	device   gatt.Device
	scanning bool
}

func newBLETracker() Tracker {
	d, err := gatt.NewDevice(defaultClientOptions...)
	if err != nil {
		log.Fatal("Failed to create BLE device: ", err)
	}
	return &bleTracker{device: d, scanning: false}
}

func (t *bleTracker) onDeviceStateChanged(d gatt.Device, s gatt.State) {
	log.Debug("BLE device state:", s)
	if s == gatt.StatePoweredOn {
		d.Scan([]gatt.UUID{}, false)
		t.scanning = true
	}
}
func (t *bleTracker) onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	log.Debugf("Discovered %s %s %s", p.ID(), p.Name(), a.LocalName)
}

func (t *bleTracker) Track(devices []config.Device, presence chan string, stopping chan struct{}) {
	log.Info("Starting: ble tracker")
	t.device.Handle(gatt.PeripheralDiscovered(t.onPeripheralDiscovered))
	err := t.device.Init(t.onDeviceStateChanged)
	if err != nil {
		log.Error("Failed to init BLE device: ", err)
	}
	for {
		select {
		case <-stopping:
			if t.scanning {
				t.device.StopScanning()
			}
			log.Info("Stopped: ble tracker")
			return
		}
	}
}
