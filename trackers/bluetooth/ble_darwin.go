package bluetooth

import "github.com/bettercap/gatt"

var defaultClientOptions = []gatt.Option{
	gatt.MacDeviceRole(gatt.CentralManager),
}
