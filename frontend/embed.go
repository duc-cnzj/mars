package frontend

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/duc-cnzj/mars/v4/internal/middlewares"
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
	subrouter.Handle("/{any}", toWebRoute())
}

func toWebRoute() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		writer.Write(index)
	})
}
