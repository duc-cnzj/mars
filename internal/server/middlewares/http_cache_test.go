package middlewares

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/util/hasher"
	"github.com/duc-cnzj/mars/v6/internal/version"
	"github.com/stretchr/testify/assert"
)

func TestHttpCache(t *testing.T) {
	m := &mockHandler{}
	rw := &mockResponseWriter{h: map[string][]string{}}
	etag = ""
	HttpCache(m).ServeHTTP(rw, &http.Request{})
	assert.Len(t, rw.h, 0)
	etag = "xxx"
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

func Test_setEtag(t *testing.T) {
	defer func(t string) {
		etag = t
	}(etag)
	etag = ""
	setEtag(version.GetVersion())
	assert.Empty(t, etag)
	v := version.Version{
		GitCommit: "xxx",
		BuildDate: time.Now().Format("2006-01-02T15:04:05Z"),
	}
	setEtag(v)
	assert.Equal(t, hasher.Hash(fmt.Sprintf("%s-%s", v.GitCommit, v.BuildDate)), etag)
}
