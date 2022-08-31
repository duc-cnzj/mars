package middlewares

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestDeletePatternHeader(t *testing.T) {
	rw := &mockResponseWriter{h: map[string][]string{}}
	m := &mockHandler{}
	rw.Header().Set("pattern", "/api/xxx")
	assert.Equal(t, "/api/xxx", rw.Header().Get("pattern"))
	DeletePatternHeader(m).ServeHTTP(rw, &http.Request{})
	assert.Equal(t, "", rw.Header().Get("pattern"))
	assert.Equal(t, 1, m.serverCalled)
}
