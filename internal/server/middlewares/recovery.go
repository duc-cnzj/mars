package middlewares

import (
	"net/http"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// Recovery 是 HTTP 兜底中间件：defer logger.HandlePanic 捕获 handler panic 并记录，
// 避免单个请求的 panic 击穿整个服务进程（mlog 后端 debug 模式会重抛便于排查）。
func Recovery(logger mlog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer logger.HandlePanic("Api-Gateway-Recovery")
		h.ServeHTTP(w, r)
	})
}
