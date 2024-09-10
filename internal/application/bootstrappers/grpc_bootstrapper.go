package bootstrappers

import (
	"fmt"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/server"
)

type GrpcBootstrapper struct{}

func (g *GrpcBootstrapper) Tags() []string {
	return []string{"api", "grpc"}
}

func (g *GrpcBootstrapper) Bootstrap(app application.App) error {
	app.AddServer(server.NewGrpcRunner(fmt.Sprintf("0.0.0.0:%v", app.Config().GrpcPort), app))

	return nil
}
