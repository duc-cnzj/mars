package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Name            string `json:"name" gorm:"size:100;not null;comment:'项目名'"`
	GitlabProjectId int    `json:"gitlab_project_id" gorm:"not null;type:integer;"`
	GitlabBranch    string `json:"gitlab_branch" gorm:"not null;size:255;"`
	GitlabCommit    string `json:"gitlab_commit" gorm:"not null;size:255;"`
	Config          string `json:"config"`

	DockerImage string `json:"docker_image" gorm:"not null;size:255;default:''"`

	NamespaceId int `json:"namespace_id"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Namespace Namespace
}
