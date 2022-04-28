package services

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/duc-cnzj/mars-client/v4/git"
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/internal/utils/date"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		git.RegisterGitServer(s, new(GitSvc))
	})
	RegisterEndpoint(git.RegisterGitHandlerFromEndpoint)
}

type GitSvc struct {
	git.UnimplementedGitServer
}

func (g *GitSvc) EnableProject(ctx context.Context, request *git.EnableProjectRequest) (*git.EnableProjectResponse, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}
	project, _ := plugins.GetGitServer().GetProject(request.GitProjectId)

	var gp models.GitProject
	if app.DB().Where("`git_project_id` = ?", request.GitProjectId).First(&gp).Error == nil {
		app.DB().Model(&gp).Updates(map[string]any{
			"enabled":        true,
			"default_branch": project.GetDefaultBranch(),
			"name":           project.GetName(),
		})
	} else {
		atoi, _ := strconv.Atoi(request.GitProjectId)
		app.DB().Create(&models.GitProject{
			DefaultBranch: project.GetDefaultBranch(),
			Name:          project.GetName(),
			GitProjectId:  atoi,
			Enabled:       true,
		})
	}
	AuditLog(MustGetUser(ctx).Name, types.EventActionType_Create, fmt.Sprintf("启用项目: %s", project.GetName()))

	return &git.EnableProjectResponse{}, nil
}

func (g *GitSvc) DisableProject(ctx context.Context, request *git.DisableProjectRequest) (*git.DisableProjectResponse, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}
	project, _ := plugins.GetGitServer().GetProject(request.GitProjectId)
	var gp models.GitProject
	if app.DB().Where("`git_project_id` = ?", request.GitProjectId).First(&gp).Error == nil {
		app.DB().Model(&gp).Updates(map[string]any{
			"enabled":        false,
			"default_branch": project.GetDefaultBranch(),
			"name":           project.GetName(),
		})
	} else {
		itoa, _ := strconv.Atoi(request.GitProjectId)
		app.DB().Create(&models.GitProject{
			DefaultBranch: project.GetDefaultBranch(),
			Name:          project.GetName(),
			GitProjectId:  itoa,
			Enabled:       false,
		})
	}
	AuditLog(MustGetUser(ctx).Name, types.EventActionType_Create, fmt.Sprintf("关闭项目: %s", project.GetName()))

	return &git.DisableProjectResponse{}, nil
}

func (g *GitSvc) All(ctx context.Context, req *git.AllProjectsRequest) (*git.AllProjectsResponse, error) {
	projects, err := plugins.GetGitServer().AllProjects()
	if err != nil {
		return nil, err
	}

	var gps []models.GitProject
	app.DB().Find(&gps)

	var m = map[int]models.GitProject{}
	for _, gp := range gps {
		m[gp.GitProjectId] = gp
	}

	var infos = make([]*git.ProjectItem, 0)

	for _, project := range projects {
		var enabled, GlobalEnabled bool
		if gitProject, ok := m[int(project.GetID())]; ok {
			enabled = gitProject.Enabled
			GlobalEnabled = gitProject.GlobalEnabled
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
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Id > infos[j].Id
	})

	return &git.AllProjectsResponse{Items: infos}, nil
}

const (
	OptionTypeProject string = "project"
	OptionTypeBranch  string = "branch"
	OptionTypeCommit  string = "commit"
)

func (g *GitSvc) ProjectOptions(ctx context.Context, request *git.ProjectOptionsRequest) (*git.ProjectOptionsResponse, error) {
	remember, err := app.Cache().Remember("ProjectOptions", 30, func() ([]byte, error) {
		var (
			enabledProjects []models.GitProject
			ch              = make(chan *git.Option)
			wg              = sync.WaitGroup{}
		)

		app.DB().Where("`enabled` = ?", true).Find(&enabledProjects)
		wg.Add(len(enabledProjects))
		for _, project := range enabledProjects {
			go func(project models.GitProject) {
				defer wg.Done()
				defer utils.HandlePanic("ProjectOptions")
				if !project.GlobalEnabled {
					if _, err := GetProjectMarsConfig(project.GitProjectId, project.DefaultBranch); err != nil {
						mlog.Debug(err)
						return
					}
				}
				ch <- &git.Option{
					Value:        fmt.Sprintf("%d", project.GitProjectId),
					Label:        project.Name,
					IsLeaf:       false,
					Type:         OptionTypeProject,
					GitProjectId: strconv.Itoa(project.GitProjectId),
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
	return res, nil
}

func (g *GitSvc) BranchOptions(ctx context.Context, request *git.BranchOptionsRequest) (*git.BranchOptionsResponse, error) {
	branches, err := plugins.GetGitServer().AllBranches(request.GitProjectId)
	if err != nil {
		return nil, err
	}

	res := make([]*git.Option, 0, len(branches))
	for _, branch := range branches {
		res = append(res, &git.Option{
			Value:        branch.GetName(),
			Label:        branch.GetName(),
			IsLeaf:       false,
			Type:         OptionTypeBranch,
			Branch:       branch.GetName(),
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
		if utils.BranchPass(config, op.Value) {
			filteredRes = append(filteredRes, op)
		}
	}

	return &git.BranchOptionsResponse{Items: filteredRes}, nil
}

func (g *GitSvc) CommitOptions(ctx context.Context, request *git.CommitOptionsRequest) (*git.CommitOptionsResponse, error) {
	commits, err := plugins.GetGitServer().ListCommits(request.GitProjectId, request.Branch)
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

func (g *GitSvc) Commit(ctx context.Context, request *git.CommitRequest) (*git.CommitResponse, error) {
	remember, err := app.Cache().Remember(fmt.Sprintf("Commit:%s-%s", request.GitProjectId, request.Commit), 60*60, func() ([]byte, error) {
		commit, err := plugins.GetGitServer().GetCommit(request.GitProjectId, request.Commit)
		if err != nil {
			return nil, err
		}
		res := &git.CommitResponse{
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
		}
		return proto.Marshal(res)
	})
	if err != nil {
		return nil, err
	}
	msg := &git.CommitResponse{}
	_ = proto.Unmarshal(remember, msg)
	return msg, nil
}

func (g *GitSvc) PipelineInfo(ctx context.Context, request *git.PipelineInfoRequest) (*git.PipelineInfoResponse, error) {
	pipeline, err := plugins.GetGitServer().GetCommitPipeline(request.GitProjectId, request.Commit)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &git.PipelineInfoResponse{
		Status: pipeline.GetStatus(),
		WebUrl: pipeline.GetWebURL(),
	}, nil
}

func (g *GitSvc) MarsConfigFile(ctx context.Context, request *git.ConfigFileRequest) (*git.ConfigFileResponse, error) {
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
		return &git.ConfigFileResponse{
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

	if utils.IsRemoteConfigFile(marsC) {
		split := strings.Split(configFile, "|")
		pid = split[0]
		branch = split[1]
		filename = split[2]
	} else {
		pid = fmt.Sprintf("%v", request.GitProjectId)
		branch = request.Branch
		filename = configFile
	}

	content, err := plugins.GetGitServer().GetFileContentWithBranch(pid, branch, filename)
	if err != nil {
		mlog.Debug(err)
		return &git.ConfigFileResponse{
			Data:     marsC.ConfigFileValues,
			Type:     marsC.ConfigFileType,
			Elements: marsC.Elements,
		}, nil
	}

	return &git.ConfigFileResponse{
		Data:     content,
		Type:     marsC.ConfigFileType,
		Elements: marsC.Elements,
	}, nil
}
