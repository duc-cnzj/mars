package database

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestManager_DB(t *testing.T) {
	manager := NewManager(nil)
	db := gorm.DB{}
	manager.SetDB(&db)
	assert.Same(t, &db, manager.DB())
}

func TestManager_SetDB(t *testing.T) {
	manager := NewManager(nil)
	db := gorm.DB{}
	manager.SetDB(&db)
	assert.Same(t, &db, manager.db)
}

func TestNewManager(t *testing.T) {
	assert.Implements(t, (*contracts.DBManager)(nil), NewManager(nil))
}
