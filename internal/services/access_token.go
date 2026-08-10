package services

import (
	"context"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/token"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/biz"
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

// List 分页列出当前用户的 access token（含软删除，只返回本人数据）。
func (a *accessTokenSvc) List(ctx context.Context, request *token.ListRequest) (*token.ListResponse, error) {
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	tokens, p, err := a.repo.List(ctx, &biz.ListAccessTokenInput{
		Page:           page,
		PageSize:       size,
		Email:          biz.MustGetUser(ctx).Email,
		WithSoftDelete: true,
	})
	if err != nil {
		return nil, logError(ctx, a.logger, err)
	}

	return &token.ListResponse{
		Page:     p.Page,
		PageSize: p.PageSize,
		Items:    slice.Map(tokens, transformer.FromAccessToken),
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
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 删除 token "%s".`, user.Name, maskToken(request.Token)),
		request,
	)

	return &token.RevokeResponse{}, nil
}
