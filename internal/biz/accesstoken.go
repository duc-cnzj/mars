package biz

import (
	"context"
	"errors"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
)

// AccessTokenBiz 收口 access token 生命周期业务：签发、续租、撤销与使用回写。
type AccessTokenBiz interface {
	// List 分页列出 access token。
	List(ctx context.Context, input *ListAccessTokenInput) ([]*AccessToken, *pagination.Pagination, error)
	// Grant 签发新的 access token。
	Grant(ctx context.Context, input *GrantAccessTokenInput) (*AccessToken, error)
	// Lease 续租 token 的过期时间。
	Lease(ctx context.Context, token string, expireSeconds int32) (*AccessToken, error)
	// Revoke 撤销（软删除）指定 access token。
	Revoke(ctx context.Context, token string) error
	// FindByToken 按 token 原文查询一个 access token。
	FindByToken(ctx context.Context, token string) (*AccessToken, error)
	// TouchLastUsedAt 回写 token 的最近使用时间。
	TouchLastUsedAt(ctx context.Context, token string, t time.Time) error
}

type accessTokenBiz struct {
	logger mlog.Logger
	timer  timer.Timer
	repo   AccessTokenRepo
}

// NewAccessTokenBiz 构造 access token biz。
func NewAccessTokenBiz(logger mlog.Logger, timer timer.Timer, repo AccessTokenRepo) AccessTokenBiz {
	return &accessTokenBiz{logger: logger.WithModule("biz/accessToken"), timer: timer, repo: repo}
}

// List 分页列出 access token（透传 repo）。
func (a *accessTokenBiz) List(ctx context.Context, input *ListAccessTokenInput) ([]*AccessToken, *pagination.Pagination, error) {
	return a.repo.List(ctx, input)
}

// Grant 校验输入后签发新的 access token。
func (a *accessTokenBiz) Grant(ctx context.Context, input *GrantAccessTokenInput) (*AccessToken, error) {
	if input == nil || input.User == nil {
		return nil, errs.WrapInvalidArgument(errors.New("grant input 或 user 不能为空"), "grant access token")
	}
	if input.ExpireSeconds <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("expireSeconds 必须大于 0"), "grant access token")
	}
	return a.repo.Grant(ctx, input)
}

// Lease 校验 token 后续租：查询原 token，已过期则拒绝，否则延长过期时间。
func (a *accessTokenBiz) Lease(ctx context.Context, token string, expireSeconds int32) (*AccessToken, error) {
	if token == "" {
		return nil, errs.WrapInvalidArgument(errors.New("token 不能为空"), "lease access token")
	}
	at, err := a.repo.FindByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if at.IsExpired(a.timer.Now()) {
		return nil, errs.WrapInvalidArgument(errors.New("token 已经过期"), "lease expired token")
	}
	return a.repo.UpdateExpiresAt(ctx, token, a.timer.Now().Add(time.Duration(expireSeconds)*time.Second))
}

// Revoke 校验 token 后撤销（软删除）指定 access token。
func (a *accessTokenBiz) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return errs.WrapInvalidArgument(errors.New("token 不能为空"), "revoke access token")
	}
	return a.repo.Revoke(ctx, token)
}

// FindByToken 按 token 查询一个 access token（透传 repo）。
func (a *accessTokenBiz) FindByToken(ctx context.Context, token string) (*AccessToken, error) {
	return a.repo.FindByToken(ctx, token)
}

// TouchLastUsedAt 校验 token 后回写最近使用时间。
func (a *accessTokenBiz) TouchLastUsedAt(ctx context.Context, token string, t time.Time) error {
	if token == "" {
		return errs.WrapInvalidArgument(errors.New("token 不能为空"), "touch last used at")
	}
	return a.repo.TouchLastUsedAt(ctx, token, t)
}

// AccessTokenRepo 是访问令牌（access token）持久化与生命周期操作的仓库端口。
type AccessTokenRepo interface {
	// List 分页列出 access token（可按邮箱过滤、含软删除项）。
	List(ctx context.Context, input *ListAccessTokenInput) ([]*AccessToken, *pagination.Pagination, error)
	// Grant 签发新的 access token。
	Grant(ctx context.Context, input *GrantAccessTokenInput) (*AccessToken, error)
	// FindByToken 按 token 原文查询一个 access token。
	FindByToken(ctx context.Context, token string) (*AccessToken, error)
	// UpdateExpiresAt 更新 token 的过期时间（续租）。
	UpdateExpiresAt(ctx context.Context, token string, t time.Time) (*AccessToken, error)
	// Revoke 撤销（软删除）指定 access token。
	Revoke(ctx context.Context, token string) error
	// TouchLastUsedAt 回写 token 的最近使用时间。
	TouchLastUsedAt(ctx context.Context, token string, t time.Time) error
}
