package models

import (
	"bytes"
	"encoding/json"
	"time"

	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/pkg/mars"
)

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

func (g GitlabProject) PrettyYaml() string {
	cfg := mars.Config{}
	json.Unmarshal([]byte(g.GlobalConfig), &cfg)
	clone := proto.Clone(&cfg).(*mars.Config)
	var v map[string]interface{}
	yaml.Unmarshal([]byte(cfg.ValuesYaml), &v)
	var data = struct {
		ConfigFile       string                 `yaml:"config_file"`
		ConfigFileValues string                 `yaml:"config_file_values"`
		ConfigField      string                 `yaml:"config_field"`
		IsSimpleEnv      bool                   `yaml:"is_simple_env"`
		ConfigFileType   string                 `yaml:"config_file_type"`
		LocalChartPath   string                 `yaml:"local_chart_path"`
		Branches         []string               `yaml:"branches"`
		ValuesYaml       map[string]interface{} `yaml:"values_yaml"`
	}{
		ConfigFile:       clone.ConfigFile,
		ConfigFileValues: clone.ConfigFileValues,
		ConfigField:      clone.ConfigField,
		IsSimpleEnv:      clone.IsSimpleEnv,
		ConfigFileType:   clone.ConfigFileType,
		LocalChartPath:   clone.LocalChartPath,
		Branches:         clone.Branches,
		ValuesYaml:       v,
	}
	bf := bytes.Buffer{}
	yaml.NewEncoder(&bf).Encode(data)
	return bf.String()
}

func (g *GitlabProject) GlobalMarsConfig() *mars.Config {
	if g.GlobalConfig == "" {
		return &mars.Config{}
	}

	var c = &mars.Config{}
	json.Unmarshal([]byte(g.GlobalConfig), c)
	return c
}
