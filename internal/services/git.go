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

// PipelineInfo 返回指定提交对应 CI 流水线的状态、URL 与各 job 状态，status 即流水线整体状态，
// 不做 pass 规则判定。需要按仓库配置的 pass 规则判定 status 时用 PipelineInfoByRepoId。
func (g *gitSvc) PipelineInfo(ctx context.Context, request *git.PipelineInfoRequest) (*git.PipelineInfoResponse, error) {
	projectID := cast.ToInt(request.GitProjectId)
	pipeline, err := g.gitBiz.GetCommitPipeline(ctx, projectID, request.Branch, request.Commit)
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}

	jobs := make([]*git.PipelineJob, 0, len(pipeline.Jobs))
	for _, j := range pipeline.Jobs {
		jobs = append(jobs, &git.PipelineJob{Name: j.Name, Status: j.Status, StageName: j.StageName})
	}
	return &git.PipelineInfoResponse{
		Status: pipeline.Status,
		WebUrl: pipeline.WebURL,
		Jobs:   jobs,
	}, nil
}

// PipelineInfoByRepoId 返回指定 repo 下某提交对应 CI 流水线的状态、URL 与各 job 状态。
// 与 PipelineInfo 的差异在入参主语：本方法按 repo_id（mars 部署仓库主键）解析出
// git 项目 ID 与通过规则，再判定流水线状态；repo 不存在时透传 404。仓库配置 pass
// 规则时 status 由规则判定（见 biz.PipelinePassStatus），未配置时返回整体流水线状态。
func (g *gitSvc) PipelineInfoByRepoId(ctx context.Context, request *git.PipelineInfoByRepoIdRequest) (*git.PipelineInfoResponse, error) {
	repoID := int(request.RepoId)
	repo, err := g.repoBiz.Get(ctx, repoID)
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}

	pipeline, err := g.gitBiz.GetCommitPipeline(ctx, int(repo.GitProjectID), request.Branch, request.Commit)
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}

	status := pipeline.Status
	if rules := repo.GetMarsConfig().PipelinePassRules; len(rules) > 0 {
		status = biz.PipelinePassStatus(pipeline, rules)
	}

	jobs := make([]*git.PipelineJob, 0, len(pipeline.Jobs))
	for _, j := range pipeline.Jobs {
		jobs = append(jobs, &git.PipelineJob{Name: j.Name, Status: j.Status, StageName: j.StageName})
	}
	return &git.PipelineInfoResponse{
		Status: status,
		WebUrl: pipeline.WebURL,
		Jobs:   jobs,
	}, nil
}

// PipelineJobOptions 返回项目流水线的 stage/job 去重选项，供配置通过规则下拉选择。
// branch 为空时取项目最近 pipeline，非空时取该分支最近 pipeline。
func (g *gitSvc) PipelineJobOptions(ctx context.Context, request *git.PipelineJobOptionsRequest) (*git.PipelineJobOptionsResponse, error) {
	stages, jobs, err := g.gitBiz.PipelineJobOptions(ctx, int(request.GitProjectId), request.Branch)
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}
	return &git.PipelineJobOptionsResponse{Stages: stages, Jobs: jobs}, nil
}

// GetChartValuesYaml 解析 helm chart 的 values.yaml：把前端提交的表单项映射为 values 返回。
func (g *gitSvc) GetChartValuesYaml(ctx context.Context, req *git.GetChartValuesYamlRequest) (*git.GetChartValuesYamlResponse, error) {
	yaml, err := g.gitBiz.GetChartValuesYaml(ctx, req.GetInput())
	if err != nil {
		return nil, logError(ctx, g.logger, err)
	}
	return &git.GetChartValuesYamlResponse{Values: yaml}, nil
}
