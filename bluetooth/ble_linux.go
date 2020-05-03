package package bluetooth

import "github.com/bettercap/gatt"

var defaultClientOptions = []gatt.Option{
	gatt.LnxMaxConnections(1),
	gatt.LnxDeviceID(-1, true),
}
