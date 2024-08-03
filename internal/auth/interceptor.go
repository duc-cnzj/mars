package auth

import (
	"context"
	"errors"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
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

func StreamServerInterceptor() grpc.StreamServerInterceptor {
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

type ctxTokenInfo struct{}

func SetUser(ctx context.Context, info *UserInfo) context.Context {
	return context.WithValue(ctx, &ctxTokenInfo{}, info)
}

func GetUser(ctx context.Context) (*UserInfo, error) {
	if info, ok := ctx.Value(&ctxTokenInfo{}).(*UserInfo); ok {
		return info, nil
	}

	return nil, errors.New("user not found")
}

func MustGetUser(ctx context.Context) *UserInfo {
	info, _ := ctx.Value(&ctxTokenInfo{}).(*UserInfo)
	return info
}
