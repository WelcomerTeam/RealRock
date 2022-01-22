package interfacecache

import (
	"sync"
	"time"

	"go.uber.org/atomic"
)

type interfaceCache struct {
	resMu sync.Mutex
	res   interface{}

	lastRequest *atomic.Time
}

// InterfaceCache allows for easy reuse of a specific interface value
// for a specified duration of time. Does not reset the lastRequest when
// a new request is made.
func NewInterfaceCache() (ic *interfaceCache) {
	return &interfaceCache{
		lastRequest: &atomic.Time{},
	}
}

func (ic *interfaceCache) Result(dur time.Duration, getter func() interface{}) (res interface{}) {
	if now := time.Now().UTC(); ic.lastRequest.Load().Add(dur).Before(now) {
		ic.resMu.Lock()
		ic.res = getter()
		ic.resMu.Unlock()

		ic.lastRequest.Store(now)
	}

	ic.resMu.Lock()
	res = ic.res
	ic.resMu.Unlock()

	return res
}
