package bluetooth

import (
	"github.com/bettercap/gatt"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/model"
)

func (t *btTracker) onDeviceStateChanged(d gatt.Device, s gatt.State) {
	log.Debug("Local Bluetooth device state changed: ", s)
	if s == gatt.StatePoweredOn {
		if !t.scanning {
			d.Scan([]gatt.UUID{}, true)
			t.scanning = true
		}
	}
}

func (t *btTracker) onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	log.Debugf("Discovered a Bluetooth device: %s %s %s", p.ID(), p.Name(), a.LocalName)
	t.deviceReport(model.Interface{Type: model.InterfaceBluetoothLowEnergy, MACAddress: p.ID()})
}

func (t *btTracker) startScanning() {
	log.Debug("Start scanning for Bluetooth devices...")
	d, err := gatt.NewDevice(defaultClientOptions...)
	if err != nil {
		log.Fatal("Failed to create the local Bluetooth device: ", err)
		return
	}
	if err := d.Init(t.onDeviceStateChanged); err != nil {
		log.Error("Failed to initialise the local Bluetooth device: ", err)
		return
	}
	t.device = d
	t.device.Handle(gatt.PeripheralDiscovered(t.onPeripheralDiscovered))
}

func (t *btTracker) stopScanning() {
	log.Debug("Stop scanning for Bluetooth devices...")
	if t.scanning {
		t.device.StopScanning()
		t.scanning = false
	}
	err := t.device.Stop()
	if err != nil {
		log.Error("Failed to stop the local Bluetooth device: ", err)
	}
	t.device = nil
}
