package middlewares

// login_http_test.go 覆盖 LoginHTTP 中间件契约：校验通过放行并把用户注入新 ctx、
// 校验失败统一 401 且不进入业务 handler、Authorization header 原样传给 verify。

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoginHTTP_Success 验证校验通过时中间件放行，且 next 收到携带注入用户的新 ctx。
func TestLoginHTTP_Success(t *testing.T) {
	type ctxKey struct{}
	user := "duc"
	verify := func(ctx context.Context, token string) (context.Context, error) {
		return context.WithValue(ctx, ctxKey{}, user), nil
	}

	called := false
	LoginHTTP(verify)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		assert.Equal(t, user, r.Context().Value(ctxKey{}))
	})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))

	assert.True(t, called, "next 必须被调用")
}

// TestLoginHTTP_Unauthorized 验证校验失败时中间件写 401，且不进入业务 handler。
func TestLoginHTTP_Unauthorized(t *testing.T) {
	verify := func(ctx context.Context, token string) (context.Context, error) {
		return nil, errors.New("invalid token")
	}

	called := false
	rec := httptest.NewRecorder()
	LoginHTTP(verify)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, called, "校验失败时 next 不得被调用")
}

// TestLoginHTTP_TokenFromHeader 验证 verify 收到的是请求头 Authorization 的原始值。
func TestLoginHTTP_TokenFromHeader(t *testing.T) {
	var gotToken string
	verify := func(ctx context.Context, token string) (context.Context, error) {
		gotToken = token
		return ctx, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer abc123")
	LoginHTTP(verify)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).
		ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, "Bearer abc123", gotToken)
}
