package services

import (
	"context"
	"fmt"

	hibernatepb "github.com/duc-cnzj/mars/api/v6/proto/hibernate"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

var _ hibernatepb.HibernateServer = (*hibernateSvc)(nil)

// hibernateSvc 是 hibernate.HibernateServer 的 gRPC 实现：管理员在后台维护全局共享的
// 休眠时间窗（命名时段），各仓库按名字引用。经 Authorize 分级门禁：**查询**（ListWindows）
// 登录即可读——项目卡片对普通用户可见，需按 window_name 展示时段，故查询放行到用户级；
// **增删改**仍限管理员（时间窗是仓库级休眠配置的前置依赖，维护权限对齐后台）。
type hibernateSvc struct {
	hibernatepb.UnimplementedHibernateServer

	hibernateWindowBiz biz.HibernateWindowBiz
	accessBiz          biz.AccessBiz
	eventBiz           biz.EventBiz
	logger             mlog.Logger
}

// HibernateSvcDeps 收口 NewHibernateSvc 的构造依赖，由 wire 按字段注入。
type HibernateSvcDeps struct {
	HibernateWindowBiz biz.HibernateWindowBiz
	AccessBiz          biz.AccessBiz
	EventBiz           biz.EventBiz
	Logger             mlog.Logger
}

// NewHibernateSvc 构造全局休眠时间窗管理服务，依赖由 wire 注入。
func NewHibernateSvc(deps HibernateSvcDeps) hibernatepb.HibernateServer {
	return &hibernateSvc{
		hibernateWindowBiz: deps.HibernateWindowBiz,
		accessBiz:          deps.AccessBiz,
		eventBiz:           deps.EventBiz,
		logger:             deps.Logger.WithModule("services/hibernate"),
	}
}

// Authorize 是服务级授权门禁：ListWindows 走 RequireAdmin 的 allowlist 豁免放行到
// 用户级（任何登录用户可读），Create/Update/Delete 未命中豁免故仍要求管理员。
// ListWindows 豁免的依据：时间窗只是命名的启止时段，不含敏感信息；项目卡片对普通
// 用户可见并需展示「窗名（时段）」，若锁死在管理员会让普通用户永远看不到时段。
func (s *hibernateSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	return s.accessBiz.RequireAdmin(ctx, fullMethodName, hibernatepb.Hibernate_ListWindows_FullMethodName)
}

// ListWindows 返回全部全局休眠时间窗（按名称升序）。登录即可读（见 Authorize）。
func (s *hibernateSvc) ListWindows(ctx context.Context, _ *hibernatepb.ListWindowsRequest) (*hibernatepb.ListWindowsResponse, error) {
	list, err := s.hibernateWindowBiz.List(ctx)
	if err != nil {
		return nil, logError(ctx, s.logger, err)
	}
	resp := &hibernatepb.ListWindowsResponse{Items: make([]*hibernatepb.HibernateWindow, 0, len(list))}
	for _, w := range list {
		resp.Items = append(resp.Items, toHibernateWindowProto(w))
	}
	return resp, nil
}

// CreateWindow 新建全局休眠时间窗：名称重复返回 409，时段非法返回 400，落创建审计日志。
func (s *hibernateSvc) CreateWindow(ctx context.Context, req *hibernatepb.CreateWindowRequest) (*hibernatepb.CreateWindowResponse, error) {
	created, err := s.hibernateWindowBiz.Create(ctx, &biz.HibernateWindow{
		Name:        req.Name,
		Start:       req.Start,
		End:         req.End,
		Description: req.Description,
	})
	if err != nil {
		return nil, logError(ctx, s.logger, err)
	}
	user := biz.MustGetUser(ctx)
	s.eventBiz.AuditLogWithRequest(
		types.EventActionType_Create,
		user.Name,
		user.Email,
		fmt.Sprintf("创建休眠时间窗: %d: %s", created.ID, created.Name),
		req,
	)
	return &hibernatepb.CreateWindowResponse{Window: toHibernateWindowProto(created)}, nil
}

// UpdateWindow 更新全局休眠时间窗：不存在返回 404，名称重复返回 409，时段非法返回 400，落更新审计日志。
// 改名时按 req.Cascade 分流：级联则迁移引用旧名的仓库并回报迁移数，否则旧名仍被引用则拒绝（400）。
func (s *hibernateSvc) UpdateWindow(ctx context.Context, req *hibernatepb.UpdateWindowRequest) (*hibernatepb.UpdateWindowResponse, error) {
	updated, migrated, err := s.hibernateWindowBiz.Update(ctx, &biz.HibernateWindow{
		ID:          int(req.Id),
		Name:        req.Name,
		Start:       req.Start,
		End:         req.End,
		Description: req.Description,
	}, req.Cascade)
	if err != nil {
		return nil, logError(ctx, s.logger, err)
	}
	user := biz.MustGetUser(ctx)
	s.eventBiz.AuditLogWithRequest(
		types.EventActionType_Update,
		user.Name,
		user.Email,
		hibernateWindowUpdateAuditMsg(updated, migrated),
		req,
	)
	return &hibernatepb.UpdateWindowResponse{
		Window:           toHibernateWindowProto(updated),
		UpdatedRepoCount: int32(migrated),
	}, nil
}

// hibernateWindowUpdateAuditMsg 拼更新审计消息：级联迁移了引用仓库时附上迁移数量，便于事后追溯。
func hibernateWindowUpdateAuditMsg(w *biz.HibernateWindow, migrated int) string {
	if migrated > 0 {
		return fmt.Sprintf("更新休眠时间窗并级联迁移 %d 个引用仓库: %d: %s", migrated, w.ID, w.Name)
	}
	return fmt.Sprintf("更新休眠时间窗: %d: %s", w.ID, w.Name)
}

// DeleteWindow 删除全局休眠时间窗（物理删除、幂等）：仍被仓库按名引用时拒绝（400），
// 无下游关联时删除；窗口不存在时幂等成功，落删除审计日志。
func (s *hibernateSvc) DeleteWindow(ctx context.Context, req *hibernatepb.DeleteWindowRequest) (*hibernatepb.DeleteWindowResponse, error) {
	if err := s.hibernateWindowBiz.Delete(ctx, int(req.Id)); err != nil {
		return nil, logError(ctx, s.logger, err)
	}
	user := biz.MustGetUser(ctx)
	s.eventBiz.AuditLogWithRequest(
		types.EventActionType_Delete,
		user.Name,
		user.Email,
		fmt.Sprintf("删除休眠时间窗: %d", req.Id),
		req,
	)
	return &hibernatepb.DeleteWindowResponse{}, nil
}

// toHibernateWindowProto 把领域对象映射为 proto（nil 安全）。
func toHibernateWindowProto(w *biz.HibernateWindow) *hibernatepb.HibernateWindow {
	if w == nil {
		return nil
	}
	return &hibernatepb.HibernateWindow{
		Id:          int32(w.ID),
		Name:        w.Name,
		Start:       w.Start,
		End:         w.End,
		Description: w.Description,
	}
}
