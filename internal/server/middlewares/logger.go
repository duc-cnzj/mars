package middlewares

import (
	"context"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"google.golang.org/grpc"
)

func LoggerUnaryServerInterceptor(logger mlog.Logger) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if request, ok := req.(interface {
			String() string
		}); ok {
			logger.Debugf("[request logger]: method=%s body=%v", info.FullMethod, request.String())
		}
		return handler(ctx, req)
	}
}
