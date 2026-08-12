package middlewares

import (
	"net/http"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestRecovery(t *testing.T) {
	// handler panic：Recovery 兜底捕获（HandlePanic），不 panic 击穿，handler 已被调用。
	rw := &mockResponseWriter{h: map[string][]string{}}
	m := &mockHandler{
		fn: func(writer http.ResponseWriter, request *http.Request) {
			panic("err")
		},
	}
	assert.NotPanics(t, func() {
		Recovery(mlog.NewForConfig(nil), m).ServeHTTP(rw, &http.Request{})
	})
	assert.Equal(t, 1, m.serverCalled, "handler 应被调用一次")

	// handler 无 panic：正常透传。
	m2 := &mockHandler{}
	rw2 := &mockResponseWriter{h: map[string][]string{}}
	Recovery(mlog.NewForConfig(nil), m2).ServeHTTP(rw2, &http.Request{})
	assert.Equal(t, 1, m2.serverCalled)
}
