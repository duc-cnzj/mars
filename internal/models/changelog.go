package models

import (
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/utils/date"

	"gorm.io/gorm"
)

type Changelog struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Version       uint8  `json:"version" gorm:"not null;default:1;"`
	Username      string `json:"username" gorm:"size:100;not null;comment:修改人"`
	Manifest      string `json:"manifest" gorm:"type:longtext;"`
	Config        string `json:"config" gorm:"type:text;commit:用户提交的配置"`
	ConfigChanged bool   `json:"config_changed"`

	ProjectID    int `json:"project_id" gorm:"not null;default:0;"`
	GitProjectID int `json:"git_project_id" gorm:"not null;default:0;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Project    Project
	GitProject GitProject
}

func (c *Changelog) ProtoTransform() *types.ChangelogModel {
	return &types.ChangelogModel{
		Id:            int64(c.ID),
		Version:       int64(c.Version),
		Username:      c.Username,
		Manifest:      c.Manifest,
		Config:        c.Config,
		ConfigChanged: c.ConfigChanged,
		ProjectId:     int64(c.ProjectID),
		GitProjectId:  int64(c.GitProjectID),
		Project:       c.Project.ProtoTransform(),
		GitProject:    c.GitProject.ProtoTransform(),
		Date:          date.ToHumanizeDatetimeString(&c.CreatedAt),
		CreatedAt:     date.ToRFC3339DatetimeString(&c.CreatedAt),
		UpdatedAt:     date.ToRFC3339DatetimeString(&c.UpdatedAt),
		DeletedAt:     date.ToRFC3339DatetimeString(&c.DeletedAt.Time),
	}
}
