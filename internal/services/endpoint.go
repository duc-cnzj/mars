package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/endpoint"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

var _ endpoint.EndpointServer = (*endpointSvc)(nil)

// endpointSvc 是 endpoint.EndpointServer 的 gRPC 实现：查询命名空间/项目下的端点列表，
// 经 access 校验访问权限，由 NewEndpointSvc 构造。
type endpointSvc struct {
	endpoint.UnimplementedEndpointServer

	logger    mlog.Logger
	epBiz     biz.EndpointBiz
	accessBiz biz.AccessBiz
}

// EndpointSvcDeps 收口 NewEndpointSvc 的构造依赖，由 wire 按字段注入。
type EndpointSvcDeps struct {
	Logger    mlog.Logger
	EpBiz     biz.EndpointBiz
	AccessBiz biz.AccessBiz
}

// NewEndpointSvc 收口端点服务的构造依赖，由 wire 按字段注入。
func NewEndpointSvc(deps EndpointSvcDeps) endpoint.EndpointServer {
	logger := deps.Logger.WithModule("services/endpoint")
	return &endpointSvc{
		logger:    logger,
		epBiz:     deps.EpBiz,
		accessBiz: deps.AccessBiz,
	}
}

// InNamespace 返回命名空间下全部可用端点，响应前做命名空间级访问控制。
func (e *endpointSvc) InNamespace(ctx context.Context, request *endpoint.InNamespaceRequest) (*endpoint.InNamespaceResponse, error) {
	if _, nserr := e.accessBiz.RequireNamespaceAccessByID(ctx, int(request.NamespaceId)); nserr != nil {
		return nil, nserr
	}

	res, err := e.epBiz.InNamespace(ctx, int(request.NamespaceId))
	if err != nil {
		return nil, logError(ctx, e.logger, err)
	}
	return &endpoint.InNamespaceResponse{Items: res}, nil
}

// InProject 返回项目关联的全部端点，响应前做项目级访问控制。
func (e *endpointSvc) InProject(ctx context.Context, request *endpoint.InProjectRequest) (*endpoint.InProjectResponse, error) {
	if _, err := e.accessBiz.RequireProjectAccess(ctx, int(request.ProjectId)); err != nil {
		return nil, err
	}

	res, err := e.epBiz.InProject(ctx, int(request.ProjectId))
	if err != nil {
		return nil, logError(ctx, e.logger, err)
	}
	return &endpoint.InProjectResponse{Items: res}, nil
}
