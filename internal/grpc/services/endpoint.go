package services

import (
	"context"

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

	var res = []*types.ServiceEndpoint{}
	nodePortMapping := utils.GetNodePortMappingByProjects(ns.Name, ns.Projects...)
	ingMapping := utils.GetIngressMappingByProjects(ns.Name, ns.Projects...)
	for _, hosts := range ingMapping {
		res = append(res, hosts...)
	}

	for _, hosts := range nodePortMapping {
		res = append(res, hosts...)
	}

	return &endpoint.InNamespaceResponse{Items: res}, nil
}

func (e *EndpointSvc) InProject(ctx context.Context, request *endpoint.InProjectRequest) (*endpoint.InProjectResponse, error) {
	var p models.Project
	if err := app.DB().Where("`id` = ?", request.ProjectId).First(&p).Error; err != nil {
		return nil, err
	}
	sd, err := e.InNamespace(ctx, &endpoint.InNamespaceRequest{NamespaceId: int64(p.NamespaceId)})
	if err != nil {
		return nil, err
	}
	var res = []*types.ServiceEndpoint{}

	for _, re := range sd.Items {
		if re.Name == p.Name {
			res = append(res, re)
		}
	}

	return &endpoint.InProjectResponse{Items: res}, nil
}
