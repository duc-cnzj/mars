package bootstrappers

import (
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
)

type PluginsBootstrapper struct{}

func (a *PluginsBootstrapper) Tags() []string {
	return []string{}
}

func (a *PluginsBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.SetPlugins(plugins.GetPlugins())

	return nil
}
