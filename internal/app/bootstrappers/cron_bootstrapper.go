package bootstrappers

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
)

// CronBootstrapper 启动定时任务管理器 server。
type CronBootstrapper struct{}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (c *CronBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	deps.AddServer(deps.CronManager())
	return nil
}

// Tags 实现 Bootstrapper 接口的 Tags。
func (c *CronBootstrapper) Tags() []string {
	return []string{"cron"}
}
