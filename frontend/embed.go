package frontend

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/gorilla/mux"
)

//go:embed build/*
var staticFs embed.FS

var index []byte

func LoadFrontendRoutes(mux *mux.Router) {
	subrouter := mux.PathPrefix("").Subrouter()
	subrouter.Use(middlewares.HttpCache)

	sub, _ := fs.Sub(staticFs, "build")
	subrouter.PathPrefix("/resources/").Handler(
		http.StripPrefix("/resources/",
			http.FileServer(http.FS(sub)),
		),
	)

	index, _ = staticFs.ReadFile("build/index.html")
	subrouter.Handle("/",
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			writer.WriteHeader(http.StatusOK)
			writer.Write(index)
		}),
	)
	subrouter.Handle("/auth/callback", toWebRoute())
	// SPA 兜底：/{any:.*} 匹配任意深度的前端路由（如 /admin/cluster），刷新/深链接时
	// 都回 index.html 交由前端 Router 接管。必须显式给 .* 通配——gorilla 的命名变量
	// 默认只匹配 [^/]+ 单段路径，多段路由会漏到 grpc-gateway 返回 404（历史 P0：
	// /admin/* 刷新即 {"code":5,"message":"Not Found"}）。/api /ws /docs 等后端路径
	// 已在 server 层先于本兜底注册（见 http.go initServer），不会被吞。
	subrouter.Handle("/{any:.*}", toWebRoute())
}

func toWebRoute() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		writer.Write(index)
	})
}
