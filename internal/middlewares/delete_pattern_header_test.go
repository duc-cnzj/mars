package middlewares

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestGetPattern(t *testing.T) {
	rw := &mockResponseWriter{h: map[string][]string{}}
	assert.Equal(t, "", GetPattern(rw))
}

func TestSetPattern(t *testing.T) {
	rw := &mockResponseWriter{h: map[string][]string{}}
	SetPattern(rw, "aaa")
	assert.Equal(t, "aaa", GetPattern(rw))
}
