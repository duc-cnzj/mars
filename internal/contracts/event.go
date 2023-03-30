package contracts

import "github.com/duc-cnzj/mars-client/v4/types"

//go:generate mockgen -destination ../mock/mock_event.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts DispatcherInterface

type Listener func(any, Event) error

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
	Dispatch(Event, any) error

	// Forget Remove a set of listeners from the dispatcher.
	Forget(Event)

	// GetListeners get all listeners by event.
	GetListeners(Event) []Listener
}

type AuditLogImpl interface {
	// GetUsername 获取用户
	GetUsername() string
	// GetAction 行为
	GetAction() types.EventActionType
	// GetMsg desc
	GetMsg() string
	// GetOldStr old config str
	GetOldStr() string
	// GetNewStr new config str
	GetNewStr() string
	// GetFileID file id
	GetFileID() int
}
