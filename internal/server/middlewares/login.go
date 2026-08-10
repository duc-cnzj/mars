package middlewares

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// PublicMethods 是 mars 全部免登录 gRPC 方法的 fullMethodName 白名单：登录拦截器命中
// 白名单时跳过 Bearer token 校验直接放行，未命中一律要求有效 token。
//
// 白名单与 doc/access_control.md §4.1「免登录服务」清单逐行一致：auth.Login/
// Settings/Exchange/Info（Info 入口免登录、方法内自校验 token）、cluster.ClusterInfo、
// picture.Background、version.Version。新增免登录方法必须同时更新本白名单与文档，
// public_methods_test.go 的契约测试会在二者漂移时失败。
//
// 相比原先的 guest 内嵌（AuthFuncOverride 无条件放行整个服务）：白名单把"公开"从
// 服务粒度收窄到方法粒度——新方法默认私有（安全默认），"公开"判定单一归属本处。
var PublicMethods = map[string]struct{}{
	"/auth.Auth/Login":             {},
	"/auth.Auth/Settings":          {},
	"/auth.Auth/Exchange":          {},
	"/auth.Auth/Info":              {},
	"/cluster.Cluster/ClusterInfo": {},
	"/picture.Picture/Background":  {},
	"/version.Version/Version":     {},
}

// IsPublicMethod 判断 fullMethodName 是否命中公开白名单（免登录放行）。
func IsPublicMethod(fullMethodName string) bool {
	_, ok := PublicMethods[fullMethodName]
	return ok
}

// LoginUnaryServerInterceptor 是登录校验的 Unary 拦截器：命中 PublicMethods 白名单的
// 方法跳过 token 校验直接进 handler，其余方法先 authenticate 注入用户上下文；失败
// （未携带/无效 token）返回 Unauthenticated，与原先 grpc_auth 的语义一致，仅把
// "公开"判定从服务内嵌收进本层。
func LoginUnaryServerInterceptor(authenticate func(ctx context.Context) (context.Context, error)) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if IsPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}
		ctx, err = authenticate(ctx)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// LoginStreamServerInterceptor 是登录校验的 Stream 拦截器：公开方法直接放行，其余
// 方法 authenticate 后通过 WrapServerStream 把注入用户的新 context 传递给 handler。
func LoginStreamServerInterceptor(authenticate func(ctx context.Context) (context.Context, error)) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if IsPublicMethod(info.FullMethod) {
			return handler(srv, ss)
		}
		newCtx, err := authenticate(ss.Context())
		if err != nil {
			return err
		}
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}
