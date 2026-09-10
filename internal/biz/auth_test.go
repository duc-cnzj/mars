package biz

// auth_test.go 覆盖 authBiz 的凭据校验用例：Login/VerifyToken/Sign/Settings。
// authBiz 依赖 Auth 端口（Sign/VerifyToken），用 fakeAuthForBiz 替身注入，不触达真实 JWT。

import (
	"context"
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeAuthForBiz 只覆写 authBiz 用到的两个 Auth 方法，其余由接口兜底。
type fakeAuthForBiz struct {
	Auth
	verifyToken func(token string) (*JwtClaims, bool)
	sign        func(*UserInfo) (*SignData, error)
}

func (f *fakeAuthForBiz) VerifyToken(token string) (*JwtClaims, bool) {
	return f.verifyToken(token)
}
func (f *fakeAuthForBiz) Sign(u *UserInfo) (*SignData, error) {
	return f.sign(u)
}

// fakeAuthConfigProvider 是 AuthConfigProvider 的测试替身：直接返回注入的
// adminPassword 与 oidcConfig（oidcConfig 可为 nil，此时 OidcConfig() 返回 nil）。
type fakeAuthConfigProvider struct {
	adminPassword string
	oidcConfig    func() OidcConfig
}

func (f fakeAuthConfigProvider) AdminPassword() string {
	return f.adminPassword
}

func (f fakeAuthConfigProvider) OidcConfig() OidcConfig {
	if f.oidcConfig == nil {
		return nil
	}
	return f.oidcConfig()
}

// fakeRolesProvider 是 EffectiveRolesProvider 的测试替身：记录入参并按注入的
// 结果/错误返回，供 authBiz.EffectiveRoles 用例验证「邮箱 trim + 透传」编排。
type fakeRolesProvider struct {
	inEmail    string
	inSSORoles []string
	out        []string
	err        error
}

func (f *fakeRolesProvider) EffectiveRoles(_ context.Context, email string, ssoRoles []string) ([]string, error) {
	f.inEmail = email
	f.inSSORoles = ssoRoles
	return f.out, f.err
}

// newAuthBizForTest 组装 authBiz 测试实例：auth 与 oidcConfig 均可为 nil，依赖由测试注入。
// roles 为 nil 时 authBiz 的生效角色解析保持 nil——既有用例不触碰 EffectiveRoles，不触发解引用。
func newAuthBizForTest(auth Auth, adminPassword string, oidcConfig func() OidcConfig) *authBiz {
	return NewAuthBiz(auth, fakeAuthConfigProvider{adminPassword: adminPassword, oidcConfig: oidcConfig}, nil, mlog.NewForConfig(nil)).(*authBiz)
}

// newAuthBizForTestWithRoles 组装带生效角色解析器的 authBiz 测试实例，供 EffectiveRoles 用例注入 fakeRolesProvider。
func newAuthBizForTestWithRoles(auth Auth, roles EffectiveRolesProvider) *authBiz {
	return NewAuthBiz(auth, fakeAuthConfigProvider{}, roles, mlog.NewForConfig(nil)).(*authBiz)
}

func TestAuthBiz_Login_Success(t *testing.T) {
	a := newAuthBizForTest(&fakeAuthForBiz{sign: func(u *UserInfo) (*SignData, error) {
		assert.Equal(t, "超级管理员", u.Name)
		return &SignData{Token: "token-1", ExpiredIn: 3600}, nil
	}}, "secret", nil)
	resp, err := a.Login(context.TODO(), &LoginInput{Username: "admin", Password: "secret"})
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, "token-1", resp.Token)
		assert.Equal(t, int64(3600), resp.ExpiredIn)
		assert.Equal(t, adminUserInfo, resp.UserInfo)
	}
}

func TestAuthBiz_Login_WrongPassword(t *testing.T) {
	a := newAuthBizForTest(&fakeAuthForBiz{}, "secret", nil)
	_, err := a.Login(context.TODO(), &LoginInput{Username: "admin", Password: "wrong"})
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthBiz_Login_SignError(t *testing.T) {
	a := newAuthBizForTest(&fakeAuthForBiz{sign: func(u *UserInfo) (*SignData, error) {
		return nil, errors.New("sign boom")
	}}, "secret", nil)
	_, err := a.Login(context.TODO(), &LoginInput{Username: "admin", Password: "secret"})
	assert.Error(t, err)
}

func TestAuthBiz_VerifyToken_Success(t *testing.T) {
	a := newAuthBizForTest(&fakeAuthForBiz{verifyToken: func(token string) (*JwtClaims, bool) {
		assert.Equal(t, "t", token)
		return &JwtClaims{UserInfo: &UserInfo{Email: "DUC@EXAMPLE.COM", Name: "duc"}}, true
	}}, "", nil)
	u, err := a.VerifyToken(context.TODO(), "t")
	assert.NoError(t, err)
	if assert.NotNil(t, u) {
		// email 统一小写，供 OIDC 用户匹配。
		assert.Equal(t, "duc@example.com", u.Email)
	}
}

func TestAuthBiz_VerifyToken_Fail(t *testing.T) {
	a := newAuthBizForTest(&fakeAuthForBiz{verifyToken: func(token string) (*JwtClaims, bool) {
		return nil, false
	}}, "", nil)
	_, err := a.VerifyToken(context.TODO(), "bad")
	assert.Error(t, err)
}

func TestAuthBiz_Sign_Success(t *testing.T) {
	u := &UserInfo{Name: "duc"}
	a := newAuthBizForTest(&fakeAuthForBiz{sign: func(input *UserInfo) (*SignData, error) {
		assert.Equal(t, u, input)
		return &SignData{Token: "token-2", ExpiredIn: 7200}, nil
	}}, "", nil)
	resp, err := a.Sign(context.TODO(), u)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, "token-2", resp.Token)
		assert.Equal(t, int64(7200), resp.ExpiredIn)
		assert.Equal(t, u, resp.UserInfo)
	}
}

func TestAuthBiz_Sign_Error(t *testing.T) {
	a := newAuthBizForTest(&fakeAuthForBiz{sign: func(input *UserInfo) (*SignData, error) {
		return nil, errors.New("sign boom")
	}}, "", nil)
	_, err := a.Sign(context.TODO(), &UserInfo{Name: "duc"})
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthBiz_Settings_NilConfig(t *testing.T) {
	a := newAuthBizForTest(&fakeAuthForBiz{}, "", nil)
	got, err := a.Settings(context.TODO())
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestAuthBiz_Settings_WithConfig(t *testing.T) {
	a := newAuthBizForTest(&fakeAuthForBiz{}, "", func() OidcConfig {
		return OidcConfig{"x": OidcConfigItem{EndSessionEndpoint: "https://logout"}}
	})
	got, err := a.Settings(context.TODO())
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "https://logout", got["x"].EndSessionEndpoint)
}

func TestAuthBiz_Login_NilInput(t *testing.T) {
	a := newAuthBizForTest(nil, "secret", nil)
	got, err := a.Login(context.TODO(), nil)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "login input 不能为空", status.Convert(err).Message())
}

func TestAuthBiz_Sign_NilInput(t *testing.T) {
	a := newAuthBizForTest(nil, "secret", nil)
	got, err := a.Sign(context.TODO(), nil)
	assert.Nil(t, got)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, "sign user 不能为空", status.Convert(err).Message())
}

// TestAuthBiz_EffectiveRoles_Success 成功路径：邮箱 trim 后透传 email/SSO 角色到 provider，
// 生效角色取 provider 结果。
func TestAuthBiz_EffectiveRoles_Success(t *testing.T) {
	roles := &fakeRolesProvider{out: []string{}}
	a := newAuthBizForTestWithRoles(nil, roles)

	got, err := a.EffectiveRoles(context.TODO(), "  a@b.c  ", []string{MarsAdmin})
	assert.NoError(t, err)
	assert.Equal(t, "a@b.c", roles.inEmail, "邮箱应 trim 后传给 provider")
	assert.Equal(t, []string{MarsAdmin}, roles.inSSORoles, "SSO 角色应原样透传")
	assert.Equal(t, []string{}, got)
}

// TestAuthBiz_EffectiveRoles_EmptyEmailFallsBack 空邮箱回落登录身份角色（鉴权路径不阻断）：
// 返回原角色且不触达 provider。
func TestAuthBiz_EffectiveRoles_EmptyEmailFallsBack(t *testing.T) {
	roles := &fakeRolesProvider{}
	a := newAuthBizForTestWithRoles(nil, roles)

	got, err := a.EffectiveRoles(context.TODO(), "  ", []string{MarsAdmin})
	assert.NoError(t, err)
	assert.Equal(t, []string{MarsAdmin}, got, "空邮箱应回落登录身份角色")
	assert.Empty(t, roles.inEmail, "空邮箱不应触达 provider")
}

// TestAuthBiz_EffectiveRoles_RepoError 透传 provider 错误。
func TestAuthBiz_EffectiveRoles_RepoError(t *testing.T) {
	roles := &fakeRolesProvider{err: errors.New("boom")}
	a := newAuthBizForTestWithRoles(nil, roles)

	_, err := a.EffectiveRoles(context.TODO(), "a@b.c", []string{})
	assert.EqualError(t, err, "boom")
}
