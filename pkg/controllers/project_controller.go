package controllers

import (
	"bytes"
	"encoding/base64"
	"net/http"

	"github.com/DuC-cnZj/mars/pkg/mars"
	"github.com/DuC-cnZj/mars/pkg/mlog"
	"github.com/DuC-cnZj/mars/pkg/models"
	"github.com/DuC-cnZj/mars/pkg/response"
	"github.com/DuC-cnZj/mars/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli/values"
)

type ProjectController struct{}

func NewProjectController() *ProjectController {
	return &ProjectController{}
}

type ProjectStoreInput struct {
	NamespaceId int `uri:"namespace_id"`

	Name            string `json:"name"`
	GitlabProjectId int    `json:"gitlab_project_id"`
	GitlabBranch    string `json:"gitlab_branch"`
	GitlabCommit    string `json:"gitlab_commit"`
	Config          string `json:"config"`
}

func (p *ProjectController) Store(ctx *gin.Context) {
	var input ProjectStoreInput
	if err := ctx.ShouldBindUri(&input); err != nil {
		response.Error(ctx, 422, err)
		return
	}
	if err := ctx.ShouldBind(&input); err != nil {
		response.Error(ctx, 422, err)
		return
	}

	var ns models.Namespace
	if err := utils.DB().Where("`id` = ?", input.NamespaceId).First(&ns).Error; err != nil {
		response.Error(ctx, 500, err)

		return
	}

	input.Name = slug.Make(input.Name)

	project := models.Project{
		Name:            input.Name,
		GitlabProjectId: input.GitlabProjectId,
		GitlabBranch:    input.GitlabBranch,
		GitlabCommit:    input.GitlabCommit,
		Config:          input.Config,
		NamespaceId:     ns.ID,
	}
	utils.DB().Create(&project)

	marsC, err := GetProjectMarsConfig(input.GitlabProjectId, input.GitlabBranch)
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}
	file, _, err := utils.GitlabClient().RepositoryFiles.GetFile(input.GitlabProjectId, marsC.LocalChartPath, &gitlab.GetFileOptions{Ref: gitlab.String(input.GitlabBranch)})
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}
	archive, _ := base64.StdEncoding.DecodeString(file.Content)

	loadArchive, err := loader.LoadArchive(bytes.NewReader(archive))
	if err != nil {
		response.Error(ctx, 500, err)

		return
	}

	filePath, deleteFn, err := marsC.GenerateConfigYamlFileByInput(input.Config)
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}
	defer deleteFn()

	var valueOpts = &values.Options{
		ValueFiles: []string{filePath},
	}
	//input.Config
	if _, err := utils.UpgradeOrInstall(input.Name, ns.Name, loadArchive, valueOpts); err != nil {
		mlog.Error(err)
	}

	response.Success(ctx, http.StatusCreated, project)
}

type ProjectDetailItem struct {
	ID int `json:"id"`

	Name            string `json:"name"`
	GitlabProjectId int    `json:"gitlab_project_id"`
	GitlabBranch    string `json:"gitlab_branch"`
	GitlabCommit    string `json:"gitlab_commit"`
	Config          string `json:"config"`

	GitlabCommitWebURL string `json:"gitlab_commit_web_url"`
	GitlabCommitTitle  string `json:"gitlab_commit_title"`
	GitlabCommitAuthor string `json:"gitlab_commit_author"`

	Namespace struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"namespace"`

	Cpu    string `json:"cpu"`
	Memory string `json:"memory"`

	CreatedAt string `json:"created_at"`
}

func (p *ProjectController) Show(ctx *gin.Context) {
	var uri ProjectUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		response.Error(ctx, 422, err)
		return
	}

	var project models.Project
	if err := utils.DB().Preload("Namespace").Where("`id` = ?", uri.ProjectId).First(&project).Error; err != nil {
		response.Error(ctx, 500, err)
		return
	}
	cpu, memory := utils.GetCpuAndMemoryInNamespaceByRelease(project.Namespace.Name, project.Name)
	commit, _, err := utils.GitlabClient().Commits.GetCommit(project.GitlabProjectId, project.GitlabCommit)
	if err != nil {
		mlog.Error(err)
		response.Error(ctx, 500, err)
		return
	}
	// TODO add pipeline infp
	mlog.Warning("commit.LastPipeline", commit.LastPipeline)
	response.Success(ctx, 200, ProjectDetailItem{
		ID:                 project.ID,
		Name:               project.Name,
		GitlabProjectId:    project.GitlabProjectId,
		GitlabBranch:       project.GitlabBranch,
		GitlabCommit:       project.GitlabCommit,
		Config:             project.Config,
		GitlabCommitWebURL: commit.WebURL,
		GitlabCommitTitle:  commit.Title,
		GitlabCommitAuthor: commit.AuthorName,
		Namespace: struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{project.NamespaceId, project.Namespace.Name},
		Cpu:       cpu,
		Memory:    memory,
		CreatedAt: utils.ToHumanizeDatetimeString(&project.CreatedAt),
	})
}

type ProjectUri struct {
	NamespaceId int `uri:"namespace_id"`
	ProjectId   int `uri:"project_id"`
}

func (p *ProjectController) Destroy(ctx *gin.Context) {
	var uri ProjectUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		response.Error(ctx, 422, err)
		return
	}

	var project models.Project
	if err := utils.DB().Preload("Namespace").Where("`id` = ?", uri.ProjectId).First(&project).Error; err != nil {
		response.Error(ctx, 500, err)
		return
	}
	settings := utils.GetSettings(project.Namespace.Name)
	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), "", mlog.Debugf); err != nil {
		response.Error(ctx, 500, err)
		return
	}
	uninstall := action.NewUninstall(actionConfig)
	if _, err := uninstall.Run(project.Name); err != nil {
		mlog.Error(err)
	}
	utils.DB().Delete(&project)
	response.Success(ctx, 204, "")
}

func GetProjectMarsConfig(projectId int, branch string) (*mars.Config, error) {
	var marsC mars.Config

	// 获取 mars.yaml
	file, _, err := utils.GitlabClient().RepositoryFiles.GetFile(projectId, ".mars.yaml", &gitlab.GetFileOptions{Ref: gitlab.String(branch)})
	if err != nil {
		return nil, err
	}
	data, _ := base64.StdEncoding.DecodeString(file.Content)
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&marsC); err != nil {
		return nil, err
	}

	return &marsC, nil
}
