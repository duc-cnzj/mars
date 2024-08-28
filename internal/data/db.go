package data

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"

	_ "github.com/mattn/go-sqlite3"
)

func NewSqliteDB() (*ent.Client, error) {
	client, _ := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&loc=Local")
	_ = client.Schema.Create(context.TODO())

	return client, nil
}

type slowLogDriver struct {
	dialect.Driver
	slowThreshold time.Duration
	logger        mlog.Logger
	timer         timer.Timer
}

func (d *slowLogDriver) Exec(ctx context.Context, query string, args, v any) error {
	start := d.timer.Now()
	err := d.Driver.Exec(ctx, query, args, v)
	elapsed := time.Since(start)
	if elapsed > d.slowThreshold {
		d.logger.Infof("slow query: %s, args: %v, took: %s", query, args, elapsed)
	}
	return err
}

func (d *slowLogDriver) Query(ctx context.Context, query string, args, v any) error {
	start := d.timer.Now()
	err := d.Driver.Query(ctx, query, args, v)
	elapsed := time.Since(start)
	if elapsed > d.slowThreshold {
		d.logger.Infof("slow query: %s, args: %v, took: %s", query, args, elapsed)
	}
	return err
}

func OpenDB(config *config.Config) (*sql.Driver, error) {
	switch config.DBDriver {
	case "sqlite":
		return sql.Open("sqlite3", fmt.Sprintf("file:%v?cache=shared&_fk=1", config.DBDatabase))
	case "mysql":
		return sql.Open("mysql", config.DSN())
	}
	return nil, fmt.Errorf("unsupported database driver %v", config.DBDriver)
}

func InitDB(drv dialect.Driver, logger mlog.Logger, slogLogEnabled bool, slowLogThreshold time.Duration, timer timer.Timer) (*ent.Client, error) {
	db := drv.(*sql.Driver).DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)
	if slogLogEnabled {
		drv = &slowLogDriver{
			timer:         timer,
			Driver:        drv,
			slowThreshold: slowLogThreshold,
			logger:        logger.WithModule("SlowLog"),
		}
	}
	dbCli := ent.NewClient(
		ent.Driver(drv),
		ent.Log(func(a ...any) {
			logger.Debug(a...)
		}),
	)

	return dbCli, nil
}
