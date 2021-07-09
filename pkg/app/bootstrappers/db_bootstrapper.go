package bootstrappers

import (
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/pkg/models"

	"github.com/duc-cnzj/mars/pkg/adapter"
	"github.com/duc-cnzj/mars/pkg/contracts"
	"github.com/duc-cnzj/mars/pkg/mlog"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Models = []interface{}{
	&models.Namespace{},
	&models.Project{},
	&models.GitlabProject{},
}

type DBBootstrapper struct{}

func (D *DBBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
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
		db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	}

	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	db.Logger = &adapter.GormLoggerAdapter{}

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

		mlog.Info("db closed.")
	})
	app.DBManager().SetDB(db)

	if err := app.DBManager().AutoMigrate(Models...); err != nil {
		return err
	}

	return nil
}
