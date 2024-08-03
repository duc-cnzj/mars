package middlewares

import (
	"net/http"
	"sync/atomic"

	"github.com/duc-cnzj/mars/v4/internal/mlog"

	"github.com/duc-cnzj/mars/v4/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var metricsIgnoreFn = tracingIgnoreFn

type CustomResponseWriter struct {
	http.ResponseWriter
	bytes atomic.Value
}

func (c *CustomResponseWriter) Write(bytes []byte) (int, error) {
	c.bytes.Store(len(bytes))
	return c.ResponseWriter.Write(bytes)
}

func ResponseMetrics(logger mlog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if metricsIgnoreFn(path) {
			h.ServeHTTP(w, r)
			return
		}
		rw := &CustomResponseWriter{ResponseWriter: w}
		defer func() {
			pattern := GetPatternHeader(rw)
			bytes := rw.bytes.Load()
			if pattern != "" && bytes != nil {
				metrics.HttpResponseSize.With(prometheus.Labels{"path": pattern}).Observe(float64(bytes.(int)))
			}
		}()
		h.ServeHTTP(rw, r)
	})
}
