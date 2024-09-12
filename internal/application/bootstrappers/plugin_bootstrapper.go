package bootstrappers

import (
	"github.com/duc-cnzj/mars/v5/internal/application"
)

type PluginBootstrapper struct{}

func (d *PluginBootstrapper) Tags() []string {
	return []string{}
}

func (d *PluginBootstrapper) Bootstrap(app application.App) error {
	pl := app.PluginMgr()
	if err := pl.Load(app); err != nil {
		return err
	}

	app.RegisterAfterShutdownFunc(func(app application.App) {
		pl.Ws().Destroy()
		pl.Domain().Destroy()
		pl.Git().Destroy()
		pl.Picture().Destroy()
	})
	return nil
}
