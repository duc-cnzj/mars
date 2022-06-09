package events

import (
	"sync"

	"github.com/duc-cnzj/mars/internal/contracts"
)

var (
	registry = make(map[contracts.Event][]contracts.Listener)
	mu       sync.RWMutex
)

func Register(e contracts.Event, l contracts.Listener) {
	mu.Lock()
	defer mu.Unlock()
	registry[e] = append(registry[e], l)
}

func RegisteredEvents() map[contracts.Event][]contracts.Listener {
	mu.RLock()
	defer mu.RUnlock()
	return registry
}
