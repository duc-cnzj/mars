package bootstrappers

import (
	"github.com/duc-cnzj/mars/v5/internal/application"
)

type PluginBootstrapper struct{}

func (d *PluginBootstrapper) Tags() []string {
	return []string{}
}

func (d *PluginBootstrapper) Bootstrap(app application.App) error {
	if err := app.PluginMgr().Load(app); err != nil {
		return err
	}

	app.RegisterAfterShutdownFunc(func(app application.App) {
		app.PluginMgr().Ws().Destroy()
		app.PluginMgr().Domain().Destroy()
		app.PluginMgr().Git().Destroy()
		app.PluginMgr().Picture().Destroy()
	})
	return nil
}
