package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// TestLoginUnaryServerInterceptor 覆盖 Unary 登录拦截器的三分支：公开方法跳过校验、
// 私有方法 authenticate 成功注入用户后放行、私有方法 authenticate 失败不进 handler。
func TestLoginUnaryServerInterceptor(t *testing.T) {
	authCalled := 0
	authFn := func(ctx context.Context) (context.Context, error) {
		authCalled++
		return biz.SetUser(ctx, &biz.UserInfo{Name: "duc"}), nil
	}
	login := LoginUnaryServerInterceptor(authFn)

	// 公开方法：不调 authenticate，直接进 handler。
	handled := 0
	_, err := login(context.TODO(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/auth.Auth/Login",
	}, func(ctx context.Context, req any) (any, error) {
		handled++
		return "ok", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, authCalled)
	assert.Equal(t, 1, handled)

	// 私有方法：authenticate 注入用户后进 handler，handler 可见注入的 user。
	handled = 0
	_, err = login(context.TODO(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/namespace.Namespace/List",
	}, func(ctx context.Context, req any) (any, error) {
		handled++
		assert.Equal(t, "duc", biz.MustGetUser(ctx).Name)
		return "ok", nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, authCalled)
	assert.Equal(t, 1, handled)

	// 私有方法：authenticate 失败，返回错误且不进 handler。
	handled = 0
	_, err = LoginUnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
		return nil, errors.New("no token")
	})(context.TODO(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/namespace.Namespace/List",
	}, func(ctx context.Context, req any) (any, error) {
		handled++
		return nil, nil
	})
	assert.Error(t, err)
	assert.Equal(t, 0, handled)
}

// TestLoginStreamServerInterceptor 覆盖 Stream 登录拦截器的三分支：公开方法跳过校验、
// 私有方法 authenticate 成功注入用户后放行、私有方法 authenticate 失败不进 handler。
// ss 复用 validator_test.go 中的测试桩实现 grpc.ServerStream。
func TestLoginStreamServerInterceptor(t *testing.T) {
	authCalled := 0
	authFn := func(ctx context.Context) (context.Context, error) {
		authCalled++
		return biz.SetUser(ctx, &biz.UserInfo{Name: "duc"}), nil
	}
	login := LoginStreamServerInterceptor(authFn)

	// 公开方法：不调 authenticate，直接进 handler。
	handled := 0
	err := login(nil, &ss{}, &grpc.StreamServerInfo{
		FullMethod: "/version.Version/Version",
	}, func(srv any, stream grpc.ServerStream) error {
		handled++
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, authCalled)
	assert.Equal(t, 1, handled)

	// 私有方法：authenticate 注入用户后进 handler，handler 从 stream.Context() 可见注入的 user。
	handled = 0
	err = login(nil, &ss{}, &grpc.StreamServerInfo{
		FullMethod: "/container.Container/StreamContainerLog",
	}, func(srv any, stream grpc.ServerStream) error {
		handled++
		assert.Equal(t, "duc", biz.MustGetUser(stream.Context()).Name)
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, authCalled)
	assert.Equal(t, 1, handled)

	// 私有方法：authenticate 失败，返回错误且不进 handler。
	handled = 0
	err = LoginStreamServerInterceptor(func(ctx context.Context) (context.Context, error) {
		return nil, errors.New("no token")
	})(nil, &ss{}, &grpc.StreamServerInfo{
		FullMethod: "/container.Container/StreamContainerLog",
	}, func(srv any, stream grpc.ServerStream) error {
		handled++
		return nil
	})
	assert.Error(t, err)
	assert.Equal(t, 0, handled)
}

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
