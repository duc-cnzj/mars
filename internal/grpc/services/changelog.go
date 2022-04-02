package services

import (
	"context"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars-client/v4/changelog"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		changelog.RegisterChangelogServer(s, new(ChangelogSvc))
	})
	RegisterEndpoint(changelog.RegisterChangelogHandlerFromEndpoint)
}

type ChangelogSvc struct {
	changelog.UnimplementedChangelogServer
}

func (c *ChangelogSvc) Show(ctx context.Context, request *changelog.ChangelogShowRequest) (*changelog.ChangelogShowResponse, error) {
	var logs []models.Changelog
	err := app.DB().
		Scopes(func(db *gorm.DB) *gorm.DB {
			if request.OnlyChanged {
				return db.Where("`config_changed` = ?", true)
			}
			return db
		}).
		Where("`project_id` = ?", request.ProjectId).
		Order("`version` DESC").
		Limit(5).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	items := make([]*changelog.ChangelogShowItem, 0, len(logs))
	for _, log := range logs {
		items = append(items, &changelog.ChangelogShowItem{
			Version:  int32(log.Version),
			Config:   log.Config,
			Date:     utils.ToHumanizeDatetimeString(&log.CreatedAt),
			Username: log.Username,
		})
	}

	return &changelog.ChangelogShowResponse{Items: items}, nil
}
