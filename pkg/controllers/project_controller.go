package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"

	"github.com/duc-cnzj/mars/pkg/mars"
	"github.com/duc-cnzj/mars/pkg/mlog"
	"github.com/duc-cnzj/mars/pkg/models"
	"github.com/duc-cnzj/mars/pkg/response"
	"github.com/duc-cnzj/mars/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
	v12 "k8s.io/api/core/v1"
)

type ProjectController struct{}

func NewProjectController() *ProjectController {
	return &ProjectController{}
}

type ProjectDetailItem struct {
	ID int `json:"id"`

	Name            string `json:"name"`
	GitlabProjectId int    `json:"gitlab_project_id"`
	GitlabBranch    string `json:"gitlab_branch"`
	GitlabCommit    string `json:"gitlab_commit"`
	Config          string `json:"config"`
	DockerImage     string `json:"docker_image"`

	GitlabCommitWebURL string `json:"gitlab_commit_web_url"`
	GitlabCommitTitle  string `json:"gitlab_commit_title"`
	GitlabCommitAuthor string `json:"gitlab_commit_author"`

	Namespace struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"namespace"`

	Cpu            string `json:"cpu"`
	Memory         string `json:"memory"`
	OverrideValues string `json:"override_values"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
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
	cpu, memory := utils.GetCpuAndMemory(project.GetAllPodMetrics())
	commit, _, err := utils.GitlabClient().Commits.GetCommit(project.GitlabProjectId, project.GitlabCommit)
	if err != nil {
		mlog.Error(err)
		response.Error(ctx, 500, err)
		return
	}
	mlog.Debug("commit.LastPipeline", commit.LastPipeline)
	response.Success(ctx, 200, ProjectDetailItem{
		ID:                 project.ID,
		Name:               project.Name,
		GitlabProjectId:    project.GitlabProjectId,
		GitlabBranch:       project.GitlabBranch,
		GitlabCommit:       project.GitlabCommit,
		DockerImage:        project.DockerImage,
		Config:             project.Config,
		GitlabCommitWebURL: commit.WebURL,
		GitlabCommitTitle:  commit.Title,
		GitlabCommitAuthor: commit.AuthorName,
		Namespace: struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}{project.NamespaceId, project.Namespace.Name},
		Cpu:            cpu,
		Memory:         memory,
		OverrideValues: project.OverrideValues,
		CreatedAt:      utils.ToHumanizeDatetimeString(&project.CreatedAt),
		UpdatedAt:      utils.ToHumanizeDatetimeString(&project.UpdatedAt),
	})
}

type ContainerLogsResponse struct {
	List []struct {
		PodName       string `json:"pod_name"`
		ContainerName string `json:"container_name"`
	} `json:"list"`

	Log struct {
		PodName       string `json:"pod_name"`
		ContainerName string `json:"container_name"`
		Log           string `json:"log"`
	} `json:"logs"`
}

type ContainerLogsQuery struct {
	Pod       string `form:"pod"`
	Container string `form:"container"`
}

type PodContainerResponse struct {
	PodName       string `json:"pod_name"`
	ContainerName string `json:"container_name"`

	Log string `json:"log,omitempty"`
}

func (p *ProjectController) AllPodContainers(ctx *gin.Context) {
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

	var list = project.GetAllPods()

	var containerList []PodContainerResponse
	for _, item := range list {
		for _, c := range item.Spec.Containers {
			containerList = append(containerList, PodContainerResponse{
				PodName:       item.Name,
				ContainerName: c.Name,
			})
		}
	}

	response.Success(ctx, 200, containerList)
}

type PodContainerLogUri struct {
	NamespaceId int `uri:"namespace_id"`
	ProjectId   int `uri:"project_id"`

	Pod       string `uri:"pod"`
	Container string `uri:"container"`
}

func (p *ProjectController) PodContainerLog(ctx *gin.Context) {
	var uri PodContainerLogUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		response.Error(ctx, 422, err)
		return
	}

	var project models.Project
	if err := utils.DB().Preload("Namespace").Where("`id` = ?", uri.ProjectId).First(&project).Error; err != nil {
		response.Error(ctx, 500, err)
		return
	}

	var limit int64 = 10000
	logs := utils.K8sClientSet().CoreV1().Pods(project.Namespace.Name).GetLogs(uri.Pod, &v12.PodLogOptions{
		Container: uri.Container,
		TailLines: &limit,
	})
	var raw = []byte("未找到日志")
	do := logs.Do(context.Background())
	raw, err := do.Raw()
	if err == nil {
		split := strings.Split(string(raw), "\n")
		var reverseLog []string
		for i := len(split) - 1; i > 0; i-- {
			reverseLog = append(reverseLog, split[i])
		}

		raw = bytes.Trim([]byte(strings.Join(reverseLog, "\n")), "\n")
	}

	response.Success(ctx, 200, PodContainerResponse{
		PodName:       uri.Pod,
		ContainerName: uri.Container,
		Log:           string(raw),
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
	if err := utils.UninstallRelease(project.Name, project.Namespace.Name); err != nil {
		mlog.Error(err)
	}
	utils.DB().Delete(&project)
	response.Success(ctx, 204, "")
}

func GetProjectMarsConfig(projectId int, branch string) (*mars.Config, error) {
	var marsC mars.Config

	// 获取 .mars.yaml
	opt := &gitlab.GetFileOptions{}
	if branch != "" {
		opt.Ref = gitlab.String(branch)
	}
	file, _, err := utils.GitlabClient().RepositoryFiles.GetFile(projectId, ".mars.yaml", opt)
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
