package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/duc-cnzj/mars/pkg/mars"
	"gopkg.in/yaml.v2"

	"gorm.io/gorm"
)

var emptyConfigString string

func init() {
	sb := strings.Builder{}
	yaml.NewEncoder(&sb).Encode(&mars.Config{})
	emptyConfigString = sb.String()
}

type GitlabProject struct {
	ID int `json:"id" gorm:"primaryKey;"`

	DefaultBranch   string `json:"default_branch" gorm:"type:varchar(255);not null;default:'';"`
	Name            string `json:"name" gorm:"type:varchar(255);not null;default:'';"`
	GitlabProjectId int    `json:"gitlab_project_id" gorm:"not null;type:integer;"`
	Enabled         bool   `json:"enabled" gorm:"not null;default:false;"`
	GlobalEnabled   bool   `json:"global_enabled" gorm:"not null;default:false;"`
	GlobalConfig    string `json:"global_config" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

func (g *GitlabProject) GlobalMarsConfig() *mars.Config {
	if g.GlobalConfig == "" {
		return &mars.Config{}
	}

	var c = &mars.Config{}
	json.Unmarshal([]byte(g.GlobalConfig), c)
	return c
}
