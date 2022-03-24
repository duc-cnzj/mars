package contracts

import "gorm.io/gorm"

type DBManager interface {
	DB() *gorm.DB
	SetDB(*gorm.DB)

	AutoMigrate(dst ...any) error
}
