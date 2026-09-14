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
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// oidcStateCtx 模拟「grpc-gateway 把 HTTP Cookie 头映射成 metadata」之后的 ctx：
// DefaultHeaderMatcher 给 Cookie 加 "grpcgateway-" 前缀，故键名为 grpcgateway-cookie。
// rawCookie 传原始 Cookie 头串（非单值），便于覆盖解析失败的用例。
func oidcStateCtx(rawCookie string) context.Context {
	return metadata.NewIncomingContext(context.TODO(), metadata.Pairs(oidcCookieMetadataKey, rawCookie))
}

// oidcStateCtxWithState 是 oidcStateCtx 的常用形态：只带一个 mars_oidc_state 的合法 Cookie 头。
func oidcStateCtxWithState(state string) context.Context {
	return oidcStateCtx(fmt.Sprintf("%s=%s", middlewares.OidcStateCookieName, state))
}

func TestNewAuthSvc(t *testing.T) {
	svc, _ := newAuthSvcWithMocks(t)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.eventBiz)
	assert.NotNil(t, svc.authBiz)
	assert.NotNil(t, svc.userBiz)
}

// Test_authSvc_Info 覆盖 Info 成功路径：用户由鉴权拦截器经 biz.SetUser 注入 ctx，
// Info 不再自行验签，仅做「取 ctx 用户 → 映射响应」+ 读一次库取权威灰度标记。
func Test_authSvc_Info(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	user := &biz.UserInfo{
		ID:        "123",
		Email:     "duc@example.com",
		Name:      "duc",
		Picture:   "https://example.com/avatar.png",
		Roles:     []string{"admin", "dev"},
		LogoutUrl: "https://logout.example",
	}
	// IsGray 是 Info 里唯一读库字段：灰度标记不在 JWT 里，每次都要取当前库值。
	mocks.userBiz.EXPECT().IsGray(gomock.Any(), "duc@example.com").Return(true, nil)
	resp, err := svc.Info(biz.SetUser(context.TODO(), user), nil)
	assert.Nil(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, int32(123), resp.Id)
		assert.Equal(t, "https://example.com/avatar.png", resp.Avatar)
		assert.Equal(t, "duc", resp.Name)
		assert.Equal(t, "duc@example.com", resp.Email)
		assert.Equal(t, "https://logout.example", resp.LogoutUrl)
		assert.Equal(t, []string{"admin", "dev"}, resp.Roles)
		assert.False(t, resp.IsSuperAdmin)
		assert.True(t, resp.IsGray)
	}
}

// Test_authSvc_Info_SuperAdmin 内置超级管理员固定邮箱登录 → is_super_admin = true。
func Test_authSvc_Info_SuperAdmin(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	mocks.userBiz.EXPECT().IsGray(gomock.Any(), biz.SuperAdminEmail).Return(false, nil)
	resp, err := svc.Info(biz.SetUser(context.TODO(), &biz.UserInfo{
		Email: biz.SuperAdminEmail,
		Roles: []string{biz.MarsAdmin},
	}), nil)
	assert.Nil(t, err)
	if assert.NotNil(t, resp) {
		assert.True(t, resp.IsSuperAdmin)
		assert.False(t, resp.IsGray)
	}
}

// Test_authSvc_Info_IsGrayFailOpen 读库失败降级为「非灰度」且不阻断 /api/auth/info：
// 灰度只是发布通道，一次查询失败不该把整个登录态恢复打断（用户会看到「打不开页面」，
// 而问题其实只是发布通道未知）。
func Test_authSvc_Info_IsGrayFailOpen(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	mocks.userBiz.EXPECT().IsGray(gomock.Any(), "duc@example.com").Return(false, errors.New("db boom"))
	resp, err := svc.Info(biz.SetUser(context.TODO(), &biz.UserInfo{
		Email: "duc@example.com",
	}), nil)
	assert.NoError(t, err, "灰度读库失败不得阻断 info")
	if assert.NotNil(t, resp) {
		assert.False(t, resp.IsGray, "降级为非灰度（稳定版）")
		assert.Equal(t, "duc@example.com", resp.Email)
	}
}

func TestAuthSvc_Login_Success(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	eventRepo := mocks.eventRepo
	authBizMock := mocks.authBiz
	userBizMock := mocks.userBiz

	resp := &biz.LoginResponse{
		Token:     "test-token",
		ExpiredIn: 100,
		UserInfo: &biz.UserInfo{
			Name:  biz.SuperAdminName,
			Email: biz.SuperAdminEmail,
			Roles: []string{biz.MarsAdmin},
		},
	}
	eventRepo.EXPECT().AuditLog(
		types.EventActionType_Login,
		resp.UserInfo.Name,
		resp.UserInfo.Email,
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", resp.UserInfo.Name, resp.UserInfo.Email),
	)
	// 登录成功即同步用户投影（admin 同样落 users 表，与 OIDC Exchange 行为对齐），角色随投影写入
	userBizMock.EXPECT().SyncLoginUser(gomock.Any(), biz.SuperAdminEmail, biz.SuperAdminName, []string{biz.MarsAdmin}).Return(nil)

	authBizMock.EXPECT().Login(gomock.Any(), &biz.LoginInput{
		Username: "admin",
		Password: "password",
	}).Return(resp, nil)

	res, err := svc.Login(context.TODO(), &apiauth.LoginRequest{
		Username: "admin",
		Password: "password",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test-token", res.Token)
}

// TestAuthSvc_Login_SyncUserErrorNotBlocking 投影写库失败不阻断登录（与 OIDC Exchange 一致）：
// 凭证已校验、登录事件已落库，users 只是管理投影，该用户下次登录会由 SyncLoginUser 自动补回。
func TestAuthSvc_Login_SyncUserErrorNotBlocking(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz
	eventRepo := mocks.eventRepo
	userBizMock := mocks.userBiz

	resp := &biz.LoginResponse{
		Token:     "test-token",
		ExpiredIn: 100,
		UserInfo: &biz.UserInfo{
			Name:  biz.SuperAdminName,
			Email: biz.SuperAdminEmail,
			Roles: []string{biz.MarsAdmin},
		},
	}
	eventRepo.EXPECT().AuditLog(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	userBizMock.EXPECT().SyncLoginUser(gomock.Any(), biz.SuperAdminEmail, biz.SuperAdminName, []string{biz.MarsAdmin}).Return(errors.New("projection boom"))
	authBizMock.EXPECT().Login(gomock.Any(), &biz.LoginInput{
		Username: "admin",
		Password: "password",
	}).Return(resp, nil)

	res, err := svc.Login(context.TODO(), &apiauth.LoginRequest{
		Username: "admin",
		Password: "password",
	})
	assert.NoError(t, err, "投影写库失败不得阻断登录")
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
	userBizMock := mocks.userBiz

	// SSO id_token 携带管理员角色：同步用户投影时应把角色一并传入
	userinfo := &biz.UserInfo{Name: "duc", Email: "DUC@example.com", Roles: []string{biz.MarsAdmin}}
	authBizMock.EXPECT().Exchange(gomock.Any(), "code").Return(userinfo, nil)
	authBizMock.EXPECT().Sign(gomock.Any(), userinfo).Return(&biz.LoginResponse{Token: "signed", ExpiredIn: 3600}, nil)
	eventRepo.EXPECT().AuditLogWithRequest(
		types.EventActionType_Login,
		userinfo.Name,
		userinfo.Email,
		fmt.Sprintf("用户 '%s' email: '%s' 登录了系统", userinfo.Name, userinfo.Email),
		gomock.Any(),
	)
	// 登录成功即同步用户投影（不存在则创建、存在则推进最近登录），SSO 角色随投影写入
	userBizMock.EXPECT().SyncLoginUser(gomock.Any(), "DUC@example.com", "duc", []string{biz.MarsAdmin}).Return(nil)

	resp, err := svc.Exchange(oidcStateCtxWithState("state"), &apiauth.ExchangeRequest{Code: "code", State: "state"})
	assert.NoError(t, err)
	assert.Equal(t, "signed", resp.Token)
	assert.Equal(t, int64(3600), resp.ExpiresIn)
}

// TestAuthSvc_Exchange_SyncUserErrorNotBlocking 投影写库失败不阻断登录：OIDC 凭证已校验、
// 登录事件已落库，users 只是管理投影，该用户下次登录会由 SyncLoginUser 自动补回。
func TestAuthSvc_Exchange_SyncUserErrorNotBlocking(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz
	eventRepo := mocks.eventRepo
	userBizMock := mocks.userBiz

	userinfo := &biz.UserInfo{Name: "duc", Email: "duc@example.com", Roles: []string{biz.MarsAdmin}}
	authBizMock.EXPECT().Exchange(gomock.Any(), "code").Return(userinfo, nil)
	authBizMock.EXPECT().Sign(gomock.Any(), userinfo).Return(&biz.LoginResponse{Token: "signed", ExpiredIn: 3600}, nil)
	eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	userBizMock.EXPECT().SyncLoginUser(gomock.Any(), "duc@example.com", "duc", []string{biz.MarsAdmin}).Return(errors.New("projection boom"))

	resp, err := svc.Exchange(oidcStateCtxWithState("state"), &apiauth.ExchangeRequest{Code: "code", State: "state"})
	assert.NoError(t, err, "投影写库失败不得阻断登录")
	assert.Equal(t, "signed", resp.Token)
}

func TestAuthSvc_Exchange_Error(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	authBizMock.EXPECT().Exchange(gomock.Any(), "code").Return(nil, errors.New("boom"))

	_, err := svc.Exchange(oidcStateCtxWithState("state"), &apiauth.ExchangeRequest{Code: "code", State: "state"})
	assert.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

func TestAuthSvc_Exchange_CodeNotEchoed(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	// 换发编排失败（biz 侧）返回 InvalidArgument，且错误信息不回显一次性 code。
	code := "auth-code-SECRET-abc123"
	authBizMock.EXPECT().Exchange(gomock.Any(), code).Return(nil, status.Errorf(codes.InvalidArgument, "invalid code"))

	_, err := svc.Exchange(oidcStateCtxWithState("state"), &apiauth.ExchangeRequest{Code: code, State: "state"})
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), code)
}

func TestAuthSvc_Exchange_SignError(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	authBizMock := mocks.authBiz

	userinfo := &biz.UserInfo{Name: "duc"}
	authBizMock.EXPECT().Exchange(gomock.Any(), "code").Return(userinfo, nil)
	authBizMock.EXPECT().Sign(gomock.Any(), userinfo).Return(nil, errors.New("sign boom"))

	_, err := svc.Exchange(oidcStateCtxWithState("state"), &apiauth.ExchangeRequest{Code: "code", State: "state"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sign boom")
}

// TestAuthSvc_Exchange_MissingStateCookie 无 state Cookie 时必须拒绝，且不得触达 IdP。
// 这正是「登录 CSRF」的形态：攻击者只拿得到自己的 code，拿不到受害者浏览器上的 Cookie。
// 不设 authBiz.Exchange 期望——门卫一旦漏过、换发被调用，gomock 会直接判失败。
func TestAuthSvc_Exchange_MissingStateCookie(t *testing.T) {
	svc, _ := newAuthSvcWithMocks(t)

	_, err := svc.Exchange(context.TODO(), &apiauth.ExchangeRequest{Code: "code", State: "state"})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestAuthSvc_Exchange_MetadataWithoutCookieKey 带 metadata 但不含 cookie 键时必须拒绝：
// 跨站表单 POST 不带任何 Cookie 时 grpc-gateway 不会写出 grpcgateway-cookie 键，
// 这是攻击请求最常见的形态。不设 authBiz.Exchange 期望，门卫漏过即 gomock 判失败。
func TestAuthSvc_Exchange_MetadataWithoutCookieKey(t *testing.T) {
	svc, _ := newAuthSvcWithMocks(t)

	ctx := metadata.NewIncomingContext(context.TODO(), metadata.Pairs("grpcgateway-authorization", "Bearer x"))
	_, err := svc.Exchange(ctx, &apiauth.ExchangeRequest{Code: "code", State: "state"})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestAuthSvc_Exchange_StateMismatch Cookie 携带的 state 与请求体回传的不一致时必须拒绝
// （攻击者拿自己申请到的合法 state 塞给受害者浏览器的场景）。
func TestAuthSvc_Exchange_StateMismatch(t *testing.T) {
	svc, _ := newAuthSvcWithMocks(t)

	_, err := svc.Exchange(oidcStateCtxWithState("cookie-state"), &apiauth.ExchangeRequest{Code: "code", State: "body-state"})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestAuthSvc_Exchange_MalformedCookieHeader Cookie 头无法解析时按校验失败处理，不 panic。
func TestAuthSvc_Exchange_MalformedCookieHeader(t *testing.T) {
	svc, _ := newAuthSvcWithMocks(t)

	_, err := svc.Exchange(oidcStateCtx("malformed-cookie-without-equals"), &apiauth.ExchangeRequest{Code: "code", State: "state"})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestAuthSvc_Exchange_CookieWithoutStateKey Cookie 可解析但缺少 mars_oidc_state 键时拒绝。
func TestAuthSvc_Exchange_CookieWithoutStateKey(t *testing.T) {
	svc, _ := newAuthSvcWithMocks(t)

	_, err := svc.Exchange(oidcStateCtx("other=1"), &apiauth.ExchangeRequest{Code: "code", State: "state"})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// TestAuthSvc_Exchange_CookieAmongOthers 同一个 Cookie 头里混有其他 Cookie 时仍能取出 state。
func TestAuthSvc_Exchange_CookieAmongOthers(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)

	userinfo := &biz.UserInfo{Name: "duc", Email: "duc@example.com"}
	mocks.authBiz.EXPECT().Exchange(gomock.Any(), "code").Return(userinfo, nil)
	mocks.authBiz.EXPECT().Sign(gomock.Any(), userinfo).Return(&biz.LoginResponse{Token: "signed", ExpiredIn: 1}, nil)
	mocks.eventRepo.EXPECT().AuditLogWithRequest(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	mocks.userBiz.EXPECT().SyncLoginUser(gomock.Any(), "duc@example.com", "duc", gomock.Any()).Return(nil)

	ctx := oidcStateCtx(fmt.Sprintf("other=1; %s=state; another=2", middlewares.OidcStateCookieName))
	resp, err := svc.Exchange(ctx, &apiauth.ExchangeRequest{Code: "code", State: "state"})
	assert.NoError(t, err)
	assert.Equal(t, "signed", resp.Token)
}

// TestAuthSvc_Settings_SharedState 所有 provider 必须共用同一个 state：state 绑定的是
// 「这次浏览器发起的登录尝试」而非某个 provider，且 state Cookie 只有一个槽位，
// per-provider 各发一个会让回调时无从比对（旧实现正是每 provider 各生成一个）。
func TestAuthSvc_Settings_SharedState(t *testing.T) {
	svc, mocks := newAuthSvcWithMocks(t)
	mocks.authBiz.EXPECT().Settings(gomock.Any()).Return(biz.OidcConfig{
		"b-provider": {Config: oauth2.Config{
			ClientID: "b",
			Endpoint: oauth2.Endpoint{AuthURL: "https://b.example/auth"},
		}},
		"a-provider": {Config: oauth2.Config{
			ClientID: "a",
			Endpoint: oauth2.Endpoint{AuthURL: "https://a.example/auth"},
		}},
	}, nil)

	resp, err := svc.Settings(context.TODO(), &apiauth.SettingsRequest{})
	assert.NoError(t, err)
	if !assert.Len(t, resp.Items, 2) {
		return
	}
	// 排序稳定：按 provider 名字升序
	assert.Equal(t, "a-provider", resp.Items[0].Name)
	assert.Equal(t, "b-provider", resp.Items[1].Name)

	state := resp.Items[0].State
	assert.NotEmpty(t, state)
	assert.Equal(t, state, resp.Items[1].State, "所有 provider 必须共用同一个 state")
	assert.Contains(t, resp.Items[0].Url, state)
	assert.Contains(t, resp.Items[1].Url, state)
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
	userBiz   *biz.MockUserBiz
}

func newAuthSvcWithMocks(t *testing.T) (*authSvc, *authSvcMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mocks := &authSvcMocks{
		ctrl:      ctrl,
		eventRepo: data.NewMockEventRepo(ctrl),
		authBiz:   biz.NewMockAuthBiz(ctrl),
		userBiz:   biz.NewMockUserBiz(ctrl),
	}
	s, ok := NewAuthSvc(AuthSvcDeps{
		EventBiz: biz.NewEventBiz(mocks.eventRepo),
		Logger:   mlog.NewForConfig(nil),
		AuthBiz:  mocks.authBiz,
		UserBiz:  mocks.userBiz,
	}).(*authSvc)
	if !ok {
		panic("NewAuthSvc returned unexpected type")
	}
	return s, mocks
}
