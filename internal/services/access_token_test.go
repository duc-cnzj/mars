package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/v5/internal/util/date"

	"github.com/duc-cnzj/mars/api/v5/token"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/ent/schema/schematype"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var adminEmail = "1025434218@qq.com"

func newAdminUserCtx() context.Context {
	return auth.SetUser(context.TODO(), &auth.UserInfo{
		ID:    "1",
		Email: adminEmail,
		Name:  "admin",
		Roles: []string{schematype.MarsAdmin},
	})
}

func newOtherUserCtx() context.Context {
	return auth.SetUser(context.TODO(), &auth.UserInfo{
		ID:    "2",
		Email: "user@mars.com",
		Name:  "user1",
	})
}

func TestNewAccessTokenSvc(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	tokenRepo := repo.NewMockAccessTokenRepo(m)
	svc := NewAccessTokenSvc(mlog.NewForConfig(nil), eventRepo, timer.NewRealTimer(), tokenRepo)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.(*accessTokenSvc).logger)
	assert.NotNil(t, svc.(*accessTokenSvc).eventRepo)
	assert.NotNil(t, svc.(*accessTokenSvc).timer)
	assert.NotNil(t, svc.(*accessTokenSvc).repo)
}

func Test_accessTokenSvc_Grant(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	tokenRepo := repo.NewMockAccessTokenRepo(m)
	svc := NewAccessTokenSvc(mlog.NewForConfig(nil), eventRepo, timer.NewRealTimer(), tokenRepo)

	tokenRepo.EXPECT().Grant(gomock.Any(), &repo.GrantAccessTokenInput{
		ExpireSeconds: 100,
		Usage:         "usage",
		User:          MustGetUser(newAdminUserCtx()),
	}).Return(nil, errors.New("xx"))
	_, err := svc.Grant(newAdminUserCtx(), &token.GrantRequest{
		ExpireSeconds: 100,
		Usage:         "usage",
	})
	assert.Equal(t, "xx", err.Error())
}

func TestAccessTokenSvc_Grant_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	tokenRepo := repo.NewMockAccessTokenRepo(m)
	svc := NewAccessTokenSvc(mlog.NewForConfig(nil), eventRepo, timer.NewRealTimer(), tokenRepo)

	req := &token.GrantRequest{
		ExpireSeconds: 100,
		Usage:         "usage",
	}

	resp := &repo.AccessToken{}
	user := MustGetUser(newAdminUserCtx())
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Create,
		"admin",
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 创建了一个 token "%s", 过期时间是 "%s".`, user.Name, resp.Token, resp.ExpiredAt.Format("2006-01-02 15:04:05")),
		req,
	)
	tokenRepo.EXPECT().Grant(gomock.Any(), &repo.GrantAccessTokenInput{
		ExpireSeconds: 100,
		Usage:         "usage",
		User:          user,
	}).Return(resp, nil)

	_, err := svc.Grant(newAdminUserCtx(), req)
	assert.NoError(t, err)
}

func TestAccessTokenSvc_Lease_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	tokenRepo := repo.NewMockAccessTokenRepo(m)
	svc := NewAccessTokenSvc(mlog.NewForConfig(nil), eventRepo, timer.NewRealTimer(), tokenRepo)

	req := &token.LeaseRequest{
		Token:         "token",
		ExpireSeconds: 100,
	}
	resp := &repo.AccessToken{}

	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Update,
		"admin",
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 续租了 token "%s", 增加了 "%s", 过期时间是 "%s".`, MustGetUser(newAdminUserCtx()).Name, resp.Token, date.HumanDuration(time.Second*time.Duration(req.ExpireSeconds)), resp.ExpiredAt.Format("2006-01-02 15:04:05")),
		req,
	)

	tokenRepo.EXPECT().Lease(gomock.Any(), "token", int32(100)).Return(resp, nil)
	_, err := svc.Lease(newAdminUserCtx(), req)
	assert.NoError(t, err)
}

func TestAccessTokenSvc_Lease_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	tokenRepo := repo.NewMockAccessTokenRepo(m)
	svc := NewAccessTokenSvc(mlog.NewForConfig(nil), eventRepo, timer.NewRealTimer(), tokenRepo)

	tokenRepo.EXPECT().Lease(gomock.Any(), "token", int32(100)).Return(nil, errors.New("error"))

	_, err := svc.Lease(newAdminUserCtx(), &token.LeaseRequest{
		Token:         "token",
		ExpireSeconds: 100,
	})
	assert.Error(t, err)
}

func TestAccessTokenSvc_Revoke_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	tokenRepo := repo.NewMockAccessTokenRepo(m)
	svc := NewAccessTokenSvc(mlog.NewForConfig(nil), eventRepo, timer.NewRealTimer(), tokenRepo)

	req := &token.RevokeRequest{
		Token: "token",
	}
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Delete,
		"admin",
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 删除 token "%s".`, MustGetUser(newAdminUserCtx()).Name, req.Token),
		req,
	)
	tokenRepo.EXPECT().Revoke(gomock.Any(), "token").Return(nil)

	_, err := svc.Revoke(newAdminUserCtx(), req)
	assert.NoError(t, err)
}

func TestAccessTokenSvc_Revoke_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	tokenRepo := repo.NewMockAccessTokenRepo(m)
	svc := NewAccessTokenSvc(mlog.NewForConfig(nil), eventRepo, timer.NewRealTimer(), tokenRepo)

	tokenRepo.EXPECT().Revoke(gomock.Any(), "token").Return(errors.New("error"))

	_, err := svc.Revoke(newAdminUserCtx(), &token.RevokeRequest{
		Token: "token",
	})
	assert.Error(t, err)
}

func TestAccessTokenSvc_List_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	tokenRepo := repo.NewMockAccessTokenRepo(m)
	svc := NewAccessTokenSvc(mlog.NewForConfig(nil), eventRepo, timer.NewRealTimer(), tokenRepo)

	tokenRepo.EXPECT().List(gomock.Any(), &repo.ListAccessTokenInput{
		Page:           1,
		PageSize:       10,
		Email:          MustGetUser(newAdminUserCtx()).Email,
		WithSoftDelete: true,
	}).Return([]*repo.AccessToken{}, pagination.NewPagination(1, 10, 0), nil)

	_, err := svc.List(newAdminUserCtx(), &token.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(10)),
	})
	assert.NoError(t, err)
}

func TestAccessTokenSvc_List_Failure(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	eventRepo := repo.NewMockEventRepo(m)
	tokenRepo := repo.NewMockAccessTokenRepo(m)
	svc := NewAccessTokenSvc(mlog.NewForConfig(nil), eventRepo, timer.NewRealTimer(), tokenRepo)

	tokenRepo.EXPECT().List(gomock.Any(), &repo.ListAccessTokenInput{
		Page:           1,
		PageSize:       10,
		Email:          MustGetUser(newAdminUserCtx()).Email,
		WithSoftDelete: true,
	}).Return(nil, nil, errors.New("error"))

	_, err := svc.List(newAdminUserCtx(), &token.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(10)),
	})
	assert.Error(t, err)
}
