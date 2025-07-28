package ipv4

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/touchardv/myhome-presence/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.Settings{}

	tr := newIPTracker(cfg)
	tracker := tr.(*ipTracker)
	assert.Equal(t, 5, tracker.pingPacketCount)
	assert.Equal(t, 100*time.Millisecond, tracker.pingPacketDelay)
	assert.Equal(t, 0, tracker.sequenceNumber)

	cfg["ping_packet_count"] = "1"
	cfg["ping_packet_delay"] = "250ms"
	tr = newIPTracker(cfg)
	tracker = tr.(*ipTracker)
	assert.Equal(t, 1, tracker.pingPacketCount)
	assert.Equal(t, 250*time.Millisecond, tracker.pingPacketDelay)
}
