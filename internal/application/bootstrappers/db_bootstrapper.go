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

func (d *DBBootstrapper) Bootstrap(appli application.App) error {
	appli.Logger().Info("[DB]: auto migrate database")
	return appli.Data().DB.Schema.Create(
		context.TODO(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
}
