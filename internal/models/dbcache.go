package models

import (
	"time"
)

type DBCache struct {
	Key       string    `gorm:"size:255;not null;primaryKey;"`
	Value     string    `json:"value"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (c DBCache) TableName() string {
	return "db_cache"
}
