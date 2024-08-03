package services

import (
	"context"

	reposerver "github.com/duc-cnzj/mars/api/v4/repo"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
	"github.com/samber/lo"
)

var _ reposerver.RepoServer = (*repoSvc)(nil)

type repoSvc struct {
	guest

	gitRepo  repo.GitRepo
	logger   mlog.Logger
	repoRepo repo.Repo
	reposerver.UnimplementedRepoServer
}

func NewRepoSvc(logger mlog.Logger, gitRepo repo.GitRepo, repoRepo repo.Repo) reposerver.RepoServer {
	return &repoSvc{logger: logger, repoRepo: repoRepo, gitRepo: gitRepo}
}

func (r *repoSvc) List(ctx context.Context, request *reposerver.ListRequest) (*reposerver.ListResponse, error) {
	pagination.InitByDefault(&request.Page, &request.PageSize)
	list, pag, err := r.repoRepo.List(ctx, &repo.ListRepoRequest{
		Page:          request.Page,
		PageSize:      request.PageSize,
		Enabled:       request.Enabled,
		OrderByIDDesc: lo.ToPtr(true),
	})
	if err != nil {
		return nil, err
	}
	return &reposerver.ListResponse{
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Count:    pag.Count,
		Items:    serialize.Serialize(list, transformer.FromRepo),
	}, nil
}

func (r *repoSvc) Create(ctx context.Context, req *reposerver.CreateRequest) (*reposerver.CreateResponse, error) {
	var defaultBranch *string

	if req.GitProjectId != nil {
		project, err := r.gitRepo.GetByProjectID(ctx, int(*req.GitProjectId))
		if err != nil {
			return nil, err
		}
		defaultBranch = lo.ToPtr(project.GetDefaultBranch())
	}
	create, err := r.repoRepo.Create(ctx, &repo.CreateRepoInput{
		Name:          req.Name,
		Enabled:       req.Enabled,
		GitProjectID:  req.GitProjectId,
		MarsConfig:    req.MarsConfig,
		DefaultBranch: defaultBranch,
	})
	if err != nil {
		return nil, err
	}
	return &reposerver.CreateResponse{
		Item: transformer.FromRepo(create),
	}, nil
}

func (r *repoSvc) Show(ctx context.Context, request *reposerver.ShowRequest) (*reposerver.ShowResponse, error) {
	show, err := r.repoRepo.Show(ctx, int(request.Id))
	if err != nil {
		return nil, err
	}
	return &reposerver.ShowResponse{
		Item: transformer.FromRepo(show),
	}, nil
}

func (r *repoSvc) ToggleEnabled(ctx context.Context, request *reposerver.ToggleEnabledRequest) (*reposerver.ToggleEnabledResponse, error) {
	toggle, err := r.repoRepo.ToggleEnabled(ctx, int(request.Id), request.Enabled)
	if err != nil {
		return nil, err
	}
	return &reposerver.ToggleEnabledResponse{
		Item: transformer.FromRepo(toggle),
	}, nil
}
