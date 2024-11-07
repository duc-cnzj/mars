package services

import (
	"context"
	"fmt"

	"github.com/duc-cnzj/mars/api/v5/git"
	"github.com/duc-cnzj/mars/v5/internal/cache"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/util/date"
	mars2 "github.com/duc-cnzj/mars/v5/internal/util/mars"
	"github.com/duc-cnzj/mars/v5/internal/util/serialize"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

const (
	OptionTypeProject string = "project"
	OptionTypeBranch  string = "branch"
	OptionTypeCommit  string = "commit"
)

var _ git.GitServer = (*gitSvc)(nil)

type gitSvc struct {
	git.UnimplementedGitServer

	eventRepo repo.EventRepo
	logger    mlog.Logger
	gitRepo   repo.GitRepo
	cache     cache.Cache
	repoRepo  repo.RepoRepo
}

func NewGitSvc(repoRepo repo.RepoRepo, eventRepo repo.EventRepo, logger mlog.Logger, gitRepo repo.GitRepo, cache cache.Cache) git.GitServer {
	return &gitSvc{
		eventRepo: eventRepo,
		logger:    logger.WithModule("services/git"),
		gitRepo:   gitRepo,
		cache:     cache,
		repoRepo:  repoRepo,
	}
}

func (g *gitSvc) AllRepos(ctx context.Context, req *git.AllReposRequest) (*git.AllReposResponse, error) {
	projects, err := g.gitRepo.AllProjects(ctx, false)
	if err != nil {
		g.logger.ErrorCtx(ctx, err)
		return nil, err
	}
	return &git.AllReposResponse{
		Items: serialize.Serialize(projects, func(v *repo.GitProject) *git.AllReposResponse_Item {
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

func (g *gitSvc) ProjectOptions(ctx context.Context, request *git.ProjectOptionsRequest) (*git.ProjectOptionsResponse, error) {
	all, err := g.repoRepo.All(context.TODO(), &repo.AllRepoRequest{Enabled: lo.ToPtr(true)})
	if err != nil {
		g.logger.ErrorCtx(ctx, err)
		return nil, err
	}
	var gitOptions []*git.Option
	for _, repo := range all {
		gitOptions = append(gitOptions, &git.Option{
			Value:        cast.ToString(repo.ID),
			Label:        repo.Name,
			Type:         OptionTypeProject,
			IsLeaf:       false,
			GitProjectId: repo.GitProjectID,
			NeedGitRepo:  repo.NeedGitRepo,
			Description:  repo.Description,
		})
	}

	return &git.ProjectOptionsResponse{Items: gitOptions}, nil
}

func (g *gitSvc) BranchOptions(ctx context.Context, request *git.BranchOptionsRequest) (*git.BranchOptionsResponse, error) {
	branches, err := g.gitRepo.AllBranches(ctx, cast.ToInt(request.GitProjectId), false)
	if err != nil {
		g.logger.ErrorCtx(ctx, err)
		return nil, err
	}
	res := make([]*git.Option, 0, len(branches))
	for _, branch := range branches {
		branchName := branch.Name
		res = append(res, &git.Option{
			Value:        branchName,
			Label:        branchName,
			IsLeaf:       false,
			Type:         OptionTypeBranch,
			Branch:       branchName,
			GitProjectId: request.GitProjectId,
		})
	}
	if request.RepoId > 0 {
		show, err := g.repoRepo.Get(ctx, int(request.RepoId))
		if err != nil {
			return nil, err
		}
		res = lo.Filter(res, func(b *git.Option, _ int) bool {
			return mars2.BranchPass(show.MarsConfig.Branches, b.Branch)
		})
	}

	return &git.BranchOptionsResponse{Items: res}, nil
}

func (g *gitSvc) CommitOptions(ctx context.Context, request *git.CommitOptionsRequest) (*git.CommitOptionsResponse, error) {
	commits, err := g.gitRepo.ListCommits(ctx, cast.ToInt(request.GitProjectId), request.Branch)
	if err != nil {
		g.logger.ErrorCtx(ctx, err)
		return nil, err
	}
	res := make([]*git.Option, 0, len(commits))
	for _, commit := range commits {
		res = append(res, &git.Option{
			Value:        commit.ID,
			IsLeaf:       true,
			Label:        fmt.Sprintf("[%s]: %s", date.ToHumanizeDatetimeString(commit.CommittedDate), commit.Title),
			Type:         OptionTypeCommit,
			GitProjectId: request.GitProjectId,
			Branch:       request.Branch,
		})
	}

	return &git.CommitOptionsResponse{Items: res}, nil
}

func (g *gitSvc) Commit(ctx context.Context, request *git.CommitRequest) (*git.CommitResponse, error) {
	commit, err := g.gitRepo.GetCommit(ctx, cast.ToInt(request.GitProjectId), request.Commit)
	if err != nil {
		g.logger.ErrorCtx(ctx, err)
		return nil, err
	}
	return &git.CommitResponse{
		Id:             commit.ID,
		ShortId:        commit.ShortID,
		GitProjectId:   request.GitProjectId,
		Label:          fmt.Sprintf("[%s]: %s", date.ToHumanizeDatetimeString(commit.CommittedDate), commit.Title),
		Title:          commit.Title,
		Branch:         request.Branch,
		AuthorName:     commit.AuthorName,
		AuthorEmail:    commit.AuthorEmail,
		CommitterName:  commit.CommitterName,
		CommitterEmail: commit.CommitterEmail,
		WebUrl:         commit.WebURL,
		Message:        commit.Message,
		CommittedDate:  date.ToRFC3339DatetimeString(commit.CommittedDate),
		CreatedAt:      date.ToRFC3339DatetimeString(commit.CreatedAt),
	}, nil
}

func (g *gitSvc) PipelineInfo(ctx context.Context, request *git.PipelineInfoRequest) (*git.PipelineInfoResponse, error) {
	pipeline, err := g.gitRepo.GetCommitPipeline(ctx, cast.ToInt(request.GitProjectId), request.Branch, request.Commit)
	if err != nil {
		return nil, err
	}

	return &git.PipelineInfoResponse{
		Status: pipeline.Status,
		WebUrl: pipeline.WebURL,
	}, nil
}

func (g *gitSvc) GetChartValuesYaml(ctx context.Context, req *git.GetChartValuesYamlRequest) (*git.GetChartValuesYamlResponse, error) {
	yaml, err := g.gitRepo.GetChartValuesYaml(ctx, req.GetInput())
	if err != nil {
		return nil, err
	}
	return &git.GetChartValuesYamlResponse{Values: yaml}, nil
}
