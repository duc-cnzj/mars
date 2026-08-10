package middlewares

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// TestPublicMethods_AlignsWithAccessControlDoc 是公开白名单的契约测试：断言白名单与
// doc/access_control.md §4.1「免登录服务」清单逐行一致。新增免登录方法必须同时更新
// 本测试、PublicMethods 与文档三处，任何一处漏改都会在此失败（防契约与实现漂移）。
func TestPublicMethods_AlignsWithAccessControlDoc(t *testing.T) {
	want := []string{
		"/auth.Auth/Exchange",
		"/auth.Auth/Info",
		"/auth.Auth/Login",
		"/auth.Auth/Settings",
		"/cluster.Cluster/ClusterInfo",
		"/picture.Picture/Background",
		"/version.Version/Version",
	}
	// 白名单全 key 排序后与文档清单比对（排序逻辑内联于此，生产代码不为此保留导出函数）。
	got := make([]string, 0, len(PublicMethods))
	for name := range PublicMethods {
		got = append(got, name)
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PublicMethods 与 doc §4.1 漂移\n got: %v\nwant: %v", got, want)
	}
	for _, name := range want {
		if !IsPublicMethod(name) {
			t.Errorf("白名单缺少公开方法: %s", name)
		}
	}
	// 非白名单方法（如私有服务的命名空间级方法）一律不命中，防止误放行。
	if IsPublicMethod("/namespace.Namespace/List") {
		t.Errorf("私有方法不应命中白名单")
	}
}

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
