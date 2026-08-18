package biz

import (
	"context"
	"errors"
	"fmt"

	"github.com/duc-cnzj/mars/v6/internal/errs"
)

// GitBiz 收口 git 托管平台只读查询业务：项目/分支/提交/流水线与文件内容。
type GitBiz interface {
	// AllProjects 返回全部 git 项目（forceFresh 控制是否绕过缓存强刷）。
	AllProjects(ctx context.Context, forceFresh bool) ([]*GitProject, error)
	// AllBranches 返回项目分支列表（forceFresh 控制是否绕过缓存强刷）。
	AllBranches(ctx context.Context, projectID int, forceFresh bool) ([]*Branch, error)
	// ListCommits 返回项目分支的提交列表。
	ListCommits(ctx context.Context, projectID int, branch string) ([]*Commit, error)
	// GetCommit 查询单个提交详情。
	GetCommit(ctx context.Context, projectID int, sha string) (*Commit, error)
	// GetCommitPipeline 查询提交关联的 CI/CD 流水线。
	GetCommitPipeline(ctx context.Context, projectID int, branch, sha string) (*Pipeline, error)
	// PipelineJobOptions 返回项目流水线的 stage/job 去重选项。
	PipelineJobOptions(ctx context.Context, projectID int, branch string) (stages []string, jobs []string, err error)
	// GetByProjectID 按 git 项目 id 查询项目。
	GetByProjectID(ctx context.Context, id int) (*GitProject, error)
	// GetFileContentWithBranch 读取项目指定分支下某路径的文件内容。
	GetFileContentWithBranch(ctx context.Context, projectID int, branch, path string) (string, error)
	// GetProject 按 id 查询 git 项目。
	GetProject(ctx context.Context, id int) (*GitProject, error)
	// GetChartValuesYaml 读取本地 chart 目录的 values.yaml 内容。
	GetChartValuesYaml(ctx context.Context, localChartPath string) (string, error)
	// EnsureBranchAndCommit 解析部署用分支与 commit：未传分支使用仓库默认分支，
	// 未传 commit 使用该分支最新 commit。返回用户可见的提示消息，由上层发送。
	EnsureBranchAndCommit(ctx context.Context, show *Repo, inBranch, inCommit string) (branch string, commit string, msgs []string, err error)
}

type gitBiz struct {
	gitRepo GitRepo
}

// NewGitBiz 构造 git biz。
func NewGitBiz(gitRepo GitRepo) GitBiz {
	return &gitBiz{gitRepo: gitRepo}
}

// AllProjects 返回全部 git 项目（透传 repo）。
func (g *gitBiz) AllProjects(ctx context.Context, forceFresh bool) ([]*GitProject, error) {
	return g.gitRepo.AllProjects(ctx, forceFresh)
}

// AllBranches 返回项目全部分支（透传 repo）。
func (g *gitBiz) AllBranches(ctx context.Context, projectID int, forceFresh bool) ([]*Branch, error) {
	return g.gitRepo.AllBranches(ctx, projectID, forceFresh)
}

// ListCommits 返回项目指定分支的提交列表（透传 repo）。
func (g *gitBiz) ListCommits(ctx context.Context, projectID int, branch string) ([]*Commit, error) {
	return g.gitRepo.ListCommits(ctx, projectID, branch)
}

// GetCommit 按 sha 查询单个提交（透传 repo）。
func (g *gitBiz) GetCommit(ctx context.Context, projectID int, sha string) (*Commit, error) {
	return g.gitRepo.GetCommit(ctx, projectID, sha)
}

// GetCommitPipeline 查询提交对应的 CI 流水线（透传 repo）。
func (g *gitBiz) GetCommitPipeline(ctx context.Context, projectID int, branch, sha string) (*Pipeline, error) {
	return g.gitRepo.GetCommitPipeline(ctx, projectID, branch, sha)
}

// PipelineJobOptions 返回项目流水线的 stage/job 去重选项（透传 repo）。
func (g *gitBiz) PipelineJobOptions(ctx context.Context, projectID int, branch string) (stages []string, jobs []string, err error) {
	return g.gitRepo.PipelineJobOptions(ctx, projectID, branch)
}

// GetByProjectID 按 id 查询 git 项目（透传 repo）。
func (g *gitBiz) GetByProjectID(ctx context.Context, id int) (*GitProject, error) {
	return g.gitRepo.GetByProjectID(ctx, id)
}

// GetFileContentWithBranch 读取分支下指定路径的文件内容（透传 repo）。
func (g *gitBiz) GetFileContentWithBranch(ctx context.Context, projectID int, branch, path string) (string, error) {
	return g.gitRepo.GetFileContentWithBranch(ctx, projectID, branch, path)
}

// GetProject 按 id 查询 git 项目（透传 repo）。
func (g *gitBiz) GetProject(ctx context.Context, id int) (*GitProject, error) {
	return g.gitRepo.GetProject(ctx, id)
}

// GetChartValuesYaml 读取本地 chart 的 values 配置（透传 repo）。
func (g *gitBiz) GetChartValuesYaml(ctx context.Context, localChartPath string) (string, error) {
	return g.gitRepo.GetChartValuesYaml(ctx, localChartPath)
}

// EnsureBranchAndCommit 校验 repo 非空后解析部署分支与 commit：
// 未传分支使用默认分支，未传 commit 使用该分支最新提交。
func (g *gitBiz) EnsureBranchAndCommit(ctx context.Context, show *Repo, inBranch, inCommit string) (branch string, commit string, msgs []string, err error) {
	if show == nil {
		return "", "", nil, errs.WrapInvalidArgument(errors.New("repo 不能为空"), "ensure branch and commit")
	}
	projectID := int(show.GitProjectID)
	branch = inBranch
	commit = inCommit
	if branch == "" {
		branch = show.DefaultBranch
		msgs = append(msgs, fmt.Sprintf("项目 %d 未传入分支，使用默认分支 %s", projectID, branch))
	}
	if commit == "" {
		commits, err := g.gitRepo.ListCommits(ctx, projectID, branch)
		if err != nil {
			return "", "", nil, err
		}
		if len(commits) < 1 {
			return "", "", nil, errs.NotFound(fmt.Sprintf("项目 %d 分支 %s 没有可用的 commit", projectID, branch))
		}
		lastCommit := commits[0]
		commit = lastCommit.ID
		msgs = append(msgs, fmt.Sprintf("项目 %d 分支 %s 未传入commit，使用最新的commit [%s](%s)", projectID, branch, lastCommit.Title, lastCommit.WebURL))
	}
	return branch, commit, msgs, nil
}

// GitRepo 是 git 托管平台操作端口，抽象项目/分支/提交/流水线等只读查询。
type GitRepo interface {
	// AllProjects 返回全部 git 项目（forceFresh 控制是否绕过缓存强刷）。
	AllProjects(ctx context.Context, forceFresh bool) (projects []*GitProject, err error)
	// AllBranches 返回项目的分支列表（forceFresh 控制是否绕过缓存强刷）。
	AllBranches(ctx context.Context, projectID int, forceFresh bool) (branches []*Branch, err error)
	// ListCommits 返回项目分支的提交列表。
	ListCommits(ctx context.Context, projectID int, branch string) ([]*Commit, error)
	// GetCommit 查询单个提交详情。
	GetCommit(ctx context.Context, projectID int, sha string) (*Commit, error)
	// GetCommitPipeline 查询提交关联的 CI/CD 流水线。
	GetCommitPipeline(ctx context.Context, projectID int, branch, sha string) (*Pipeline, error)
	// PipelineJobOptions 返回项目流水线的 stage/job 去重选项。
	PipelineJobOptions(ctx context.Context, projectID int, branch string) (stages []string, jobs []string, err error)
	// GetByProjectID 按 git 项目 id 查询项目。
	GetByProjectID(ctx context.Context, id int) (project *GitProject, err error)
	// GetFileContentWithBranch 读取项目指定分支下某路径的文件内容。
	GetFileContentWithBranch(ctx context.Context, projectID int, branch, path string) (string, error)
	// GetProject 按 id 查询 git 项目。
	GetProject(ctx context.Context, id int) (project *GitProject, err error)
	// GetChartValuesYaml 读取本地 chart 目录的 values.yaml 内容。
	GetChartValuesYaml(ctx context.Context, localChartPath string) (string, error)
}
