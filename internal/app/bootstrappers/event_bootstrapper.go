package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	mevent "github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type EventBootstrapper struct{}

func (e *EventBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	for e, listeners := range mevent.RegisteredEvents() {
		for _, listener := range listeners {
			app.EventDispatcher().Listen(e, listener)
		}
	}

	mlog.Debug("EventBootstrapper booted.")

	return nil
}
