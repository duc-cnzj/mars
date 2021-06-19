package frontend

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed build/*
var staticFs embed.FS

var index []byte

func LoadFrontendRoutes(e *gin.Engine) {
	e.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/web")
	})
	sub, _ := fs.Sub(staticFs, "build")
	e.StaticFS("/resources/", http.FS(sub))
	index, _ = staticFs.ReadFile("build/index.html")
	e.Any("/web", serve)
	e.Any("/web/*action", serve)
}

func serve(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", index)
}
