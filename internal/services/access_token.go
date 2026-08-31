package services

import (
	"context"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/token"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
)

// dateTimeFormat 审计日志统一的时间格式。
const dateTimeFormat = "2006-01-02 15:04:05"

var _ token.AccessTokenServer = (*accessTokenSvc)(nil)

// accessTokenSvc 是 token.AccessTokenServer 的 gRPC 实现：聚合 access token 仓库与
// 事件审计（签发/续租/撤销均落审计日志），由 NewAccessTokenSvc 构造。
type accessTokenSvc struct {
	token.UnimplementedAccessTokenServer

	logger   mlog.Logger
	repo     biz.AccessTokenBiz
	eventBiz biz.EventBiz
}

// AccessTokenSvcDeps 收口 NewAccessTokenSvc 的构造依赖，由 wire 按字段注入。
type AccessTokenSvcDeps struct {
	Logger   mlog.Logger
	EventBiz biz.EventBiz
	Repo     biz.AccessTokenBiz
}

// NewAccessTokenSvc 收口 access token 服务的构造依赖，由 wire 按字段注入。
func NewAccessTokenSvc(deps AccessTokenSvcDeps) token.AccessTokenServer {
	return &accessTokenSvc{
		logger:   deps.Logger.WithModule("services/accessToken"),
		eventBiz: deps.EventBiz,
		repo:     deps.Repo,
	}
}

// List 分页列出 access token（含软删除）：默认收敛到本人——所有用户（含 admin）都
// 只见自己创建的令牌（最小权限，延续旧非 admin 契约）；仅 admin 显式传 all 才展开
// 全部用户令牌的全量视图，非 admin 传 all 等效无操作。admin 判定复用
// UserInfo.IsAdmin（同 access 门卫判据）。
// 令牌返回完整值——列表承载复制/撤销/续租三个功能，若返回掩码则三者全废（撤销匹配
// 0 行假成功、续租 NotFound、复制无效密钥）；安全权衡：完整值只对 admin 全量视图与
// 本人可见（admin 为最高可信角色，管理操作本就需要完整值），视觉脱敏由前端展示层
// 承担（列表防肩窥），maskToken 仅保留给审计日志，不让明文密钥落日志。
func (a *accessTokenSvc) List(ctx context.Context, request *token.ListRequest) (*token.ListResponse, error) {
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	user := biz.MustGetUser(ctx)
	// 状态过滤参数边界校验：公开 HTTP 参数，未知值直接 400 拒绝（对齐 valid/expired/revoked 三态），
	// 避免静默吞错把「打错的值」当成「不过滤」改变查询语义。
	switch request.Status {
	case "", "valid", "expired", "revoked":
	default:
		return nil, logError(ctx, a.logger, errs.WrapInvalidArgument(fmt.Errorf("unknown access token status %q", request.Status), "list access tokens"))
	}
	// 默认：全部用户只看本人（Email=当前用户，延续旧非 admin 契约）；仅 admin 显式传
	// all 才展开全量（Email="" 不过滤），普通用户传 all 等效无操作（仍收敛到本人）。
	input := &biz.ListAccessTokenInput{
		Page:           page,
		PageSize:       size,
		Email:          user.Email,
		Search:         request.Search,
		WithSoftDelete: true,
		Status:         request.Status,
	}
	if user.IsAdmin() && request.All {
		input.Email = ""
	}
	tokens, p, err := a.repo.List(ctx, input)
	if err != nil {
		return nil, logError(ctx, a.logger, err)
	}
	items := slice.Map(tokens, transformer.FromAccessToken)

	return &token.ListResponse{
		Page:     p.Page,
		PageSize: p.PageSize,
		Items:    items,
		Count:    p.Count,
	}, nil
}

// Grant 为当前用户签发 access token 并落创建审计日志，token 以掩码形式入日志。
func (a *accessTokenSvc) Grant(ctx context.Context, request *token.GrantRequest) (*token.GrantResponse, error) {
	user := biz.MustGetUser(ctx)
	at, err := a.repo.Grant(ctx, &biz.GrantAccessTokenInput{
		ExpireSeconds: request.ExpireSeconds,
		Usage:         request.Usage,
		User:          user,
	})
	if err != nil {
		return nil, logError(ctx, a.logger, err)
	}

	a.eventBiz.AuditLogWithRequest(
		types.EventActionType_Create,
		user.Name,
		user.Email,
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 创建了一个 token "%s", 过期时间是 "%s".`, user.Name, maskToken(at.Token), at.ExpiredAt.Format(dateTimeFormat)),
		request,
	)

	return &token.GrantResponse{Token: transformer.FromAccessToken(at)}, nil
}

// Lease 续租现有 token 的过期时间并落更新审计日志。
func (a *accessTokenSvc) Lease(ctx context.Context, request *token.LeaseRequest) (*token.LeaseResponse, error) {
	user := biz.MustGetUser(ctx)

	at, err := a.repo.Lease(ctx, request.Token, request.ExpireSeconds)
	if err != nil {
		return nil, logError(ctx, a.logger, err)
	}
	a.eventBiz.AuditLogWithRequest(
		types.EventActionType_Update,
		user.Name,
		user.Email,
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 续租了 token "%s", 增加了 "%s", 过期时间是 "%s".`, user.Name, maskToken(at.Token), date.HumanDuration(time.Second*time.Duration(request.ExpireSeconds)), at.ExpiredAt.Format(dateTimeFormat)),
		request,
	)

	return &token.LeaseResponse{Token: transformer.FromAccessToken(at)}, nil
}

// maskToken 只保留 token 前后各 4 位字符，避免完整密钥落入审计日志。
func maskToken(token string) string {
	if len(token) <= 8 {
		return "******"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// Revoke 撤销指定 token（软删除）并落删除审计日志。
func (a *accessTokenSvc) Revoke(ctx context.Context, request *token.RevokeRequest) (*token.RevokeResponse, error) {
	user := biz.MustGetUser(ctx)
	if err := a.repo.Revoke(ctx, request.Token); err != nil {
		return nil, logError(ctx, a.logger, err)
	}
	a.eventBiz.AuditLogWithRequest(
		types.EventActionType_Delete,
		user.Name,
		user.Email,
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 删除 token "%s".`, user.Name, maskToken(request.Token)),
		request,
	)

	return &token.RevokeResponse{}, nil
}
