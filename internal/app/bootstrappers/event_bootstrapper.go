package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event"
	mevent "github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type EventBootstrapper struct{}

func (e *EventBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.SetEventDispatcher(event.NewDispatcher(app))

	for e, listeners := range mevent.RegisteredEvents() {
		for _, listener := range listeners {
			app.EventDispatcher().Listen(e, listener)
		}
	}

	mlog.Debug("EventBootstrapper booted.")

	return nil
}
