package event

import (
	"sync"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

type Dispatcher struct {
	sync.RWMutex

	app       contracts.ApplicationInterface
	listeners map[contracts.Event][]contracts.Listener
}

func NewDispatcher(app contracts.ApplicationInterface) *Dispatcher {
	return &Dispatcher{listeners: map[contracts.Event][]contracts.Listener{}, app: app}
}

func (d *Dispatcher) Listen(event contracts.Event, listener contracts.Listener) {
	d.Lock()
	defer d.Unlock()

	if listeners, ok := d.listeners[event]; ok {
		d.listeners[event] = append(listeners, listener)
	} else {
		d.listeners[event] = []contracts.Listener{listener}
	}
}

func (d *Dispatcher) HasListeners(event contracts.Event) bool {
	d.RLock()
	defer d.RUnlock()

	if listeners, ok := d.listeners[event]; ok {
		return len(listeners) > 0
	}

	return false
}

func (d *Dispatcher) Dispatch(event contracts.Event, payload any) error {
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

func (d *Dispatcher) Forget(event contracts.Event) {
	d.Lock()
	defer d.Unlock()
	delete(d.listeners, event)
}

func (d *Dispatcher) GetListeners(event contracts.Event) []contracts.Listener {
	d.RLock()
	defer d.RUnlock()

	return d.listeners[event]
}
