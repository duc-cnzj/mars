package services

import (
	"context"
	"fmt"

	"github.com/duc-cnzj/mars/api/v6/proto/git"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/date"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

const (
	optionTypeProject string = "project"
	optionTypeBranch  string = "branch"
	optionTypeCommit  string = "commit"
)

var _ git.GitServer = (*gitSvc)(nil)

// gitSvc 是 git.GitServer 的 gRPC 实现：提供仓库/分支/提交选项查询与提交操作、
// CI 流水线信息与 Chart values 获取，由 NewGitSvc 构造。
type gitSvc struct {
	git.UnimplementedGitServer

	logger  mlog.Logger
	gitBiz  biz.GitBiz
	repoBiz biz.RepoBiz
}

// GitSvcDeps 收口 NewGitSvc 的构造依赖，由 wire 按字段注入。
type GitSvcDeps struct {
	RepoBiz biz.RepoBiz
	Logger  mlog.Logger
	GitBiz  biz.GitBiz
}

// NewGitSvc 收口 git 服务的构造依赖，由 wire 按字段注入。
func NewGitSvc(deps GitSvcDeps) git.GitServer {
	return &gitSvc{
		logger:  deps.Logger.WithModule("services/git"),
		gitBiz:  deps.GitBiz,
		repoBiz: deps.RepoBiz,
	}
}

// AllRepos 返回 git 侧全部项目（含未启用 repo 的项目），供全局搜索/下拉使用。
func (g *gitSvc) AllRepos(ctx context.Context, req *git.AllReposRequest) (*git.AllReposResponse, error) {
	projects, err := g.gitBiz.AllProjects(ctx, false)
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}
	return &git.AllReposResponse{
		Items: slice.Map(projects, func(v *biz.GitProject) *git.AllReposResponse_Item {
			if v == nil {
				return nil
			}
			return &git.AllReposResponse_Item{
				Id:          int32(v.ID),
				Name:        v.Name,
				Description: v.Description,
			}
		}),
	}, nil
}

// ProjectOptions 返回已启用的 repo 项目，作为部署表单的 project 级下拉选项。
func (g *gitSvc) ProjectOptions(ctx context.Context, request *git.ProjectOptionsRequest) (*git.ProjectOptionsResponse, error) {
	all, err := g.repoBiz.All(ctx, &biz.AllRepoRequest{Enabled: lo.ToPtr(true)})
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}
	var gitOptions []*git.Option
	for _, repo := range all {
		gitOptions = append(gitOptions, &git.Option{
			Value:        cast.ToString(repo.ID),
			Label:        repo.Name,
			Type:         optionTypeProject,
			IsLeaf:       false,
			GitProjectId: repo.GitProjectID,
			NeedGitRepo:  repo.NeedGitRepo,
			Description:  repo.Description,
		})
	}

	return &git.ProjectOptionsResponse{Items: gitOptions}, nil
}

// BranchOptions 返回指定 git 项目的分支列表；带 repo 时基于其分支白名单过滤。
func (g *gitSvc) BranchOptions(ctx context.Context, request *git.BranchOptionsRequest) (*git.BranchOptionsResponse, error) {
	branches, err := g.gitBiz.AllBranches(ctx, cast.ToInt(request.GitProjectId), false)
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}
	res := make([]*git.Option, 0, len(branches))
	for _, branch := range branches {
		branchName := branch.Name
		res = append(res, &git.Option{
			Value:        branchName,
			Label:        branchName,
			IsLeaf:       false,
			Type:         optionTypeBranch,
			Branch:       branchName,
			GitProjectId: request.GitProjectId,
		})
	}
	if request.RepoId > 0 {
		show, err := g.repoBiz.Get(ctx, int(request.RepoId))
		if err != nil {
			return nil, logError(ctx, g.logger, err)
		}
		res = lo.Filter(res, func(b *git.Option, _ int) bool {
			return biz.MatchBranch(show.GetMarsConfig().Branches, b.Branch)
		})
	}

	return &git.BranchOptionsResponse{Items: res}, nil
}

// CommitOptions 返回指定分支的提交历史，作为部署表单的 commit 级下拉选项。
func (g *gitSvc) CommitOptions(ctx context.Context, request *git.CommitOptionsRequest) (*git.CommitOptionsResponse, error) {
	commits, err := g.gitBiz.ListCommits(ctx, cast.ToInt(request.GitProjectId), request.Branch)
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}
	res := make([]*git.Option, 0, len(commits))
	for _, commit := range commits {
		res = append(res, &git.Option{
			Value:        commit.ID,
			IsLeaf:       true,
			Label:        fmt.Sprintf("[%s]: %s", date.ToHumanizeDateTime(commit.CommittedDate), commit.Title),
			Type:         optionTypeCommit,
			GitProjectId: request.GitProjectId,
			Branch:       request.Branch,
		})
	}

	return &git.CommitOptionsResponse{Items: res}, nil
}

// Commit 按 git 项目 + 提交号返回单个提交的完整信息（含 author/committer/时间/URL）。
func (g *gitSvc) Commit(ctx context.Context, request *git.CommitRequest) (*git.CommitResponse, error) {
	commit, err := g.gitBiz.GetCommit(ctx, cast.ToInt(request.GitProjectId), request.Commit)
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}
	return &git.CommitResponse{
		Id:             commit.ID,
		ShortId:        commit.ShortID,
		GitProjectId:   request.GitProjectId,
		Label:          fmt.Sprintf("[%s]: %s", date.ToHumanizeDateTime(commit.CommittedDate), commit.Title),
		Title:          commit.Title,
		Branch:         request.Branch,
		AuthorName:     commit.AuthorName,
		AuthorEmail:    commit.AuthorEmail,
		CommitterName:  commit.CommitterName,
		CommitterEmail: commit.CommitterEmail,
		WebUrl:         commit.WebURL,
		Message:        commit.Message,
		CommittedDate:  date.ToRFC3339(commit.CommittedDate),
		CreatedAt:      date.ToRFC3339(commit.CreatedAt),
	}, nil
}

// PipelineInfo 返回指定提交对应 CI 流水线的状态与 URL。
func (g *gitSvc) PipelineInfo(ctx context.Context, request *git.PipelineInfoRequest) (*git.PipelineInfoResponse, error) {
	pipeline, err := g.gitBiz.GetCommitPipeline(ctx, cast.ToInt(request.GitProjectId), request.Branch, request.Commit)
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}

	return &git.PipelineInfoResponse{
		Status: pipeline.Status,
		WebUrl: pipeline.WebURL,
	}, nil
}

// GetChartValuesYaml 解析 helm chart 的 values.yaml：把前端提交的表单项映射为 values 返回。
func (g *gitSvc) GetChartValuesYaml(ctx context.Context, req *git.GetChartValuesYamlRequest) (*git.GetChartValuesYamlResponse, error) {
	yaml, err := g.gitBiz.GetChartValuesYaml(ctx, req.GetInput())
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}
	return &git.GetChartValuesYamlResponse{Values: yaml}, nil
}
