package middlewares

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseMetrics(t *testing.T) {

	var tests = []struct {
		path string
		fn   func(writer http.ResponseWriter, request *http.Request)
	}{
		{
			path: "/api/xxxx?xxx=xxx",
			fn: func(writer http.ResponseWriter, request *http.Request) {
				_, ok := writer.(*CustomResponseWriter)
				assert.True(t, ok)
			},
		},
		{
			path: "/ws?xx=xx",
			fn: func(writer http.ResponseWriter, request *http.Request) {
				_, ok := writer.(*CustomResponseWriter)
				assert.False(t, ok)
			},
		},
		{
			path: "/api/metrics/namespace/xxx/stream?xxx=xxx",
			fn: func(writer http.ResponseWriter, request *http.Request) {
				_, ok := writer.(*CustomResponseWriter)
				assert.False(t, ok)
			},
		},
		{
			path: "/resources/app.js",
			fn: func(writer http.ResponseWriter, request *http.Request) {
				_, ok := writer.(*CustomResponseWriter)
				assert.False(t, ok)
			},
		},
		{
			path: "/resources/app.css",
			fn: func(writer http.ResponseWriter, request *http.Request) {
				_, ok := writer.(*CustomResponseWriter)
				assert.False(t, ok)
			},
		},
		{
			path: "/api/containers/namespaces/xxx/stream_logs?xxx=xxx",
			fn: func(writer http.ResponseWriter, request *http.Request) {
				_, ok := writer.(*CustomResponseWriter)
				assert.False(t, ok)
			},
		},
	}
	for _, test := range tests {
		rw := &mockResponseWriter{h: map[string][]string{}}
		m := &mockHandler{
			fn: func(writer http.ResponseWriter, request *http.Request) {
				test.fn(writer, request)
			},
		}
		parse, _ := url.Parse(test.path)
		req := &http.Request{URL: parse}
		ResponseMetrics(m).ServeHTTP(rw, req)
	}
}
