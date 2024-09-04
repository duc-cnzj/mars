package repo

import (
	"context"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/data"
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
	nsRepo   NamespaceRepo
}

func NewEndpointRepo(logger mlog.Logger, data data.Data, projRepo ProjectRepo, nsRepo NamespaceRepo) EndpointRepo {
	return &endpointRepo{logger: logger.WithModule("repo/endpoint"), data: data, projRepo: projRepo, nsRepo: nsRepo}
}

func (repo *endpointRepo) InProject(ctx context.Context, projectID int) ([]*types.ServiceEndpoint, error) {
	show, err := repo.projRepo.Show(ctx, projectID)
	if err != nil {
		return nil, ToError(404, err)
	}

	return repo.projRepo.GetProjectEndpointsInNamespace(ctx, show.Namespace.Name, show.ID)
}

func (repo *endpointRepo) InNamespace(ctx context.Context, namespaceID int) ([]*types.ServiceEndpoint, error) {
	show, err := repo.nsRepo.Show(ctx, namespaceID)
	if err != nil {
		return nil, ToError(404, err)
	}

	return repo.projRepo.GetProjectEndpointsInNamespace(ctx, show.Name, serialize.Serialize(show.Projects, func(v *Project) int { return v.ID })...)
}
