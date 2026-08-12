package bootstrappers

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
)

// PluginBootstrapper 加载与销毁配置的插件。
type PluginBootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
func (p *PluginBootstrapper) Tags() []string {
	return []string{}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (p *PluginBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	pl := deps.PluginManager()
	if err := pl.Load(deps); err != nil {
		return err
	}

	deps.RegisterAfterShutdownFunc(func() {
		// 插件回收的 nil 守卫、逆序与错误日志统一收敛在 manager.Destroy 中。
		pl.Destroy()
	})
	return nil
}
