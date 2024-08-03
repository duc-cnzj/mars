package bootstrappers

import (
	"github.com/duc-cnzj/mars/v4/internal/application"
)

type EventBootstrapper struct{}

func (e *EventBootstrapper) Tags() []string {
	return []string{}
}

func (e *EventBootstrapper) Bootstrap(app application.App) error {
	app.AddServer(app.Dispatcher())
	app.Logger().Debug("EventBootstrapper booted.")

	return nil
}
