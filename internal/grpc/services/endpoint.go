package services

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars-client/v4/endpoint"
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		endpoint.RegisterEndpointServer(s, new(EndpointSvc))
	})
	RegisterEndpoint(endpoint.RegisterEndpointHandlerFromEndpoint)
}

type EndpointSvc struct {
	endpoint.UnimplementedEndpointServer
}

func (e *EndpointSvc) InNamespace(ctx context.Context, request *endpoint.InNamespaceRequest) (*endpoint.InNamespaceResponse, error) {
	var ns models.Namespace
	if err := app.DB().Preload("Projects", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "manifest", "namespace_id", "name")
	}).Where("`id` = ?", request.NamespaceId).First(&ns).Error; err != nil {
		return nil, err
	}

	return &endpoint.InNamespaceResponse{Items: e.endpoints(ns.Name, ns.Projects...)}, nil
}

func (e *EndpointSvc) InProject(ctx context.Context, request *endpoint.InProjectRequest) (*endpoint.InProjectResponse, error) {
	var p models.Project
	if err := app.DB().
		Preload("Namespace").
		Select("id", "manifest", "namespace_id", "name").
		Where("`id` = ?", request.ProjectId).
		First(&p).Error; err != nil {
		return nil, err
	}
	if p.Namespace.Name == "" {
		return nil, errors.New("namespace not exists")
	}

	return &endpoint.InProjectResponse{Items: e.endpoints(p.Namespace.Name, p)}, nil
}

func (e *EndpointSvc) endpoints(ns string, projects ...models.Project) []*types.ServiceEndpoint {
	var res = []*types.ServiceEndpoint{}
	nodePortMapping := utils.GetNodePortMappingByProjects(ns, projects...)
	ingMapping := utils.GetIngressMappingByProjects(ns, projects...)
	lbMapping := utils.GetLoadBalancerMappingByProjects(ns, projects...)
	res = append(res, ingMapping.AllEndpoints()...)
	res = append(res, lbMapping.AllEndpoints()...)
	res = append(res, nodePortMapping.AllEndpoints()...)
	return res
}
