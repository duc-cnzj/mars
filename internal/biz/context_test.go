package biz

// context_test.go 覆盖用户 context 存取（原 internal/auth 包的测试迁移）：
// 注入/取回、缺失 panic、nil 注入、key 类型隔离四条契约。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestContextWithUser 验证 SetUser 注入用户后，GetUser / MustGetUser 能取回同一用户。
func TestContextWithUser(t *testing.T) {
	ctx := context.TODO()
	userInfo := &UserInfo{
		ID:    "1",
		Name:  "Test User",
		Email: "test@example.com",
	}

	ctxWithUser := SetUser(ctx, userInfo)

	retrievedUser, err := GetUser(ctxWithUser)
	assert.Nil(t, err)
	assert.Equal(t, userInfo, retrievedUser)

	retrievedUser = MustGetUser(ctxWithUser)
	assert.Equal(t, userInfo, retrievedUser)
}

// TestContextWithoutUser 验证未注入用户的 context 上 GetUser 返回错误、
// MustGetUser 必须 panic——用户缺失是编程错误，不能返回 nil 向下游传递。
func TestContextWithoutUser(t *testing.T) {
	ctx := context.TODO()

	_, err := GetUser(ctx)
	assert.NotNil(t, err)

	assert.Panics(t, func() { MustGetUser(ctx) })
}

// TestContextWithNilUser 验证 SetUser 注入 nil 用户（上游编程错误）时：
// GetUser 不 panic 而是返回错误，MustGetUser 因用户缺失而 panic。
// 同时覆盖带类型 nil 指针 (*UserInfo)(nil)：断言成功但 info==nil，
// 与未注入（断言失败）是 GetUser 里两条不同分支。
func TestContextWithNilUser(t *testing.T) {
	// untyped nil：ctx.Value 返回接口 nil，类型断言失败（ok=false 分支）。
	ctx := SetUser(context.TODO(), nil)
	_, err := GetUser(ctx)
	assert.Error(t, err)
	assert.Panics(t, func() { MustGetUser(ctx) })

	// 带类型 nil 指针 (*UserInfo)(nil)：断言成功但 info==nil（info!=nil 分支）。
	ctx2 := SetUser(context.TODO(), (*UserInfo)(nil))
	_, err2 := GetUser(ctx2)
	assert.Error(t, err2)
	assert.Panics(t, func() { MustGetUser(ctx2) })
}

// TestContextKeyIsolation 验证 key 类型隔离：用其他类型作 key 塞入的值，
// GetUser 必须拿不到——context key 依赖的是类型唯一性而非值内容。
func TestContextKeyIsolation(t *testing.T) {
	// 用一个与 ctxTokenInfo 无关的私有 key 类型注入假用户。
	type unrelatedKey struct{}
	ctx := context.WithValue(context.TODO(), unrelatedKey{}, &UserInfo{ID: "1"})

	_, err := GetUser(ctx)
	assert.Error(t, err)
	assert.Panics(t, func() { MustGetUser(ctx) })
}

// Test_authenticate_Success 验证 authenticate 校验通过后把用户注入 ctx：
// 返回的新 ctx 里 MustGetUser 能取回同一用户（纯 token 校验基座，角色取登录身份/JWT）。
func Test_authenticate_Success(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	auth := NewMockAuthBiz(m)
	user := &UserInfo{ID: "1", Name: "duc", Email: "duc@example.com"}
	auth.EXPECT().VerifyToken(gomock.Any(), "token-1").Return(user, nil)

	ctx, err := authenticate(context.TODO(), auth, "token-1")
	assert.NoError(t, err)
	assert.Equal(t, user, MustGetUser(ctx))
}

// Test_authenticate_InvalidToken 验证 token 校验失败时 authenticate 返回原始错误，
// 且不产生可消费的 ctx（返回 nil——失败契约是"不进入业务逻辑"，中间件据此直接 401）。
func Test_authenticate_InvalidToken(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	auth := NewMockAuthBiz(m)
	verifyErr := errors.New("verify failed")
	auth.EXPECT().VerifyToken(gomock.Any(), "bad-token").Return(nil, verifyErr)

	ctx, err := authenticate(context.TODO(), auth, "bad-token")
	assert.ErrorIs(t, err, verifyErr)
	assert.Nil(t, ctx)
}

// TestAuthenticate_AppliesEffectiveRoles 验证有效角色解析链路：VerifyToken 通过后，
// authBiz.EffectiveRoles 按 users 表接管状态返回生效角色并覆盖注入用户（后台降权后 JWT
// 仍带的 mars_admin 不再生效，RequireAdmin 据此拒绝）。
func TestAuthenticate_AppliesEffectiveRoles(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	auth := NewMockAuthBiz(m)
	user := &UserInfo{Email: "duc@x.com", Roles: []string{MarsAdmin}}
	auth.EXPECT().VerifyToken(gomock.Any(), "token-1").Return(user, nil)
	auth.EXPECT().EffectiveRoles(gomock.Any(), "duc@x.com", []string{MarsAdmin}).Return([]string{}, nil)

	ctx, err := Authenticate(context.TODO(), auth, "token-1")
	assert.NoError(t, err)
	assert.Equal(t, []string{}, MustGetUser(ctx).Roles, "降权后生效角色不应含 mars_admin")
}

// TestAuthenticate_RepoErrorFallsBack 用户表读取失败回落登录身份角色：
// 鉴权不阻断，返回携带原角色的 ctx（DB 恢复后手动接管由下一次请求自动生效）。
func TestAuthenticate_RepoErrorFallsBack(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	auth := NewMockAuthBiz(m)
	user := &UserInfo{Email: "duc@x.com", Roles: []string{MarsAdmin}}
	auth.EXPECT().VerifyToken(gomock.Any(), "token-1").Return(user, nil)
	auth.EXPECT().EffectiveRoles(gomock.Any(), "duc@x.com", []string{MarsAdmin}).Return(nil, errors.New("db boom"))

	ctx, err := Authenticate(context.TODO(), auth, "token-1")
	assert.NoError(t, err, "用户表读取失败不阻断鉴权")
	assert.Equal(t, []string{MarsAdmin}, MustGetUser(ctx).Roles, "回落登录身份角色")
}

// TestAuthenticate_InvalidToken_EffectiveRolesUntouched VerifyToken 失败透传原始错误并返回 nil ctx，
// 不触达 EffectiveRoles。
func TestAuthenticate_InvalidToken_EffectiveRolesUntouched(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	auth := NewMockAuthBiz(m)
	verifyErr := errors.New("verify failed")
	auth.EXPECT().VerifyToken(gomock.Any(), "bad").Return(nil, verifyErr)

	ctx, err := Authenticate(context.TODO(), auth, "bad")
	assert.ErrorIs(t, err, verifyErr)
	assert.Nil(t, ctx)
}
