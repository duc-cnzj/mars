package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Action   uint8  `json:"action" gorm:"type:tinyint;not null;default:0;"`
	Username string `json:"username" gorm:"size:255;not null;default:'';comment:用户名称"`
	Message  string `json:"message" gorm:"size:255;not null;default:'';"`

	Old string `json:"old" gorm:"type:text;"`
	New string `json:"new" gorm:"type:text;"`

	FileID *int `json:"file_id" gorm:"nullable;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	File     *File
	Commands []*Command
}
