package event

import (
	"sync"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

type dispatcher struct {
	sync.RWMutex

	app       contracts.ApplicationInterface
	listeners map[contracts.Event][]contracts.Listener
}

// NewDispatcher return contracts.DispatcherInterface.
func NewDispatcher(app contracts.ApplicationInterface) contracts.DispatcherInterface {
	return &dispatcher{listeners: map[contracts.Event][]contracts.Listener{}, app: app}
}

// Listen Register an event listener with the dispatcher.
func (d *dispatcher) Listen(event contracts.Event, listener contracts.Listener) {
	d.Lock()
	defer d.Unlock()

	if listeners, ok := d.listeners[event]; ok {
		d.listeners[event] = append(listeners, listener)
	} else {
		d.listeners[event] = []contracts.Listener{listener}
	}
}

// HasListeners Determine if a given event has listeners.
func (d *dispatcher) HasListeners(event contracts.Event) bool {
	d.RLock()
	defer d.RUnlock()

	if listeners, ok := d.listeners[event]; ok {
		return len(listeners) > 0
	}

	return false
}

// Dispatch Fire an event and call the listeners.
func (d *dispatcher) Dispatch(event contracts.Event, payload any) error {
	d.RLock()
	defer d.RUnlock()
	if listeners, ok := d.listeners[event]; ok {
		for _, listener := range listeners {
			if err := listener(payload, event); err != nil {
				return err
			}
		}
	}

	return nil
}

// Forget Remove a set of listeners from the dispatcher.
func (d *dispatcher) Forget(event contracts.Event) {
	d.Lock()
	defer d.Unlock()
	delete(d.listeners, event)
}

// GetListeners get all listeners by event.
func (d *dispatcher) GetListeners(event contracts.Event) []contracts.Listener {
	d.RLock()
	defer d.RUnlock()

	return d.listeners[event]
}
