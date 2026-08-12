package bootstrappers

import (
	"fmt"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/server"
)

// GrpcBootstrapper 启动 grpc server。
type GrpcBootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
func (g *GrpcBootstrapper) Tags() []string {
	return []string{"api", "grpc"}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (g *GrpcBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	deps.AddServer(server.NewGrpcRunner(fmt.Sprintf("0.0.0.0:%v", deps.Config().GrpcPort), deps))

	return nil
}
