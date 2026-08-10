package middlewares

import (
	"net/http"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// preflightHandler 处理 CORS 预检（OPTIONS）请求：声明允许的请求头与方法。
func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept", "X-Requested-With", "Authorization", "Accept-Language"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
}

// AllowCORS 是跨域中间件：带 Origin 的请求回写 Allow-Origin 与预检头，预检直接短路，
// 其余透传给下游 handler。logger 参数与其它中间件签名对齐（统一注入日志，但本中间件不落日志）。
func AllowCORS(logger mlog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
