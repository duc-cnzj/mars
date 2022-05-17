package middlewares

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpCache(t *testing.T) {
	m := &mockHandler{}
	rw := &mockResponseWriter{h: map[string][]string{}}
	Etag = ""
	HttpCache(m).ServeHTTP(rw, &http.Request{})
	assert.Len(t, rw.h, 0)
	Etag = "xxx"
	HttpCache(m).ServeHTTP(rw, &http.Request{})
	assert.Equal(t, "xxx", rw.h["Etag"][0])
	rw = &mockResponseWriter{h: map[string][]string{}}
	HttpCache(m).ServeHTTP(rw, &http.Request{
		Header: map[string][]string{
			"If-None-Match": {"xxx"},
		},
	})
	assert.Equal(t, 304, rw.code)
}
