package repo

import (
	"context"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/filters"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
)

type Repo interface {
	List(ctx context.Context, in *ListRepoRequest) ([]*ent.Repo, *pagination.Pagination, error)
	Show(ctx context.Context, id int) (*ent.Repo, error)
	ToggleEnabled(ctx context.Context, id int, enabled bool) (*ent.Repo, error)
	Create(ctx context.Context, in *CreateRepoInput) (*ent.Repo, error)
}

var _ Repo = (*repo)(nil)

type repo struct {
	logger mlog.Logger
	db     *ent.Client
}

func NewRepo(logger mlog.Logger, data *data.Data) Repo {
	return &repo{logger: logger, db: data.DB}
}

func (r *repo) ToggleEnabled(ctx context.Context, id int, enabled bool) (*ent.Repo, error) {
	get, err := r.db.Repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return get.Update().SetEnabled(enabled).Save(ctx)
}

type ListRepoRequest struct {
	Page, PageSize int64
	Enabled        *bool
	OrderByIDDesc  *bool
}

func (r *repo) List(ctx context.Context, in *ListRepoRequest) ([]*ent.Repo, *pagination.Pagination, error) {
	query := r.db.Repo.Query().Where(
		filters.IfOrderByIDDesc(in.OrderByIDDesc),
		filters.IfEnabled(in.Enabled),
	)
	all, err := query.Clone().
		Offset(pagination.GetPageOffset(in.Page, in.PageSize)).
		Limit(int(in.PageSize)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	count := query.Clone().CountX(ctx)

	return all, &pagination.Pagination{
		Page:     in.Page,
		PageSize: in.PageSize,
		Count:    int64(count),
	}, nil
}

type CreateRepoInput struct {
	Name          string
	Enabled       bool
	GitProjectID  *int64
	MarsConfig    *mars.Config
	DefaultBranch *string
}

func (r *repo) Create(ctx context.Context, in *CreateRepoInput) (*ent.Repo, error) {
	return r.db.Repo.Create().
		SetName(in.Name).
		SetNillableDefaultBranch(in.DefaultBranch).
		SetEnabled(in.Enabled).
		SetNillableGitProjectID(in.GitProjectID).
		SetMarsConfig(in.MarsConfig).
		Save(ctx)
}

func (r *repo) Show(ctx context.Context, id int) (*ent.Repo, error) {
	return r.db.Repo.Get(ctx, id)
}
