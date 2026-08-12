package biz

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
)

// EndpointBiz 封装端点编排逻辑。
type EndpointBiz interface {
	// InProject 返回单个项目的全部 endpoint。
	InProject(ctx context.Context, projectID int) ([]*types.ServiceEndpoint, error)
	// InNamespace 返回 namespace 下全部项目的 endpoint。
	InNamespace(ctx context.Context, namespaceID int) ([]*types.ServiceEndpoint, error)
}

type endpointBiz struct {
	projBiz ProjectBiz
	nsRepo  NamespaceRepo
	logger  mlog.Logger
}

// NewEndpointBiz 构造 endpoint biz。
func NewEndpointBiz(logger mlog.Logger, projBiz ProjectBiz, nsRepo NamespaceRepo) EndpointBiz {
	return &endpointBiz{
		logger:  logger.WithModule("biz/endpoint"),
		projBiz: projBiz,
		nsRepo:  nsRepo,
	}
}

// InProject 返回单个项目的全部 endpoint。
func (b *endpointBiz) InProject(ctx context.Context, projectID int) ([]*types.ServiceEndpoint, error) {
	show, err := b.projBiz.Show(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return b.projBiz.GetProjectEndpointsInNamespace(ctx, show.Namespace.Name, show.ID)
}

// InNamespace 返回 namespace 下全部项目的 endpoint。
func (b *endpointBiz) InNamespace(ctx context.Context, namespaceID int) ([]*types.ServiceEndpoint, error) {
	show, err := b.nsRepo.Show(ctx, namespaceID)
	if err != nil {
		return nil, err
	}
	return b.projBiz.GetProjectEndpointsInNamespace(ctx, show.Name, slice.Map(show.Projects, func(v *Project) int { return v.ID })...)
}
