package services

import (
	"context"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/util/pagination"

	"github.com/duc-cnzj/mars/api/v4/token"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
)

var _ token.AccessTokenServer = (*accessTokenSvc)(nil)

type accessTokenSvc struct {
	token.UnimplementedAccessTokenServer

	timer timer.Timer

	repo      repo.AccessTokenRepo
	eventRepo repo.EventRepo
}

func NewAccessTokenSvc(eventRepo repo.EventRepo, timer timer.Timer, repo repo.AccessTokenRepo) token.AccessTokenServer {
	return &accessTokenSvc{
		eventRepo: eventRepo,
		timer:     timer,
		repo:      repo,
	}
}

func (a *accessTokenSvc) List(ctx context.Context, request *token.ListRequest) (*token.ListResponse, error) {
	pagination.InitByDefault(&request.Page, &request.PageSize)
	tokens, p, err := a.repo.List(ctx, &repo.ListAccessTokenInput{
		Page:           request.Page,
		PageSize:       request.PageSize,
		Email:          MustGetUser(ctx).Email,
		WithSoftDelete: true,
	})
	if err != nil {
		return nil, err
	}
	var res = make([]*types.AccessTokenModel, 0, len(tokens))
	for _, accessToken := range tokens {
		res = append(res, transformer.FromAccessToken(accessToken))
	}
	return &token.ListResponse{
		Page:     p.Page,
		PageSize: p.PageSize,
		Items:    res,
		Count:    p.Count,
	}, nil
}

func (a *accessTokenSvc) Grant(ctx context.Context, request *token.GrantRequest) (*token.GrantResponse, error) {
	var user = MustGetUser(ctx)
	at, err := a.repo.Grant(ctx, &repo.GrantAccessTokenInput{
		ExpireSeconds: request.ExpireSeconds,
		Usage:         request.Usage,
		User:          user,
	})
	if err != nil {
		return nil, err
	}

	a.eventRepo.AuditLog(types.EventActionType_Create, user.Name, fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 创建了一个 token "%s", 过期时间是 "%s".`, user.Name, at.Token, at.ExpiredAt.Format("2006-01-02 15:04:05")))

	return &token.GrantResponse{Token: transformer.FromAccessToken(at)}, nil
}

func (a *accessTokenSvc) Lease(ctx context.Context, request *token.LeaseRequest) (*token.LeaseResponse, error) {
	var (
		user = MustGetUser(ctx)
	)

	at, err := a.repo.Lease(ctx, request.Token, request.ExpireSeconds)
	if err != nil {
		return nil, err
	}
	a.eventRepo.AuditLog(
		types.EventActionType_Update,
		user.Name,
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 续租了 token "%s", 增加了 "%s", 过期时间是 "%s".`, user.Name, at.Token, date.HumanDuration(time.Second*time.Duration(request.ExpireSeconds)), at.ExpiredAt.Format("2006-01-02 15:04:05")),
	)

	return &token.LeaseResponse{Token: transformer.FromAccessToken(at)}, nil
}

func (a *accessTokenSvc) Revoke(ctx context.Context, request *token.RevokeRequest) (*token.RevokeResponse, error) {
	var user = MustGetUser(ctx)
	if err := a.repo.Revoke(ctx, request.Token); err != nil {
		return nil, err
	}
	a.eventRepo.AuditLog(types.EventActionType_Delete, user.Name, fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 删除 token "%s".`, user.Name, request.Token))

	return &token.RevokeResponse{}, nil
}
