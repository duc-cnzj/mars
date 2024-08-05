package repo

import (
	"context"
	"sort"
	"time"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/gitproject"
	"github.com/duc-cnzj/mars/v4/internal/filters"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
	"gopkg.in/yaml.v3"
)

type GitProject struct {
	ID            int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	Name          string
	DefaultBranch string
	GitProjectID  int
	Enabled       bool
	GlobalEnabled bool
	GlobalConfig  *mars.Config
}

func (g *GitProject) PrettyYaml() string {
	cfg := g.GlobalConfig
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
		ConfigFile:       cfg.ConfigFile,
		ConfigFileValues: cfg.ConfigFileValues,
		ConfigField:      cfg.ConfigField,
		IsSimpleEnv:      cfg.IsSimpleEnv,
		ConfigFileType:   cfg.ConfigFileType,
		LocalChartPath:   cfg.LocalChartPath,
		Branches:         cfg.Branches,
		ValuesYaml:       v,
		Elements:         cfg.Elements,
		DisplayName:      cfg.DisplayName,
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
	if g.GlobalConfig == nil {
		return &mars.Config{}
	}

	sort.Sort(sortedElements(g.GlobalConfig.Elements))
	return g.GlobalConfig
}

func ToGitProject(g *ent.GitProject) *GitProject {
	if g == nil {
		return nil
	}
	return &GitProject{
		ID:            g.ID,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
		DeletedAt:     g.DeletedAt,
		Name:          g.Name,
		DefaultBranch: g.DefaultBranch,
		GitProjectID:  g.GitProjectID,
		Enabled:       g.Enabled,
		GlobalEnabled: g.GlobalEnabled,
		GlobalConfig:  g.GlobalConfig,
	}
}

type GitProjectRepo interface {
	All(ctx context.Context, input *AllGitProjectInput) ([]*GitProject, error)
	GetByID(ctx context.Context, id int) (*GitProject, error)
	Upsert(ctx context.Context, input *UpsertGitProjectInput) error
	DisabledByProjectID(ctx context.Context, projID int) (*GitProject, error)
	GetByProjectID(ctx context.Context, projID int) (*GitProject, error)
	ToggleEnabled(ctx context.Context, projID int) (*GitProject, error)
	UpdateGlobalConfig(ctx context.Context, projID int, cfg *mars.Config) (*GitProject, error)
}

var _ GitProjectRepo = (*gitProjectRepo)(nil)

type gitProjectRepo struct {
	logger mlog.Logger
	data   data.Data
}

func NewGitProjectRepo(logger mlog.Logger, data data.Data) GitProjectRepo {
	return &gitProjectRepo{logger: logger, data: data}
}

type AllGitProjectInput struct {
	Enabled *bool
}

func (g *gitProjectRepo) All(ctx context.Context, input *AllGitProjectInput) ([]*GitProject, error) {
	var db = g.data.DB()
	all, err := db.GitProject.Query().Where(filters.IfBool("id")(input.Enabled)).All(ctx)
	return serialize.Serialize(all, ToGitProject), err
}

type UpsertGitProjectInput struct {
	DefaultBranch string
	Name          string
	GitProjectId  int
	Enabled       bool
}

func (g *gitProjectRepo) Upsert(ctx context.Context, input *UpsertGitProjectInput) error {
	var db = g.data.DB()
	return db.GitProject.Create().SetName(input.Name).
		SetGitProjectID(input.GitProjectId).
		SetDefaultBranch(input.DefaultBranch).
		SetEnabled(input.Enabled).
		OnConflict().
		UpdateNewValues().
		Exec(ctx)
}

func (g *gitProjectRepo) GetByProjectID(ctx context.Context, projID int) (project *GitProject, err error) {
	var db = g.data.DB()
	only, err := db.GitProject.Query().Where(gitproject.GitProjectID(projID)).Only(ctx)
	return ToGitProject(only), err
}

func (g *gitProjectRepo) GetByID(ctx context.Context, id int) (project *GitProject, err error) {
	var db = g.data.DB()
	only, err := db.GitProject.Query().Where(gitproject.ID(id)).Only(ctx)
	return ToGitProject(only), err
}

func (g *gitProjectRepo) UpdateGlobalConfig(ctx context.Context, projID int, cfg *mars.Config) (*GitProject, error) {
	var db = g.data.DB()
	first, err := db.GitProject.Query().Where(gitproject.GitProjectID(projID)).First(ctx)
	if err != nil {
		return nil, err
	}
	save, err := first.Update().SetGlobalConfig(cfg).Save(ctx)
	return ToGitProject(save), err
}

func (g *gitProjectRepo) ToggleEnabled(ctx context.Context, projID int) (*GitProject, error) {
	var db = g.data.DB()
	only, err := db.GitProject.Query().Where(gitproject.GitProjectID(projID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	save, err := only.Update().SetGlobalEnabled(!only.Enabled).Save(ctx)
	return ToGitProject(save), err
}

func (g *gitProjectRepo) DisabledByProjectID(ctx context.Context, projID int) (*GitProject, error) {
	var db = g.data.DB()
	first, err := db.GitProject.Query().Where(gitproject.GitProjectID(projID)).First(ctx)
	if err != nil {
		return nil, err
	}
	save, err := first.Update().SetEnabled(false).Save(ctx)
	return ToGitProject(save), err
}
