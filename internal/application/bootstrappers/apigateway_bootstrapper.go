package bootstrappers

import (
	"context"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/server"
)

type ApiGatewayBootstrapper struct{}

func (a *ApiGatewayBootstrapper) Tags() []string {
	return []string{"api", "gateway"}
}

func (a *ApiGatewayBootstrapper) Bootstrap(appli application.App) error {
	appli.AddServer(server.NewApiGateway(fmt.Sprintf("localhost:%s", appli.Config().GrpcPort), appli))
	appli.RegisterAfterShutdownFunc(func(app application.App) {
		ctx, cancelFunc := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancelFunc()
		if err := appli.WsServer().Shutdown(ctx); err != nil {
			app.Logger().Warning("shutdown ws server error: ", err.Error())
		}
	})
	appli.BeforeServerRunHooks(func(app application.App) {
		go app.WsServer().TickClusterHealth(app.Done())
	})

	return nil
}
