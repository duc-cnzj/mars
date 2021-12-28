package middlewares

import (
	"fmt"
	"net/http"

	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/version"
)

var Etag string

func init() {
	v := version.GetVersion()
	if v.HasBuildInfo() {
		Etag = utils.Md5(fmt.Sprintf("%s-%s", v.GitCommit, v.BuildDate))
	}
}

type HttpCacheBody struct {
	Etag string
}

func HttpCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Etag != "" {
			if r.Header.Get("If-None-Match") == Etag {
				w.WriteHeader(304)
				return
			}
			w.Header().Set("Etag", Etag)
		}

		h.ServeHTTP(w, r)
	})
}
