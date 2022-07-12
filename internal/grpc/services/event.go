package services

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars-client/v4/event"
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/scopes"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		event.RegisterEventServer(s, new(EventSvc))
	})
	RegisterEndpoint(event.RegisterEventHandlerFromEndpoint)
}

type EventSvc struct {
	event.UnsafeEventServer
}

func (e *EventSvc) List(ctx context.Context, request *event.ListRequest) (*event.ListResponse, error) {
	var (
		page     = int(request.Page)
		pageSize = int(request.PageSize)
		events   []models.Event
		count    int64
	)

	queryScope := func(db *gorm.DB) *gorm.DB {
		if request.ActionType != types.EventActionType_Unknown {
			db = db.Where("`action` = ?", request.GetActionType())
		}

		if request.Message != "" {
			db = db.Where("`message` LIKE ?", "%"+request.Message+"%")
		}

		return db
	}

	if err := app.DB().Preload("File", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID")
	}).Scopes(queryScope, scopes.Paginate(&page, &pageSize)).Order("`id` DESC").Find(&events).Error; err != nil {
		return nil, err
	}
	app.DB().Model(&models.Event{}).Scopes(queryScope).Count(&count)
	res := make([]*types.EventModel, 0, len(events))
	for _, m := range events {
		res = append(res, m.ProtoTransform())
	}

	return &event.ListResponse{
		Page:     int64(page),
		PageSize: int64(pageSize),
		Items:    res,
		Count:    count,
	}, nil
}

func (e *EventSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}
