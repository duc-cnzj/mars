package bootstrappers

import (
	"github.com/duc-cnzj/mars/pkg/contracts"
	"github.com/duc-cnzj/mars/pkg/event"
	"github.com/duc-cnzj/mars/pkg/mlog"
)

type EventBootstrapper struct{}

var events map[contracts.Event][]contracts.Listener = map[contracts.Event][]contracts.Listener{}

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
