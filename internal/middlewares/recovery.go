package middlewares

import (
	"net/http"

	"github.com/duc-cnzj/mars/v4/internal/utils/recovery"
)

func Recovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recovery.HandlePanic("Api-Gateway-Recovery")
		h.ServeHTTP(w, r)
	})
}
