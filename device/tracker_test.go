package device

import (
	"context"
	"sync"

	"github.com/touchardv/myhome-presence/model"
)

type dummyTracker struct {
	pingCount int
}

var tracker dummyTracker

func newDummyTracker() Tracker {
	return &tracker
}

func (t *dummyTracker) Loop(f ReportPresenceFunc, ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	<-ctx.Done()
	return nil
}

func (t *dummyTracker) Ping(device model.Device) {
	t.pingCount++
}

func init() {
	Register("dummy", newDummyTracker)
}
