package bootstrappers

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/server"
)

// PprofBootstrapper 启动 pprof http server。
type PprofBootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
// tag 用 "profile"（与 --exclude_server 文档/README 一致），而非类型名 "pprof"，
// 否则用户按文档执行 --exclude_server profile 会静默失效。
func (p *PprofBootstrapper) Tags() []string {
	return []string{"profile"}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (p *PprofBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	deps.AddServer(server.NewPprofRunner(deps.Logger()))

	return nil
}
