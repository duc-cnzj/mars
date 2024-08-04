package services

import (
	"context"

	"github.com/duc-cnzj/mars/v4/internal/util/pagination"

	"github.com/duc-cnzj/mars/api/v4/event"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ event.EventServer = (*eventSvc)(nil)

type eventSvc struct {
	event.UnimplementedEventServer

	eventRepo repo.EventRepo
}

func NewEventSvc(eventRepo repo.EventRepo) event.EventServer {
	return &eventSvc{eventRepo: eventRepo}
}

func (e *eventSvc) List(ctx context.Context, request *event.ListRequest) (*event.ListResponse, error) {
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	events, pag, err := e.eventRepo.List(ctx, &repo.ListEventInput{
		Page:        page,
		PageSize:    size,
		ActionType:  request.ActionType,
		Search:      request.Search,
		OrderIDDesc: lo.ToPtr(true),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*types.EventModel, 0, len(events))
	for _, m := range events {
		res = append(res, transformer.FromEvent(m))
	}

	return &event.ListResponse{
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Items:    res,
	}, nil
}

func (e *eventSvc) Show(ctx context.Context, request *event.ShowRequest) (*event.ShowResponse, error) {
	show, err := e.eventRepo.Show(ctx, int(request.Id))
	if err != nil {
		return nil, err
	}

	return &event.ShowResponse{Event: transformer.FromEvent(show)}, nil
}

func (e *eventSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}
