package bluetooth

import (
	"context"

	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	linux_device "github.com/muka/go-bluetooth/bluez/profile/device"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/device"
	"github.com/touchardv/myhome-presence/model"
)

type btLinuxManager struct {
	a      *adapter.Adapter1
	cancel func()
}

func newBtManager() btManager {
	a, err := adapter.GetDefaultAdapter()
	if err != nil {
		log.Error(err)
		return nil
	}
	return &btLinuxManager{
		a: a,
	}
}

func (mgr *btLinuxManager) scan(report device.ReportPresenceFunc, ctx context.Context) error {
	mgr.a.FlushDevices()

	discovery, cancel, err := api.Discover(mgr.a, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	mgr.cancel = cancel

	go func() {
		for ev := range discovery {
			if ev.Type == adapter.DeviceRemoved {
				continue
			}

			dev, err := linux_device.NewDevice1(ev.Path)
			if err != nil {
				log.Errorf("%s: %s", ev.Path, err)
				continue
			}

			if dev == nil {
				log.Errorf("%s: not found", ev.Path)
				continue
			}

			log.Debug("Address: ", dev.Properties.Address, " AddressType: ", dev.Properties.AddressType, " Name: ", dev.Properties.Name, " Alias: ", dev.Properties.Alias)
			for _, u := range dev.Properties.UUIDs {
				log.Debug("UUIDs: ", u)
			}
			for uuid, d := range dev.Properties.ServiceData {
				log.Debug("ServiceData: ", uuid, " -> ", d)
			}
			report(model.Interface{
				Type:       model.InterfaceBluetoothLowEnergy,
				MACAddress: dev.Properties.Address,
			})
		}
	}()

	return nil
}

func (mgr *btLinuxManager) stopScan() {
	mgr.cancel()
	api.Exit()
}
