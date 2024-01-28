package bluetooth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JuulLabs-OSS/cbgo"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/internal/device"
	"github.com/touchardv/myhome-presence/pkg/model"
)

type btDarwinManager struct {
	cm     cbgo.CentralManager
	report device.ReportPresenceFunc
}

func newBtManager() btManager {
	mgr := &btDarwinManager{
		cm: cbgo.NewCentralManager(nil),
	}

	mgr.cm.SetDelegate(mgr)
	return mgr
}

func (mgr *btDarwinManager) scan(report device.ReportPresenceFunc, ctx context.Context) error {
	log.Debug("Start scanning for Bluetooth devices...")
	for {
		state := mgr.cm.State()
		if state == cbgo.ManagerStatePoweredOn {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.NewTimer(time.Second).C:
		}
	}

	mgr.report = report
	mgr.cm.Scan(nil, &cbgo.CentralManagerScanOpts{
		AllowDuplicates: false,
	})

	return nil
}

func (mgr *btDarwinManager) DidConnectPeripheral(cm cbgo.CentralManager, p cbgo.Peripheral) {
	log.Debug("DidConnectPeripheral:", p.Name(), " - ", p.Identifier())
}

func (mgr *btDarwinManager) DidFailToConnectPeripheral(cm cbgo.CentralManager, p cbgo.Peripheral, err error) {
	log.Debug("DidFailToConnectPeripheral")
}

func (mgr *btDarwinManager) DidDisconnectPeripheral(cm cbgo.CentralManager, p cbgo.Peripheral, err error) {
	log.Debug("DidDisconnectPeripheral: ", p.Name(), " - ", p.Identifier())
}

func (mgr *btDarwinManager) CentralManagerDidUpdateState(cm cbgo.CentralManager) {
	log.Debug("CentralManagerDidUpdateState")
}

func (mgr *btDarwinManager) CentralManagerWillRestoreState(cmgr cbgo.CentralManager, opts cbgo.CentralManagerRestoreOpts) {
	log.Debug("CentralManagerWillRestoreState")
}

func (mgr *btDarwinManager) DidDiscoverPeripheral(cmgr cbgo.CentralManager, p cbgo.Peripheral, f cbgo.AdvFields, rssi int) {
	log.Debug("DidDiscoverPeripheral")
	props := make(map[string]string)
	props[device.ReportDataSuggestedIdentifier] = p.Identifier().String()

	v := strings.TrimSpace(p.Name())
	if len(v) > 0 {
		props[device.ReportDataSuggestedDescription] = v
	} else {
		v = strings.TrimSpace(f.LocalName)
		if len(v) > 0 {
			props[device.ReportDataSuggestedDescription] = v
		}
	}

	for i, u := range f.ServiceUUIDs {
		props[fmt.Sprintf("ServiceUUID-%d", i)] = u.String()
	}
	for _, data := range f.ServiceData {
		props[fmt.Sprintf("ServiceData-%s", data.UUID.String())] = string(data.Data)
	}
	itf := model.Interface{
		Type:       model.InterfaceBluetooth,
		MACAddress: p.Identifier().String(), // on OSX this is not a MACAddress but a UUID
	}
	mgr.report(itf, props)
}

func (mgr *btDarwinManager) stopScan() {
	log.Debug("Stop scanning for Bluetooth devices...")
	mgr.cm.StopScan()
}
