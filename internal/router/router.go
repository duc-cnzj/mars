package router

import (
	"net/http"

	"github.com/duc-cnzj/mars/frontend"

	"github.com/duc-cnzj/mars/internal/controllers"
	t "github.com/duc-cnzj/mars/internal/translator"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	JSONContentType = "application/json"
)

func Init(e *gin.Engine) {
	var cd = cors.DefaultConfig()
	cd.AllowAllOrigins = true
	cd.AddAllowHeaders("X-Requested-With", "Authorization", "Accept-Language", "Access-Control-Allow-Credentials")
	e.Use(cors.New(cd))

	frontend.LoadFrontendRoutes(e)

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

		marsController := controllers.NewMarsController()
		{
			api.GET("/gitlab/projects/:project_id/mars_config", marsController.Show)
			api.GET("/gitlab/projects/:project_id/global_config", marsController.GlobalConfig)
			api.POST("/gitlab/projects/:project_id/toggle_enabled", marsController.ToggleEnabled)
			api.PUT("/gitlab/projects/:project_id/mars_config", marsController.Update)
		}

		wsC := controllers.NewWebsocketController()
		{
			e.GET("/ws", wsC.Ws)
			api.GET("/ws_info", wsC.Info)
		}

		cc := controllers.NewClusterController()
		{
			api.GET("/cluster_info", cc.Info)
		}
	}
}
