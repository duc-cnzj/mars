package bootstrappers

import (
	"fmt"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/server"
)

// ApiGatewayBootstrapper 启动 api 网关 server。
type ApiGatewayBootstrapper struct{}

// Tags 实现 Bootstrapper 接口的 Tags。
func (a *ApiGatewayBootstrapper) Tags() []string {
	return []string{"api", "gateway"}
}

// Bootstrap 实现 Bootstrapper 接口的 Bootstrap。
func (a *ApiGatewayBootstrapper) Bootstrap(deps app.BootstrapDeps) error {
	deps.AddServer(server.NewApiGateway(fmt.Sprintf("localhost:%s", deps.Config().GrpcPort), deps))

	return nil
}
