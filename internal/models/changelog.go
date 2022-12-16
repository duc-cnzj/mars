package models

import (
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/utils/date"

	"gorm.io/gorm"
)

type Changelog struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Version  int    `json:"version" gorm:"not null;default:1;type:integer;index:idx_projectid_config_changed_deleted_at_version,priority:4;"`
	Username string `json:"username" gorm:"size:100;not null;comment:修改人"`
	Manifest string `json:"manifest" gorm:"type:longtext;"`
	Config   string `json:"config" gorm:"type:text;commit:用户提交的配置"`

	ConfigType       string     `json:"config_type" gorm:"size:255;nullable;"`
	GitBranch        string     `json:"git_branch" gorm:"not null;size:255;"`
	GitCommit        string     `json:"git_commit" gorm:"not null;size:255;"`
	DockerImage      string     `json:"docker_image" gorm:"not null;size:255;default:''"`
	EnvValues        string     `json:"env_values" gorm:"type:text;nullable;comment:可用的环境变量值"`
	ExtraValues      string     `json:"extra_values" gorm:"type:text;nullable;comment:用户表单传入的额外值"`
	FinalExtraValues string     `json:"final_extra_values" gorm:"type:text;nullable;comment:用户表单传入的额外值 + 系统默认的额外值"`
	GitCommitWebUrl  string     `json:"git_commit_web_url" gorm:"size:255;nullable;"`
	GitCommitTitle   string     `json:"git_commit_title" gorm:"size:255;nullable;"`
	GitCommitAuthor  string     `json:"git_commit_author" gorm:"size:255;nullable;"`
	GitCommitDate    *time.Time `json:"git_commit_date"`

	ConfigChanged bool `json:"config_changed" gorm:"index:idx_projectid_config_changed_deleted_at_version,priority:2;"`

	ProjectID    int `json:"project_id" gorm:"not null;default:0;index:idx_projectid_config_changed_deleted_at_version,priority:1;"`
	GitProjectID int `json:"git_project_id" gorm:"not null;default:0;"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index:idx_projectid_config_changed_deleted_at_version,priority:3;"`

	Project    Project
	GitProject GitProject
}

func (c *Changelog) ProtoTransform() *types.ChangelogModel {
	return &types.ChangelogModel{
		Id:               int64(c.ID),
		Version:          int64(c.Version),
		Username:         c.Username,
		Manifest:         c.Manifest,
		Config:           c.Config,
		ConfigChanged:    c.ConfigChanged,
		ProjectId:        int64(c.ProjectID),
		GitProjectId:     int64(c.GitProjectID),
		Project:          c.Project.ProtoTransform(),
		GitProject:       c.GitProject.ProtoTransform(),
		Date:             date.ToHumanizeDatetimeString(&c.CreatedAt),
		ConfigType:       c.ConfigType,
		GitBranch:        c.GitBranch,
		GitCommit:        c.GitCommit,
		DockerImage:      c.DockerImage,
		EnvValues:        c.EnvValues,
		ExtraValues:      c.ExtraValues,
		FinalExtraValues: c.FinalExtraValues,
		GitCommitWebUrl:  c.GitCommitWebUrl,
		GitCommitTitle:   c.GitCommitTitle,
		GitCommitAuthor:  c.GitCommitAuthor,
		GitCommitDate:    date.ToHumanizeDatetimeString(c.GitCommitDate),
		CreatedAt:        date.ToRFC3339DatetimeString(&c.CreatedAt),
		UpdatedAt:        date.ToRFC3339DatetimeString(&c.UpdatedAt),
		DeletedAt:        date.ToRFC3339DatetimeString(&c.DeletedAt.Time),
	}
}
