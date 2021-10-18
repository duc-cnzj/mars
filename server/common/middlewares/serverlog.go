package middlewares

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/internal/mlog"
	"google.golang.org/grpc"
)

func ServerLog() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func(t time.Time) {
			mlog.Debugf("method: %v, use %v", info.FullMethod, time.Since(t))
		}(time.Now())
		return handler(ctx, req)
	}
}
