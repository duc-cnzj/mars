package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
)

type CronBootstrapper struct{}

func (c *CronBootstrapper) Tags() []string {
	return []string{"cron"}
}

func (c *CronBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(app.CronManager())
	return nil
}
