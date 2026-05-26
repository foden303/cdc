package pool

import (
	"sync"

	"github.com/foden/cdc/internal/core/domain"
)

var eventPool = sync.Pool{
	New: func() any {
		return new(domain.Event)
	},
}

// GetEvent retrieves an Event from the pool or allocates a new one.
func GetEvent() *domain.Event {
	return eventPool.Get().(*domain.Event)
}

// PutEvent resets the Event fields and returns it to the pool.
func PutEvent(ev *domain.Event) {
	if ev == nil {
		return
	}
	ev.Reset() // Explicitly clear fields
	eventPool.Put(ev)
}
