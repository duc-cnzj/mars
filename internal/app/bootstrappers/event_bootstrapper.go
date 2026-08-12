package bootstrappers

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
)

// EventBootstrapper 启动事件分发 server。
type EventBootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
func (e *EventBootstrapper) Tags() []string {
	return []string{}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (e *EventBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	deps.AddServer(deps.Dispatcher())
	deps.Logger().Debug("EventBootstrapper booted.")

	return nil
}
