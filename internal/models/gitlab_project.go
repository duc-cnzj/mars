package models

import (
	"time"

	"gorm.io/gorm"
)

type GitlabProject struct {
	ID int `json:"id" gorm:"primaryKey;"`

	GitlabProjectId int  `json:"gitlab_project_id" gorm:"not null;type:integer;"`
	Enabled         bool `json:"enabled" gorm:"not null;default:false;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}
