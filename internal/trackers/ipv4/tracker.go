package ipv4

import (
	"github.com/touchardv/myhome-presence/internal/config"
	"github.com/touchardv/myhome-presence/internal/device"
	"golang.org/x/net/icmp"
)

// EnableTracker registers the "ipv4" tracker so that it can be used.
func EnableTracker() {
	device.Register("ipv4", newIPTracker)
}

type ipTracker struct {
	sequenceNumber int
	socket         *icmp.PacketConn
	stopReceiving  bool
}

func newIPTracker(config.Settings) device.Tracker {
	return &ipTracker{
		sequenceNumber: 0,
		stopReceiving:  false,
	}
}
