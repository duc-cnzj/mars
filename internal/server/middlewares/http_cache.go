package middlewares

import (
	"fmt"
	"net/http"

	"github.com/duc-cnzj/mars/v4/internal/utils/hash"

	"github.com/duc-cnzj/mars/v4/version"
)

var Etag string

func init() {
	setEtag(version.GetVersion())
}

func setEtag(v version.Version) {
	if v.HasBuildInfo() {
		Etag = hash.Hash(fmt.Sprintf("%s-%s", v.GitCommit, v.BuildDate))
	}
}

func HttpCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Etag != "" {
			if r.Header.Get("If-None-Match") == Etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
			w.Header().Set("Etag", Etag)
		}

		h.ServeHTTP(w, r)
	})
}
