package models

type CacheLock struct {
	Key        string `gorm:"size:255;not null;primaryKey;"`
	Owner      string `gorm:"size:255;not null;"`
	Expiration int64  `gorm:"type:int;not null;"`
}
