package models

import (
	"encoding/json"
	"sort"
	"time"

	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars-client/v4/mars"
	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"
)

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

func (g GitProject) PrettyYaml() string {
	cfg := mars.Config{}
	json.Unmarshal([]byte(g.GlobalConfig), &cfg)
	clone := proto.Clone(&cfg).(*mars.Config)
	var v map[string]any
	yaml.Unmarshal([]byte(cfg.ValuesYaml), &v)
	var data = struct {
		ConfigFile       string          `yaml:"config_file"`
		ConfigFileValues string          `yaml:"config_file_values"`
		ConfigField      string          `yaml:"config_field"`
		IsSimpleEnv      bool            `yaml:"is_simple_env"`
		ConfigFileType   string          `yaml:"config_file_type"`
		LocalChartPath   string          `yaml:"local_chart_path"`
		Branches         []string        `yaml:"branches"`
		ValuesYaml       map[string]any  `yaml:"values_yaml"`
		Elements         []*mars.Element `yaml:"elements"`
		DisplayName      string          `yaml:"display_name"`
	}{
		ConfigFile:       clone.ConfigFile,
		ConfigFileValues: clone.ConfigFileValues,
		ConfigField:      clone.ConfigField,
		IsSimpleEnv:      clone.IsSimpleEnv,
		ConfigFileType:   clone.ConfigFileType,
		LocalChartPath:   clone.LocalChartPath,
		Branches:         clone.Branches,
		ValuesYaml:       v,
		Elements:         clone.Elements,
		DisplayName:      clone.DisplayName,
	}

	out, _ := yaml.Marshal(data)
	return string(out)
}

type sortedElements []*mars.Element

func (s sortedElements) Len() int {
	return len(s)
}

func (s sortedElements) Less(i, j int) bool {
	return s[i].Order < s[j].Order
}

func (s sortedElements) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (g *GitProject) GlobalMarsConfig() *mars.Config {
	if g.GlobalConfig == "" {
		return &mars.Config{}
	}

	var c = &mars.Config{}
	json.Unmarshal([]byte(g.GlobalConfig), c)
	sort.Sort(sortedElements(c.Elements))
	return c
}

func (g *GitProject) ProtoTransform() *types.GitProjectModel {
	return &types.GitProjectModel{
		Id:            int64(g.ID),
		DefaultBranch: g.DefaultBranch,
		Name:          g.Name,
		GitProjectId:  int64(g.GitProjectId),
		Enabled:       g.Enabled,
		GlobalEnabled: g.GlobalEnabled,
		GlobalConfig:  g.GlobalConfig,
		CreatedAt:     date.ToRFC3339DatetimeString(&g.CreatedAt),
		UpdatedAt:     date.ToRFC3339DatetimeString(&g.UpdatedAt),
		DeletedAt:     date.ToRFC3339DatetimeString(&g.DeletedAt.Time),
	}
}
