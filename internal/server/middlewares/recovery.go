package middlewares

import (
	"net/http"

	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

func Recovery(logger mlog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer logger.HandlePanic("Api-Gateway-Recovery")
		h.ServeHTTP(w, r)
	})
}
