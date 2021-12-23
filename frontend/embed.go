package frontend

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/duc-cnzj/mars/internal/middlewares"
	"github.com/gorilla/mux"
)

//go:embed build/*
var staticFs embed.FS

var index []byte

func LoadFrontendRoutes(mux *mux.Router) {
	sub, _ := fs.Sub(staticFs, "build")
	mux.PathPrefix("/resources/").Handler(
		middlewares.HttpCache(
			http.StripPrefix("/resources/",
				http.FileServer(http.FS(sub)),
			),
		),
	)

	index, _ = staticFs.ReadFile("build/index.html")
	mux.Handle("/",
		middlewares.HttpCache(
			http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "text/html; charset=utf-8")
				writer.WriteHeader(http.StatusOK)
				writer.Write(index)
			}),
		),
	)
	mux.Handle("/auth/callback", middlewares.HttpCache(toWebRoute()))
	mux.Handle("/{any}", middlewares.HttpCache(toWebRoute()))
}

func toWebRoute() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		writer.Write(index)
	})
}
