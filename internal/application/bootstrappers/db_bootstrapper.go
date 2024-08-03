package bootstrappers

import (
	"context"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/ent/migrate"
)

type DBBootstrapper struct{}

func (d *DBBootstrapper) Tags() []string {
	return []string{}
}

func (d *DBBootstrapper) Bootstrap(app application.App) error {
	app.Logger().Info("[DB]: auto migrate database")
	return app.Data().DB.Schema.Create(
		context.TODO(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
}
