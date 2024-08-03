package bootstrappers

import (
	"github.com/duc-cnzj/mars/v4/internal/application"
)

type PluginBootstrapper struct{}

func (d *PluginBootstrapper) Tags() []string {
	return []string{}
}

func (d *PluginBootstrapper) Bootstrap(app application.App) error {
	return app.PluginMgr().Load(app)
}
