package services

import (
	"context"
	"errors"

	"github.com/duc-cnzj/mars/api/v4/event"
	"github.com/duc-cnzj/mars/api/v4/types"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/scopes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		event.RegisterEventServer(s, new(eventSvc))
	})
	RegisterEndpoint(event.RegisterEventHandlerFromEndpoint)
}

type eventSvc struct {
	event.UnimplementedEventServer
}

func (e *eventSvc) List(ctx context.Context, request *event.ListRequest) (*event.ListResponse, error) {
	var (
		page     = int(request.Page)
		pageSize = int(request.PageSize)
		events   []eventDiff
	)

	queryScope := func(db *gorm.DB) *gorm.DB {
		if request.ActionType != types.EventActionType_Unknown {
			db = db.Where("`action` = ?", request.GetActionType())
		}

		// 全表扫了，很慢
		if request.Search != "" {
			db = db.Where("`message` LIKE ? or `username` LIKE ?", "%"+request.Search+"%", request.Search+"%")
		}

		return db
	}

	if err := app.DB().Table("events").Preload("File", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID", "Size")
	}).Scopes(queryScope, scopes.Paginate(&page, &pageSize)).Select([]string{
		"id", "action", "username", "message", "duration", "file_id", "created_at", "updated_at",
		"(`old` != `new`) as has_diff",
	}).Order("`id` DESC").Find(&events).Error; err != nil {
		return nil, err
	}
	res := make([]*types.EventModel, 0, len(events))
	for _, m := range events {
		protoModel := m.ProtoTransform()
		protoModel.HasDiff = m.HasDiff
		res = append(res, protoModel)
	}

	return &event.ListResponse{
		Page:     int64(page),
		PageSize: int64(pageSize),
		Items:    res,
	}, nil
}

func (e *eventSvc) Show(ctx context.Context, request *event.ShowRequest) (*event.ShowResponse, error) {
	var eventModel models.Event
	err := app.DB().Preload("File").First(&eventModel, request.Id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &event.ShowResponse{Event: eventModel.ProtoTransform()}, nil
}

func (e *eventSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}

type eventDiff struct {
	models.Event
	HasDiff bool `json:"has_diff"`
}

func (eventDiff) TableName() string {
	return "events"
}
