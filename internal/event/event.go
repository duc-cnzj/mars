package event

import (
	"context"
	"sync"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// Event is a named application event dispatched through the Dispatcher.
type Event string

// String implements fmt.Stringer for Event.
func (e Event) String() string {
	return string(e)
}

// Listener handles a dispatched event payload. The returned error is logged
// by the dispatcher and does not stop the remaining listeners of the event.
type Listener func(any, Event) error

// eventChannelBuffer is the capacity of the dispatcher's event channel.
const eventChannelBuffer = 800

// Dispatcher is an in-process, fire-and-forget event bus.
type Dispatcher interface {
	// Listen registers a listener for the given event.
	Listen(Event, Listener)

	// Dispatch enqueues an event asynchronously. It never blocks: when the
	// internal buffer is full, the event is dropped and a warning is logged.
	Dispatch(Event, any)

	// GetListeners returns a copy of the listeners registered for the event.
	GetListeners(Event) []Listener

	// Run starts the dispatcher's processing loop in a background goroutine.
	// It must be invoked exactly once.
	Run(context.Context) error

	// Shutdown stops the dispatcher's processing loop.
	Shutdown(context.Context) error

	// List returns a copy of all listeners, keyed by event.
	List() map[Event][]Listener
}

// eventBody is a queued event and its payload.
type eventBody struct {
	event   Event
	payload any
}

// dispatcher is the in-memory Dispatcher implementation. All access to the
// listeners map is guarded by the embedded sync.RWMutex.
type dispatcher struct {
	sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	ch        chan *eventBody
	logger    mlog.Logger
	listeners map[Event][]Listener
}

var _ Dispatcher = (*dispatcher)(nil)

// NewDispatcher returns a Dispatcher backed by an eventChannelBuffer-capacity
// event channel.
func NewDispatcher(logger mlog.Logger) Dispatcher {
	ctx, cancelFunc := context.WithCancel(context.TODO())

	return &dispatcher{
		ctx:       ctx,
		cancel:    cancelFunc,
		ch:        make(chan *eventBody, eventChannelBuffer),
		logger:    logger.WithModule("event/dispatcher"),
		listeners: map[Event][]Listener{},
	}
}

// Run starts the processing loop in a background goroutine and returns nil.
// The loop drains the event channel and runs each event's listeners; it stops
// when the dispatcher or the caller's context is done, or the channel closes.
// The error return satisfies application.Server: starting the loop cannot fail.
// Callers must invoke it exactly once.
func (d *dispatcher) Run(ctx context.Context) error {
	d.logger.Info("[Event]: dispatcher running")
	go func() {
		for {
			select {
			case <-d.ctx.Done():
				d.logger.Warning("[Shutdown]: event dispatcher context done")
				return
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
						if err := fn(obj.payload, obj.event); err != nil {
							d.logger.Error(err)
						}
					}
				}()
			}
		}
	}()
	return nil
}

// Shutdown cancels the dispatcher's internal context, stopping the processing
// loop started by Run. Shutdown is instantaneous; the passed context is not
// consulted (reserved for a future graceful-drain implementation).
func (d *dispatcher) Shutdown(ctx context.Context) error {
	d.logger.Info("[Event]: dispatcher shutdown")
	d.cancel()
	return nil
}

// Listen registers a listener for the given event.
func (d *dispatcher) Listen(event Event, listener Listener) {
	d.Lock()
	defer d.Unlock()

	// append on a nil slice yields []Listener{listener}, covering the first
	// registration for an event.
	d.listeners[event] = append(d.listeners[event], listener)
}

// Dispatch enqueues an event. Non-blocking: when the buffer is full, the
// event is dropped and a warning is logged.
func (d *dispatcher) Dispatch(event Event, payload any) {
	select {
	case d.ch <- &eventBody{
		event:   event,
		payload: payload,
	}:
	default:
		d.logger.Warningf("event dispatcher channel full drop event %v %v", event.String(), payload)
	}
}

// GetListeners returns a copy of the listeners registered for the event, so
// callers cannot mutate the dispatcher's internal registration.
func (d *dispatcher) GetListeners(event Event) []Listener {
	d.RLock()
	defer d.RUnlock()

	return append([]Listener(nil), d.listeners[event]...)
}

// List returns a copy of all listeners, keyed by event, so callers cannot
// mutate the dispatcher's internal registration.
func (d *dispatcher) List() map[Event][]Listener {
	d.RLock()
	defer d.RUnlock()

	out := make(map[Event][]Listener, len(d.listeners))
	for event, listeners := range d.listeners {
		out[event] = append([]Listener(nil), listeners...)
	}
	return out
}
