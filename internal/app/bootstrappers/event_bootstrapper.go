package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event"
	mevent "github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type EventBootstrapper struct{}

var events map[contracts.Event][]contracts.Listener = map[contracts.Event][]contracts.Listener{
	mevent.EventNamespaceDeleted: {
		mevent.HandleNamespaceDeleted,
	},
	mevent.EventProjectDeleted: {
		mevent.HandleProjectDeleted,
	},
	mevent.EventProjectChanged: {
		mevent.HandleProjectChanged,
	},
}

func (e *EventBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.SetEventDispatcher(event.NewDispatcher(app))

	for e, listeners := range events {
		for _, listener := range listeners {
			app.EventDispatcher().Listen(e, listener)
		}
	}

	mlog.Debug("EventBootstrapper booted.")

	return nil
}
