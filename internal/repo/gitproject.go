package repo

import (
	"context"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/gitproject"
	"github.com/duc-cnzj/mars/v4/internal/filters"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type GitProjectRepo interface {
	All(ctx context.Context, input *AllGitProjectInput) ([]*ent.GitProject, error)
	GetByID(ctx context.Context, id int) (*ent.GitProject, error)
	Upsert(ctx context.Context, input *UpsertGitProjectInput) error
	DisabledByProjectID(ctx context.Context, projID int) (*ent.GitProject, error)
	GetByProjectID(ctx context.Context, projID int) (*ent.GitProject, error)
	ToggleEnabled(ctx context.Context, projID int) (*ent.GitProject, error)
	UpdateGlobalConfig(ctx context.Context, projID int, cfg *mars.Config) (*ent.GitProject, error)
}

var _ GitProjectRepo = (*gitProjectRepo)(nil)

type gitProjectRepo struct {
	logger mlog.Logger
	db     *ent.Client
}

func NewGitProjectRepo(logger mlog.Logger, data *data.Data) GitProjectRepo {
	return &gitProjectRepo{logger: logger, db: data.DB}
}

type AllGitProjectInput struct {
	Enabled *bool
}

func (g *gitProjectRepo) All(ctx context.Context, input *AllGitProjectInput) ([]*ent.GitProject, error) {
	return g.db.GitProject.Query().Where(filters.IfBool("id")(input.Enabled)).All(ctx)
}

type UpsertGitProjectInput struct {
	DefaultBranch string
	Name          string
	GitProjectId  int
	Enabled       bool
}

func (g *gitProjectRepo) Upsert(ctx context.Context, input *UpsertGitProjectInput) error {
	return g.db.GitProject.Create().SetName(input.Name).
		SetGitProjectID(input.GitProjectId).
		SetDefaultBranch(input.DefaultBranch).
		SetEnabled(input.Enabled).
		OnConflict().
		UpdateNewValues().
		Exec(ctx)
}

func (g *gitProjectRepo) GetByProjectID(ctx context.Context, projID int) (project *ent.GitProject, err error) {
	return g.db.GitProject.Query().Where(gitproject.GitProjectID(projID)).Only(ctx)
}

func (g *gitProjectRepo) GetByID(ctx context.Context, id int) (project *ent.GitProject, err error) {
	return g.db.GitProject.Query().Where(gitproject.ID(id)).Only(ctx)
}

func (g *gitProjectRepo) UpdateGlobalConfig(ctx context.Context, projID int, cfg *mars.Config) (*ent.GitProject, error) {
	first, err := g.db.GitProject.Query().Where(gitproject.GitProjectID(projID)).First(ctx)
	if err != nil {
		return nil, err
	}
	return first.Update().SetGlobalConfig(cfg).Save(ctx)
}

func (g *gitProjectRepo) ToggleEnabled(ctx context.Context, projID int) (*ent.GitProject, error) {
	only, err := g.db.GitProject.Query().Where(gitproject.GitProjectID(projID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return only.Update().SetGlobalEnabled(!only.Enabled).Save(ctx)
}

func (g *gitProjectRepo) DisabledByProjectID(ctx context.Context, projID int) (*ent.GitProject, error) {
	first, err := g.db.GitProject.Query().Where(gitproject.GitProjectID(projID)).First(ctx)
	if err != nil {
		return nil, err
	}
	return first.Update().SetEnabled(false).Save(ctx)
}
