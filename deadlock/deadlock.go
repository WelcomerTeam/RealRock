package deadlock

import (
	"sync"

	"go.uber.org/atomic"
)

// Simple orchestrator to close long running tasks
// and wait for them to acknowledge completion.
type DeadSignal struct {
	sync.Mutex
	waiting sync.WaitGroup

	alreadyClosed atomic.Bool
	dead          chan bool
}

func (ds *DeadSignal) init() {
	ds.Lock()
	if ds.dead == nil {
		ds.alreadyClosed = *atomic.NewBool(false)
		ds.dead = make(chan bool, 1)
		ds.waiting = sync.WaitGroup{}
	}
	ds.Unlock()
}

// Returns the dead channel.
func (ds *DeadSignal) Dead() chan bool {
	ds.init()

	ds.Lock()
	defer ds.Unlock()

	return ds.dead
}

// Signifies the goroutine has started.
// When calling open, done should be called on end.
func (ds *DeadSignal) Started() {
	ds.init()
	ds.waiting.Add(1)
}

// Signifies the goroutine is done.
func (ds *DeadSignal) Done() {
	ds.init()
	ds.waiting.Done()
}

// Close closes the dead channel and
// waits for other goroutines waiting on Dead() to call Done().
// When Close returns, it is designed that any goroutines will no
// longer be using it.
func (ds *DeadSignal) Close(t string) {
	ds.init()

	ds.Lock()
	if !ds.alreadyClosed.Load() {
		close(ds.dead)
		ds.alreadyClosed.Store(true)
	}
	ds.Unlock()

	ds.waiting.Wait()
}

// Revive makes a closed DeadSignal create
// a new dead channel to allow for it to be reused. You should not
// revive on a closed channel if it is still being actively used.
func (ds *DeadSignal) Revive() {
	ds.init()

	ds.Lock()
	ds.dead = make(chan bool, 1)
	ds.alreadyClosed.Store(false)
	ds.Unlock()
}

// Similar to Close however does not wait for goroutines to finish.
// Both should not be ran.
func (ds *DeadSignal) Kill() {
	ds.init()

	ds.Lock()
	defer ds.Unlock()

	close(ds.dead)
}
