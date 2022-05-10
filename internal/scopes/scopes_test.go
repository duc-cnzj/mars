package scopes

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestOrderByIdDesc(t *testing.T) {
	sqlDB, _, _ := sqlmock.New()
	defer sqlDB.Close()
	gormDB, _ := gorm.Open(mysql.New(mysql.Config{SkipInitializeWithVersion: true, Conn: sqlDB}), &gorm.Config{})
	db := OrderByIdDesc()(gormDB)
	_, ok := db.Statement.Clauses["ORDER BY"]
	assert.True(t, ok)
}

func TestPaginate(t *testing.T) {
	var page, pageSize int
	sqlDB, _, _ := sqlmock.New()
	defer sqlDB.Close()
	gormDB, _ := gorm.Open(mysql.New(mysql.Config{SkipInitializeWithVersion: true, Conn: sqlDB}), &gorm.Config{})
	Paginate(&page, &pageSize)(gormDB)
	assert.Equal(t, 1, page)
	assert.Equal(t, 15, pageSize)
}
