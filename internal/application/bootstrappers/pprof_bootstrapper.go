package bootstrappers

import (
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/server"
)

type PprofBootstrapper struct{}

func (p *PprofBootstrapper) Tags() []string {
	return []string{"profile"}
}

func (p *PprofBootstrapper) Bootstrap(app application.App) error {
	app.AddServer(server.NewPprofRunner(app.Logger()))

	return nil
}
