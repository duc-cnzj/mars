package contracts

//go:generate mockgen -destination ../mock/mock_db.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts DBManager

import "gorm.io/gorm"

type DBManager interface {
	DB() *gorm.DB
	SetDB(*gorm.DB)

	AutoMigrate(dst ...any) error
}
