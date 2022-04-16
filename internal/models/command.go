package models

import (
	"time"

	"gorm.io/gorm"
)

type Command struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Namespace string `json:"namespace" gorm:"size:255;not null;default:''"`
	Pod       string `json:"pod" gorm:"size:255;not null;default:''"`
	Container string `json:"container" gorm:"size:255;not null;default:''"`
	Command   string `json:"command" gorm:"type:text"`
	EventID   int    `json:"event_id" gorm:"index;not null;default:0"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Event Event
}
