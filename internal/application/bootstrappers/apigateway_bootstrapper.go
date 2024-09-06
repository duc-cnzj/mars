package bootstrappers

import (
	"fmt"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/server"
)

type ApiGatewayBootstrapper struct{}

func (a *ApiGatewayBootstrapper) Tags() []string {
	return []string{"api", "gateway"}
}

func (a *ApiGatewayBootstrapper) Bootstrap(appli application.App) error {
	appli.AddServer(server.NewApiGateway(fmt.Sprintf("localhost:%s", appli.Config().GrpcPort), appli))

	return nil
}
