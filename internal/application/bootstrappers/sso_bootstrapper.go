package bootstrappers

import (
	"github.com/duc-cnzj/mars/v4/internal/application"
)

type SSOBootstrapper struct{}

func (d *SSOBootstrapper) Tags() []string {
	return []string{}
}

func (d *SSOBootstrapper) Bootstrap(app application.App) error {
	app.Data().InitOidcProvider()
	return nil
}
