package bootstrappers

import (
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/server"
)

type PprofBootstrapper struct{}

func (p *PprofBootstrapper) Tags() []string {
	return []string{"pprof"}
}

func (p *PprofBootstrapper) Bootstrap(app application.App) error {
	app.AddServer(server.NewPprofRunner(app.Logger()))

	return nil
}
