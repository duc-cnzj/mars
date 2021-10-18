package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/gitlab"
	gitlab2 "github.com/xanzy/go-gitlab"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Gitlab struct {
	gitlab.UnimplementedGitlabServer
}

func (g Gitlab) EnableProject(ctx context.Context, request *gitlab.EnableProjectRequest) (*emptypb.Empty, error) {
	project, _, _ := app.GitlabClient().Projects.GetProject(request.GitlabProjectId, &gitlab2.GetProjectOptions{})

	var gp models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", request.GitlabProjectId).First(&gp).Error == nil {
		app.DB().Model(&gp).Updates(map[string]interface{}{
			"enabled":        true,
			"default_branch": project.DefaultBranch,
			"name":           project.Name,
		})
	} else {
		atoi, _ := strconv.Atoi(request.GitlabProjectId)
		app.DB().Create(&models.GitlabProject{
			DefaultBranch:   project.DefaultBranch,
			Name:            project.Name,
			GitlabProjectId: atoi,
			Enabled:         true,
		})
	}

	return &emptypb.Empty{}, nil
}

func (g Gitlab) DisableProject(ctx context.Context, request *gitlab.DisableProjectRequest) (*emptypb.Empty, error) {
	project, _, _ := app.GitlabClient().Projects.GetProject(request.GitlabProjectId, &gitlab2.GetProjectOptions{})
	var gp models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", request.GitlabProjectId).First(&gp).Error == nil {
		app.DB().Model(&gp).Updates(map[string]interface{}{
			"enabled":        false,
			"default_branch": project.DefaultBranch,
			"name":           project.Name,
		})
	} else {
		itoa, _ := strconv.Atoi(request.GitlabProjectId)
		app.DB().Create(&models.GitlabProject{
			DefaultBranch:   project.DefaultBranch,
			Name:            project.Name,
			GitlabProjectId: itoa,
			Enabled:         false,
		})
	}

	return &emptypb.Empty{}, nil
}

func (g Gitlab) ProjectList(ctx context.Context, empty *emptypb.Empty) (*gitlab.ProjectListResponse, error) {
	projects, _, err := app.GitlabClient().Projects.ListProjects(&gitlab2.ListProjectsOptions{
		MinAccessLevel: gitlab2.AccessLevel(gitlab2.DeveloperPermissions),
		ListOptions: gitlab2.ListOptions{
			PerPage: 100,
		},
	})
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

	for _, project := range projects {
		var enabled, GlobalEnabled bool
		if gitlabProject, ok := m[project.ID]; ok {
			enabled = gitlabProject.Enabled
			GlobalEnabled = gitlabProject.GlobalEnabled
		}
		infos = append(infos, &gitlab.GitlabProjectInfo{
			Id:            int64(project.ID),
			Name:          project.Name,
			Path:          project.Path,
			WebUrl:        project.WebURL,
			AvatarUrl:     project.AvatarURL,
			Description:   project.Description,
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

func (g Gitlab) Projects(ctx context.Context, empty *emptypb.Empty) (*gitlab.ProjectsResponse, error) {
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

func (g Gitlab) Branches(ctx context.Context, request *gitlab.BranchesRequest) (*gitlab.BranchesResponse, error) {
	branches, err := utils.GetAllBranches(request.ProjectId)
	if err != nil {
		return nil, err
	}

	var defaultBranch string
	for _, branch := range branches {
		if branch.Default {
			defaultBranch = branch.Name
		}
	}

	config, err := GetProjectMarsConfig(request.ProjectId, defaultBranch)
	if err != nil {
		return &gitlab.BranchesResponse{Data: make([]*gitlab.Option, 0)}, nil
	}

	res := make([]*gitlab.Option, 0, len(branches))
	for _, branch := range branches {
		if config.BranchPass(branch.Name) {
			res = append(res, &gitlab.Option{
				Value:     branch.Name,
				Label:     branch.Name,
				IsLeaf:    false,
				Type:      OptionTypeBranch,
				Branch:    branch.Name,
				ProjectId: request.ProjectId,
			})
		}
	}

	return &gitlab.BranchesResponse{Data: res}, nil
}

func (g Gitlab) Commits(ctx context.Context, request *gitlab.CommitsRequest) (*gitlab.CommitsResponse, error) {
	commits, _, err := app.GitlabClient().Commits.ListCommits(request.ProjectId, &gitlab2.ListCommitsOptions{RefName: gitlab2.String(request.Branch), ListOptions: gitlab2.ListOptions{PerPage: 100}})
	if err != nil {
		return nil, err
	}

	res := make([]*gitlab.Option, 0, len(commits))
	for _, commit := range commits {
		res = append(res, &gitlab.Option{
			Value:     commit.ID,
			IsLeaf:    true,
			Label:     fmt.Sprintf("[%s]: %s", utils.ToHumanizeDatetimeString(commit.CommittedDate), commit.Title),
			Type:      OptionTypeCommit,
			ProjectId: request.ProjectId,
			Branch:    request.Branch,
		})
	}

	return &gitlab.CommitsResponse{Data: res}, nil
}

func (g Gitlab) Commit(ctx context.Context, request *gitlab.CommitRequest) (*gitlab.CommitResponse, error) {
	commit, _, err := app.GitlabClient().Commits.GetCommit(request.ProjectId, request.Commit)
	if err != nil {
		return nil, err
	}

	return &gitlab.CommitResponse{
		Data: &gitlab.Option{
			Value:     commit.ID,
			IsLeaf:    true,
			Label:     fmt.Sprintf("[%s]: %s", utils.ToHumanizeDatetimeString(commit.CommittedDate), commit.Title),
			Type:      OptionTypeCommit,
			ProjectId: request.ProjectId,
			Branch:    request.Branch,
		},
	}, nil
}

func (g Gitlab) PipelineInfo(ctx context.Context, request *gitlab.PipelineInfoRequest) (*gitlab.PipelineInfoResponse, error) {
	commit, _, err := app.GitlabClient().Commits.GetCommit(request.ProjectId, request.Commit)
	if err != nil {
		return nil, err
	}
	if commit.LastPipeline == nil {
		return nil, status.Errorf(codes.NotFound, "pipeline not found")
	}

	return &gitlab.PipelineInfoResponse{
		Status: commit.LastPipeline.Status,
		WebUrl: commit.LastPipeline.WebURL,
	}, nil
}

func (g Gitlab) ConfigFile(ctx context.Context, request *gitlab.ConfigFileRequest) (*gitlab.ConfigFileResponse, error) {
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

	if marsC.IsRemoteConfigFile() {
		split := strings.Split(configFile, "|")
		pid = split[0]
		branch = split[1]
		filename = split[2]
	} else {
		pid = fmt.Sprintf("%d", request.ProjectId)
		branch = request.Branch
		filename = configFile
	}

	f, _, err := app.GitlabClient().RepositoryFiles.GetFile(pid, filename, &gitlab2.GetFileOptions{Ref: gitlab2.String(branch)})
	if err != nil {
		mlog.Debug(err)
		return &gitlab.ConfigFileResponse{
			Data: marsC.ConfigFileValues,
			Type: marsC.ConfigFileType,
		}, nil
	}
	fdata, _ := base64.StdEncoding.DecodeString(f.Content)
	return &gitlab.ConfigFileResponse{
		Data: string(fdata),
		Type: marsC.ConfigFileType,
	}, nil
}
