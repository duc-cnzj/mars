package biz

import (
	"context"
	"errors"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
)

// ProjectBiz 收口项目领域业务：增删改查、版本管理、部署状态与容器/端点派生。
type ProjectBiz interface {
	// GetAllActiveContainers 返回项目当前活跃容器状态。
	GetAllActiveContainers(ctx context.Context, id int) ([]*types.StateContainer, error)
	// GetProjectEndpointsInNamespace 汇总命名空间内项目的服务端点。
	GetProjectEndpointsInNamespace(ctx context.Context, namespace string, projectIDs ...int) ([]*types.ServiceEndpoint, error)
	// List 分页列出项目。
	List(ctx context.Context, input *ListProjectInput) ([]*Project, *pagination.Pagination, error)
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
}

type projectBiz struct {
	logger   mlog.Logger
	projRepo ProjectRepo
	k8sRepo  K8sRepo
}

// NewProjectBiz 构造 project biz。
func NewProjectBiz(logger mlog.Logger, projRepo ProjectRepo, k8sRepo K8sRepo) ProjectBiz {
	return &projectBiz{
		logger:   logger.WithModule("biz/project"),
		projRepo: projRepo,
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

// ProjectRepo 是项目仓库端口。
type ProjectRepo interface {
	// FindProjectsByIDs 按主键批量取项目（endpoint 编排需要，按 Manifest 匹配集群内对象）。
	FindProjectsByIDs(ctx context.Context, ids ...int) ([]*Project, error)
	// List 分页列出项目（可按命名空间/名称/访问谓词过滤）。
	List(ctx context.Context, input *ListProjectInput) ([]*Project, *pagination.Pagination, error)
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
