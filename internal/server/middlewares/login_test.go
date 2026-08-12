package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

// TestLoginUnaryServerInterceptor 覆盖 Unary 登录拦截器的三分支：公开方法跳过校验、
// 私有方法 authenticate 成功注入用户后放行、私有方法 authenticate 失败不进 handler；
// 失败分支另以 mock logger 承重断言 [auth audit] Warning 审计日志落盘（401 审计兜底）。
func TestLoginUnaryServerInterceptor(t *testing.T) {
	authCalled := 0
	authFn := func(ctx context.Context) (context.Context, error) {
		authCalled++
		return biz.SetUser(ctx, &biz.UserInfo{Name: "duc"}), nil
	}
	login := LoginUnaryServerInterceptor(authFn, mlog.NewForConfig(nil))

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

	// 私有方法：authenticate 失败，返回错误、不进 handler，且打一条 [auth audit] 审计日志。
	handled = 0
	m := gomock.NewController(t)
	defer m.Finish()
	audit := mlog.NewMockLogger(m)
	audit.EXPECT().Warningf("[auth audit]: method=%s auth failed: %v", "/namespace.Namespace/List", gomock.Any()).Times(1)
	_, err = LoginUnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
		return nil, errors.New("no token")
	}, audit)(context.TODO(), nil, &grpc.UnaryServerInfo{
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
	login := LoginStreamServerInterceptor(authFn, mlog.NewForConfig(nil))

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

	// 私有方法：authenticate 失败，返回错误、不进 handler，且打一条 [auth audit] 审计日志。
	handled = 0
	m := gomock.NewController(t)
	defer m.Finish()
	audit := mlog.NewMockLogger(m)
	audit.EXPECT().Warningf("[auth audit]: method=%s auth failed: %v", "/container.Container/StreamContainerLog", gomock.Any()).Times(1)
	err = LoginStreamServerInterceptor(func(ctx context.Context) (context.Context, error) {
		return nil, errors.New("no token")
	}, audit)(nil, &ss{}, &grpc.StreamServerInfo{
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
	LoginHTTP(verify, mlog.NewForConfig(nil))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		assert.Equal(t, user, r.Context().Value(ctxKey{}))
	})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))

	assert.True(t, called, "next 必须被调用")
}

// TestLoginHTTP_Unauthorized 验证校验失败时中间件写 401、不进入业务 handler，
// 并以 mock logger 承重断言 [auth audit] Warning 审计日志落盘（401 审计兜底）。
func TestLoginHTTP_Unauthorized(t *testing.T) {
	verify := func(ctx context.Context, token string) (context.Context, error) {
		return nil, errors.New("invalid token")
	}

	m := gomock.NewController(t)
	defer m.Finish()
	audit := mlog.NewMockLogger(m)
	audit.EXPECT().Warningf("[auth audit]: path=%s auth failed: %v", "/x", gomock.Any()).Times(1)

	called := false
	rec := httptest.NewRecorder()
	LoginHTTP(verify, audit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	LoginHTTP(verify, mlog.NewForConfig(nil))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).
		ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, "Bearer abc123", gotToken)
}
