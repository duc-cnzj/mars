package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v4/changelog"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
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
	items := make([]*types.ChangelogModel, 0, len(logs))
	for _, log := range logs {
		items = append(items, transformer.FromChangeLog(log))
	}

	return &changelog.FindLastChangelogsByProjectIDResponse{Items: items}, nil
}
