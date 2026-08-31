package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/event"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/samber/lo"
)

var _ event.EventServer = (*eventSvc)(nil)

// eventSvc 是 event.EventServer 的 gRPC 实现：分页查询事件审计日志，由 NewEventSvc 构造。
// 鉴权约定：admin 可读全量事件；普通登录用户仅可读操作人邮箱（operator_email）为自己
// 的事件，归属过滤在方法体内用 ctx 用户身份推导，请求参数不参与，防止越权枚举他人事件。
type eventSvc struct {
	event.UnimplementedEventServer

	logger   mlog.Logger
	eventBiz biz.EventBiz
}

// EventSvcDeps 收口 NewEventSvc 的构造依赖，由 wire 按字段注入。
type EventSvcDeps struct {
	Logger   mlog.Logger
	EventBiz biz.EventBiz
}

// NewEventSvc 收口事件/审计日志服务的构造依赖，由 wire 按字段注入。
func NewEventSvc(deps EventSvcDeps) event.EventServer {
	return &eventSvc{eventBiz: deps.EventBiz, logger: deps.Logger.WithModule("services/event")}
}

// normalizeActionTypes 归一化动作类型过滤：优先多值 action_types，否则回退单值 action_type（Unknown=全部）。
func normalizeActionTypes(actionType types.EventActionType, actionTypes []types.EventActionType) []types.EventActionType {
	if len(actionTypes) > 0 {
		return actionTypes
	}
	if actionType != types.EventActionType_Unknown {
		return []types.EventActionType{actionType}
	}
	return nil
}

// List 分页查询事件（审计日志），支持按动作类型过滤与关键字搜索，按 id 倒序返回。
// 默认只看操作人邮箱（event.operator_email）为自己的事件，归属过滤条件由 ctx 用户身份
// 推导注入，不接受请求参数——否则可传他人邮箱枚举全部事件。仅 admin 显式传 all 才展开
// 全量（镜像 access_token 的 all 语义），普通用户传 all 等效无操作（仍收敛本人）。
func (e *eventSvc) List(ctx context.Context, request *event.ListRequest) (*event.ListResponse, error) {
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	input := &biz.ListEventInput{
		Page:        page,
		PageSize:    size,
		ActionTypes: normalizeActionTypes(request.ActionType, request.ActionTypes),
		Search:      request.Search,
		OrderIDDesc: lo.ToPtr(true),
	}
	user := biz.MustGetUser(ctx)
	if !user.IsAdmin() || !request.All {
		// 归属过滤按邮箱等值（filters.IfStrEQ 对空串不过滤）：无邮箱的普通用户
		// 没有可归属事件，直接返回空列表，防止空串触发"不过滤"语义导致全量可见。
		if user.Email == "" {
			// 与正常空结果路径（slice.Map 返回空 slice）一致：Items 用空 slice 而非 nil，
			// 避免 JSON 序列化出现 null 与 [] 的响应形状不一致。
			return &event.ListResponse{Page: page, PageSize: size, Items: []*types.EventModel{}}, nil
		}
		input.OperatorEmail = &user.Email
	}
	events, pag, err := e.eventBiz.List(ctx, input)
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
// admin 可看任意事件；普通用户仅可查看操作人邮箱为自己的事件，越权访问与不存在
// 一律返回 404（视同不存在，防审计日志 id 可枚举导致的存在性侧信道）。
// 无邮箱的普通用户一律 404，与 List 的"空邮箱返回空列表"对齐：operator_email 为空的
// 事件（迁移前历史行 / cron 系统事件）无法与本人事件区分，等值比较会退化为全量可见。
func (e *eventSvc) Show(ctx context.Context, request *event.ShowRequest) (*event.ShowResponse, error) {
	show, err := e.eventBiz.Show(ctx, int(request.Id))
	if err != nil {
		return nil, logError(ctx, e.logger, err)
	}
	user := biz.MustGetUser(ctx)
	if !user.IsAdmin() && (user.Email == "" || show.OperatorEmail != user.Email) {
		return nil, errs.NotFound("事件不存在")
	}

	return &event.ShowResponse{Item: transformer.FromEvent(show)}, nil
}
