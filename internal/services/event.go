package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/event"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/samber/lo"
)

var _ event.EventServer = (*eventSvc)(nil)

// eventSvc 是 event.EventServer 的 gRPC 实现：分页查询事件审计日志，经 access 校验
// 访问权限，由 NewEventSvc 构造。
type eventSvc struct {
	event.UnimplementedEventServer

	logger    mlog.Logger
	eventBiz  biz.EventBiz
	accessBiz biz.AccessBiz
}

// EventSvcDeps 收口 NewEventSvc 的构造依赖，由 wire 按字段注入。
type EventSvcDeps struct {
	Logger    mlog.Logger
	EventBiz  biz.EventBiz
	AccessBiz biz.AccessBiz
}

// NewEventSvc 收口事件/审计日志服务的构造依赖，由 wire 按字段注入。
func NewEventSvc(deps EventSvcDeps) event.EventServer {
	return &eventSvc{eventBiz: deps.EventBiz, logger: deps.Logger.WithModule("services/event"), accessBiz: deps.AccessBiz}
}

// List 分页查询事件（审计日志），支持按动作类型过滤与关键字搜索，按 id 倒序返回。
func (e *eventSvc) List(ctx context.Context, request *event.ListRequest) (*event.ListResponse, error) {
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	events, pag, err := e.eventBiz.List(ctx, &biz.ListEventInput{
		Page:        page,
		PageSize:    size,
		ActionType:  request.ActionType,
		Search:      request.Search,
		OrderIDDesc: lo.ToPtr(true),
	})
	if err != nil {
		return nil, logError(ctx, e.logger, err)
	}

	return &event.ListResponse{
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Items:    slice.Map(events, transformer.FromEvent),
	}, nil
}

// Show 按 id 查询单条事件详情。
func (e *eventSvc) Show(ctx context.Context, request *event.ShowRequest) (*event.ShowResponse, error) {
	show, err := e.eventBiz.Show(ctx, int(request.Id))
	if err != nil {
		return nil, logError(ctx, e.logger, err)
	}

	return &event.ShowResponse{Item: transformer.FromEvent(show)}, nil
}

// Authorize 是事件服务的 admin 门禁：审计日志仅管理员可读，其余调用一律拒绝。
func (e *eventSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	return e.accessBiz.RequireAdmin(ctx, fullMethodName)
}
