package controllers

import (
	"encoding/base64"
	"fmt"

	"github.com/DuC-cnZj/mars/pkg/response"
	"github.com/DuC-cnZj/mars/pkg/utils"
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

	res := make([]Options, 0, len(projects))
	for _, project := range projects {
		res = append(res, Options{
			Value:     fmt.Sprintf("%d", project.ID),
			Label:     project.Name,
			IsLeaf:    false,
			Type:      OptionTypeProject,
			ProjectId: project.ID,
		})
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

	res := make([]Options, 0, len(branches))
	for _, branch := range branches {
		res = append(res, Options{
			Value:     branch.Name,
			Label:     branch.Name,
			IsLeaf:    false,
			Type:      OptionTypeBranch,
			Branch:    branch.Name,
			ProjectId: uri.ProjectId,
		})
	}

	response.Success(ctx, 200, res)
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
