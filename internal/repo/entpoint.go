package repo

import (
	"context"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/ent"
	"github.com/duc-cnzj/mars/v5/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v5/internal/ent/project"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/serialize"
)

type EndpointRepo interface {
	InNamespace(ctx context.Context, namespaceID int) (res []*types.ServiceEndpoint, err error)
	InProject(ctx context.Context, projectID int) (res []*types.ServiceEndpoint, err error)
}

var _ EndpointRepo = (*endpointRepo)(nil)

type endpointRepo struct {
	logger mlog.Logger
	data   data.Data

	projRepo ProjectRepo
}

func NewEndpointRepo(logger mlog.Logger, data data.Data, projRepo ProjectRepo) EndpointRepo {
	return &endpointRepo{logger: logger.WithModule("repo/endpoint"), data: data, projRepo: projRepo}
}

func (repo *endpointRepo) InProject(ctx context.Context, projectID int) (res []*types.ServiceEndpoint, err error) {
	var db = repo.data.DB()
	first, err := db.Project.Query().
		WithNamespace().
		Select(
			project.FieldID,
			project.FieldName,
			project.FieldNamespaceID,
		).
		Where(project.ID(projectID)).
		First(ctx)
	if err != nil {
		return nil, ToError(404, err)
	}
	nodePortMapping := repo.projRepo.GetNodePortMappingByProjects(ctx, first.Edges.Namespace.Name, ToProject(first))
	ingMapping := repo.projRepo.GetIngressMappingByProjects(ctx, first.Edges.Namespace.Name, ToProject(first))
	lbMapping := repo.projRepo.GetLoadBalancerMappingByProjects(ctx, first.Edges.Namespace.Name, ToProject(first))
	res = append(res, ingMapping.AllEndpoints()...)
	res = append(res, lbMapping.AllEndpoints()...)
	res = append(res, nodePortMapping.AllEndpoints()...)
	return
}

func (repo *endpointRepo) InNamespace(ctx context.Context, namespaceID int) (res []*types.ServiceEndpoint, err error) {
	var db = repo.data.DB()
	first, err := db.Namespace.Query().
		WithProjects(func(query *ent.ProjectQuery) {
			query.Select(
				project.FieldID,
				project.FieldName,
				project.FieldNamespaceID,
			)
		}).
		Where(namespace.ID(namespaceID)).
		First(ctx)
	if err != nil {
		return nil, ToError(404, err)
	}

	nodePortMapping := repo.projRepo.GetNodePortMappingByProjects(ctx, first.Name, serialize.Serialize(first.Edges.Projects, ToProject)...)
	ingMapping := repo.projRepo.GetIngressMappingByProjects(ctx, first.Name, serialize.Serialize(first.Edges.Projects, ToProject)...)
	lbMapping := repo.projRepo.GetLoadBalancerMappingByProjects(ctx, first.Name, serialize.Serialize(first.Edges.Projects, ToProject)...)
	res = append(res, ingMapping.AllEndpoints()...)
	res = append(res, lbMapping.AllEndpoints()...)
	res = append(res, nodePortMapping.AllEndpoints()...)
	return
}
