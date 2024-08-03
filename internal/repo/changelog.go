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
	"github.com/samber/lo"
)

type ChangelogRepo interface {
	Show(ctx context.Context, input *ShowChangeLogInput) ([]*ent.Changelog, error)
	Create(ctx context.Context, input *CreateChangeLogInput) (*ent.Changelog, error)
	FindLastChangeByProjectID(ctx context.Context, projectID int) (*ent.Changelog, error)
}

var _ ChangelogRepo = (*changelogRepo)(nil)

type changelogRepo struct {
	logger mlog.Logger
	db     *ent.Client
}

func NewChangelogRepo(logger mlog.Logger, data *data.Data) ChangelogRepo {
	return &changelogRepo{logger: logger, db: data.DB}
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

func (c *changelogRepo) Create(ctx context.Context, input *CreateChangeLogInput) (*ent.Changelog, error) {
	return c.db.Changelog.Create().
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
}

type ShowChangeLogInput struct {
	OnlyChanged        bool
	ProjectID          int
	OrderByVersionDesc *bool
}

func (c *changelogRepo) Show(ctx context.Context, input *ShowChangeLogInput) ([]*ent.Changelog, error) {
	return c.db.Changelog.Query().
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
}

func (c *changelogRepo) FindLastChangeByProjectID(ctx context.Context, projectID int) (*ent.Changelog, error) {
	return c.db.Changelog.Query().Where(changelog.ProjectID(projectID)).Order(ent.Desc(changelog.FieldID)).First(ctx)
}
