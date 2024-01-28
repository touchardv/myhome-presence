package bluetooth

import (
	"context"
	"fmt"
	"strings"

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

			props := make(map[string]string)
			props[device.ReportDataSuggestedIdentifier] = dev.Properties.Address
			props["AddressType"] = dev.Properties.AddressType

			v := strings.TrimSpace(dev.Properties.Name)
			if len(v) > 0 {
				props[device.ReportDataSuggestedDescription] = v
			} else {
				v = strings.TrimSpace(dev.Properties.Alias)
				if len(v) > 0 {
					props[device.ReportDataSuggestedDescription] = v
				}
			}

			for i, u := range dev.Properties.UUIDs {
				props[fmt.Sprintf("ServiceUUID-%d", i)] = u
			}
			for uuid, d := range dev.Properties.ServiceData {
				props[fmt.Sprintf("ServiceData-%s", uuid)] = fmt.Sprint(d)
			}
			report(model.Interface{
				Type:       model.InterfaceBluetooth,
				MACAddress: dev.Properties.Address,
			}, props)
		}
	}()

	return nil
}

func (mgr *btLinuxManager) stopScan() {
	mgr.cancel()
	api.Exit()
}
