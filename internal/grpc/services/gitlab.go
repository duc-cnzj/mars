package services

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/event"
	"github.com/duc-cnzj/mars/pkg/gitlab"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Gitlab struct {
	gitlab.UnimplementedGitlabServer
}

func (g *Gitlab) EnableProject(ctx context.Context, request *gitlab.EnableProjectRequest) (*emptypb.Empty, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}
	project, _ := plugins.GetGitServer().GetProject(request.GitlabProjectId)

	var gp models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", request.GitlabProjectId).First(&gp).Error == nil {
		app.DB().Model(&gp).Updates(map[string]interface{}{
			"enabled":        true,
			"default_branch": project.GetDefaultBranch(),
			"name":           project.GetName(),
		})
	} else {
		atoi, _ := strconv.Atoi(request.GitlabProjectId)
		app.DB().Create(&models.GitlabProject{
			DefaultBranch:   project.GetDefaultBranch(),
			Name:            project.GetName(),
			GitlabProjectId: atoi,
			Enabled:         true,
		})
	}
	AuditLog(MustGetUser(ctx).Name, event.ActionType_Create, fmt.Sprintf("启用项目: %s", project.GetName()))

	return &emptypb.Empty{}, nil
}

func (g *Gitlab) DisableProject(ctx context.Context, request *gitlab.DisableProjectRequest) (*emptypb.Empty, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}
	project, _ := plugins.GetGitServer().GetProject(request.GitlabProjectId)
	var gp models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", request.GitlabProjectId).First(&gp).Error == nil {
		app.DB().Model(&gp).Updates(map[string]interface{}{
			"enabled":        false,
			"default_branch": project.GetDefaultBranch(),
			"name":           project.GetName(),
		})
	} else {
		itoa, _ := strconv.Atoi(request.GitlabProjectId)
		app.DB().Create(&models.GitlabProject{
			DefaultBranch:   project.GetDefaultBranch(),
			Name:            project.GetName(),
			GitlabProjectId: itoa,
			Enabled:         false,
		})
	}
	AuditLog(MustGetUser(ctx).Name, event.ActionType_Create, fmt.Sprintf("关闭项目: %s", project.GetName()))

	return &emptypb.Empty{}, nil
}

func (g *Gitlab) ProjectList(ctx context.Context, empty *emptypb.Empty) (*gitlab.ProjectListResponse, error) {
	projects, err := plugins.GetGitServer().ListProjects(1, 100)
	if err != nil {
		return nil, err
	}

	var gps []models.GitlabProject
	app.DB().Find(&gps)

	var m = map[int]models.GitlabProject{}
	for _, gp := range gps {
		m[gp.GitlabProjectId] = gp
	}

	var infos = make([]*gitlab.GitlabProjectInfo, 0)

	for _, project := range projects.GetItems() {
		var enabled, GlobalEnabled bool
		if gitlabProject, ok := m[int(project.GetID())]; ok {
			enabled = gitlabProject.Enabled
			GlobalEnabled = gitlabProject.GlobalEnabled
		}
		infos = append(infos, &gitlab.GitlabProjectInfo{
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

	return &gitlab.ProjectListResponse{Data: infos}, nil
}

const (
	OptionTypeProject string = "project"
	OptionTypeBranch  string = "branch"
	OptionTypeCommit  string = "commit"
)

func (g *Gitlab) Projects(ctx context.Context, empty *emptypb.Empty) (*gitlab.ProjectsResponse, error) {
	var (
		enabledProjects []models.GitlabProject
		ch              = make(chan *gitlab.Option)
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
			ch <- &gitlab.Option{
				Value:     fmt.Sprintf("%d", project.GitlabProjectId),
				Label:     project.Name,
				IsLeaf:    false,
				Type:      OptionTypeProject,
				ProjectId: strconv.Itoa(project.GitlabProjectId),
			}
		}(project)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	res := make([]*gitlab.Option, 0)

	for options := range ch {
		res = append(res, options)
	}

	return &gitlab.ProjectsResponse{Data: res}, nil
}

func (g *Gitlab) Branches(ctx context.Context, request *gitlab.BranchesRequest) (*gitlab.BranchesResponse, error) {
	branches, err := plugins.GetGitServer().AllBranches(request.ProjectId)
	if err != nil {
		return nil, err
	}

	res := make([]*gitlab.Option, 0, len(branches))
	for _, branch := range branches {
		res = append(res, &gitlab.Option{
			Value:     branch.GetName(),
			Label:     branch.GetName(),
			IsLeaf:    false,
			Type:      OptionTypeBranch,
			Branch:    branch.GetName(),
			ProjectId: request.ProjectId,
		})
	}
	if request.All {
		return &gitlab.BranchesResponse{Data: res}, nil
	}

	var defaultBranch string
	for _, branch := range branches {
		if branch.IsDefault() {
			defaultBranch = branch.GetName()
		}
	}

	config, err := GetProjectMarsConfig(request.ProjectId, defaultBranch)
	if err != nil {
		return &gitlab.BranchesResponse{Data: make([]*gitlab.Option, 0)}, nil
	}

	filteredRes := make([]*gitlab.Option, 0)
	for _, op := range res {
		if utils.BranchPass(config, op.Value) {
			filteredRes = append(filteredRes, op)
		}
	}

	return &gitlab.BranchesResponse{Data: filteredRes}, nil
}

func (g *Gitlab) Commits(ctx context.Context, request *gitlab.CommitsRequest) (*gitlab.CommitsResponse, error) {
	commits, err := plugins.GetGitServer().ListCommits(request.ProjectId, request.Branch)
	if err != nil {
		return nil, err
	}

	res := make([]*gitlab.Option, 0, len(commits))
	for _, commit := range commits {
		res = append(res, &gitlab.Option{
			Value:     commit.GetID(),
			IsLeaf:    true,
			Label:     fmt.Sprintf("[%s]: %s", utils.ToHumanizeDatetimeString(commit.GetCommittedDate()), commit.GetTitle()),
			Type:      OptionTypeCommit,
			ProjectId: request.ProjectId,
			Branch:    request.Branch,
		})
	}

	return &gitlab.CommitsResponse{Data: res}, nil
}

func (g *Gitlab) Commit(ctx context.Context, request *gitlab.CommitRequest) (*gitlab.CommitResponse, error) {
	commit, err := plugins.GetGitServer().GetCommit(request.ProjectId, request.Commit)
	if err != nil {
		return nil, err
	}

	return &gitlab.CommitResponse{
		Data: &gitlab.Option{
			Value:     commit.GetID(),
			IsLeaf:    true,
			Label:     fmt.Sprintf("[%s]: %s", utils.ToHumanizeDatetimeString(commit.GetCommittedDate()), commit.GetTitle()),
			Type:      OptionTypeCommit,
			ProjectId: request.ProjectId,
			Branch:    request.Branch,
		},
	}, nil
}

func (g *Gitlab) PipelineInfo(ctx context.Context, request *gitlab.PipelineInfoRequest) (*gitlab.PipelineInfoResponse, error) {
	commit, err := plugins.GetGitServer().GetCommit(request.ProjectId, request.Commit)
	if err != nil {
		return nil, err
	}
	if commit.GetLastPipeline() == nil {
		return nil, status.Errorf(codes.NotFound, "pipeline not found")
	}

	return &gitlab.PipelineInfoResponse{
		Status: commit.GetLastPipeline().GetStatus(),
		WebUrl: commit.GetLastPipeline().GetWebURL(),
	}, nil
}

func (g *Gitlab) ConfigFile(ctx context.Context, request *gitlab.ConfigFileRequest) (*gitlab.ConfigFileResponse, error) {
	marsC, err := GetProjectMarsConfig(request.ProjectId, request.Branch)
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
		return &gitlab.ConfigFileResponse{
			Data: marsC.ConfigFileValues,
			Type: ct,
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
		pid = fmt.Sprintf("%v", request.ProjectId)
		branch = request.Branch
		filename = configFile
	}

	content, err := plugins.GetGitServer().GetFileContentWithBranch(pid, branch, filename)
	if err != nil {
		mlog.Debug(err)
		return &gitlab.ConfigFileResponse{
			Data: marsC.ConfigFileValues,
			Type: marsC.ConfigFileType,
		}, nil
	}

	return &gitlab.ConfigFileResponse{
		Data: content,
		Type: marsC.ConfigFileType,
	}, nil
}
