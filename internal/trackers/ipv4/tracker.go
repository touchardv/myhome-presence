package ipv4

import (
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/internal/config"
	"github.com/touchardv/myhome-presence/internal/device"
	"golang.org/x/net/icmp"
)

// EnableTracker registers the "ipv4" tracker so that it can be used.
func EnableTracker() {
	device.Register("ipv4", newIPTracker)
}

type ipTracker struct {
	pingPacketCount int
	pingPacketDelay time.Duration
	sequenceNumber  int
	socket          *icmp.PacketConn
	stopReceiving   bool
}

func newIPTracker(settings config.Settings) device.Tracker {
	count := defaultPingPacketCount
	if v, ok := settings["ping_packet_count"]; ok {
		c, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal("Invalid ping_packet_count setting value: ", err)
		}
		count = c
	}
	delay := defaultPingPacketDelay
	if v, ok := settings["ping_packet_delay"]; ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			log.Fatal("Invalid ping_packet_delay setting value: ", err)
		}
		delay = d
	}
	return &ipTracker{
		pingPacketCount: count,
		pingPacketDelay: delay,
		sequenceNumber:  0,
		stopReceiving:   false,
	}
}
