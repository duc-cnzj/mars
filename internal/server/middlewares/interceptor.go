package middlewares

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

type Authorize interface {
	Authorize(ctx context.Context, fullMethodName string) (context.Context, error)
}

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
