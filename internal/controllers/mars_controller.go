package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mars"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/response"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

type MarsController struct{}

func NewMarsController() *MarsController {
	return &MarsController{}
}

type ShowInput struct {
	// 必填
	ProjectId int `uri:"project_id"`

	// 可选，如果传分支，则显示对应分支的配置，不传则显示 master
	Branch string `form:"branch" json:"branch"`
}

// Show 获取项目配置
func (*MarsController) Show(ctx *gin.Context) {
	var input ShowInput
	if err := ctx.BindUri(&input); err != nil {
		response.Error(ctx, http.StatusBadRequest, err)

		return
	}
	if err := ctx.BindQuery(&input); err != nil {
		response.Error(ctx, http.StatusBadRequest, err)

		return
	}
	var branch string = input.Branch
	if branch == "" {
		branch, _ = getDefaultBranch(input.ProjectId)
	}
	config, err := GetProjectMarsConfig(input.ProjectId, branch)
	if err != nil {
		response.Success(ctx, http.StatusOK, gin.H{
			"branch": branch,
			"config": "",
		})
		return
	}
	bf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(bf)
	if err := encoder.Encode(config); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err)

		return
	}

	response.Success(ctx, http.StatusOK, gin.H{
		"branch": branch,
		"config": bf.String(),
	})
}

type GlobalConfigInput struct {
	// 必填
	ProjectId int `uri:"project_id"`
}

func (*MarsController) GlobalConfig(ctx *gin.Context) {
	var uri GlobalConfigInput
	if err := ctx.ShouldBindUri(&uri); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err)
		return
	}
	var project models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", uri.ProjectId).First(&project).Error != nil {
		response.Success(ctx, 200, gin.H{
			"enabled": false,
			"config":  project.GlobalConfigString(),
		})
		return
	}

	response.Success(ctx, 200, gin.H{
		"enabled": project.GlobalEnabled,
		"config":  project.GlobalConfigString(),
	})
}

type ToggleInput struct {
	// 必填
	ProjectId int `uri:"project_id"`

	Enabled bool `json:"enabled"`
}

func (*MarsController) ToggleEnabled(ctx *gin.Context) {
	var input ToggleInput
	if err := ctx.ShouldBindUri(&input); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err)
		return
	}
	if err := ctx.ShouldBind(&input); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err)
		return
	}

	var project models.GitlabProject
	if err := app.DB().Where("`gitlab_project_id` = ?", input.ProjectId).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			app.DB().Create(&models.GitlabProject{
				GitlabProjectId: input.ProjectId,
				Enabled:         false,
				GlobalEnabled:   input.Enabled,
			})
		}
		response.Success(ctx, 204, nil)
		return
	}

	app.DB().Model(&project).UpdateColumn("global_enabled", input.Enabled)
	response.Success(ctx, 204, nil)
}

type UpdateInput struct {
	// 必填
	ProjectId int `uri:"project_id"`

	Config string `json:"config"`
}

func (*MarsController) Update(ctx *gin.Context) {
	var input UpdateInput
	if err := ctx.ShouldBindUri(&input); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err)
		return
	}
	if err := ctx.ShouldBind(&input); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err)
		return
	}

	var project models.GitlabProject
	if err := app.DB().Where("`gitlab_project_id` = ?", input.ProjectId).First(&project).Error; err != nil {
		response.Success(ctx, 500, err)
		return
	}

	mc := mars.Config{}
	if err := yaml.Unmarshal([]byte(input.Config), &mc); err != nil {
		response.Error(ctx, 500, fmt.Errorf("配置不正确 %w", err))
		return
	}

	app.DB().Model(&project).UpdateColumn("global_config", input.Config)
	response.Success(ctx, 200, &project)
}
