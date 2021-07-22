package controllers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"sync"

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
	projects, _, err := utils.GitlabClient().Projects.ListProjects(&gitlab.ListProjectsOptions{Membership: gitlab.Bool(true)})
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}

	var enabledProjects []models.GitlabProject
	utils.DB().Where("`enabled` = ?", true).Find(&enabledProjects)

	var ids = map[int]models.GitlabProject{}

	for _, project := range enabledProjects {
		ids[project.GitlabProjectId] = project
	}

	ch := make(chan Options)

	res := make([]Options, 0, len(projects))
	wg := sync.WaitGroup{}
	wg.Add(len(projects))
	for _, project := range projects {
		go func(project *gitlab.Project) {
			defer wg.Done()
			if p, ok := ids[project.ID]; ok {
				if !p.GlobalEnabled {
					defaultBranch, _ := getDefaultBranch(p.GitlabProjectId)
					if _, err := GetProjectMarsConfig(p.GitlabProjectId, defaultBranch); err != nil {
						mlog.Debug(err)
						return
					}
				}

				ch <- Options{
					Value:     fmt.Sprintf("%d", project.ID),
					Label:     project.Name,
					IsLeaf:    false,
					Type:      OptionTypeProject,
					ProjectId: project.ID,
				}
			}
		}(project)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

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

	branches, _, err := utils.GitlabClient().Branches.ListBranches(uri.ProjectId, &gitlab.ListBranchesOptions{})
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
	branches, _, err := utils.GitlabClient().Branches.ListBranches(projectId, &gitlab.ListBranchesOptions{})
	if err != nil {
		return "", err
	}

	for _, branch := range branches {
		if branch.Default {
			return branch.Name, nil
		}
	}

	return "", errors.New(fmt.Sprintf("project %d cant find default branch", projectId))
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

	commits, _, err := utils.GitlabClient().Commits.ListCommits(uri.ProjectId, &gitlab.ListCommitsOptions{RefName: gitlab.String(uri.Branch)})
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
	configFile := marsC.ConfigFile
	if configFile == "" {
		ct := marsC.ConfigFileType
		if marsC.ConfigFileType == "" {
			ct = "yaml"
		}
		response.Success(ctx, 200, gin.H{
			"data": "",
			"type": ct,
		})
		return
	}
	f, _, err := utils.GitlabClient().RepositoryFiles.GetFile(uri.ProjectId, configFile, &gitlab.GetFileOptions{Ref: gitlab.String(uri.Branch)})
	if err != nil {
		response.Success(ctx, 200, "")
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

	commit, _, err := utils.GitlabClient().Commits.GetCommit(uri.ProjectId, uri.Commit)
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

	commit, _, err := utils.GitlabClient().Commits.GetCommit(uri.ProjectId, uri.Commit)
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
	projects, _, err := utils.GitlabClient().Projects.ListProjects(&gitlab.ListProjectsOptions{Membership: gitlab.Bool(true)})
	if err != nil {
		response.Error(ctx, 500, err)
		return
	}

	var gps []models.GitlabProject
	utils.DB().Find(&gps)

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

	var gp models.GitlabProject
	if utils.DB().Where("`gitlab_project_id` = ?", input.GitlabProjectID).First(&gp).Error == nil {
		utils.DB().Model(&gp).UpdateColumn("enabled", true)
	} else {
		utils.DB().Create(&models.GitlabProject{
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

	var gp models.GitlabProject
	if utils.DB().Where("`gitlab_project_id` = ?", input.GitlabProjectID).First(&gp).Error == nil {
		utils.DB().Model(&gp).UpdateColumn("enabled", false)
	} else {
		utils.DB().Create(&models.GitlabProject{
			GitlabProjectId: input.GitlabProjectID,
			Enabled:         false,
		})
	}

	response.Success(ctx, 200, nil)
}
