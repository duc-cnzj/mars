package middlewares

import (
	"fmt"
	"net/http"

	"github.com/duc-cnzj/mars/v6/internal/util/hasher"
	"github.com/duc-cnzj/mars/v6/internal/version"
)

var Etag string

func init() {
	setEtag(version.GetVersion())
}

// setEtag 根据构建信息生成静态资源 ETag：有构建信息（git commit + 构建时间）才设置，
// 保证版本化资源的缓存指纹与发布一致。
func setEtag(v version.Version) {
	if v.HasBuildInfo() {
		Etag = hasher.Hash(fmt.Sprintf("%s-%s", v.GitCommit, v.BuildDate))
	}
}

// HttpCache 是静态资源缓存中间件：请求带 If-None-Match 且命中 Etag 时返回 304，
// 否则回写 Etag 头后透传下游 handler（前端静态资源的浏览器缓存协调）。
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
