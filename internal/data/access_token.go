package data

import (
	"context"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
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

// List 分页查询 access token 列表，支持邮箱精确过滤 + 邮箱/创建人显示名模糊搜索（admin 后台
// 按用户查令牌）+ 状态过滤（valid/expired/revoked，语义见 ListAccessTokenInput.Status 注释）；
// WithSoftDelete 时跳过软删除过滤，包含已删除记录。
func (r *accessTokenRepo) List(ctx context.Context, input *biz.ListAccessTokenInput) (out []*biz.AccessToken, pag *pagination.Pagination, err error) {
	ctx, span := tracer.Start(ctx, "accessTokenRepo/List")
	defer func() { endSpan(span, err) }()
	db := r.data.DB()
	if input.WithSoftDelete {
		ctx = mixin.SkipSoftDelete(ctx)
	}
	query := db.AccessToken.Query().Where(filters.IfEmail(input.Email))
	// Search：按创建人邮箱或显示名（user_info JSON 的 name 字段）模糊匹配——admin 后台
	// 「按创建人搜索」搜邮箱/name 都命中（对齐 userRepo.List 的 email OR name 语义）。
	// name 存于 JSON 列，用 JSON_EXTRACT('$.name') LIKE（大小写不敏感）命中；历史令牌
	// user_info 缺失/为空时 JSON 提取返回 NULL，NULL LIKE 恒为假，仅邮箱列兜底不误伤。
	if s := strings.TrimSpace(input.Search); s != "" {
		query = query.Where(accesstoken.Or(
			accesstoken.EmailContainsFold(s),
			func(sel *sql.Selector) {
				sel.Where(sqljson.StringContains(accesstoken.FieldUserInfo, s, sqljson.Path("name")))
			},
		))
	}
	// 状态过滤（对齐前端状态标签优先级：已撤销 > 已过期 > 有效）：
	// valid=未撤销且未过期；expired=未撤销但已过期；revoked=已撤销（软删除优先，即使同时已过期）。
	// now 取注入时钟保证可测；query 是下方 data 与 count 的共享基座，一处加、两端生效。
	switch input.Status {
	case "revoked":
		query = query.Where(accesstoken.DeletedAtNotNil())
	case "expired":
		query = query.Where(accesstoken.DeletedAtIsNil(), accesstoken.ExpiredAtLT(r.timer.Now()))
	case "valid":
		query = query.Where(accesstoken.DeletedAtIsNil(), accesstoken.ExpiredAtGTE(r.timer.Now()))
	}

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
func (r *accessTokenRepo) Grant(ctx context.Context, input *biz.GrantAccessTokenInput) (out *biz.AccessToken, err error) {
	ctx, span := tracer.Start(ctx, "accessTokenRepo/Grant")
	defer func() { endSpan(span, err) }()
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
func (r *accessTokenRepo) FindByToken(ctx context.Context, token string) (out *biz.AccessToken, err error) {
	ctx, span := tracer.Start(ctx, "accessTokenRepo/FindByToken")
	defer func() { endSpan(span, err) }()
	db := r.data.DB()
	first, err := db.AccessToken.Query().Where(accesstoken.Token(token)).First(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query access token")
	}
	return toAccessToken(first), nil
}

// UpdateExpiresAt 更新指定 token 的过期时间并返回更新后的记录。
func (r *accessTokenRepo) UpdateExpiresAt(ctx context.Context, token string, t time.Time) (out *biz.AccessToken, err error) {
	ctx, span := tracer.Start(ctx, "accessTokenRepo/UpdateExpiresAt")
	defer func() { endSpan(span, err) }()
	db := r.data.DB()
	first, err := db.AccessToken.Query().Where(accesstoken.Token(token)).First(ctx)
	if err != nil {
		return nil, errs.Wrap(err, "query access token")
	}
	save, err := first.Update().SetExpiredAt(t).Save(ctx)
	return toAccessToken(save), errs.Wrap(err, "update access token expire")
}

// Revoke 撤销一个 access token（软删除）。
func (r *accessTokenRepo) Revoke(ctx context.Context, token string) (err error) {
	ctx, span := tracer.Start(ctx, "accessTokenRepo/Revoke")
	defer func() { endSpan(span, err) }()
	db := r.data.DB()
	_, err = db.AccessToken.Delete().Where(accesstoken.Token(token)).Exec(ctx)
	return errs.Wrap(err, "revoke access token")
}

// TouchLastUsedAt 回写 token 的最近使用时间。
func (r *accessTokenRepo) TouchLastUsedAt(ctx context.Context, token string, t time.Time) (err error) {
	ctx, span := tracer.Start(ctx, "accessTokenRepo/TouchLastUsedAt")
	defer func() { endSpan(span, err) }()
	return errs.Wrap(r.data.DB().AccessToken.Update().Where(accesstoken.Token(token)).SetLastUsedAt(t).Exec(ctx), "touch access token last used")
}
