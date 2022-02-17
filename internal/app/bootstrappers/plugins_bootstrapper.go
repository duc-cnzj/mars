package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/plugins"
)

type PluginsBootstrapper struct{}

func (a *PluginsBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.SetPlugins(plugins.GetPlugins())

	// 预加载插件
	plugins.GetWsSender()
	plugins.GetPicture()
	plugins.GetGitServer()
	plugins.GetDomainManager()

	return nil
}
