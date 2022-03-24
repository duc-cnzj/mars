package auth

import (
	"context"
	"errors"

	"github.com/duc-cnzj/mars/internal/contracts"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if authorizeInterface, ok := info.Server.(contracts.AuthorizeInterface); ok {
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
		if authorizeInterface, ok := srv.(contracts.AuthorizeInterface); ok {
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

type CtxTokenInfo struct{}

func SetUser(ctx context.Context, info *contracts.UserInfo) context.Context {
	return context.WithValue(ctx, &CtxTokenInfo{}, info)
}

func GetUser(ctx context.Context) (*contracts.UserInfo, error) {
	if info, ok := ctx.Value(&CtxTokenInfo{}).(*contracts.UserInfo); ok {
		return info, nil
	}

	return nil, errors.New("user not found")
}

func MustGetUser(ctx context.Context) *contracts.UserInfo {
	info, _ := ctx.Value(&CtxTokenInfo{}).(*contracts.UserInfo)
	return info
}
