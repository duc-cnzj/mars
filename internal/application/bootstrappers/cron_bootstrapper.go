package bootstrappers

import (
	"github.com/duc-cnzj/mars/v5/internal/application"
)

type CronBootstrapper struct{}

func (c *CronBootstrapper) Bootstrap(app application.App) error {
	app.AddServer(app.CronManager())
	return nil
}

func (c *CronBootstrapper) Tags() []string {
	return []string{"cron"}
}
