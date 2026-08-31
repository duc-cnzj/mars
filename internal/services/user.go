package services

import (
	"context"

	"github.com/duc-cnzj/mars/api/v6/proto/user"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
)

var _ user.UserServer = (*userSvc)(nil)

// userSvc 是 user.UserServer 的 gRPC 实现：管理员后台用户管理（列表/角色升降级），
// 整服务 admin 专用，由 NewUserSvc 构造。
type userSvc struct {
	user.UnimplementedUserServer

	userBiz   biz.UserBiz
	accessBiz biz.AccessBiz
	logger    mlog.Logger
}

// UserSvcDeps 收口 NewUserSvc 的构造依赖，由 wire 按字段注入。
type UserSvcDeps struct {
	UserBiz   biz.UserBiz
	AccessBiz biz.AccessBiz
	Logger    mlog.Logger
}

// NewUserSvc 收口用户管理服务的构造依赖，由 wire 按字段注入。
func NewUserSvc(deps UserSvcDeps) user.UserServer {
	return &userSvc{userBiz: deps.UserBiz, accessBiz: deps.AccessBiz, logger: deps.Logger.WithModule("services/user")}
}

// Authorize 校验访问权限：用户管理为管理员专用，无白名单放行方法。
func (s *userSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	return s.accessBiz.RequireAdmin(ctx, fullMethodName)
}

// List 分页返回用户列表（含统计），role=admin 时只看管理员。
func (s *userSvc) List(ctx context.Context, request *user.ListRequest) (*user.ListResponse, error) {
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	list, err := s.userBiz.List(ctx, &biz.ListUserInput{
		Page:      page,
		PageSize:  size,
		Search:    request.Search,
		AdminOnly: request.Role == "admin",
	})
	if err != nil {
		return nil, logError(ctx, s.logger, err)
	}

	items := make([]*user.UserModel, 0, len(list.Items))
	for _, u := range list.Items {
		items = append(items, transformer.FromUser(u))
	}
	return &user.ListResponse{
		Page:     list.Pag.Page,
		PageSize: list.Pag.PageSize,
		Count:    list.Pag.Count,
		Items:    items,
		Stats: &user.UserStats{
			Total:   list.Stats.Total,
			Admins:  list.Stats.Admins,
			Regular: list.Stats.Regular,
		},
	}, nil
}

// ToggleAdmin 设置/移除指定用户的管理员角色。
func (s *userSvc) ToggleAdmin(ctx context.Context, request *user.ToggleAdminRequest) (*user.ToggleAdminResponse, error) {
	if err := s.userBiz.ToggleAdmin(ctx, request.Email, request.Admin); err != nil {
		return nil, logError(ctx, s.logger, err)
	}
	return &user.ToggleAdminResponse{}, nil
}
