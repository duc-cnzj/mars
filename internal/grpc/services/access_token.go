package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/internal/utils/date"

	"github.com/duc-cnzj/mars-client/v4/token"
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/scopes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		token.RegisterAccessTokenServer(s, &AccessToken{
			nowFunc: func() time.Time { return time.Now() },
		})
	})
	RegisterEndpoint(token.RegisterAccessTokenHandlerFromEndpoint)
}

type AccessToken struct {
	token.UnimplementedAccessTokenServer

	nowFunc func() time.Time
}

func (a *AccessToken) List(ctx context.Context, request *token.ListRequest) (*token.ListResponse, error) {
	var (
		page     = int(request.Page)
		pageSize = int(request.PageSize)
		tokens   []models.AccessToken
		count    int64
	)
	queryScope := func(db *gorm.DB) *gorm.DB {
		db = db.Where("`email` = ?", MustGetUser(ctx).Email)
		return db
	}

	app.DB().
		Unscoped().
		Scopes(queryScope, scopes.Paginate(&page, &pageSize)).
		Order("`ID` DESC").
		Find(&tokens)
	app.DB().Model(&models.AccessToken{}).Unscoped().Scopes(queryScope).Count(&count)
	var res = make([]*types.AccessTokenModel, 0, len(tokens))
	for _, accessToken := range tokens {
		res = append(res, accessToken.ProtoTransform())
	}
	return &token.ListResponse{
		Page:     int64(page),
		PageSize: int64(pageSize),
		Items:    res,
		Count:    count,
	}, nil
}

func (a *AccessToken) Grant(ctx context.Context, request *token.GrantRequest) (*token.GrantResponse, error) {
	var (
		user = MustGetUser(ctx)
		at   = models.NewAccessToken(request.Usage, a.nowFunc().Add(time.Second*time.Duration(request.ExpireSeconds)), user)
	)
	app.DB().Create(&at)
	AuditLog(user.Name, types.EventActionType_Create, fmt.Sprintf(`[AccessToken]: 用户 "%s" 创建了一个 token "%s", 过期时间是 "%s".`, user.Name, at.Token, at.ExpiredAt.Format("2006-01-02 15:04:05")))
	return &token.GrantResponse{Token: at.ProtoTransform()}, nil
}

func (a *AccessToken) Lease(ctx context.Context, request *token.LeaseRequest) (*token.LeaseResponse, error) {
	var (
		at   models.AccessToken
		user = MustGetUser(ctx)
	)
	if err := app.DB().Where("`email` = ? AND `token` = ?", user.Email, request.Token).First(&at).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "token not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if at.Expired() {
		return nil, status.Error(codes.Aborted, "token 已经过期")
	}
	app.DB().Model(&at).Update("expired_at", at.ExpiredAt.Add(time.Duration(request.ExpireSeconds)*time.Second))
	AuditLog(user.Name, types.EventActionType_Update, fmt.Sprintf(`[AccessToken]: 用户 "%s" 续租了 token "%s", 增加了 "%s", 过期时间是 "%s".`, user.Name, at.Token, date.HumanDuration(time.Second*time.Duration(request.ExpireSeconds)), at.ExpiredAt.Format("2006-01-02 15:04:05")))

	return &token.LeaseResponse{Token: at.ProtoTransform()}, nil
}

func (a *AccessToken) Revoke(ctx context.Context, request *token.RevokeRequest) (*token.RevokeResponse, error) {
	var user = MustGetUser(ctx)
	app.DB().Where("`email` = ? AND `token` = ?", user.Email, request.Token).Delete(&models.AccessToken{})
	AuditLog(user.Name, types.EventActionType_Delete, fmt.Sprintf(`[AccessToken]: 用户 "%s" 删除 token "%s".`, user.Name, request.Token))

	return &token.RevokeResponse{}, nil
}
