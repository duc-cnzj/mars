package models

import (
	"time"

	"gorm.io/gorm"
)

type DBCache struct {
	Key       string    `gorm:"size:255;not null;primaryKey;"`
	Value     string    `json:"value"`
	ExpiredAt time.Time `json:"expired_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func (c DBCache) TableName() string {
	return "db_cache"
}
