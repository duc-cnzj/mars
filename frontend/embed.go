package frontend

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"
)

//go:embed build/*
var staticFs embed.FS

var index []byte

func LoadFrontendRoutes(mux *mux.Router) {
	sub, _ := fs.Sub(staticFs, "build")
	mux.PathPrefix("/resources/").Handler(http.StripPrefix("/resources/", http.FileServer(http.FS(sub))))

	index, _ = staticFs.ReadFile("build/index.html")
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		writer.Write(index)
	})
	mux.HandleFunc("/auth/callback", toWebRoute())
	mux.HandleFunc("/{any}", toWebRoute())
}

func toWebRoute() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		writer.Write(index)
	}
}
