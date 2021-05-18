package router

import (
	"encoding/base64"
	"net/http"

	"github.com/DuC-cnZj/mars/pkg/controllers"
	"github.com/DuC-cnZj/mars/pkg/mlog"
	t "github.com/DuC-cnZj/mars/pkg/translator"
	"github.com/DuC-cnZj/mars/pkg/utils"
	"github.com/xanzy/go-gitlab"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

const (
	JSONContentType = "application/json"
)

func Init(e *gin.Engine) {
	var cd = cors.DefaultConfig()
	cd.AllowAllOrigins = true
	cd.AddAllowHeaders("X-Requested-With", "Authorization", "Accept-Language")
	e.Use(cors.New(cd))

	e.GET("/", func(ctx *gin.Context) {
		file, r, err := utils.GitlabClient().RepositoryFiles.GetFile(21409590, ".drone.yml", &gitlab.GetFileOptions{Ref: gitlab.String("master")})
		if err != nil {
			if r != nil && r.Status == "404" {
				ctx.JSON(404, gin.H{"ok": err.Error()})
				return
			}
			ctx.JSON(500, gin.H{"ok": err.Error()})
			return
		}
		decodeString, _ := base64.StdEncoding.DecodeString(file.Content)
		mlog.Info(string(decodeString))
		ctx.JSON(200, gin.H{"ok": true})
	})
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
		api.POST("/namespaces/:namespace_id/projects", proC.Store)
		api.DELETE("/namespaces/:namespace_id/projects/:project_id", proC.Destroy)
		api.GET("/namespaces/:namespace_id/projects/:project_id", proC.Show)

		gitlabController := controllers.NewGitlabController()
		api.GET("/gitlab/projects", gitlabController.Projects)
		api.GET("/gitlab/projects/:project_id/branches", gitlabController.Branches)
		api.GET("/gitlab/projects/:project_id/branches/:branch/commits", gitlabController.Commits)

		api.GET("/gitlab/projects/:project_id/branches/:branch/config_file", gitlabController.ConfigFile)
	}
}
