package bootstrappers

import (
	"errors"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/adapter"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var initModels = []any{
	&models.Namespace{},
	&models.Project{},
	&models.GitProject{},
	&models.File{},
	&models.Changelog{},
	&models.Event{},
	&models.DBCache{},
}

type DBBootstrapper struct{}

func (d *DBBootstrapper) Tags() []string {
	return []string{}
}

func (d *DBBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	var (
		db  *gorm.DB
		err error
	)
	cfg := app.Config()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.DBUsername, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBDatabase)
	switch cfg.DBDriver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DBDatabase), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	default:
		return errors.New("db_driver must in ['sqlite', 'mysql']")
	}

	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	db.Logger = adapter.NewGormLoggerAdapter(adapter.GormLoggerWithSlowLog(cfg.DBSlowLogEnabled, cfg.DBSlowLogThreshold))

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)

	if app.IsDebug() {
		db.Logger.LogMode(logger.Info)
	} else {
		db.Logger.LogMode(logger.Error)
	}

	app.RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
		if err := sqlDB.Close(); err != nil {
			mlog.Error(err)
		}

		mlog.Info("[Database]: db closed.")
	})
	app.DBManager().SetDB(db)

	return app.DBManager().AutoMigrate(initModels...)
}
