package event

import (
	"context"
	"sync"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type Event string

func (e Event) String() string {
	return string(e)
}

func (e Event) Is(event Event) bool {
	return e.String() == event.String()
}

type Listener func(any, Event) error

type Dispatcher interface {
	// Listen Register an event listener with the dispatcher.
	Listen(Event, Listener)

	// HasListeners Determine if a given event has listeners.
	HasListeners(Event) bool

	// Dispatch Fire an event and call the listeners.
	Dispatch(Event, any) error

	// Forget Shutdown a set of listeners from the dispatcher.
	Forget(Event)

	// GetListeners get all listeners by
	GetListeners(Event) []Listener

	// Run server.
	Run(context.Context) error

	// Shutdown server.
	Shutdown(context.Context) error

	List() map[Event][]Listener
}

type eventBody struct {
	event   Event
	Payload any
}

type dispatcher struct {
	sync.RWMutex

	ch        chan *eventBody
	logger    mlog.Logger
	listeners map[Event][]Listener
}

var _ Dispatcher = (*dispatcher)(nil)

// NewDispatcher return Dispatcher.
func NewDispatcher(logger mlog.Logger) Dispatcher {
	return &dispatcher{
		ch:        make(chan *eventBody, 1000),
		logger:    logger,
		listeners: map[Event][]Listener{},
	}
}

func (d *dispatcher) Run(ctx context.Context) error {
	d.logger.Info("[Event]: dispatcher running")
	go func() {
		for {
			select {
			case <-ctx.Done():
				d.logger.Warning("event dispatcher context done")
				return
			case obj, ok := <-d.ch:
				if !ok {
					d.logger.Warning("event dispatcher channel closed")
					return
				}
				go func() {
					defer d.logger.HandlePanic("event dispatcher")
					for _, fn := range d.GetListeners(obj.event) {
						if err := fn(obj.Payload, obj.event); err != nil {
							d.logger.Error(err)
						}
					}
				}()
			}
		}
	}()
	return nil
}

func (d *dispatcher) Shutdown(ctx context.Context) error {
	d.logger.Info("[Event]: dispatcher shutdown")
	return nil
}

// Listen Register an event listener with the dispatcher.
func (d *dispatcher) Listen(event Event, listener Listener) {
	d.Lock()
	defer d.Unlock()

	if listeners, ok := d.listeners[event]; ok {
		d.listeners[event] = append(listeners, listener)
	} else {
		d.listeners[event] = []Listener{listener}
	}
}

// HasListeners Determine if a given event has listeners.
func (d *dispatcher) HasListeners(event Event) bool {
	d.RLock()
	defer d.RUnlock()

	if listeners, ok := d.listeners[event]; ok {
		return len(listeners) > 0
	}

	return false
}

// Dispatch Fire an event and call the listeners.
func (d *dispatcher) Dispatch(event Event, payload any) error {
	select {
	case d.ch <- &eventBody{
		event:   event,
		Payload: payload,
	}:
	default:
	}

	return nil
}

// Forget Shutdown a set of listeners from the dispatcher.
func (d *dispatcher) Forget(event Event) {
	d.Lock()
	defer d.Unlock()
	delete(d.listeners, event)
}

// GetListeners get all listeners by
func (d *dispatcher) GetListeners(event Event) []Listener {
	d.RLock()
	defer d.RUnlock()

	return d.listeners[event]
}

func (d *dispatcher) List() map[Event][]Listener {
	d.RLock()
	defer d.RUnlock()
	return d.listeners
}
