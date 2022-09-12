package device

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/touchardv/myhome-presence/model"
)

func TestStartStop(t *testing.T) {
	registry := NewRegistry(cfg)
	watchdog := newWatchDog(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	go watchdog.loop(registry, ctx)

	cancel()
	watchdog.stop()
}

func TestPingMissingDevices(t *testing.T) {
	registry := NewRegistry(cfg)
	watchdog := newWatchDog(cfg)

	// No ping: device is not tracked
	watchdog.pingMissingDevices(registry, time.Now())
	assert.Equal(t, 0, tracker.pingCount)

	// No ping: device is present and seen less than 5 minutes ago
	device.Status = model.StatusTracked
	device.LastSeenAt = time.Now().Add(-3 * time.Minute)
	device.Present = true
	watchdog.pingMissingDevices(registry, time.Now())
	assert.Equal(t, 0, tracker.pingCount)
	assert.True(t, device.Present)

	// Ping: device is present and seen more than 5 minutes ago
	device.LastSeenAt = time.Now().Add(-7 * time.Minute)
	watchdog.pingMissingDevices(registry, time.Now())
	assert.Equal(t, 1, tracker.pingCount)
	assert.True(t, device.Present)

	// Ping: device is absent and seen more than 10 minutes ago
	device.LastSeenAt = time.Now().Add(-15 * time.Minute)
	device.Present = false
	watchdog.pingMissingDevices(registry, time.Now())
	assert.Equal(t, 2, tracker.pingCount)
	assert.False(t, device.Present)
}
