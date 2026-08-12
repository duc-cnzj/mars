package data

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/accesstoken"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v6/internal/data/filters"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/google/uuid"
)

// toAccessToken 把 ent.AccessToken 转换为 biz.AccessToken（nil 安全）。
func toAccessToken(token *ent.AccessToken) *biz.AccessToken {
	if token == nil {
		return nil
	}
	return &biz.AccessToken{
		ID:         token.ID,
		CreatedAt:  token.CreatedAt,
		UpdatedAt:  token.UpdatedAt,
		DeletedAt:  token.DeletedAt,
		Token:      token.Token,
		Usage:      token.Usage,
		Email:      token.Email,
		ExpiredAt:  token.ExpiredAt,
		LastUsedAt: token.LastUsedAt,
		UserInfo:   token.UserInfo,
	}
}

var _ biz.AccessTokenRepo = (*accessTokenRepo)(nil)

// accessTokenRepo 是 access token 的持久化实现：负责访问令牌的列表/签发/
// 查询/续期/撤销/最近使用回写，基于 ent 访问数据存储。
type accessTokenRepo struct {
	data  dataStore
	timer timer.Timer
}

// NewAccessTokenRepo 构造 access token repo。
func NewAccessTokenRepo(data dataStore, timer timer.Timer) biz.AccessTokenRepo {
	return &accessTokenRepo{data: data, timer: timer}
}

// List 分页查询 access token 列表，支持邮箱过滤；
// WithSoftDelete 时跳过软删除过滤，包含已删除记录。
func (r *accessTokenRepo) List(ctx context.Context, input *biz.ListAccessTokenInput) ([]*biz.AccessToken, *pagination.Pagination, error) {
	db := r.data.DB()
	if input.WithSoftDelete {
		ctx = mixin.SkipSoftDelete(ctx)
	}
	query := db.AccessToken.Query().
		Where(filters.IfEmail(input.Email))

	tokens, err := query.Clone().
		Order(ent.Desc(accesstoken.FieldID)).
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).
		All(ctx)
	if err != nil {
		return nil, nil, errs.Wrap(err, "list access tokens")
	}
	count := query.Clone().CountX(ctx)
	return slice.Map(tokens, toAccessToken), pagination.NewPagination(input.Page, input.PageSize, count), nil
}

// Grant 签发一个新的 access token：生成随机 token，写入用户/用途/过期时间。
func (r *accessTokenRepo) Grant(ctx context.Context, input *biz.GrantAccessTokenInput) (*biz.AccessToken, error) {
	db := r.data.DB()
	save, err := db.AccessToken.Create().
		SetToken(uuid.NewString()).
		SetEmail(input.User.Email).
		SetUsage(input.Usage).
		SetNillableUserInfo(input.User).
		SetExpiredAt(r.timer.Now().Add(time.Duration(input.ExpireSeconds) * time.Second)).
		Save(ctx)
	return toAccessToken(save), errs.Wrap(err, "grant access token")
}

// FindByToken 按 token 精确查询一个 access token，不存在返回 NotFound 错误。
func (r *accessTokenRepo) FindByToken(ctx context.Context, token string) (*biz.AccessToken, error) {
	db := r.data.DB()
	first, err := db.AccessToken.Query().Where(accesstoken.Token(token)).First(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query access token")
	}
	return toAccessToken(first), nil
}

// UpdateExpiresAt 更新指定 token 的过期时间并返回更新后的记录。
func (r *accessTokenRepo) UpdateExpiresAt(ctx context.Context, token string, t time.Time) (*biz.AccessToken, error) {
	db := r.data.DB()
	first, err := db.AccessToken.Query().Where(accesstoken.Token(token)).First(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query access token")
	}
	save, err := first.Update().SetExpiredAt(t).Save(ctx)
	return toAccessToken(save), errs.Wrap(err, "update access token expire")
}

// Revoke 撤销一个 access token（软删除）。
func (r *accessTokenRepo) Revoke(ctx context.Context, token string) error {
	db := r.data.DB()
	_, err := db.AccessToken.Delete().Where(accesstoken.Token(token)).Exec(ctx)
	return errs.Wrap(err, "revoke access token")
}

// TouchLastUsedAt 回写 token 的最近使用时间。
func (r *accessTokenRepo) TouchLastUsedAt(ctx context.Context, token string, t time.Time) error {
	return errs.Wrap(r.data.DB().AccessToken.Update().Where(accesstoken.Token(token)).SetLastUsedAt(t).Exec(ctx), "touch access token last used")
}
