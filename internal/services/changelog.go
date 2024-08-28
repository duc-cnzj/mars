package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v5/changelog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/transformer"
	"github.com/duc-cnzj/mars/v5/internal/util/serialize"
	"github.com/samber/lo"
)

var _ changelog.ChangelogServer = (*changelogSvc)(nil)

type changelogSvc struct {
	changelog.UnimplementedChangelogServer

	repo repo.ChangelogRepo
}

func NewChangelogSvc(repo repo.ChangelogRepo) changelog.ChangelogServer {
	return &changelogSvc{repo: repo}
}

func (c *changelogSvc) FindLastChangelogsByProjectID(ctx context.Context, request *changelog.FindLastChangelogsByProjectIDRequest) (*changelog.FindLastChangelogsByProjectIDResponse, error) {
	logs, err := c.repo.FindLastChangelogsByProjectID(ctx, &repo.FindLastChangelogsByProjectIDChangeLogInput{
		OnlyChanged:        request.OnlyChanged,
		ProjectID:          int(request.ProjectId),
		OrderByVersionDesc: lo.ToPtr(true),
		Limit:              5,
	})
	if err != nil {
		return nil, err
	}

	return &changelog.FindLastChangelogsByProjectIDResponse{
		Items: serialize.Serialize(logs, transformer.FromChangeLog),
	}, nil
}
