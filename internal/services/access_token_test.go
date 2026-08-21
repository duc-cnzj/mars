package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/util/date"

	"github.com/duc-cnzj/mars/api/v6/proto/token"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var adminEmail = "1025434218@qq.com"

func newAdminUserCtx() context.Context {
	return biz.SetUser(context.TODO(), &biz.UserInfo{
		ID:    "1",
		Email: adminEmail,
		Name:  "admin",
		Roles: []string{schematype.MarsAdmin},
	})
}

func newOtherUserCtx() context.Context {
	return biz.SetUser(context.TODO(), &biz.UserInfo{
		ID:    "2",
		Email: "user@mars.com",
		Name:  "user1",
	})
}

func TestMaskToken(t *testing.T) {
	assert.Equal(t, "******", maskToken(""))
	assert.Equal(t, "******", maskToken("12345678"))
	assert.Equal(t, "1234****9012", maskToken("123456789012"))
}

func TestNewAccessTokenSvc(t *testing.T) {
	svc, _ := newAccessTokenSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.eventBiz)
	assert.NotNil(t, svc.repo)
}

func Test_accessTokenSvc_Grant(t *testing.T) {
	svc, mocks := newAccessTokenSvcWithMocks(t)
	tokenRepo := mocks.accessTokenBiz

	tokenRepo.EXPECT().Grant(gomock.Any(), &biz.GrantAccessTokenInput{
		ExpireSeconds: 100,
		Usage:         "usage",
		User:          biz.MustGetUser(newAdminUserCtx()),
	}).Return(nil, errors.New("xx"))
	_, err := svc.Grant(newAdminUserCtx(), &token.GrantRequest{
		ExpireSeconds: 100,
		Usage:         "usage",
	})
	assert.Equal(t, "xx", err.Error())
}

func TestAccessTokenSvc_Grant_Success(t *testing.T) {
	svc, mocks := newAccessTokenSvcWithMocks(t)
	eventRepo := mocks.eventRepo
	tokenRepo := mocks.accessTokenBiz

	req := &token.GrantRequest{
		ExpireSeconds: 100,
		Usage:         "usage",
	}

	now := time.Now()
	resp := &biz.AccessToken{Token: "secret-token", Email: "admin", Usage: "usage", ExpiredAt: now, CreatedAt: now, UpdatedAt: now}
	user := biz.MustGetUser(newAdminUserCtx())
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Create,
		"admin",
		user.Email,
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 创建了一个 token "%s", 过期时间是 "%s".`, user.Name, maskToken(resp.Token), resp.ExpiredAt.Format("2006-01-02 15:04:05")),
		req,
	)
	tokenRepo.EXPECT().Grant(gomock.Any(), &biz.GrantAccessTokenInput{
		ExpireSeconds: 100,
		Usage:         "usage",
		User:          user,
	}).Return(resp, nil)

	grantResp, err := svc.Grant(newAdminUserCtx(), req)
	assert.NoError(t, err)
	if assert.NotNil(t, grantResp) && assert.NotNil(t, grantResp.Token) {
		assert.Equal(t, "secret-token", grantResp.Token.Token)
		assert.Equal(t, "admin", grantResp.Token.Email)
		assert.Equal(t, "usage", grantResp.Token.Usage)
	}
}

// 用真实长度的 token 锁定 Fire #86 的安全属性：
// 审计日志必须含脱敏形式、且绝不能出现完整 token。
// 现有空 token 测试的 mask 恒为 "******"，判别力不足。
func TestAccessTokenSvc_Grant_AuditLogMasksFullToken(t *testing.T) {
	svc, mocks := newAccessTokenSvcWithMocks(t)
	eventRepo := mocks.eventRepo
	tokenRepo := mocks.accessTokenBiz

	fullToken := "0123456789abcdef0123456789abcdef" // 真实 token（uuid 长度同量级）
	resp := &biz.AccessToken{Token: fullToken, ExpiredAt: time.Now().Add(time.Hour)}
	user := biz.MustGetUser(newAdminUserCtx())

	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Create,
		user.Name,
		user.Email,
		gomock.Any(),
		gomock.Any(),
	).Do(func(_ types.EventActionType, _ string, _ string, msg string, _ any) {
		assert.NotContains(t, msg, fullToken, "审计日志不得出现完整 token")
		assert.Contains(t, msg, maskToken(fullToken), "审计日志应包含脱敏 token")
	})
	tokenRepo.EXPECT().Grant(gomock.Any(), gomock.Any()).Return(resp, nil)

	_, err := svc.Grant(newAdminUserCtx(), &token.GrantRequest{ExpireSeconds: 100, Usage: "usage"})
	assert.NoError(t, err)
}

func TestAccessTokenSvc_Lease_Success(t *testing.T) {
	svc, mocks := newAccessTokenSvcWithMocks(t)
	eventRepo := mocks.eventRepo
	tokenRepo := mocks.accessTokenBiz

	req := &token.LeaseRequest{
		Token:         "token",
		ExpireSeconds: 100,
	}
	now := time.Now()
	resp := &biz.AccessToken{Token: "secret-token", Email: "admin", Usage: "usage", ExpiredAt: now, CreatedAt: now, UpdatedAt: now}

	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Update,
		"admin",
		adminEmail,
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 续租了 token "%s", 增加了 "%s", 过期时间是 "%s".`, biz.MustGetUser(newAdminUserCtx()).Name, maskToken(resp.Token), date.HumanDuration(time.Second*time.Duration(req.ExpireSeconds)), resp.ExpiredAt.Format("2006-01-02 15:04:05")),
		req,
	)

	tokenRepo.EXPECT().Lease(gomock.Any(), "token", int32(100)).Return(resp, nil)
	leaseResp, err := svc.Lease(newAdminUserCtx(), req)
	assert.NoError(t, err)
	if assert.NotNil(t, leaseResp) && assert.NotNil(t, leaseResp.Token) {
		assert.Equal(t, "secret-token", leaseResp.Token.Token)
		assert.Equal(t, "admin", leaseResp.Token.Email)
		assert.Equal(t, "usage", leaseResp.Token.Usage)
	}
}

func TestAccessTokenSvc_Lease_Failure(t *testing.T) {
	svc, mocks := newAccessTokenSvcWithMocks(t)
	tokenRepo := mocks.accessTokenBiz

	tokenRepo.EXPECT().Lease(gomock.Any(), "token", int32(100)).Return(nil, errors.New("error"))

	_, err := svc.Lease(newAdminUserCtx(), &token.LeaseRequest{
		Token:         "token",
		ExpireSeconds: 100,
	})
	assert.Error(t, err)
}

func TestAccessTokenSvc_Revoke_Success(t *testing.T) {
	svc, mocks := newAccessTokenSvcWithMocks(t)
	eventRepo := mocks.eventRepo
	tokenRepo := mocks.accessTokenBiz

	req := &token.RevokeRequest{
		Token: "token",
	}
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Delete,
		"admin",
		adminEmail,
		fmt.Sprintf(`[accessTokenSvc]: 用户 "%s" 删除 token "%s".`, biz.MustGetUser(newAdminUserCtx()).Name, maskToken(req.Token)),
		req,
	)
	tokenRepo.EXPECT().Revoke(gomock.Any(), "token").Return(nil)

	_, err := svc.Revoke(newAdminUserCtx(), req)
	assert.NoError(t, err)
}

func TestAccessTokenSvc_Revoke_Failure(t *testing.T) {
	svc, mocks := newAccessTokenSvcWithMocks(t)
	tokenRepo := mocks.accessTokenBiz

	tokenRepo.EXPECT().Revoke(gomock.Any(), "token").Return(errors.New("error"))

	_, err := svc.Revoke(newAdminUserCtx(), &token.RevokeRequest{
		Token: "token",
	})
	assert.Error(t, err)
}

func TestAccessTokenSvc_List_Success(t *testing.T) {
	svc, mocks := newAccessTokenSvcWithMocks(t)
	tokenRepo := mocks.accessTokenBiz

	now := time.Now()
	tokenRepo.EXPECT().List(gomock.Any(), &biz.ListAccessTokenInput{
		Page:           1,
		PageSize:       10,
		Email:          biz.MustGetUser(newAdminUserCtx()).Email,
		WithSoftDelete: true,
	}).Return([]*biz.AccessToken{
		{Token: "secret-token", Email: "admin", Usage: "usage", ExpiredAt: now, CreatedAt: now, UpdatedAt: now},
	}, pagination.NewPagination(1, 10, 1), nil)

	resp, err := svc.List(newAdminUserCtx(), &token.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(10)),
	})
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, int32(1), resp.Page)
		assert.Equal(t, int32(10), resp.PageSize)
		assert.Equal(t, int32(1), resp.Count)
		if assert.Len(t, resp.Items, 1) {
			assert.Equal(t, "secret-token", resp.Items[0].Token)
			assert.Equal(t, "admin", resp.Items[0].Email)
			assert.Equal(t, "usage", resp.Items[0].Usage)
		}
	}
}

func TestAccessTokenSvc_List_Failure(t *testing.T) {
	svc, mocks := newAccessTokenSvcWithMocks(t)
	tokenRepo := mocks.accessTokenBiz

	tokenRepo.EXPECT().List(gomock.Any(), &biz.ListAccessTokenInput{
		Page:           1,
		PageSize:       10,
		Email:          biz.MustGetUser(newAdminUserCtx()).Email,
		WithSoftDelete: true,
	}).Return(nil, nil, errors.New("error"))

	_, err := svc.List(newAdminUserCtx(), &token.ListRequest{
		Page:     lo.ToPtr(int32(1)),
		PageSize: lo.ToPtr(int32(10)),
	})
	assert.Error(t, err)
}

type accessTokenSvcMocks struct {
	ctrl           *gomock.Controller
	eventRepo      *data.MockEventRepo
	accessTokenBiz *biz.MockAccessTokenBiz
}

func newAccessTokenSvcWithMocks(t *testing.T) (*accessTokenSvc, *accessTokenSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &accessTokenSvcMocks{
		ctrl:           ctrl,
		eventRepo:      data.NewMockEventRepo(ctrl),
		accessTokenBiz: biz.NewMockAccessTokenBiz(ctrl),
	}
	s, ok := NewAccessTokenSvc(AccessTokenSvcDeps{
		Logger:   mlog.NewForConfig(nil),
		EventBiz: biz.NewEventBiz(mocks.eventRepo),
		Repo:     mocks.accessTokenBiz,
	}).(*accessTokenSvc)
	if !ok {
		panic("NewAccessTokenSvc returned unexpected type")
	}
	return s, mocks
}
