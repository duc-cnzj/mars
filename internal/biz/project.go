package biz

import (
	"context"
	"errors"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
)

// ProjectBiz 收口项目领域业务：增删改查、版本管理、部署状态与容器/端点派生。
type ProjectBiz interface {
	// GetAllActiveContainers 返回项目当前活跃容器状态。
	GetAllActiveContainers(ctx context.Context, id int) ([]*types.StateContainer, error)
	// CheckApplyStatus 判定项目最近一次部署后新版本容器是否正常运行。
	CheckApplyStatus(ctx context.Context, id int) (*ApplyStatus, error)
	// GetProjectEndpointsInNamespace 汇总命名空间内项目的服务端点。
	GetProjectEndpointsInNamespace(ctx context.Context, namespace string, projectIDs ...int) ([]*types.ServiceEndpoint, error)
	// ResourceTree 返回项目资源拓扑树（完整资源列表，供拓扑图 Tab 渲染与 pod 事件实时刷新）。
	ResourceTree(ctx context.Context, id int) (*ResourceTree, error)
	// List 分页列出项目。
	List(ctx context.Context, input *ListProjectInput) ([]*Project, *pagination.Pagination, error)
	// ListAllProjectBriefs 返回全部项目的精简投影（仅空间资源聚合消费的字段：
	// Name/PodSelectors/Namespace.Name，其余字段为零值），供聚合做 pod→项目归属映射。
	ListAllProjectBriefs(ctx context.Context) ([]*Project, error)
	// Create 校验输入后创建项目。
	Create(ctx context.Context, project *CreateProjectInput) (*Project, error)
	// Show 按 id 查询项目。
	Show(ctx context.Context, id int) (*Project, error)
	// Version 查询项目当前版本号。
	Version(ctx context.Context, id int) (int, error)
	// Delete 校验 id 后删除项目。
	Delete(ctx context.Context, id int) error
	// FindByName 按名称与命名空间查询项目。
	FindByName(ctx context.Context, name string, nsID int) (*Project, error)
	// UpdateDeployStatus 更新项目部署状态。
	UpdateDeployStatus(ctx context.Context, id int, status types.Deploy) (*Project, error)
	// UpdateVersion 更新项目版本号。
	UpdateVersion(ctx context.Context, id int, version int) (*Project, error)
	// FindByVersion 按版本号查询项目。
	FindByVersion(ctx context.Context, id, version int) (*Project, error)
	// UpdateStatusByVersion 按版本号更新部署状态。
	UpdateStatusByVersion(ctx context.Context, id int, status types.Deploy, version int) (*Project, error)
	// UpdateProject 校验输入后更新项目配置。
	UpdateProject(ctx context.Context, input *UpdateProjectInput) (*Project, error)
	// Liveness 聚合项目活跃度清单（分类/统计/部署次数）。
	Liveness(ctx context.Context, input *LivenessInput) (*LivenessResult, error)
}

type projectBiz struct {
	logger   mlog.Logger
	projRepo ProjectRepo
	clRepo   ChangelogRepo
	k8sRepo  K8sRepo
}

// NewProjectBiz 构造 project biz：projRepo 提供项目读写，k8sRepo 提供集群状态，
// clRepo 提供部署次数聚合（活跃度清单用）。
func NewProjectBiz(logger mlog.Logger, projRepo ProjectRepo, k8sRepo K8sRepo, clRepo ChangelogRepo) ProjectBiz {
	return &projectBiz{
		logger:   logger.WithModule("biz/project"),
		projRepo: projRepo,
		clRepo:   clRepo,
		k8sRepo:  k8sRepo,
	}
}

// GetAllActiveContainers 从项目 PodSelectors 推导活跃容器状态（见 buildStateContainers）。
func (p *projectBiz) GetAllActiveContainers(ctx context.Context, id int) ([]*types.StateContainer, error) {
	proj, err := p.projRepo.Show(ctx, id)
	if err != nil {
		return nil, err
	}
	return buildStateContainers(ctx, p.k8sRepo, proj)
}

// ResourceTree 返回项目资源拓扑树（见 buildResourceTree）。
func (p *projectBiz) ResourceTree(ctx context.Context, id int) (*ResourceTree, error) {
	proj, err := p.projRepo.Show(ctx, id)
	if err != nil {
		return nil, err
	}
	return buildResourceTree(ctx, p.k8sRepo, proj)
}

// GetProjectEndpointsInNamespace 汇总 Ingress/LoadBalancer/NodePort/HTTPRoute 四种来源的
// endpoint，优先级顺序为 ing → lb → nodePort → httpRoute；任一来源的 List 失败均上抛，
// 由最上层 services 统一打印，不静默返回残缺端点列表。
func (p *projectBiz) GetProjectEndpointsInNamespace(ctx context.Context, namespace string, projectIDs ...int) ([]*types.ServiceEndpoint, error) {
	projs, err := p.projRepo.FindProjectsByIDs(ctx, projectIDs...)
	if err != nil {
		return nil, err
	}
	var res []*types.ServiceEndpoint
	ing, err := BuildIngressMappingByProjects(ctx, p.logger, p.k8sRepo, namespace, projs...)
	if err != nil {
		return nil, err
	}
	res = append(res, ing.AllEndpoints()...)
	lb, err := BuildLoadBalancerMappingByProjects(ctx, p.logger, p.k8sRepo, namespace, projs...)
	if err != nil {
		return nil, err
	}
	res = append(res, lb.AllEndpoints()...)
	nodePort, err := BuildNodePortMappingByProjects(ctx, p.logger, p.k8sRepo, namespace, projs...)
	if err != nil {
		return nil, err
	}
	res = append(res, nodePort.AllEndpoints()...)
	httpRoute, err := BuildGatewayHTTPRouteMappingByProjects(ctx, p.logger, p.k8sRepo, namespace, projs...)
	if err != nil {
		return nil, err
	}
	res = append(res, httpRoute.AllEndpoints()...)
	return res, nil
}

// List 分页列出项目（透传 repo）。
func (p *projectBiz) List(ctx context.Context, input *ListProjectInput) ([]*Project, *pagination.Pagination, error) {
	return p.projRepo.List(ctx, input)
}

// ListAllProjectBriefs 返回全部项目的精简投影（透传 repo），跨命名空间做 pod→项目归属映射用。
func (p *projectBiz) ListAllProjectBriefs(ctx context.Context) ([]*Project, error) {
	return p.projRepo.ListAllProjectBriefs(ctx)
}

// Create 校验输入后创建项目。
func (p *projectBiz) Create(ctx context.Context, project *CreateProjectInput) (*Project, error) {
	if project == nil {
		return nil, errs.WrapInvalidArgument(errors.New("project 不能为空"), "create project")
	}
	if project.Name == "" {
		return nil, errs.WrapInvalidArgument(errors.New("project 名称不能为空"), "create project")
	}
	if project.NamespaceID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("project 所属 namespace 不能为空"), "create project")
	}
	if project.RepoID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("project 所属 repo 不能为空"), "create project")
	}
	return p.projRepo.Create(ctx, project)
}

// Show 按 id 查询项目（透传 repo）。
func (p *projectBiz) Show(ctx context.Context, id int) (*Project, error) {
	return p.projRepo.Show(ctx, id)
}

// Version 查询项目当前版本号（透传 repo）。
func (p *projectBiz) Version(ctx context.Context, id int) (int, error) {
	return p.projRepo.Version(ctx, id)
}

// Delete 校验 id 后删除项目。
func (p *projectBiz) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return errs.WrapInvalidArgument(errors.New("project id 不能小于等于 0"), "delete project")
	}
	return p.projRepo.Delete(ctx, id)
}

// FindByName 按名称与 namespace 查询项目（透传 repo）。
func (p *projectBiz) FindByName(ctx context.Context, name string, nsID int) (*Project, error) {
	return p.projRepo.FindByName(ctx, name, nsID)
}

// UpdateDeployStatus 更新项目部署状态（透传 repo）。
func (p *projectBiz) UpdateDeployStatus(ctx context.Context, id int, status types.Deploy) (*Project, error) {
	return p.projRepo.UpdateDeployStatus(ctx, id, status)
}

// UpdateVersion 更新项目版本号（透传 repo）。
func (p *projectBiz) UpdateVersion(ctx context.Context, id int, version int) (*Project, error) {
	return p.projRepo.UpdateVersion(ctx, id, version)
}

// FindByVersion 按版本号查询项目（透传 repo）。
func (p *projectBiz) FindByVersion(ctx context.Context, id, version int) (*Project, error) {
	return p.projRepo.FindByVersion(ctx, id, version)
}

// UpdateStatusByVersion 按版本号更新项目部署状态（透传 repo）。
func (p *projectBiz) UpdateStatusByVersion(ctx context.Context, id int, status types.Deploy, version int) (*Project, error) {
	return p.projRepo.UpdateStatusByVersion(ctx, id, status, version)
}

// UpdateProject 校验输入后更新项目。
func (p *projectBiz) UpdateProject(ctx context.Context, input *UpdateProjectInput) (*Project, error) {
	if input.ID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("project id 不能小于等于 0"), "update project")
	}
	return p.projRepo.UpdateProject(ctx, input)
}

// LivenessKind 是项目活跃度分类：active（活跃）/ dormant（休眠）/ zombie（僵尸）。
type LivenessKind string

const (
	// LivenessActive 活跃：最近 ActiveLivenessDays 天内有更新。
	LivenessActive LivenessKind = "active"
	// LivenessDormant 休眠：超过 ActiveLivenessDays 天但未达 ZombieLivenessDays 天未更新。
	LivenessDormant LivenessKind = "dormant"
	// LivenessZombie 僵尸：超过 ZombieLivenessDays 天未更新。
	LivenessZombie LivenessKind = "zombie"
)

// 活跃度分类阈值（服务端单一事实来源，data 层 SQL 边界计算共用，勿另起常量）：
// 活跃=最近 30 天有更新，僵尸=超过 90 天未更新。
const (
	ActiveLivenessDays = 30
	ZombieLivenessDays = 90
)

// ClassifyLiveness 按项目更新时间距 now 的天数分类活跃度；now 由调用方注入便于测试。
// 导出供 data 层边界奇偶性守护测试对照 SQL 分类结果，杜绝 Go/SQL 双份公式漂移。
func ClassifyLiveness(updatedAt, now time.Time) LivenessKind {
	days := int(now.Sub(updatedAt).Hours() / 24)
	switch {
	case days <= ActiveLivenessDays:
		return LivenessActive
	case days >= ZombieLivenessDays:
		return LivenessZombie
	default:
		return LivenessDormant
	}
}

// LivenessInput 是项目活跃度聚合的输入。
type LivenessInput struct {
	Page, PageSize int32
	// Search 关键词：匹配项目名/命名空间名（模糊，不分大小写）。
	Search string
	// Liveness 活跃度分类过滤：空 = 全部，否则 active/dormant/zombie。
	Liveness string
	// Sort 排序方向：空 = 按更新时间倒序（desc）；asc/desc = 指定更新时间升/降序。
	Sort string
}

// LivenessItem 是活跃度清单中的单条项目：携带部署次数。
// 活跃度分类不发线上（proto 无 kind 字段，前端依 updatedAt 自行推导），故不再承载 Kind。
type LivenessItem struct {
	Project     *Project
	DeployCount int
}

// LivenessStats 是活跃度统计（基于搜索命中全量，不随分页/过滤裁剪）。
type LivenessStats struct {
	Total, Active, Dormant, Zombie int
}

// LivenessPageQuery 是活跃度分页查询输入：搜索/分类过滤/排序/分页全部下沉 SQL。
// Now 为分类基准时间：SQL 边界（now-31d/now-90d）与 biz 行级标记共用同一基准，杜绝边界竞态。
type LivenessPageQuery struct {
	Search, Liveness, Sort string
	Page, PageSize         int32
	Now                    time.Time
}

// LivenessPageResult 是活跃度分页查询结果：已分页项目（含仓库/命名空间边）+ 全量统计。
// Count 为分类过滤后总数（未分页），Stats 为搜索命中全量统计（不随过滤/分页裁剪）。
type LivenessPageResult struct {
	Projects []*Project
	Count    int
	Stats    LivenessStats
}

// LivenessResult 是活跃度聚合结果：已分页条目 + 全量统计。
type LivenessResult struct {
	Items          []*LivenessItem
	Page, PageSize int32
	Count          int32
	Stats          LivenessStats
}

// Liveness 聚合项目活跃度清单：分类/过滤/排序/统计/分页全部由 repo 下沉 SQL（真分页），
// biz 仅补充部署次数并装配结果。
func (p *projectBiz) Liveness(ctx context.Context, input *LivenessInput) (*LivenessResult, error) {
	now := time.Now()
	page, err := p.projRepo.ListLivenessPage(ctx, &LivenessPageQuery{
		Search:   input.Search,
		Liveness: input.Liveness,
		Sort:     input.Sort,
		Page:     input.Page,
		PageSize: input.PageSize,
		Now:      now,
	})
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(page.Projects))
	for _, proj := range page.Projects {
		ids = append(ids, proj.ID)
	}
	counts, err := p.clRepo.CountByProjectIDs(ctx, ids...)
	if err != nil {
		return nil, err
	}
	items := make([]*LivenessItem, 0, len(page.Projects))
	for _, proj := range page.Projects {
		items = append(items, &LivenessItem{Project: proj, DeployCount: counts[proj.ID]})
	}
	return &LivenessResult{
		Items:    items,
		Page:     input.Page,
		PageSize: input.PageSize,
		Count:    int32(page.Count),
		Stats:    page.Stats,
	}, nil
}

// ProjectRepo 是项目仓库端口。
type ProjectRepo interface {
	// FindProjectsByIDs 按主键批量取项目（endpoint 编排需要，按 Manifest 匹配集群内对象）。
	FindProjectsByIDs(ctx context.Context, ids ...int) ([]*Project, error)
	// ListLivenessPage 分页查询活跃度聚合所需项目（含仓库/命名空间边）：分类过滤/排序/统计/
	// 分页全部下沉 SQL（真分页），Now 为分类基准时间（活跃=updated_at > now-31d，僵尸=<= now-90d）。
	ListLivenessPage(ctx context.Context, query *LivenessPageQuery) (*LivenessPageResult, error)
	// List 分页列出项目（可按命名空间/名称/访问谓词过滤）。
	List(ctx context.Context, input *ListProjectInput) ([]*Project, *pagination.Pagination, error)
	// ListAllProjectBriefs 查询全部项目的精简投影（仅 Name/PodSelectors/Namespace.Name）。
	ListAllProjectBriefs(ctx context.Context) ([]*Project, error)
	// Create 创建项目。
	Create(ctx context.Context, project *CreateProjectInput) (*Project, error)
	// Show 按 id 查询项目。
	Show(ctx context.Context, id int) (*Project, error)
	// Version 查询项目当前版本号。
	Version(ctx context.Context, id int) (int, error)
	// Delete 删除项目。
	Delete(ctx context.Context, id int) error
	// FindByName 按名称与命名空间查询项目。
	FindByName(ctx context.Context, name string, nsID int) (*Project, error)
	// ListByDeployStatus 按部署状态过滤项目并携带 namespace（cron 修复部署状态用）。
	ListByDeployStatus(ctx context.Context, statuses ...types.Deploy) ([]*Project, error)
	// UpdateDeployStatus 更新项目部署状态。
	UpdateDeployStatus(ctx context.Context, id int, status types.Deploy) (*Project, error)
	// UpdateVersion 更新项目版本号。
	UpdateVersion(ctx context.Context, id int, version int) (*Project, error)
	// FindByVersion 按版本号查询项目。
	FindByVersion(ctx context.Context, id, version int) (*Project, error)
	// UpdateStatusByVersion 按版本号更新部署状态。
	UpdateStatusByVersion(ctx context.Context, id int, status types.Deploy, version int) (*Project, error)
	// UpdateProject 更新项目配置。
	UpdateProject(ctx context.Context, input *UpdateProjectInput) (*Project, error)
}
