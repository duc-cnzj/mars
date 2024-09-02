package middlewares

import (
	"net/http"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/stretchr/testify/assert"
)

func TestRecovery(t *testing.T) {
	rw := &mockResponseWriter{h: map[string][]string{}}
	m := &mockHandler{
		fn: func(writer http.ResponseWriter, request *http.Request) {
			panic("err")
		},
	}
	Recovery(mlog.NewLogger(nil), m).ServeHTTP(rw, &http.Request{})
	assert.True(t, true)
}
