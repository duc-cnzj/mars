package middlewares

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// Authorize 是可选的服务级授权能力：实现它的 gRPC 服务（info.Server 断言）会在每次
// 调用前被询问 fullMethodName 是否可放行，由服务自身收敛业务授权判定。
type Authorize interface {
	Authorize(ctx context.Context, fullMethodName string) (context.Context, error)
}

// AuthUnaryServerInterceptor 是 Unary 授权拦截器：服务实现了 Authorize 时调用其
// Authorize 校验，失败直接返回；未实现则原样透传（无授权要求的服务不受影响）。
func AuthUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if authorizeInterface, ok := info.Server.(Authorize); ok {
			ctx, err = authorizeInterface.Authorize(ctx, info.FullMethod)
			if err != nil {
				return nil, err
			}
		}

		return handler(ctx, req)
	}
}

// AuthStreamServerInterceptor 是 Stream 授权拦截器：服务实现 Authorize 时先校验再
// 用 WrapServerStream 把注入的 context 传给 handler；未实现则原样透传。
func AuthStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var (
			newCtx context.Context
			err    error
		)
		if authorizeInterface, ok := srv.(Authorize); ok {
			newCtx, err = authorizeInterface.Authorize(ss.Context(), info.FullMethod)
			if err != nil {
				return err
			}
			wrapped := grpc_middleware.WrapServerStream(ss)
			wrapped.WrappedContext = newCtx

			return handler(srv, wrapped)
		}

		return handler(srv, ss)
	}
}
