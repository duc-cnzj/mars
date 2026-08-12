package middlewares

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"google.golang.org/grpc"
)

// AccessLogUnaryServerInterceptor 是 Unary 访问日志拦截器：每个请求打一条 Info 日志，
// 记录调用用户、方法名与耗时。须置于登录拦截器之后——Login 验签成功把用户注入新 ctx
// 传给内层拦截器，本拦截器收到的 ctx 已带用户，grpcUser 才能打印出调用用户；
// 代价：401 请求（Login 失败）不产生访问日志，401 审计由 Login 拦截器内的
// [auth audit] Warning 日志兜底、Metrics 兜住监控（见 grpc.go 链序注释）。
func AccessLogUnaryServerInterceptor(logger mlog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func(t time.Time) {
			logger.Infof("[Grpc]: user: %v, visit: %v, use: %s.", grpcUser(ctx).Name, info.FullMethod, time.Since(t))
		}(time.Now())

		return handler(ctx, req)
	}
}

// AccessLogStreamServerInterceptor 是 Stream 访问日志拦截器：流会话结束时打一条 Info 日志。
// 注意 use 是整段流会话的存活时长（长连接流如日志流可持续分钟级），而非单次请求耗时——
// 与 Unary 版的"请求耗时"语义不同，阅读日志时须区分。
func AccessLogStreamServerInterceptor(logger mlog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		defer func(t time.Time) {
			logger.Infof("[Grpc]: user: %v, visit: %v, use: %s.", grpcUser(ss.Context()).Name, info.FullMethod, time.Since(t))
		}(time.Now())

		return handler(srv, ss)
	}
}

// grpcUser 从 ctx 解析调用用户：未注入用户时返回空 UserInfo 而非 nil，保证日志侧
// 能安全取 Name 且语义上区分"匿名调用"（公开方法免登录时用户不存在）。
func grpcUser(ctx context.Context) *biz.UserInfo {
	if u, err := biz.GetUser(ctx); err == nil {
		return u
	}
	return &biz.UserInfo{}
}
