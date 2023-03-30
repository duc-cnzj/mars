package middlewares

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/utils"
	"github.com/duc-cnzj/mars/v4/version"
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

func Test_setEtag(t *testing.T) {
	defer func(t string) {
		Etag = t
	}(Etag)
	Etag = ""
	setEtag(version.GetVersion())
	assert.Empty(t, Etag)
	v := version.Version{
		GitCommit: "xxx",
		BuildDate: time.Now().Format("2006-01-02T15:04:05Z"),
	}
	setEtag(v)
	assert.Equal(t, utils.MD5(fmt.Sprintf("%s-%s", v.GitCommit, v.BuildDate)), Etag)
}
