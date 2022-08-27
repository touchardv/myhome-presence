package bluetooth

import (
	"context"
	"time"

	"github.com/JuulLabs-OSS/cbgo"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/device"
	"github.com/touchardv/myhome-presence/model"
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
	log.Debug("Identifier: ", p.Identifier(), " Name: ", p.Name(), " LocalName: ", f.LocalName)
	for _, u := range f.ServiceUUIDs {
		log.Debug("ServiceUUID: ", u)
	}
	for _, data := range f.ServiceData {
		log.Debug("ServiceData: ", data.UUID, " -> ", data.Data)
	}
	itf := model.Interface{
		Type:       model.InterfaceBluetoothLowEnergy,
		MACAddress: p.Identifier().String(), // on OSX this is not a MACAddress but a UUID
	}
	mgr.report(itf)
}

func (mgr *btDarwinManager) stopScan() {
	log.Debug("Stop scanning for Bluetooth devices...")
	mgr.cm.StopScan()
}
