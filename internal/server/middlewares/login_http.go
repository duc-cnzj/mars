package middlewares

import (
	"context"
	"net/http"
)

// LoginHTTP 是 HTTP 版登录中间件：从 Authorization header 提取 token，经 verify
// 校验并把用户注入新 ctx 后放行，失败统一返回 401。与 gRPC Login*ServerInterceptor
// 语义对齐，作为 HTTP 侧路由鉴权的统一入口——鉴权核心（校验+注入）由 verify
// 承载（通常指向 biz.Authenticate），杜绝各 HTTP handler 手写第二套鉴权实现。
func LoginHTTP(verify func(ctx context.Context, token string) (context.Context, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := verify(r.Context(), r.Header.Get("Authorization"))
			if err != nil {
				http.Error(w, "Unauthenticated", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
