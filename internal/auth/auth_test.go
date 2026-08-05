package auth

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/stretchr/testify/assert"
)

// TestContextWithUser 验证 SetUser 注入用户后，GetUser / MustGetUser 能取回同一用户。
func TestContextWithUser(t *testing.T) {
	ctx := context.TODO()
	userInfo := &biz.UserInfo{
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
func TestContextWithNilUser(t *testing.T) {
	ctx := SetUser(context.TODO(), nil)

	_, err := GetUser(ctx)
	assert.Error(t, err)

	assert.Panics(t, func() { MustGetUser(ctx) })
}

// TestContextKeyIsolation 验证 key 类型隔离：用其他类型作 key 塞入的值，
// GetUser 必须拿不到——context key 依赖的是类型唯一性而非值内容。
func TestContextKeyIsolation(t *testing.T) {
	// 用一个与本包 ctxTokenInfo 无关的私有 key 类型注入假用户。
	type unrelatedKey struct{}
	ctx := context.WithValue(context.TODO(), unrelatedKey{}, &biz.UserInfo{ID: "1"})

	_, err := GetUser(ctx)
	assert.Error(t, err)
	assert.Panics(t, func() { MustGetUser(ctx) })
}
