package middlewares

import (
	"net/http"
)

const PatternHeader = "pattern"

func GetPatternHeader(w http.ResponseWriter) string {
	return w.Header().Get(PatternHeader)
}

func SetPatternHeader(w http.ResponseWriter, val string) {
	w.Header().Set(PatternHeader, val)
}

func DeletePatternHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		w.Header().Del(PatternHeader)
	})
}
