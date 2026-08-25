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
		biz.MustGetUser(ctx).Email,
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
	old := repoAuditYaml([]*types.RepoModel{transformer.FromRepo(current)})
	out := repoAuditYaml([]*types.RepoModel{transformer.FromRepo(create)})
	r.eventBiz.AuditLogWithChange(
		types.EventActionType_Update,
		biz.MustGetUser(ctx).Name,
		biz.MustGetUser(ctx).Email,
		fmt.Sprintf("更新仓库: %d: %s", create.ID, create.Name),
		&biz.StringYamlPrettier{Str: old},
		&biz.StringYamlPrettier{Str: out},
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
		biz.MustGetUser(ctx).Email,
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
		biz.MustGetUser(ctx).Email,
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
		biz.MustGetUser(ctx).Email,
		fmt.Sprintf("克隆 repo: (id: %d, name: %s) -> (id: %d, name: %s)", show.ID, show.Name, clone.ID, clone.Name),
		req,
	)

	return &reposerver.CloneResponse{
		Item: transformer.FromRepo(clone),
	}, nil
}

// Export 导出全部 repo 为 JSON：返回体与导入请求同构（repeated RepoModel），
// 文件内容可直接回传 /api/repos/import 实现 round-trip。
// 复用 All（soft-delete 已被 interceptor 过滤，导出不含已删除 repo）。
func (r *repoSvc) Export(ctx context.Context, req *reposerver.ExportRequest) (*reposerver.ExportResponse, error) {
	repos, err := r.repoBiz.All(ctx, &biz.AllRepoRequest{})
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}
	return &reposerver.ExportResponse{
		Items: slice.Map(repos, transformer.FromRepo),
	}, nil
}

// snapshotRepos 采集全部 repo 的当前状态（与导出同构），用于导入前快照与审计。
func (r *repoSvc) snapshotRepos(ctx context.Context) ([]*types.RepoModel, error) {
	repos, err := r.repoBiz.All(ctx, &biz.AllRepoRequest{})
	if err != nil {
		return nil, err
	}
	return slice.Map(repos, transformer.FromRepo), nil
}

// Import 批量导入 repo：请求体与导出 JSON 同构，按名称幂等覆盖（biz 层实现）。
// dry_run 为 true 时只预览计数（不落库、不留审计、不产生快照）；
// 真实导入前先采集全量快照（审计日志 old 字段），导入后采集新状态
// （审计日志 new 字段），完整记录变更前后，便于事后在事件日志中对照。
func (r *repoSvc) Import(ctx context.Context, req *reposerver.ImportRequest) (*reposerver.ImportResponse, error) {
	items := slice.Map(req.GetItems(), toImportRepoItem)

	// dry_run：只预览将新建/将覆盖的数量，不触碰数据、不留审计、不产生快照。
	if req.DryRun {
		created, updated, err := r.repoBiz.PreviewImport(ctx, items)
		if err != nil {
			return nil, logError(ctx, r.logger, err)
		}
		// 采集当前全量状态，按 name 匹配出将被覆盖条目的旧值，供前端渲染 old→new diff。
		all, err := r.snapshotRepos(ctx)
		if err != nil {
			return nil, logError(ctx, r.logger, err)
		}
		byName := make(map[string]*types.RepoModel, len(all))
		for _, m := range all {
			byName[m.Name] = m
		}
		updatedOld := make([]*types.RepoModel, 0, updated)
		for _, it := range items {
			if old, ok := byName[it.Name]; ok {
				updatedOld = append(updatedOld, old)
			}
		}
		return &reposerver.ImportResponse{
			Total:      int32(len(items)),
			Created:    int32(created),
			Updated:    int32(updated),
			UpdatedOld: updatedOld,
		}, nil
	}

	// 导入前快照：写入审计日志 old 字段，供事后在事件日志中对照变更前状态。
	backup, err := r.snapshotRepos(ctx)
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}

	created, updated, err := r.repoBiz.Import(ctx, items)
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}

	// 导入后快照：审计日志 new 字段，与 old 一起完整记录本次变更前后状态。
	current, err := r.snapshotRepos(ctx)
	if err != nil {
		return nil, logError(ctx, r.logger, err)
	}
	total := len(items)
	old := repoAuditYaml(backup)
	out := repoAuditYaml(current)
	r.eventBiz.AuditLogWithChange(
		types.EventActionType_Update,
		biz.MustGetUser(ctx).Name,
		biz.MustGetUser(ctx).Email,
		fmt.Sprintf("导入 repo: 共 %d 个（新建 %d，覆盖 %d）", total, created, updated),
		&biz.StringYamlPrettier{Str: old},
		&biz.StringYamlPrettier{Str: out},
	)
	return &reposerver.ImportResponse{
		Total:   int32(total),
		Created: int32(created),
		Updated: int32(updated),
	}, nil
}

// repoAuditYaml 把 repo 状态序列化成审计日志 YAML：只保留业务字段
// （name/enabled/need_git_repo/git_project_id/description/mars_config），
// 剔除服务器生成的 id/时间戳/git 项目名，避免导入/更新审计 diff
// 出现与业务无关的噪声（如 updated_at 一刷新整个列表看着都变了）。
func repoAuditYaml(repos []*types.RepoModel) string {
	out, _ := yaml.PrettyMarshal(lo.Map(repos, func(m *types.RepoModel, _ int) map[string]any {
		return map[string]any{
			"name":           m.Name,
			"enabled":        m.Enabled,
			"need_git_repo":  m.NeedGitRepo,
			"git_project_id": m.GitProjectId,
			"description":    m.Description,
			"mars_config":    m.MarsConfig,
		}
	}))
	return string(out)
}

// toImportRepoItem 抽取 RepoModel 中可落库字段，忽略服务器生成字段
// （id/时间戳/git 项目名，后者由 data 层按 git_project_id 重新推导）。
func toImportRepoItem(m *types.RepoModel) *biz.ImportRepoItem {
	if m == nil {
		return nil
	}
	item := &biz.ImportRepoItem{
		Name:        m.Name,
		Enabled:     m.Enabled,
		NeedGitRepo: m.NeedGitRepo,
		MarsConfig:  m.MarsConfig,
		Description: m.Description,
	}
	// GitProjectID 仅在需要 git 仓库且 id 有效时携带；0 视为未配置，
	// 避免把无效 0 值作为显式指针落库（data 层会按 0 查 git 项目）。
	if m.NeedGitRepo && m.GitProjectId > 0 {
		item.GitProjectID = lo.ToPtr(m.GitProjectId)
	}
	return item
}

// Authorize 是 repo 服务的 admin 门禁：List/Show 放行给任意登录用户，
// 其余仓库管理方法（创建/更新/启停/删除/克隆/导入/导出）仅 admin 可调用。
func (r *repoSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	return r.accessBiz.RequireAdmin(ctx, fullMethodName, reposerver.Repo_List_FullMethodName, reposerver.Repo_Show_FullMethodName)
}
