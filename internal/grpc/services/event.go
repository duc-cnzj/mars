package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/scopes"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/event"
)

type EventSvc struct {
	event.UnsafeEventServer
}

func (e *EventSvc) List(ctx context.Context, request *event.EventListRequest) (*event.EventListResponse, error) {
	var (
		page     = int(request.Page)
		pageSize = int(request.PageSize)
		events   []models.Event
		count    int64
	)

	if err := app.DB().Scopes(scopes.Paginate(&page, &pageSize)).Order("`id` DESC").Find(&events).Error; err != nil {
		return nil, err
	}
	app.DB().Model(&models.Event{}).Count(&count)
	res := make([]*event.EventListItem, 0, len(events))
	for _, m := range events {
		res = append(res, &event.EventListItem{
			Id:       int64(m.ID),
			Action:   event.ActionType(m.Action),
			Username: m.Username,
			Message:  m.Message,
			Old:      m.Old,
			New:      m.New,
			EventAt:  utils.ToHumanizeDatetimeString(&m.CreatedAt),
		})
	}

	return &event.EventListResponse{
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
