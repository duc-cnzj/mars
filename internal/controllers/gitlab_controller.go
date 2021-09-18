package controllers

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"

	"github.com/duc-cnzj/mars/internal/mlog"

	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/response"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
)

type GitlabController struct{}

func NewGitlabController() *GitlabController {
	return &GitlabController{}
}

const (
	OptionTypeProject string = "project"
	OptionTypeBranch  string = "branch"
	OptionTypeCommit  string = "commit"
)

type Options struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Type  string `json:"type"`
	// isLeaf 兼容 antd
	IsLeaf bool `json:"isLeaf"`

	ProjectId int    `json:"projectId,omitempty"`
	Branch    string `json:"branch,omitempty"`
}

func (*GitlabController) Projects(ctx *gin.Context) {
	var (
		enabledProjects []models.GitlabProject
		ch              = make(chan Options)
		wg              = sync.WaitGroup{}
	)

	app.DB().Where("`enabled` = ?", true).Find(&enabledProjects)
	wg.Add(len(enabledProjects))
	for _, project := range enabledProjects {
		go func(project models.GitlabProject) {
			defer wg.Done()
			if !project.GlobalEnabled {
				if _, err := GetProjectMarsConfig(project.GitlabProjectId, project.DefaultBranch); err != nil {
					mlog.Debug(err)
					return
				}
			}
			ch <- Options{
				Value:     fmt.Sprintf("%d", project.GitlabProjectId),
				Label:     project.Name,
				IsLeaf:    false,
				Type:      OptionTypeProject,
				ProjectId: project.GitlabProjectId,
			}
		}(project)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	res := make([]Options, 0)

	for options := range ch {
		res = append(res, options)
	}

	response.Success(ctx, 200, res)
}

type BranchUri struct {
	ProjectId int `uri:"project_id"`
}

func (*GitlabController) Branches(ctx *gin.Context) {
	var uri BranchUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		response.Error(ctx, 400, err)
		return
	}

	branches, err := utils.GetAllBranches(uri.ProjectId)
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}

	var defaultBranch string
	for _, branch := range branches {
		if branch.Default {
			defaultBranch = branch.Name
		}
	}

	config, err := GetProjectMarsConfig(uri.ProjectId, defaultBranch)
	if err != nil {
		response.Success(ctx, 200, make([]Options, 0))
		return
	}

	res := make([]Options, 0, len(branches))
	for _, branch := range branches {
		if config.BranchPass(branch.Name) {
			res = append(res, Options{
				Value:     branch.Name,
				Label:     branch.Name,
				IsLeaf:    false,
				Type:      OptionTypeBranch,
				Branch:    branch.Name,
				ProjectId: uri.ProjectId,
			})
		}
	}

	response.Success(ctx, 200, res)
}

func getDefaultBranch(projectId int) (string, error) {
	p, _, err := app.GitlabClient().Projects.GetProject(projectId, &gitlab.GetProjectOptions{})
	if err != nil {
		mlog.Error(err)
		return "", err
	}
	return p.DefaultBranch, nil
}

type CommitUri struct {
	ProjectId int    `uri:"project_id"`
	Branch    string `uri:"branch"`
}

func (*GitlabController) Commits(ctx *gin.Context) {
	var uri CommitUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		response.Error(ctx, 400, err)
		return
	}

	commits, _, err := app.GitlabClient().Commits.ListCommits(uri.ProjectId, &gitlab.ListCommitsOptions{RefName: gitlab.String(uri.Branch), ListOptions: gitlab.ListOptions{PerPage: 100}})
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}

	res := make([]Options, 0, len(commits))
	for _, commit := range commits {
		res = append(res, Options{
			Value:     commit.ID,
			IsLeaf:    true,
			Label:     fmt.Sprintf("[%s]: %s", utils.ToHumanizeDatetimeString(commit.CommittedDate), commit.Title),
			Type:      OptionTypeCommit,
			ProjectId: uri.ProjectId,
			Branch:    uri.Branch,
		})
	}

	response.Success(ctx, 200, res)
}

func (*GitlabController) ConfigFile(ctx *gin.Context) {
	var uri CommitUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		response.Error(ctx, 400, err)
		return
	}

	marsC, err := GetProjectMarsConfig(uri.ProjectId, uri.Branch)
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}
	// 先拿 ConfigFile 如果没有，则拿 ConfigFileValues
	configFile := marsC.ConfigFile
	if configFile == "" {
		ct := marsC.ConfigFileType
		if marsC.ConfigFileType == "" {
			ct = "yaml"
		}
		response.Success(ctx, 200, gin.H{
			"data": marsC.ConfigFileValues,
			"type": ct,
		})
		return
	}
	// 如果有 ConfigFile，则获取内容，如果没有内容，则使用 ConfigFileValues

	var (
		pid      string
		branch   string
		filename string
	)

	if marsC.IsRemoteConfigFile() {
		split := strings.Split(configFile, "|")
		pid = split[0]
		branch = split[1]
		filename = split[2]
	} else {
		pid = fmt.Sprintf("%d", uri.ProjectId)
		branch = uri.Branch
		filename = configFile
	}

	f, _, err := app.GitlabClient().RepositoryFiles.GetFile(pid, filename, &gitlab.GetFileOptions{Ref: gitlab.String(branch)})
	if err != nil {
		mlog.Debug(err)
		response.Success(ctx, 200, gin.H{
			"data": marsC.ConfigFileValues,
			"type": marsC.ConfigFileType,
		})
		return
	}
	fdata, _ := base64.StdEncoding.DecodeString(f.Content)
	response.Success(ctx, 200, gin.H{
		"data": string(fdata),
		"type": marsC.ConfigFileType,
	})
}

type ShowCommitUri struct {
	ProjectId int    `uri:"project_id"`
	Branch    string `uri:"branch"`
	Commit    string `uri:"commit"`
}

func (*GitlabController) Commit(ctx *gin.Context) {
	var uri ShowCommitUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		response.Error(ctx, 400, err)
		return
	}

	commit, _, err := app.GitlabClient().Commits.GetCommit(uri.ProjectId, uri.Commit)
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}

	response.Success(ctx, 200, Options{
		Value:     commit.ID,
		IsLeaf:    true,
		Label:     fmt.Sprintf("[%s]: %s", utils.ToHumanizeDatetimeString(commit.CommittedDate), commit.Title),
		Type:      OptionTypeCommit,
		ProjectId: uri.ProjectId,
		Branch:    uri.Branch,
	})
}

type PipelineInfo struct {
	Status string `json:"status"`
	WebUrl string `json:"web_url"`
}

func (*GitlabController) PipelineInfo(ctx *gin.Context) {
	var uri ShowCommitUri
	if err := ctx.ShouldBindUri(&uri); err != nil {
		response.Error(ctx, 400, err)
		return
	}

	commit, _, err := app.GitlabClient().Commits.GetCommit(uri.ProjectId, uri.Commit)
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}
	if commit.LastPipeline == nil {
		response.Error(ctx, 404, "")
		return
	}

	response.Success(ctx, 200, PipelineInfo{
		Status: commit.LastPipeline.Status,
		WebUrl: commit.LastPipeline.WebURL,
	})
}

type ProjectInfo struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	WebURL        string `json:"web_url"`
	AvatarURL     string `json:"avatar_url"`
	Description   string `json:"description"`
	Enabled       bool   `json:"enabled"`
	GlobalEnabled bool   `json:"global_enabled"`
}

func (*GitlabController) ProjectList(ctx *gin.Context) {
	projects, _, err := app.GitlabClient().Projects.ListProjects(&gitlab.ListProjectsOptions{
		MinAccessLevel: gitlab.AccessLevel(gitlab.DeveloperPermissions),
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}

	var gps []models.GitlabProject
	app.DB().Find(&gps)

	var m = map[int]models.GitlabProject{}
	for _, gp := range gps {
		m[gp.GitlabProjectId] = gp
	}

	var infos = make([]ProjectInfo, 0)

	for _, project := range projects {
		var enabled, GlobalEnabled bool
		if gitlabProject, ok := m[project.ID]; ok {
			enabled = gitlabProject.Enabled
			GlobalEnabled = gitlabProject.GlobalEnabled
		}
		infos = append(infos, ProjectInfo{
			ID:            project.ID,
			Name:          project.Name,
			Path:          project.Path,
			WebURL:        project.WebURL,
			AvatarURL:     project.AvatarURL,
			Description:   project.Description,
			Enabled:       enabled,
			GlobalEnabled: GlobalEnabled,
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].ID > infos[j].ID
	})

	response.Success(ctx, 200, infos)
}

type EnableDisableInput struct {
	GitlabProjectID int `json:"gitlab_project_id" binding:"required"`
}

func (*GitlabController) EnableProject(ctx *gin.Context) {
	var input EnableDisableInput
	if err := ctx.ShouldBind(&input); err != nil {
		response.Error(ctx, 400, err)
		return
	}

	project, _, _ := app.GitlabClient().Projects.GetProject(input.GitlabProjectID, &gitlab.GetProjectOptions{})

	var gp models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", input.GitlabProjectID).First(&gp).Error == nil {
		app.DB().Model(&gp).Updates(map[string]interface{}{
			"enabled":        true,
			"default_branch": project.DefaultBranch,
			"name":           project.Name,
		})
	} else {
		app.DB().Create(&models.GitlabProject{
			DefaultBranch:   project.DefaultBranch,
			Name:            project.Name,
			GitlabProjectId: input.GitlabProjectID,
			Enabled:         true,
		})
	}

	response.Success(ctx, 200, nil)
}

func (*GitlabController) DisableProject(ctx *gin.Context) {
	var input EnableDisableInput
	if err := ctx.ShouldBind(&input); err != nil {
		response.Error(ctx, 400, err)
		return
	}

	project, _, _ := app.GitlabClient().Projects.GetProject(input.GitlabProjectID, &gitlab.GetProjectOptions{})
	var gp models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", input.GitlabProjectID).First(&gp).Error == nil {
		app.DB().Model(&gp).Updates(map[string]interface{}{
			"enabled":        false,
			"default_branch": project.DefaultBranch,
			"name":           project.Name,
		})
	} else {
		app.DB().Create(&models.GitlabProject{
			DefaultBranch:   project.DefaultBranch,
			Name:            project.Name,
			GitlabProjectId: input.GitlabProjectID,
			Enabled:         false,
		})
	}

	response.Success(ctx, 200, nil)
}
