package database

import (
	"github.com/duc-cnzj/mars/pkg/contracts"
	"gorm.io/gorm"
)

type Manager struct {
	app contracts.ApplicationInterface
	db  *gorm.DB
}

func NewManager(app contracts.ApplicationInterface) *Manager {
	return &Manager{app: app}
}

func (m *Manager) DB() *gorm.DB {
	return m.db
}

func (m *Manager) SetDB(db *gorm.DB) {
	m.db = db
}

func (m *Manager) AutoMigrate(dst ...interface{}) error {
	if err := m.db.AutoMigrate(dst...); err != nil {
		return err
	}

	return nil
}
