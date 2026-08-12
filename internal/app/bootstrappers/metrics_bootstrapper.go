package bootstrappers

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/server"
)

// MetricsBootstrapper 启动 prometheus 指标 server。
type MetricsBootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
func (m *MetricsBootstrapper) Tags() []string {
	return []string{"metrics"}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (m *MetricsBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	deps.AddServer(server.NewMetricsRunner(
		deps.Config().MetricsPort,
		deps.Logger(),
		deps.PrometheusRegistry()),
	)

	return nil
}
