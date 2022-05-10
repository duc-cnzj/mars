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
}
