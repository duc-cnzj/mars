package gorm

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type AccessToken struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Token string `json:"token" gorm:"unique;size:255;not null;"`

	Usage     string    `json:"usage" gorm:"size:50;"`
	Email     string    `json:"email" gorm:"index;not null;default:'';"`
	ExpiredAt time.Time `json:"expired_at"`

	LastUsedAt sql.NullTime `json:"last_used_at"`

	UserInfo string `json:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type GitProject struct {
	ID int `json:"id" gorm:"primaryKey;"`

	DefaultBranch string `json:"default_branch" gorm:"type:varchar(255);not null;default:'';"`
	Name          string `json:"name" gorm:"type:varchar(255);not null;default:'';"`
	GitProjectId  int    `json:"git_project_id" gorm:"not null;type:integer;default:0;"`
	Enabled       bool   `json:"enabled" gorm:"not null;default:false;"`
	GlobalEnabled bool   `json:"global_enabled" gorm:"not null;default:false;"`
	GlobalConfig  string `json:"global_config" gorm:"type:longtext"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

type Namespace struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Name             string `json:"name" gorm:"size:100;not null;comment:项目空间名"`
	ImagePullSecrets string `json:"image_pull_secrets" gorm:"size:255;not null;default:'';comment:项目空间拉取镜像的secrets，数组"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index;"`

	Projects []Project
}

type Project struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Name             string `json:"name" gorm:"size:100;not null;comment:项目名"`
	GitProjectId     int    `json:"git_project_id" gorm:"not null;type:integer;"`
	GitBranch        string `json:"git_branch" gorm:"not null;size:255;"`
	GitCommit        string `json:"git_commit" gorm:"not null;size:255;"`
	Config           string `json:"config"`
	OverrideValues   string `json:"override_values"`
	DockerImage      string `json:"docker_image" gorm:"not null;size:1024;default:''"`
	PodSelectors     string `json:"pod_selectors" gorm:"type:text;nullable;"`
	NamespaceId      int    `json:"namespace_id" gorm:"index:idx_namespace_id_deleted_at,priority:1;"`
	Atomic           bool   `json:"atomic"`
	DeployStatus     uint8  `json:"deploy_status" gorm:"index:idx_deploy_status;not null;default:0"`
	EnvValues        string `json:"env_values" gorm:"type:text;nullable;comment:可用的环境变量值"`
	ExtraValues      string `json:"extra_values" gorm:"type:longtext;nullable;comment:用户表单传入的额外值"`
	FinalExtraValues string `json:"final_extra_values" gorm:"type:longtext;nullable;comment:用户表单传入的额外值 + 系统默认的额外值"`
	Version          int    `json:"version" gorm:"type:int;not null;default:1;"`

	ConfigType string `json:"config_type" gorm:"size:255;nullable;"`
	Manifest   string `json:"manifest" gorm:"type:longtext;"`

	GitCommitWebUrl string     `json:"git_commit_web_url" gorm:"size:255;nullable;"`
	GitCommitTitle  string     `json:"git_commit_title" gorm:"size:255;nullable;"`
	GitCommitAuthor string     `json:"git_commit_author" gorm:"size:255;nullable;"`
	GitCommitDate   *time.Time `json:"git_commit_date"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index:idx_namespace_id_deleted_at,priority:2;"`

	Namespace Namespace
}
