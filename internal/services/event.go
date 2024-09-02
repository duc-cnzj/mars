package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v5/event"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/transformer"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/duc-cnzj/mars/v5/internal/util/serialize"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ event.EventServer = (*eventSvc)(nil)

type eventSvc struct {
	event.UnimplementedEventServer

	logger    mlog.Logger
	eventRepo repo.EventRepo
}

func NewEventSvc(logger mlog.Logger, eventRepo repo.EventRepo) event.EventServer {
	return &eventSvc{eventRepo: eventRepo, logger: logger.WithModule("services/event")}
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
		e.logger.ErrorCtx(ctx, err)
		return nil, err
	}

	return &event.ListResponse{
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Items:    serialize.Serialize(events, transformer.FromEvent),
	}, nil
}

func (e *eventSvc) Show(ctx context.Context, request *event.ShowRequest) (*event.ShowResponse, error) {
	show, err := e.eventRepo.Show(ctx, int(request.Id))
	if err != nil {
		e.logger.ErrorCtx(ctx, err)
		return nil, err
	}

	return &event.ShowResponse{Item: transformer.FromEvent(show)}, nil
}

func (e *eventSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}
