package repo

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/changelog"
	"github.com/duc-cnzj/mars/v4/internal/filters"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
	"github.com/samber/lo"
)

type Changelog struct {
	ID               int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	Version          int
	Username         string
	Manifest         []string
	Config           string
	ConfigType       string
	GitBranch        string
	GitCommit        string
	DockerImage      []string
	EnvValues        []*types.KeyValue
	ExtraValues      []*types.ExtraValue
	FinalExtraValues []string
	GitCommitWebURL  string
	GitCommitTitle   string
	GitCommitAuthor  string
	GitCommitDate    *time.Time
	ConfigChanged    bool
	ProjectID        int
	GitProjectID     int

	Project    *Project
	GitProject *GitProject
}

func ToChangeLog(c *ent.Changelog) *Changelog {
	if c == nil {
		return nil
	}
	return &Changelog{
		ID:               c.ID,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
		DeletedAt:        c.DeletedAt,
		Version:          c.Version,
		Username:         c.Username,
		Manifest:         c.Manifest,
		Config:           c.Config,
		ConfigType:       c.ConfigType,
		GitBranch:        c.GitBranch,
		GitCommit:        c.GitCommit,
		DockerImage:      c.DockerImage,
		EnvValues:        c.EnvValues,
		ExtraValues:      c.ExtraValues,
		FinalExtraValues: c.FinalExtraValues,
		GitCommitWebURL:  c.GitCommitWebURL,
		GitCommitTitle:   c.GitCommitTitle,
		GitCommitAuthor:  c.GitCommitAuthor,
		GitCommitDate:    c.GitCommitDate,
		ConfigChanged:    c.ConfigChanged,
		ProjectID:        c.ProjectID,
		GitProjectID:     c.GitProjectID,
		Project:          ToProject(c.Edges.Project),
		GitProject:       ToGitProject(c.Edges.GitProject),
	}
}

type ChangelogRepo interface {
	Show(ctx context.Context, input *ShowChangeLogInput) ([]*Changelog, error)
	Create(ctx context.Context, input *CreateChangeLogInput) (*Changelog, error)
	FindLastChangeByProjectID(ctx context.Context, projectID int) (*Changelog, error)
}

var _ ChangelogRepo = (*changelogRepo)(nil)

type changelogRepo struct {
	logger mlog.Logger
	data   data.Data
}

func NewChangelogRepo(logger mlog.Logger, data data.Data) ChangelogRepo {
	return &changelogRepo{logger: logger, data: data}
}

type CreateChangeLogInput struct {
	Version          int
	Username         string
	Manifest         []string
	Config           string
	ConfigType       string
	GitBranch        string
	GitCommit        string
	DockerImage      []string
	EnvValues        []*types.KeyValue
	ExtraValues      []*types.ExtraValue
	FinalExtraValues []string
	GitCommitWebURL  string
	GitCommitTitle   string
	GitCommitAuthor  string
	GitCommitDate    *time.Time
	ProjectID        int
	ConfigChanged    bool
	GitProjectID     int
}

func (c *changelogRepo) Create(ctx context.Context, input *CreateChangeLogInput) (*Changelog, error) {
	var db = c.data.DB()
	save, err := db.Changelog.Create().
		SetVersion(input.Version).
		SetUsername(input.Username).
		SetManifest(input.Manifest).
		SetConfig(input.Config).
		SetConfigType(input.ConfigType).
		SetGitBranch(input.GitBranch).
		SetGitCommit(input.GitCommit).
		SetDockerImage(input.DockerImage).
		SetEnvValues(input.EnvValues).
		SetExtraValues(input.ExtraValues).
		SetFinalExtraValues(input.FinalExtraValues).
		SetGitCommitWebURL(input.GitCommitWebURL).
		SetGitCommitTitle(input.GitCommitTitle).
		SetGitCommitAuthor(input.GitCommitAuthor).
		SetNillableGitCommitDate(input.GitCommitDate).
		SetConfigChanged(input.ConfigChanged).
		SetProjectID(input.ProjectID).
		SetGitProjectID(input.GitProjectID).
		Save(ctx)
	return ToChangeLog(save), err
}

type ShowChangeLogInput struct {
	OnlyChanged        bool
	ProjectID          int
	OrderByVersionDesc *bool
}

func (c *changelogRepo) Show(ctx context.Context, input *ShowChangeLogInput) ([]*Changelog, error) {
	var db = c.data.DB()
	all, err := db.Changelog.Query().
		WithProject(func(query *ent.ProjectQuery) {
			query.Where(filters.IfBool("config_changed")(func() *bool {
				if input.OnlyChanged {
					return lo.ToPtr(true)
				}
				return nil
			}()))
		}).
		Where(
			changelog.ProjectID(input.ProjectID),
			filters.IfOrderByDesc("version")(input.OrderByVersionDesc),
		).
		Limit(5).
		All(ctx)
	return serialize.Serialize(all, ToChangeLog), err
}

func (c *changelogRepo) FindLastChangeByProjectID(ctx context.Context, projectID int) (*Changelog, error) {
	var db = c.data.DB()
	first, err := db.Changelog.Query().Where(changelog.ProjectID(projectID)).Order(ent.Desc(changelog.FieldID)).First(ctx)
	return ToChangeLog(first), err
}
