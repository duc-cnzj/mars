package services

import (
	"context"

	"google.golang.org/grpc"

	"github.com/duc-cnzj/mars-client/v4/endpoint"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
)

func init() {
	AddServerFunc(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		endpoint.RegisterEndpointServer(s, new(EndpointSvc))
	})
	AddEndpointFunc(endpoint.RegisterEndpointHandlerFromEndpoint)
}

type EndpointSvc struct {
	endpoint.UnimplementedEndpointServer
}

func (e *EndpointSvc) InNamespace(ctx context.Context, request *endpoint.EndpointInNamespaceRequest) (*endpoint.EndpointInNamespaceResponse, error) {
	var ns models.Namespace
	if err := app.DB().Preload("Projects").Where("`id` = ?", request.NamespaceId).First(&ns).Error; err != nil {
		return nil, err
	}

	var res = []*endpoint.ServiceEndpoint{}
	nodePortMapping := utils.GetNodePortMappingByNamespace(ns.Name)
	ingMapping := utils.GetIngressMappingByNamespace(ns.Name)
	for _, hosts := range ingMapping {
		res = append(res, hosts...)
	}

	for _, hosts := range nodePortMapping {
		res = append(res, hosts...)
	}

	return &endpoint.EndpointInNamespaceResponse{Items: res}, nil
}

func (e *EndpointSvc) InProject(ctx context.Context, request *endpoint.EndpointInProjectRequest) (*endpoint.EndpointInProjectResponse, error) {
	var p models.Project
	if err := app.DB().Where("`id` = ?", request.ProjectId).First(&p).Error; err != nil {
		return nil, err
	}
	sd, err := e.InNamespace(ctx, &endpoint.EndpointInNamespaceRequest{NamespaceId: int64(p.NamespaceId)})
	if err != nil {
		return nil, err
	}
	var res = []*endpoint.ServiceEndpoint{}

	for _, re := range sd.Items {
		if re.Name == p.Name {
			res = append(res, re)
		}
	}

	return &endpoint.EndpointInProjectResponse{Items: res}, nil
}
