package models

import (
	"time"

	"gorm.io/gorm"
)

type Changelog struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Version       uint8  `json:"version" gorm:"not null;default:1;"`
	Username      string `json:"username" gorm:"size:100;not null;comment:'修改人'"`
	Manifest      string `json:"manifest" gorm:"type:text;"`
	Config        string `json:"config" gorm:"type:text;commit:用户提交的配置"`
	ConfigChanged bool   `json:"config_changed"`

	ProjectID       int `json:"project_id" gorm:"not null;default:0;"`
	GitlabProjectID int `json:"gitlab_project_id" gorm:"not null;default:0;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Project       Project
	GitlabProject GitlabProject
}
