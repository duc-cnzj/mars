package router

import (
	"net/http"

	"github.com/DuC-cnZj/mars/pkg/controllers"
	t "github.com/DuC-cnZj/mars/pkg/translator"
	"github.com/gin-contrib/cors"
	"github.com/gorilla/websocket"

	"github.com/gin-gonic/gin"
)

const (
	JSONContentType = "application/json"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Init(e *gin.Engine) {
	var cd = cors.DefaultConfig()
	cd.AllowAllOrigins = true
	cd.AddAllowHeaders("X-Requested-With", "Authorization", "Accept-Language")
	e.Use(cors.New(cd))

	wsC := controllers.NewWebsocketController()
	e.GET("/ws", wsC.Ws)

	e.NoRoute(func(ctx *gin.Context) {
		ctx.Data(http.StatusNotFound, JSONContentType, []byte(`{"code": 404, "message": "404 not found"}`))
	})

	e.GET("/ping", func(ctx *gin.Context) {
		ctx.Data(200, JSONContentType, []byte(`{"success": "true"}`))
	})

	api := e.Group("/api", t.I18nMiddleware())
	{
		nsC := controllers.NewNamespaceController()
		api.GET("/namespaces", nsC.Index)
		api.POST("/namespaces", nsC.Store)
		api.DELETE("/namespaces/:namespace_id", nsC.Destroy)

		proC := controllers.NewProjectController()
		api.DELETE("/namespaces/:namespace_id/projects/:project_id", proC.Destroy)
		api.GET("/namespaces/:namespace_id/projects/:project_id", proC.Show)

		gitlabController := controllers.NewGitlabController()
		api.GET("/gitlab/projects", gitlabController.Projects)
		api.GET("/gitlab/projects/:project_id/branches", gitlabController.Branches)
		api.GET("/gitlab/projects/:project_id/branches/:branch/commits", gitlabController.Commits)

		api.GET("/gitlab/projects/:project_id/branches/:branch/config_file", gitlabController.ConfigFile)
	}
}
