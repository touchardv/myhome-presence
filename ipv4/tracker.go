package ipv4

import (
	"github.com/touchardv/myhome-presence/device"
	"github.com/touchardv/myhome-presence/model"
	"golang.org/x/net/icmp"
)

// EnableTracker registers the "ipv4" tracker so that it can be used.
func EnableTracker() {
	device.Register("ipv4", newIPTracker)
}

type ipTracker struct {
	sequenceNumber int
	devices        map[string]model.Device
	doneReceiving  chan bool
	socket         *icmp.PacketConn
}

func newIPTracker() device.Tracker {
	return &ipTracker{
		sequenceNumber: 0,
		doneReceiving:  make(chan bool),
		devices:        make(map[string]model.Device, 10),
	}
}
