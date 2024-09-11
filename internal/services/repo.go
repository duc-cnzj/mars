package services

import (
	"context"
	"fmt"
	"strings"

	reposerver "github.com/duc-cnzj/mars/api/v5/repo"
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/transformer"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/duc-cnzj/mars/v5/internal/util/serialize"
	yaml2 "github.com/duc-cnzj/mars/v5/internal/util/yaml"
	"github.com/samber/lo"
)

var _ reposerver.RepoServer = (*repoSvc)(nil)

type repoSvc struct {
	gitRepo   repo.GitRepo
	logger    mlog.Logger
	repoRepo  repo.RepoRepo
	eventRepo repo.EventRepo

	reposerver.UnimplementedRepoServer
}

func NewRepoSvc(logger mlog.Logger, eventRepo repo.EventRepo, gitRepo repo.GitRepo, repoRepo repo.RepoRepo) reposerver.RepoServer {
	return &repoSvc{
		gitRepo:   gitRepo,
		logger:    logger.WithModule("services/repo"),
		repoRepo:  repoRepo,
		eventRepo: eventRepo,
	}
}

func (r *repoSvc) List(ctx context.Context, request *reposerver.ListRequest) (*reposerver.ListResponse, error) {
	page, pageSize := pagination.InitByDefault(request.Page, request.PageSize)

	list, pag, err := r.repoRepo.List(ctx, &repo.ListRepoRequest{
		Page:          page,
		PageSize:      pageSize,
		Enabled:       request.Enabled,
		OrderByIDDesc: lo.ToPtr(true),
		Name:          request.Name,
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
	create, err := r.repoRepo.Create(ctx, &repo.CreateRepoInput{
		Name:         req.Name,
		Enabled:      true,
		NeedGitRepo:  req.NeedGitRepo,
		GitProjectID: req.GitProjectId,
		MarsConfig:   req.MarsConfig,
		Description:  req.Description,
	})
	if err != nil {
		return nil, err
	}
	r.eventRepo.AuditLogWithRequest(
		types.EventActionType_Create,
		MustGetUser(ctx).Name,
		fmt.Sprintf("创建仓库: %d: %s", create.ID, create.Name),
		req,
	)
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

func (r *repoSvc) Update(ctx context.Context, req *reposerver.UpdateRequest) (*reposerver.UpdateResponse, error) {
	current, err := r.repoRepo.Show(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	create, err := r.repoRepo.Update(ctx, &repo.UpdateRepoInput{
		ID:           req.Id,
		Name:         req.Name,
		NeedGitRepo:  req.NeedGitRepo,
		GitProjectID: req.GitProjectId,
		MarsConfig:   req.MarsConfig,
		Description:  req.Description,
	})
	if err != nil {
		return nil, err
	}
	old, _ := yaml2.PrettyMarshal(current)
	out, _ := yaml2.PrettyMarshal(create)
	r.eventRepo.AuditLogWithChange(
		types.EventActionType_Update,
		MustGetUser(ctx).Name,
		fmt.Sprintf("更新仓库: %d: %s", create.ID, create.Name),
		&repo.StringYamlPrettier{Str: string(old)},
		&repo.StringYamlPrettier{Str: string(out)},
	)

	return &reposerver.UpdateResponse{
		Item: transformer.FromRepo(create),
	}, nil
}

func (r *repoSvc) ToggleEnabled(ctx context.Context, request *reposerver.ToggleEnabledRequest) (*reposerver.ToggleEnabledResponse, error) {
	toggle, err := r.repoRepo.ToggleEnabled(ctx, int(request.Id), request.Enabled)
	if err != nil {
		return nil, err
	}

	status := "禁用"
	if request.Enabled {
		status = "启用"
	}
	r.eventRepo.AuditLogWithRequest(
		types.EventActionType_Update,
		MustGetUser(ctx).Name,
		fmt.Sprintf("[repo 状态变动]: %s 仓库 %s", status, toggle.Name),
		request,
	)

	return &reposerver.ToggleEnabledResponse{
		Item: transformer.FromRepo(toggle),
	}, nil
}

func (r *repoSvc) Delete(ctx context.Context, request *reposerver.DeleteRequest) (*reposerver.DeleteResponse, error) {

	if err := r.repoRepo.Delete(ctx, int(request.Id)); err != nil {
		return nil, err
	}

	r.eventRepo.AuditLogWithRequest(
		types.EventActionType_Delete,
		MustGetUser(ctx).Name,
		fmt.Sprintf("删除 repo: %d", request.Id),
		request,
	)

	return &reposerver.DeleteResponse{}, nil
}

func (r *repoSvc) Clone(ctx context.Context, req *reposerver.CloneRequest) (*reposerver.CloneResponse, error) {
	clone, err := r.repoRepo.Clone(ctx, &repo.CloneRepoInput{
		ID:   int(req.Id),
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}
	show, _ := r.repoRepo.Show(ctx, int(req.Id))

	r.eventRepo.AuditLogWithRequest(
		types.EventActionType_Create,
		MustGetUser(ctx).Name,
		fmt.Sprintf("克隆 repo: (id: %d, name: %s) -> (id: %d, name: %s)", show.ID, show.Name, clone.ID, clone.Name),
		req,
	)

	return &reposerver.CloneResponse{
		Item: transformer.FromRepo(clone),
	}, nil
}

func (r *repoSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if strings.EqualFold(fullMethodName, "/repo.Repo/List") {
		return ctx, nil
	}
	if strings.EqualFold(fullMethodName, "/repo.Repo/Show") {
		return ctx, nil
	}
	if !MustGetUser(ctx).IsAdmin() {
		return nil, ErrorPermissionDenied
	}

	return ctx, nil
}
