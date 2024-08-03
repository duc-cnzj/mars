package middlewares

import (
	"net/http"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

func RouteLogger(logger mlog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(t time.Time) {
			logger.Debugf("[Http]: method: %v, url: %v, use %v", r.Method, r.URL, time.Since(t))
		}(time.Now())
		h.ServeHTTP(w, r)
	})
}
