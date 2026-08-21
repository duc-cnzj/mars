package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/types"

	apiauth "github.com/duc-cnzj/mars/api/v6/proto/auth"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewAuthSvc(t *testing.T) {
	svc, _ := newAuthSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.eventBiz)
	assert.NotNil(t, svc.authBiz)
}

// Test_authSvc_Info 覆盖 Info 成功路径：用户由鉴权拦截器经 biz.SetUser 注入 ctx，
// Info 不再自行验签，仅做「取 ctx 用户 → 映射响应」。
func Test_authSvc_Info(t *testing.T) {
	svc, _ := newAuthSvcWithMocks(t)
	user := &biz.UserInfo{
		ID:        "123",
		Email:     "duc@example.com",
		Name:      "duc",
		Picture:   "https://example.com/avatar.png",
		Roles:     []string{"admin", "dev"},
		LogoutUrl: "https://logout.example",
	}
	resp, err := svc.Info(biz.SetUser(context.TODO(), user), nil)
	assert.Nil(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, int32(123), resp.Id)
		assert.Equal(t, "https://example.com/avatar.png", resp.Avatar)
		assert.Equal(t, "duc", resp.Name)
		assert.Equal(t, "duc@example.com", resp.Email)
		assert.Equal(t, "https://logout.example", resp.LogoutUrl)
		assert.Equal(t, []string{"admin", "dev"}, resp.Roles)
	}
}

func TestAuthSvc_Login_Success(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	eventRepo := mocks.eventRepo
	authBizMock := mocks.authBiz

	resp := &biz.LoginResponse{
		Token:     "test-token",
		ExpiredIn: 100,
		UserInfo: &biz.UserInfo{
			Name: "duc",
		},
	}
	eventRepo.EXPECT().AuditLog(
		types.EventActionType_Login,
		resp.UserInfo.Name,
		resp.UserInfo.Email,
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", resp.UserInfo.Name, resp.UserInfo.Email),
	)

	authBizMock.EXPECT().Login(gomock.Any(), &biz.LoginInput{
		Username: "test",
		Password: "password",
	}).Return(resp, nil)

	res, err := svc.Login(context.TODO(), &apiauth.LoginRequest{
		Username: "test",
		Password: "password",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test-token", res.Token)
}

func TestAuthSvc_Login_Failure(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	authBizMock.EXPECT().Login(gomock.Any(), &biz.LoginInput{
		Username: "test",
		Password: "password",
	}).Return(nil, errors.New("error"))

	_, err := svc.Login(context.TODO(), &apiauth.LoginRequest{
		Username: "test",
		Password: "password",
	})
	assert.Error(t, err)
}

func TestAuthSvc_Settings_Success(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	authBizMock.EXPECT().Settings(gomock.Any()).Return(biz.OidcConfig{}, nil)

	_, err := svc.Settings(context.TODO(), &apiauth.SettingsRequest{})
	assert.NoError(t, err)
}

func TestAuthSvc_Exchange_Success(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz
	eventRepo := mocks.eventRepo

	userinfo := &biz.UserInfo{Name: "duc", Email: "DUC@example.com"}
	authBizMock.EXPECT().Exchange(gomock.Any(), "code").Return(userinfo, nil)
	authBizMock.EXPECT().Sign(gomock.Any(), userinfo).Return(&biz.LoginResponse{Token: "signed", ExpiredIn: 3600}, nil)
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Login,
		userinfo.Name,
		userinfo.Email,
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", userinfo.Name, userinfo.Email),
		gomock.Any(),
	)

	resp, err := svc.Exchange(context.TODO(), &apiauth.ExchangeRequest{Code: "code"})
	assert.NoError(t, err)
	assert.Equal(t, "signed", resp.Token)
	assert.Equal(t, int64(3600), resp.ExpiresIn)
}

func TestAuthSvc_Exchange_Error(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	authBizMock.EXPECT().Exchange(gomock.Any(), "code").Return(nil, errors.New("boom"))

	_, err := svc.Exchange(context.TODO(), &apiauth.ExchangeRequest{Code: "code"})
	assert.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

func TestAuthSvc_Exchange_CodeNotEchoed(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	// 换发编排失败（biz 侧）返回 InvalidArgument，且错误信息不回显一次性 code。
	code := "auth-code-SECRET-abc123"
	authBizMock.EXPECT().Exchange(gomock.Any(), code).Return(nil, status.Errorf(codes.InvalidArgument, "invalid code"))

	_, err := svc.Exchange(context.TODO(), &apiauth.ExchangeRequest{Code: code})
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), code)
}

func TestAuthSvc_Exchange_SignError(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	userinfo := &biz.UserInfo{Name: "duc"}
	authBizMock.EXPECT().Exchange(gomock.Any(), "code").Return(userinfo, nil)
	authBizMock.EXPECT().Sign(gomock.Any(), userinfo).Return(nil, errors.New("sign boom"))

	_, err := svc.Exchange(context.TODO(), &apiauth.ExchangeRequest{Code: "code"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sign boom")
}

func TestAuthSvc_Settings_NoSettings(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	authBizMock.EXPECT().Settings(gomock.Any()).Return(nil, nil)

	resp, err := svc.Settings(context.TODO(), &apiauth.SettingsRequest{})
	assert.NoError(t, err)
	assert.Empty(t, resp.Items)
}

func TestAuthSvc_Settings_ErrorFetchingSettings(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	authBizMock.EXPECT().Settings(gomock.Any()).Return(biz.OidcConfig{
		"b": biz.OidcConfigItem{
			Config:             oauth2.Config{},
			EndSessionEndpoint: "",
		},
		"a": biz.OidcConfigItem{
			Config:             oauth2.Config{},
			EndSessionEndpoint: "",
		},
	}, nil)

	res, err := svc.Settings(context.TODO(), &apiauth.SettingsRequest{})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res.Items))
	assert.Equal(t, "a", res.Items[0].Name)
	assert.Equal(t, "b", res.Items[1].Name)
}

func TestAuthSvc_Settings_Error(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	authBizMock.EXPECT().Settings(gomock.Any()).Return(nil, errors.New("boom"))

	resp, err := svc.Settings(context.TODO(), &apiauth.SettingsRequest{})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

type authSvcMocks struct {
	ctrl      *gomock.Controller
	eventRepo *data.MockEventRepo
	authBiz   *biz.MockAuthBiz
}

func newAuthSvcWithMocks(t *testing.T) (*authSvc, *authSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &authSvcMocks{
		ctrl:      ctrl,
		eventRepo: data.NewMockEventRepo(ctrl),
		authBiz:   biz.NewMockAuthBiz(ctrl),
	}
	s, ok := NewAuthSvc(AuthSvcDeps{
		EventBiz: biz.NewEventBiz(mocks.eventRepo),
		Logger:   mlog.NewForConfig(nil),
		AuthBiz:  mocks.authBiz,
	}).(*authSvc)
	if !ok {
		panic("NewAuthSvc returned unexpected type")
	}
	return s, mocks
}
