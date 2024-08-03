package services

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/duc-cnzj/mars/api/v4/git"
	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/utils/date"
	mars2 "github.com/duc-cnzj/mars/v4/internal/utils/mars"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
}

func NewGitSvc(eventRepo repo.EventRepo, logger mlog.Logger, gitRepo repo.GitRepo, cache cache.Cache, gitProjRepo repo.GitProjectRepo) git.GitServer {
	return &gitSvc{eventRepo: eventRepo, logger: logger, gitRepo: gitRepo, cache: cache, gitProjRepo: gitProjRepo}
}

func (g *gitSvc) EnableProject(ctx context.Context, request *git.EnableProjectRequest) (*git.EnableProjectResponse, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}
	project, err := g.gitRepo.GetByProjectID(ctx, cast.ToInt(request.GitProjectId))
	if err != nil {
		return nil, err
	}

	g.gitProjRepo.Upsert(ctx, &repo.UpsertGitProjectInput{
		DefaultBranch: project.GetDefaultBranch(),
		Name:          project.GetName(),
		GitProjectId:  int(project.GetID()),
		Enabled:       true,
	})
	g.eventRepo.AuditLog(types.EventActionType_Update, MustGetUser(ctx).Name, fmt.Sprintf("启用项目: %s", project.GetName()))

	return &git.EnableProjectResponse{}, nil
}

func (g *gitSvc) DisableProject(ctx context.Context, request *git.DisableProjectRequest) (*git.DisableProjectResponse, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}
	proj, err := g.gitProjRepo.DisabledByProjectID(ctx, cast.ToInt(request.GitProjectId))
	if err != nil {
		return nil, err
	}
	g.eventRepo.AuditLog(types.EventActionType_Update, MustGetUser(ctx).Name, fmt.Sprintf("关闭项目: %s", proj.Name))

	return &git.DisableProjectResponse{}, nil
}

func (g *gitSvc) All(ctx context.Context, req *git.AllRequest) (*git.AllResponse, error) {
	projects, err := g.gitRepo.All(ctx)
	if err != nil {
		return nil, err
	}

	gps, err := g.gitProjRepo.All(ctx, &repo.AllGitProjectInput{})
	if err != nil {
		return nil, err
	}

	var m = map[int]*ent.GitProject{}
	for _, gp := range gps {
		m[gp.GitProjectID] = gp
	}

	var infos = make([]*git.ProjectItem, 0)

	for _, project := range projects {
		var (
			enabled, GlobalEnabled bool
			displayName            string
		)
		if gitProject, ok := m[int(project.GetID())]; ok {
			enabled = gitProject.Enabled
			GlobalEnabled = gitProject.GlobalEnabled
			displayName = gitProject.GlobalMarsConfig().DisplayName
		}
		infos = append(infos, &git.ProjectItem{
			Id:            project.GetID(),
			Name:          project.GetName(),
			Path:          project.GetPath(),
			WebUrl:        project.GetWebURL(),
			AvatarUrl:     project.GetAvatarURL(),
			Description:   project.GetDescription(),
			Enabled:       enabled,
			GlobalEnabled: GlobalEnabled,
			DisplayName:   displayName,
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		if infos[i].Enabled != infos[j].Enabled {
			return infos[i].Enabled
		}

		return infos[i].Id < infos[j].Id
	})

	return &git.AllResponse{Items: infos}, nil
}

const (
	OptionTypeProject string = "project"
	OptionTypeBranch  string = "branch"
	OptionTypeCommit  string = "commit"
)

const ProjectOptionsCacheKey = "ProjectOptions"

func (g *gitSvc) ProjectOptions(ctx context.Context, request *git.ProjectOptionsRequest) (*git.ProjectOptionsResponse, error) {
	remember, err := g.cache.Remember(cache.NewKey(ProjectOptionsCacheKey), 30, func() ([]byte, error) {
		var (
			ch = make(chan *git.Option)
			wg = sync.WaitGroup{}
		)

		enabledProjects, _ := g.gitProjRepo.All(ctx, &repo.AllGitProjectInput{Enabled: lo.ToPtr(true)})
		wg.Add(len(enabledProjects))
		for _, project := range enabledProjects {
			go func(project *ent.GitProject) {
				defer wg.Done()
				defer g.logger.HandlePanic("ProjectOptions")
				var (
					marsC *mars.Config
					err   error
				)
				if marsC, err = GetProjectMarsConfig(project.GitProjectID, project.DefaultBranch); err != nil {
					g.logger.Debug(err)
					return
				}
				var (
					displayName string = project.Name
					pname       string = project.Name
				)
				if marsC.DisplayName != "" {
					displayName = marsC.DisplayName
					pname = fmt.Sprintf("%s(%s)", project.Name, displayName)
					if strings.EqualFold(project.Name, displayName) {
						pname = displayName
					}
				}
				ch <- &git.Option{
					DisplayName:  displayName,
					Value:        fmt.Sprintf("%d", project.GitProjectID),
					Label:        pname,
					IsLeaf:       false,
					Type:         OptionTypeProject,
					GitProjectId: strconv.Itoa(project.GitProjectID),
				}
			}(project)
		}
		go func() {
			wg.Wait()
			close(ch)
		}()

		res := make([]*git.Option, 0)

		for options := range ch {
			res = append(res, options)
		}

		return proto.Marshal(&git.ProjectOptionsResponse{Items: res})
	})
	if err != nil {
		return nil, err
	}
	var res = &git.ProjectOptionsResponse{}
	_ = proto.Unmarshal(remember, res)
	sort.Sort(sortableOption(res.Items))
	return res, nil
}

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
	if request.All {
		return &git.BranchOptionsResponse{Items: res}, nil
	}

	var defaultBranch string
	for _, branch := range branches {
		if branch.IsDefault() {
			defaultBranch = branch.GetName()
		}
	}

	config, err := GetProjectMarsConfig(request.GitProjectId, defaultBranch)
	if err != nil {
		return &git.BranchOptionsResponse{Items: make([]*git.Option, 0)}, nil
	}

	filteredRes := make([]*git.Option, 0)
	for _, op := range res {
		if mars2.BranchPass(config, op.Value) {
			filteredRes = append(filteredRes, op)
		}
	}

	return &git.BranchOptionsResponse{Items: filteredRes}, nil
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
