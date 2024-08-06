package repo

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/filters"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
	"github.com/samber/lo"
)

type Repo struct {
	ID             int          `json:"id"`
	CreatedAt      time.Time    `json:"-"`
	UpdatedAt      time.Time    `json:"-"`
	DeletedAt      *time.Time   `json:"-"`
	Name           string       `json:"name"`
	DefaultBranch  string       `json:"default_branch"`
	GitProjectName string       `json:"git_project_name"`
	GitProjectID   int32        `json:"git_project_id"`
	Enabled        bool         `json:"enabled"`
	NeedGitRepo    bool         `json:"need_git_repo"`
	MarsConfig     *mars.Config `json:"mars_config"`
}

func (r *Repo) GetMarsConfig() (cfg *mars.Config) {
	cfg = r.MarsConfig
	if r.MarsConfig == nil {
		cfg = &mars.Config{}
	}
	return
}

type RepoImp interface {
	All(ctx context.Context, in *AllRepoRequest) ([]*Repo, error)
	List(ctx context.Context, in *ListRepoRequest) ([]*Repo, *pagination.Pagination, error)
	Show(ctx context.Context, id int) (*Repo, error)
	ToggleEnabled(ctx context.Context, id int, enabled bool) (*Repo, error)
	Create(ctx context.Context, in *CreateRepoInput) (*Repo, error)
	Update(ctx context.Context, in *UpdateRepoInput) (*Repo, error)
}

var _ RepoImp = (*repo)(nil)

type repo struct {
	logger mlog.Logger
	data   data.Data

	gitRepo GitRepo
}

func (r *repo) All(ctx context.Context, in *AllRepoRequest) ([]*Repo, error) {
	query := r.data.DB().Repo.Query().Where(
		filters.IfEnabled(in.Enabled),
	)
	all, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	return serialize.Serialize(all, ToRepo), nil
}

func NewRepo(logger mlog.Logger, data data.Data, gitRepo GitRepo) RepoImp {
	return &repo{logger: logger, data: data, gitRepo: gitRepo}
}

type AllRepoRequest struct {
	Enabled       *bool
	OrderByIDDesc *bool
}
type ListRepoRequest struct {
	Page, PageSize int32
	Enabled        *bool
	OrderByIDDesc  *bool
	Name           string
}

func (r *repo) List(ctx context.Context, in *ListRepoRequest) ([]*Repo, *pagination.Pagination, error) {
	query := r.data.DB().Repo.Query().Where(
		filters.IfOrderByIDDesc(in.OrderByIDDesc),
		filters.IfEnabled(in.Enabled),
		filters.IfNameLike(in.Name),
	)
	all, err := query.Clone().
		Offset(pagination.GetPageOffset(in.Page, in.PageSize)).
		Limit(int(in.PageSize)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	count := query.Clone().CountX(ctx)

	return serialize.Serialize(all, ToRepo), pagination.NewPagination(in.Page, in.PageSize, count), nil
}

type CreateRepoInput struct {
	Name         string
	Enabled      bool
	NeedGitRepo  bool
	GitProjectID *int32
	MarsConfig   *mars.Config
}

func (r *repo) Create(ctx context.Context, in *CreateRepoInput) (*Repo, error) {
	var (
		projName      *string
		defaultBranch *string
		err           error
	)
	if !in.NeedGitRepo {
		in.GitProjectID = nil
	} else {
		projName, defaultBranch, err = r.GetProjNameAndBranch(ctx, int(*in.GitProjectID))
		if err != nil {
			return nil, err
		}
	}
	cr := r.data.DB().Repo.Create().
		SetName(in.Name).
		SetNeedGitRepo(in.NeedGitRepo).
		SetNillableGitProjectName(projName).
		SetNillableDefaultBranch(defaultBranch).
		SetEnabled(in.Enabled).
		SetMarsConfig(in.MarsConfig)
	if in.NeedGitRepo {
		cr.SetNillableGitProjectID(in.GitProjectID)
	}
	save, err := cr.Save(ctx)
	return ToRepo(save), err
}

func (r *repo) Show(ctx context.Context, id int) (*Repo, error) {
	get, err := r.data.DB().Repo.Get(ctx, id)
	return ToRepo(get), err
}

type UpdateRepoInput struct {
	ID           int32
	Name         string
	NeedGitRepo  bool
	GitProjectID *int32
	MarsConfig   *mars.Config
}

func (r *repo) Update(ctx context.Context, in *UpdateRepoInput) (*Repo, error) {
	var (
		projName      *string
		defaultBranch *string
		err           error
	)
	if in.NeedGitRepo {
		projName, defaultBranch, err = r.GetProjNameAndBranch(ctx, int(*in.GitProjectID))
		if err != nil {
			return nil, err
		}
	}

	up := r.data.DB().Repo.
		UpdateOneID(int(in.ID)).
		SetName(in.Name).
		SetNeedGitRepo(in.NeedGitRepo).
		SetNillableGitProjectID(in.GitProjectID).
		SetNillableGitProjectName(projName).
		SetNillableDefaultBranch(defaultBranch).
		SetMarsConfig(in.MarsConfig)
	if !in.NeedGitRepo {
		up.ClearGitProjectID().ClearGitProjectName().ClearDefaultBranch()
	}
	save, err := up.Save(ctx)
	return ToRepo(save), err
}

func (r *repo) GetProjNameAndBranch(ctx context.Context, projID int) (*string, *string, error) {
	var (
		defaultBranch *string
		projName      *string
	)

	project, err := r.gitRepo.GetByProjectID(ctx, projID)
	if err != nil {
		return nil, nil, err
	}
	defaultBranch = lo.ToPtr(project.GetDefaultBranch())
	projName = lo.ToPtr(project.GetName())
	return projName, defaultBranch, nil
}

func (r *repo) ToggleEnabled(ctx context.Context, id int, enabled bool) (*Repo, error) {
	get, err := r.data.DB().Repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	save, err := get.Update().SetEnabled(enabled).Save(ctx)
	return ToRepo(save), err
}

func ToRepo(data *ent.Repo) *Repo {
	if data == nil {
		return nil
	}
	return &Repo{
		ID:             data.ID,
		CreatedAt:      data.CreatedAt,
		UpdatedAt:      data.UpdatedAt,
		DeletedAt:      data.DeletedAt,
		Name:           data.Name,
		DefaultBranch:  data.DefaultBranch,
		GitProjectName: data.GitProjectName,
		GitProjectID:   data.GitProjectID,
		Enabled:        data.Enabled,
		NeedGitRepo:    data.NeedGitRepo,
		MarsConfig:     data.MarsConfig,
	}

}
