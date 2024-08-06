package services

import (
	"context"
	"fmt"
	gopath "path"
	"sort"
	"strings"

	"github.com/duc-cnzj/mars/api/v4/git"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/util/date"
	mars2 "github.com/duc-cnzj/mars/v4/internal/util/mars"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/proto"
)

var _ git.GitServer = (*gitSvc)(nil)

type gitSvc struct {
	git.UnimplementedGitServer

	eventRepo   repo.EventRepo
	logger      mlog.Logger
	gitRepo     repo.GitRepo
	cache       cache.Cache
	gitProjRepo repo.GitProjectRepo
	repoRepo    repo.RepoImp
}

func NewGitSvc(repoRepo repo.RepoImp, eventRepo repo.EventRepo, logger mlog.Logger, gitRepo repo.GitRepo, cache cache.Cache, gitProjRepo repo.GitProjectRepo) git.GitServer {
	return &gitSvc{repoRepo: repoRepo, eventRepo: eventRepo, logger: logger, gitRepo: gitRepo, cache: cache, gitProjRepo: gitProjRepo}
}

func (g *gitSvc) AllRepos(ctx context.Context, req *git.AllReposRequest) (*git.AllReposResponse, error) {
	remember, err := g.cache.Remember(cache.NewKey("all_repos"), 600, func() ([]byte, error) {
		projects, err := g.gitRepo.All(ctx)
		if err != nil {
			return nil, err
		}
		var res []*git.AllReposResponse_Item
		for _, project := range projects {
			res = append(res, &git.AllReposResponse_Item{
				Id:          int32(project.GetID()),
				Name:        project.GetName(),
				Description: project.GetDescription(),
			})
		}
		return proto.Marshal(&git.AllReposResponse{Items: res})
	})
	if err != nil {
		return nil, err
	}
	var res git.AllReposResponse
	proto.Unmarshal(remember, &res)
	return &res, nil
}

func (g *gitSvc) GetChartValuesYaml(ctx context.Context, req *git.GetChartValuesYamlRequest) (*git.GetChartValuesYamlResponse, error) {
	if !mars2.IsRemoteLocalChartPath(req.GetInput()) {
		return &git.GetChartValuesYamlResponse{}, nil
	}

	split := strings.Split(req.GetInput(), "|")
	if len(split) != 3 {
		return nil, fmt.Errorf("invalid input: %s", req.GetInput())
	}
	pid := split[0]
	branch := split[1]
	filename := gopath.Join(split[2], "values.yaml")

	content, err := g.gitRepo.GetFileContentWithBranch(ctx, cast.ToInt(pid), branch, filename)
	if err != nil {
		return nil, err
	}
	return &git.GetChartValuesYamlResponse{Values: content}, nil
}

func (g *gitSvc) ProjectOptions(ctx context.Context, request *git.ProjectOptionsRequest) (*git.ProjectOptionsResponse, error) {
	all, err := g.repoRepo.All(context.TODO(), &repo.AllRepoRequest{Enabled: lo.ToPtr(true)})
	if err != nil {
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
			//DisplayName:  repo.Name,
			NeedGitRepo: repo.NeedGitRepo,
		})
	}
	sort.Sort(sortableOption(gitOptions))

	return &git.ProjectOptionsResponse{Items: gitOptions}, nil
}

const (
	OptionTypeProject string = "project"
	OptionTypeBranch  string = "branch"
	OptionTypeCommit  string = "commit"
)

type sortableOption []*git.Option

func (s sortableOption) Len() int {
	return len(s)
}

func (s sortableOption) Less(i, j int) bool {
	return s[i].GitProjectId < s[j].GitProjectId
}

func (s sortableOption) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (g *gitSvc) BranchOptions(ctx context.Context, request *git.BranchOptionsRequest) (*git.BranchOptionsResponse, error) {
	remember, err := g.cache.Remember(cache.NewKey("branch_options"), 120, func() ([]byte, error) {
		branches, err := g.gitRepo.AllProjectBranches(ctx, cast.ToInt(request.GitProjectId))
		if err != nil {
			return nil, err
		}
		res := make([]*git.Option, 0, len(branches))
		for _, branch := range branches {
			branchName := branch.GetName()
			res = append(res, &git.Option{
				Value:        branchName,
				Label:        branchName,
				IsLeaf:       false,
				Type:         OptionTypeBranch,
				Branch:       branchName,
				GitProjectId: request.GitProjectId,
			})
		}
		return proto.Marshal(&git.BranchOptionsResponse{Items: res})
	})
	if err != nil {
		return nil, err
	}
	var res git.BranchOptionsResponse
	proto.Unmarshal(remember, &res)
	if request.All {
		return &res, nil
	}

	//var defaultBranch string
	//for _, branch := range branches {
	//	if branch.IsDefault() {
	//		defaultBranch = branch.GetName()
	//	}
	//}

	//config, err := GetProjectMarsConfig(request.GitProjectId, defaultBranch)
	//if err != nil {
	//	return &git.BranchOptionsResponse{Items: make([]*git.Option, 0)}, nil
	//}

	//filteredRes := make([]*git.Option, 0)
	//for _, op := range res {
	//	if mars2.BranchPass(config, op.Value) {
	//		filteredRes = append(filteredRes, op)
	//	}
	//}

	return &res, nil
}

func (g *gitSvc) CommitOptions(ctx context.Context, request *git.CommitOptionsRequest) (*git.CommitOptionsResponse, error) {
	commits, err := g.gitRepo.ListCommits(ctx, cast.ToInt(request.GitProjectId), request.Branch)
	if err != nil {
		return nil, err
	}
	res := make([]*git.Option, 0, len(commits))
	for _, commit := range commits {
		res = append(res, &git.Option{
			Value:        commit.GetID(),
			IsLeaf:       true,
			Label:        fmt.Sprintf("[%s]: %s", date.ToHumanizeDatetimeString(commit.GetCommittedDate()), commit.GetTitle()),
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
		return nil, err
	}
	return &git.CommitResponse{
		Id:             commit.GetID(),
		ShortId:        commit.GetShortID(),
		GitProjectId:   request.GitProjectId,
		Label:          fmt.Sprintf("[%s]: %s", date.ToHumanizeDatetimeString(commit.GetCommittedDate()), commit.GetTitle()),
		Title:          commit.GetTitle(),
		Branch:         request.Branch,
		AuthorName:     commit.GetAuthorName(),
		AuthorEmail:    commit.GetAuthorEmail(),
		CommitterName:  commit.GetCommitterName(),
		CommitterEmail: commit.GetCommitterEmail(),
		WebUrl:         commit.GetWebURL(),
		Message:        commit.GetMessage(),
		CommittedDate:  date.ToRFC3339DatetimeString(commit.GetCommittedDate()),
		CreatedAt:      date.ToRFC3339DatetimeString(commit.GetCreatedAt()),
	}, nil
}

func (g *gitSvc) PipelineInfo(ctx context.Context, request *git.PipelineInfoRequest) (*git.PipelineInfoResponse, error) {
	pipeline, err := g.gitRepo.GetCommitPipeline(ctx, cast.ToInt(request.GitProjectId), request.Branch, request.Commit)
	if err != nil {
		return nil, err
	}

	return &git.PipelineInfoResponse{
		Status: pipeline.GetStatus(),
		WebUrl: pipeline.GetWebURL(),
	}, nil
}

func (g *gitSvc) MarsConfigFile(ctx context.Context, request *git.MarsConfigFileRequest) (*git.MarsConfigFileResponse, error) {
	marsC, err := GetProjectMarsConfig(request.GitProjectId, request.Branch)
	if err != nil {
		return nil, err
	}
	// 先拿 ConfigFile 如果没有，则拿 ConfigFileValues
	configFile := marsC.ConfigFile
	if configFile == "" {
		ct := marsC.ConfigFileType
		if marsC.ConfigFileType == "" {
			ct = "yaml"
		}
		return &git.MarsConfigFileResponse{
			Data:     marsC.ConfigFileValues,
			Type:     ct,
			Elements: marsC.Elements,
		}, nil
	}
	// 如果有 ConfigFile，则获取内容，如果没有内容，则使用 ConfigFileValues

	var (
		pid      string
		branch   string
		filename string
	)

	if mars2.IsRemoteConfigFile(marsC) {
		split := strings.Split(configFile, "|")
		pid = split[0]
		branch = split[1]
		filename = split[2]
	} else {
		pid = fmt.Sprintf("%v", request.GitProjectId)
		branch = request.Branch
		filename = configFile
	}

	content, err := g.gitRepo.GetFileContentWithBranch(ctx, cast.ToInt(pid), branch, filename)
	if err != nil {
		g.logger.Debug(err)
		return &git.MarsConfigFileResponse{
			Data:     marsC.ConfigFileValues,
			Type:     marsC.ConfigFileType,
			Elements: marsC.Elements,
		}, nil
	}

	return &git.MarsConfigFileResponse{
		Data:     content,
		Type:     marsC.ConfigFileType,
		Elements: marsC.Elements,
	}, nil
}
