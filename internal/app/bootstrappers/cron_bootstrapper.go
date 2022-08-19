package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
)

type CronBootstrapper struct{}

func (c *CronBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	if app.Config().StartCron {
		app.AddServer(app.CronManager())
	}
	return nil
}
