package services

import (
	"context"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/api/v4/changelog"
	"github.com/duc-cnzj/mars/api/v4/types"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/models"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		changelog.RegisterChangelogServer(s, new(changelogSvc))
	})
	RegisterEndpoint(changelog.RegisterChangelogHandlerFromEndpoint)
}

type changelogSvc struct {
	changelog.UnimplementedChangelogServer
}

func (c *changelogSvc) Show(ctx context.Context, request *changelog.ShowRequest) (*changelog.ShowResponse, error) {
	var logs []models.Changelog
	err := app.DB().
		Scopes(func(db *gorm.DB) *gorm.DB {
			if request.OnlyChanged {
				return db.Where("`config_changed` = ?", true)
			}
			return db
		}).
		Select(
			"id",
			"version",
			"username",
			"config",
			"config_changed",
			"project_id",
			"git_project_id",
			"git_commit_title",
			"git_commit_web_url",
			"git_commit_date",
			"git_commit_author",
			"created_at",
		).
		Where("`project_id` = ?", request.ProjectId).
		Order("`version` DESC").
		Limit(5).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	items := make([]*types.ChangelogModel, 0, len(logs))
	for _, log := range logs {
		items = append(items, log.ProtoTransform())
	}

	return &changelog.ShowResponse{Items: items}, nil
}
