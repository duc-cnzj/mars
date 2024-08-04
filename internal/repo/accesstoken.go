package repo

import (
	"context"
	"errors"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/accesstoken"
	"github.com/duc-cnzj/mars/v4/internal/ent/schema/mixin"
	"github.com/duc-cnzj/mars/v4/internal/filters"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
	"github.com/google/uuid"
)

type AccessTokenRepo interface {
	List(ctx context.Context, input *ListAccessTokenInput) ([]*ent.AccessToken, *pagination.Pagination, error)
	Grant(ctx context.Context, input *GrantAccessTokenInput) (*ent.AccessToken, error)
	Lease(ctx context.Context, token string, expireSeconds int32) (*ent.AccessToken, error)
	Revoke(ctx context.Context, token string) error
}

var _ AccessTokenRepo = (*accessTokenRepo)(nil)

type accessTokenRepo struct {
	logger mlog.Logger
	data   data.Data
	timer  timer.Timer
}

func NewAccessTokenRepo(timer timer.Timer, logger mlog.Logger, data data.Data) AccessTokenRepo {
	return &accessTokenRepo{logger: logger, data: data, timer: timer}
}

type ListAccessTokenInput struct {
	Page, PageSize int32
	WithSoftDelete bool
	Email          string
}

func (a *accessTokenRepo) List(ctx context.Context, input *ListAccessTokenInput) ([]*ent.AccessToken, *pagination.Pagination, error) {
	var db = a.data.DB()
	if input.WithSoftDelete {
		ctx = mixin.SkipSoftDelete(ctx)
	}
	query := db.AccessToken.Query().
		Where(filters.IfEmail(input.Email))

	tokens := query.Clone().
		Order(ent.Desc(accesstoken.FieldID)).
		Offset(pagination.GetPageOffset(input.Page, input.PageSize)).
		Limit(int(input.PageSize)).
		AllX(ctx)
	count := query.Clone().CountX(ctx)
	return tokens, pagination.NewPagination(input.Page, input.PageSize, count), nil
}

type GrantAccessTokenInput struct {
	ExpireSeconds int32
	Usage         string
	User          *auth.UserInfo
}

func (a *accessTokenRepo) Grant(ctx context.Context, input *GrantAccessTokenInput) (*ent.AccessToken, error) {
	var db = a.data.DB()
	return db.AccessToken.Create().
		SetToken(uuid.NewString()).
		SetEmail(input.User.Email).
		SetUsage(input.Usage).
		SetNillableUserInfo(input.User).
		SetExpiredAt(a.timer.Now().Add(time.Duration(input.ExpireSeconds) * time.Second)).
		Save(ctx)
}

// Lease 续约
func (a *accessTokenRepo) Lease(ctx context.Context, token string, expireSeconds int32) (*ent.AccessToken, error) {
	var db = a.data.DB()
	first, err := db.AccessToken.Query().Where(accesstoken.Token(token)).First(ctx)
	if err != nil {
		return nil, err
	}
	if first.Expired() {
		return nil, errors.New("token 已经过期")
	}
	return first.Update().SetExpiredAt(a.timer.Now().Add(time.Duration(expireSeconds) * time.Second)).Save(ctx)
}

func (a *accessTokenRepo) Revoke(ctx context.Context, token string) error {
	var db = a.data.DB()
	_, err := db.AccessToken.Delete().Where(accesstoken.Token(token)).Exec(ctx)
	return err
}
