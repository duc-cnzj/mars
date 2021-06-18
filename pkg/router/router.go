package router

import (
	"net/http"

	"github.com/DuC-cnZj/mars/frontend"

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
	cd.AddAllowHeaders("X-Requested-With", "Authorization", "Accept-Language", "Access-Control-Allow-Credentials")
	e.Use(cors.New(cd))

	frontend.LoadFrontendRoutes(e)

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
		api.GET("/namespaces/:namespace_id/cpu_and_memory", nsC.CpuAndMemory)
		api.GET("/namespaces/:namespace_id/service_endpoints", nsC.ServiceEndpoints)
		api.DELETE("/namespaces/:namespace_id", nsC.Destroy)

		proC := controllers.NewProjectController()
		{
			api.DELETE("/namespaces/:namespace_id/projects/:project_id", proC.Destroy)
			api.GET("/namespaces/:namespace_id/projects/:project_id", proC.Show)

			api.GET("/namespaces/:namespace_id/projects/:project_id/containers", proC.AllPodContainers)
			api.GET("/namespaces/:namespace_id/projects/:project_id/pods/:pod/containers/:container/logs", proC.PodContainerLog)
		}

		gitlabController := controllers.NewGitlabController()
		{
			api.POST("/gitlab/projects/enable", gitlabController.EnableProject)
			api.POST("/gitlab/projects/disable", gitlabController.DisableProject)

			// 这个接口返回更加详细的项目详细
			api.GET("/gitlab/project_list", gitlabController.ProjectList)

			// 下面三个只返回级联所需的信息
			api.GET("/gitlab/projects", gitlabController.Projects)
			api.GET("/gitlab/projects/:project_id/branches", gitlabController.Branches)
			api.GET("/gitlab/projects/:project_id/branches/:branch/commits", gitlabController.Commits)
			api.GET("/gitlab/projects/:project_id/branches/:branch/commits/:commit", gitlabController.Commit)
			api.GET("/gitlab/projects/:project_id/branches/:branch/commits/:commit/pipeline_info", gitlabController.PipelineInfo)

			api.GET("/gitlab/projects/:project_id/branches/:branch/config_file", gitlabController.ConfigFile)
		}

		tc := controllers.NewTerminalController()
		{
			api.GET("/pod/:namespace/:pod/shell", tc.HandleExecShell)
			e.Any("/api/sockjs/*action", func(ctx *gin.Context) {
				tc.HandleSocket("/api/sockjs").ServeHTTP(ctx.Writer, ctx.Request)
			})
		}
	}
}
