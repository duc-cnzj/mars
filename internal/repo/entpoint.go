package repo

import (
	"context"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/namespace"
	"github.com/duc-cnzj/mars/v4/internal/ent/project"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type EndpointRepo interface {
	InNamespace(ctx context.Context, namespaceID int) (res []*types.ServiceEndpoint, err error)
	InProject(ctx context.Context, projectID int) (res []*types.ServiceEndpoint, err error)
}

var _ EndpointRepo = (*endpointRepo)(nil)

type endpointRepo struct {
	logger mlog.Logger
	db     *ent.Client

	projRepo ProjectRepo
}

func NewEndpointRepo(logger mlog.Logger, data *data.Data) EndpointRepo {
	return &endpointRepo{logger: logger, db: data.DB}
}

func (repo *endpointRepo) InProject(ctx context.Context, projectID int) (res []*types.ServiceEndpoint, err error) {
	first, err := repo.db.Project.Query().WithNamespace().Select(
		project.FieldID,
		project.FieldManifest,
		project.FieldName,
		project.FieldNamespaceID,
	).
		Where(project.ID(projectID)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	nodePortMapping := repo.projRepo.GetNodePortMappingByProjects(first.Edges.Namespace.Name, first)
	ingMapping := repo.projRepo.GetIngressMappingByProjects(first.Edges.Namespace.Name, first)
	lbMapping := repo.projRepo.GetLoadBalancerMappingByProjects(first.Edges.Namespace.Name, first)
	res = append(res, ingMapping.AllEndpoints()...)
	res = append(res, lbMapping.AllEndpoints()...)
	res = append(res, nodePortMapping.AllEndpoints()...)
	return
}

func (repo *endpointRepo) InNamespace(ctx context.Context, namespaceID int) (res []*types.ServiceEndpoint, err error) {
	first, err := repo.db.Namespace.Query().
		WithProjects(func(query *ent.ProjectQuery) {
			query.Select(
				project.FieldID,
				project.FieldManifest,
				project.FieldName,
				project.FieldNamespaceID,
			)
		}).
		Where(namespace.ID(namespaceID)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	nodePortMapping := repo.projRepo.GetNodePortMappingByProjects(first.Name, first.Edges.Projects...)
	ingMapping := repo.projRepo.GetIngressMappingByProjects(first.Name, first.Edges.Projects...)
	lbMapping := repo.projRepo.GetLoadBalancerMappingByProjects(first.Name, first.Edges.Projects...)
	res = append(res, ingMapping.AllEndpoints()...)
	res = append(res, lbMapping.AllEndpoints()...)
	res = append(res, nodePortMapping.AllEndpoints()...)
	return
}
