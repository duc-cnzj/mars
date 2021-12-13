package services

import (
	"context"

	"gorm.io/gorm"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/changelog"
)

type Changelog struct {
	changelog.UnimplementedChangelogServer
}

func (c *Changelog) Get(ctx context.Context, request *changelog.ChangelogGetRequest) (*changelog.ChangelogGetResponse, error) {
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
	items := make([]*changelog.ChangelogGetResponse_Item, 0, len(logs))
	for _, log := range logs {
		items = append(items, &changelog.ChangelogGetResponse_Item{
			Version:  int32(log.Version),
			Config:   log.Config,
			Date:     utils.ToHumanizeDatetimeString(&log.CreatedAt),
			Username: log.Username,
		})
	}

	return &changelog.ChangelogGetResponse{Items: items}, nil
}
