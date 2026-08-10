package services

import (
	"context"
	"fmt"

	reposerver "github.com/duc-cnzj/mars/api/v6/proto/repo"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/duc-cnzj/mars/v6/internal/util/yaml"
	"github.com/samber/lo"
)

var _ reposerver.RepoServer = (*repoSvc)(nil)

// repoSvc 是 repo.RepoServer 的 gRPC 实现：管理仓库源全生命周期（增删改查/启停/克隆），
// 经 access 校验访问权限，由 NewRepoSvc 构造。
type repoSvc struct {
	logger    mlog.Logger
	repoBiz   biz.RepoBiz
	eventBiz  biz.EventBiz
	accessBiz biz.AccessBiz

	reposerver.UnimplementedRepoServer
}

// RepoSvcDeps 收口 NewRepoSvc 的构造依赖，由 wire 按字段注入。
type RepoSvcDeps struct {
	Logger    mlog.Logger
	EventBiz  biz.EventBiz
	RepoBiz   biz.RepoBiz
	AccessBiz biz.AccessBiz
}

// NewRepoSvc 收口 repo 服务的构造依赖，由 wire 按字段注入。
func NewRepoSvc(deps RepoSvcDeps) reposerver.RepoServer {
	return &repoSvc{
		logger:    deps.Logger.WithModule("services/repo"),
		repoBiz:   deps.RepoBiz,
		eventBiz:  deps.EventBiz,
		accessBiz: deps.AccessBiz,
	}
}

// List 分页列出仓库，支持按名称与启用状态过滤，按 id 倒序返回。
func (r *repoSvc) List(ctx context.Context, request *reposerver.ListRequest) (*reposerver.ListResponse, error) {
	page, pageSize := pagination.InitByDefault(request.Page, request.PageSize)

	list, pag, err := r.repoBiz.List(ctx, &biz.ListRepoRequest{
		Page:          page,
		PageSize:      pageSize,
		Enabled:       request.Enabled,
		OrderByIDDesc: lo.ToPtr(true),
		Name:          request.Name,
	})
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}
	return &reposerver.ListResponse{
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Count:    pag.Count,
		Items:    slice.Map(list, transformer.FromRepo),
	}, nil
}

// Create 创建仓库（默认启用），落创建审计日志。
func (r *repoSvc) Create(ctx context.Context, req *reposerver.CreateRequest) (*reposerver.CreateResponse, error) {
	create, err := r.repoBiz.Create(ctx, &biz.CreateRepoInput{
		Name:         req.Name,
		Enabled:      true,
		NeedGitRepo:  req.NeedGitRepo,
		GitProjectID: req.GitProjectId,
		MarsConfig:   req.MarsConfig,
		Description:  req.Description,
	})
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}
	r.eventBiz.AuditLogWithRequest(
		types.EventActionType_Create,
		biz.MustGetUser(ctx).Name,
		fmt.Sprintf("创建仓库: %d: %s", create.ID, create.Name),
		req,
	)
	return &reposerver.CreateResponse{
		Item: transformer.FromRepo(create),
	}, nil
}

// Show 返回仓库详情。
func (r *repoSvc) Show(ctx context.Context, request *reposerver.ShowRequest) (*reposerver.ShowResponse, error) {
	show, err := r.repoBiz.Show(ctx, int(request.Id))
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}
	return &reposerver.ShowResponse{
		Item: transformer.FromRepo(show),
	}, nil
}

// Update 更新仓库配置，落变更审计日志（含前后 yaml diff）。
func (r *repoSvc) Update(ctx context.Context, req *reposerver.UpdateRequest) (*reposerver.UpdateResponse, error) {
	current, err := r.repoBiz.Get(ctx, int(req.Id))
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}

	create, err := r.repoBiz.Update(ctx, &biz.UpdateRepoInput{
		ID:           req.Id,
		Name:         req.Name,
		NeedGitRepo:  req.NeedGitRepo,
		GitProjectID: req.GitProjectId,
		MarsConfig:   req.MarsConfig,
		Description:  req.Description,
	})
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}
	old, _ := yaml.PrettyMarshal(current)
	out, _ := yaml.PrettyMarshal(create)
	r.eventBiz.AuditLogWithChange(
		types.EventActionType_Update,
		biz.MustGetUser(ctx).Name,
		fmt.Sprintf("更新仓库: %d: %s", create.ID, create.Name),
		&biz.StringYamlPrettier{Str: string(old)},
		&biz.StringYamlPrettier{Str: string(out)},
	)

	return &reposerver.UpdateResponse{
		Item: transformer.FromRepo(create),
	}, nil
}

// ToggleEnabled 启用/禁用仓库，落更新审计日志。
func (r *repoSvc) ToggleEnabled(ctx context.Context, request *reposerver.ToggleEnabledRequest) (*reposerver.ToggleEnabledResponse, error) {
	toggle, err := r.repoBiz.ToggleEnabled(ctx, int(request.Id), request.Enabled)
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}

	status := "禁用"
	if request.Enabled {
		status = "启用"
	}
	r.eventBiz.AuditLogWithRequest(
		types.EventActionType_Update,
		biz.MustGetUser(ctx).Name,
		fmt.Sprintf("[repo 状态变动]: %s 仓库 %s", status, toggle.Name),
		request,
	)

	return &reposerver.ToggleEnabledResponse{
		Item: transformer.FromRepo(toggle),
	}, nil
}

// Delete 删除仓库，落删除审计日志。
func (r *repoSvc) Delete(ctx context.Context, request *reposerver.DeleteRequest) (*reposerver.DeleteResponse, error) {
	if err := r.repoBiz.Delete(ctx, int(request.Id)); err != nil {
		return nil, logError(ctx, r.logger, err)
	}

	r.eventBiz.AuditLogWithRequest(
		types.EventActionType_Delete,
		biz.MustGetUser(ctx).Name,
		fmt.Sprintf("删除 repo: %d", request.Id),
		request,
	)

	return &reposerver.DeleteResponse{}, nil
}

// Clone 克隆仓库：先取源仓库记录用于审计，再执行克隆，落创建审计日志。
func (r *repoSvc) Clone(ctx context.Context, req *reposerver.CloneRequest) (*reposerver.CloneResponse, error) {
	// 先查源仓库（用于审计日志），再执行克隆副作用：源不存在/DB 故障时 fail-fast，
	// 避免"克隆已发生但 Get 失败返回错误"→ 客户端重试产生重复克隆。
	show, err := r.repoBiz.Get(ctx, int(req.Id))
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}
	clone, err := r.repoBiz.Clone(ctx, &biz.CloneRepoInput{
		ID:   int(req.Id),
		Name: req.Name,
	})
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}

	r.eventBiz.AuditLogWithRequest(
		types.EventActionType_Create,
		biz.MustGetUser(ctx).Name,
		fmt.Sprintf("克隆 repo: (id: %d, name: %s) -> (id: %d, name: %s)", show.ID, show.Name, clone.ID, clone.Name),
		req,
	)

	return &reposerver.CloneResponse{
		Item: transformer.FromRepo(clone),
	}, nil
}

// Authorize 是 repo 服务的 admin 门禁：List/Show 放行给任意登录用户，
// 其余仓库管理方法（创建/更新/启停/删除/克隆）仅 admin 可调用。
func (r *repoSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	return r.accessBiz.RequireAdmin(ctx, fullMethodName, "/repo.Repo/List", "/repo.Repo/Show")
}
