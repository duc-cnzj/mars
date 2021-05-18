package contracts

type Listener func(interface{}, Event) error

type Event string

func (e Event) String() string {
	return string(e)
}

func (e Event) Is(event Event) bool {
	return e.String() == event.String()
}

type DispatcherInterface interface {
	// Listen Register an event listener with the dispatcher.
	Listen(Event, Listener)

	// HasListeners Determine if a given event has listeners.
	HasListeners(Event) bool

	// Dispatch Fire an event and call the listeners.
	Dispatch(Event, interface{}) error

	// Forget Remove a set of listeners from the dispatcher.
	Forget(Event)

	GetListeners(Event) []Listener
}
