package controllers

import (
	"bytes"
	"net/http"

	"github.com/duc-cnzj/mars/internal/response"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
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
		response.Error(ctx, http.StatusInternalServerError, err)
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
