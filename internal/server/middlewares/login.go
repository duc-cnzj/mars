package middlewares

import (
	"context"
	"net/http"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// LoginUnaryServerInterceptor 是登录校验的 Unary 拦截器：命中 biz.IsPublicMethod 白名单的
// 免登录方法跳过 token 校验直接进 handler，其余方法先 authenticate 注入用户上下文；失败
// （未携带/无效 token）返回 Unauthenticated，与原先 grpc_auth 的语义一致。免登录白名单
// 归属 biz 层（访问控制契约，见 biz/public_methods.go），本层只消费 IsPublicMethod。
// 认证失败经 logger 打一条 Warning 审计日志（[auth audit]）：AccessLog 移置 Login 之后，
// Login 失败直接 return 时内层 AccessLog 不会执行，401 审计由本拦截器兜底（见 grpc.go 链序注释）。
func LoginUnaryServerInterceptor(authenticate func(ctx context.Context) (context.Context, error), logger mlog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if biz.IsPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}
		ctx, err = authenticate(ctx)
		if err != nil {
			logger.Warningf("[auth audit]: method=%s auth failed: %v", info.FullMethod, err)
			return nil, err
		}
		return handler(ctx, req)
	}
}

// LoginStreamServerInterceptor 是登录校验的 Stream 拦截器：命中 biz.IsPublicMethod 的
// 免登录方法直接放行，其余方法 authenticate 后通过 WrapServerStream 把注入用户的新
// context 传递给 handler。认证失败同 Unary 版打 [auth audit] Warning 审计日志兜底 401。
func LoginStreamServerInterceptor(authenticate func(ctx context.Context) (context.Context, error), logger mlog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if biz.IsPublicMethod(info.FullMethod) {
			return handler(srv, ss)
		}
		newCtx, err := authenticate(ss.Context())
		if err != nil {
			logger.Warningf("[auth audit]: method=%s auth failed: %v", info.FullMethod, err)
			return err
		}
		wrapped := grpc_middleware.WrapServerStream(ss)
		wrapped.WrappedContext = newCtx
		return handler(srv, wrapped)
	}
}

// LoginHTTP 是 HTTP 版登录中间件：从 Authorization header 提取 token，经 verify
// 校验并把用户注入新 ctx 后放行，失败统一返回 401。与 gRPC Login*ServerInterceptor
// 语义对齐，作为 HTTP 侧路由鉴权的统一入口——鉴权核心（校验+注入）由 verify
// 承载（通常指向 biz.Authenticate），杜绝各 HTTP handler 手写第二套鉴权实现。
// 校验失败经 logger 打一条 [auth audit] Warning 审计日志，与 gRPC 侧 401 审计对齐。
func LoginHTTP(verify func(ctx context.Context, token string) (context.Context, error), logger mlog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := verify(r.Context(), r.Header.Get("Authorization"))
			if err != nil {
				logger.Warningf("[auth audit]: path=%s auth failed: %v", r.URL.Path, err)
				http.Error(w, "Unauthenticated", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
