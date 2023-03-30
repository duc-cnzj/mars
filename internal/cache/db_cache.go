package cache

import (
	"encoding/base64"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/models"
)

type dbStore struct {
	db func() *gorm.DB
}

func NewDBStore(db func() *gorm.DB) contracts.Store {
	return &dbStore{db: db}
}

func (d *dbStore) Get(key string) (value []byte, err error) {
	var cache models.DBCache
	err = d.db().Where("`key` = ? and `expired_at` >= ?", key, time.Now()).First(&cache).Error
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(cache.Value)
}

func (d *dbStore) Set(key string, value []byte, seconds int) (err error) {
	toString := base64.StdEncoding.EncodeToString(value)
	cache := models.DBCache{
		Key:       key,
		Value:     toString,
		ExpiredAt: time.Now().Add(time.Duration(seconds) * time.Second),
	}

	return d.db().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "expired_at"}),
	}).Create(&cache).Error
}

func (d *dbStore) Delete(key string) error {
	return d.db().Where("`key` = ?", key).Delete(&models.DBCache{}).Error
}
