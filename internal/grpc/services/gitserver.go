package services

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/duc-cnzj/mars-client/v3/event"
	"github.com/duc-cnzj/mars-client/v3/gitserver"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type GitServer struct {
	gitserver.UnimplementedGitServerServer
}

func (g *GitServer) EnableProject(ctx context.Context, request *gitserver.GitEnableProjectRequest) (*gitserver.GitEnableProjectResponse, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}
	project, _ := plugins.GetGitServer().GetProject(request.GitProjectId)

	var gp models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", request.GitProjectId).First(&gp).Error == nil {
		app.DB().Model(&gp).Updates(map[string]interface{}{
			"enabled":        true,
			"default_branch": project.GetDefaultBranch(),
			"name":           project.GetName(),
		})
	} else {
		atoi, _ := strconv.Atoi(request.GitProjectId)
		app.DB().Create(&models.GitlabProject{
			DefaultBranch:   project.GetDefaultBranch(),
			Name:            project.GetName(),
			GitlabProjectId: atoi,
			Enabled:         true,
		})
	}
	AuditLog(MustGetUser(ctx).Name, event.ActionType_Create, fmt.Sprintf("启用项目: %s", project.GetName()))

	return &gitserver.GitEnableProjectResponse{}, nil
}

func (g *GitServer) DisableProject(ctx context.Context, request *gitserver.GitDisableProjectRequest) (*gitserver.GitDisableProjectResponse, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}
	project, _ := plugins.GetGitServer().GetProject(request.GitProjectId)
	var gp models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", request.GitProjectId).First(&gp).Error == nil {
		app.DB().Model(&gp).Updates(map[string]interface{}{
			"enabled":        false,
			"default_branch": project.GetDefaultBranch(),
			"name":           project.GetName(),
		})
	} else {
		itoa, _ := strconv.Atoi(request.GitProjectId)
		app.DB().Create(&models.GitlabProject{
			DefaultBranch:   project.GetDefaultBranch(),
			Name:            project.GetName(),
			GitlabProjectId: itoa,
			Enabled:         false,
		})
	}
	AuditLog(MustGetUser(ctx).Name, event.ActionType_Create, fmt.Sprintf("关闭项目: %s", project.GetName()))

	return &gitserver.GitDisableProjectResponse{}, nil
}

func (g *GitServer) All(ctx context.Context, req *gitserver.GitAllProjectsRequest) (*gitserver.GitAllProjectsResponse, error) {
	do, err, _ := app.Singleflight().Do("GitServerAll", func() (interface{}, error) {
		mlog.Debug("sfGitServerAll...")
		return plugins.GetGitServer().AllProjects()
	})
	if err != nil {
		return nil, err
	}
	var projects = do.([]plugins.ProjectInterface)

	var gps []models.GitlabProject
	app.DB().Find(&gps)

	var m = map[int]models.GitlabProject{}
	for _, gp := range gps {
		m[gp.GitlabProjectId] = gp
	}

	var infos = make([]*gitserver.GitProjectItem, 0)

	for _, project := range projects {
		var enabled, GlobalEnabled bool
		if gitlabProject, ok := m[int(project.GetID())]; ok {
			enabled = gitlabProject.Enabled
			GlobalEnabled = gitlabProject.GlobalEnabled
		}
		infos = append(infos, &gitserver.GitProjectItem{
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

	return &gitserver.GitAllProjectsResponse{Data: infos}, nil
}

const (
	OptionTypeProject string = "project"
	OptionTypeBranch  string = "branch"
	OptionTypeCommit  string = "commit"
)

func (g *GitServer) ProjectOptions(ctx context.Context, request *gitserver.GitProjectOptionsRequest) (*gitserver.GitProjectOptionsResponse, error) {
	remember, err := app.Cache().Remember("ProjectOptions", 30, func() ([]byte, error) {
		var (
			enabledProjects []models.GitlabProject
			ch              = make(chan *gitserver.GitOption)
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
				ch <- &gitserver.GitOption{
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

		res := make([]*gitserver.GitOption, 0)

		for options := range ch {
			res = append(res, options)
		}

		return proto.Marshal(&gitserver.GitProjectOptionsResponse{Data: res})
	})
	if err != nil {
		return nil, err
	}
	var res = &gitserver.GitProjectOptionsResponse{}
	_ = proto.Unmarshal(remember, res)
	return res, nil
}

func (g *GitServer) BranchOptions(ctx context.Context, request *gitserver.GitBranchOptionsRequest) (*gitserver.GitBranchOptionsResponse, error) {
	remember, err := app.Cache().Remember(fmt.Sprintf("BranchOptions:%v-%v", request.ProjectId, request.All), 10, func() ([]byte, error) {
		branches, err := plugins.GetGitServer().AllBranches(request.ProjectId)
		if err != nil {
			return nil, err
		}

		res := make([]*gitserver.GitOption, 0, len(branches))
		for _, branch := range branches {
			res = append(res, &gitserver.GitOption{
				Value:     branch.GetName(),
				Label:     branch.GetName(),
				IsLeaf:    false,
				Type:      OptionTypeBranch,
				Branch:    branch.GetName(),
				ProjectId: request.ProjectId,
			})
		}
		if request.All {
			return proto.Marshal(&gitserver.GitBranchOptionsResponse{Data: res})
		}

		var defaultBranch string
		for _, branch := range branches {
			if branch.IsDefault() {
				defaultBranch = branch.GetName()
			}
		}

		config, err := GetProjectMarsConfig(request.ProjectId, defaultBranch)
		if err != nil {
			return proto.Marshal(&gitserver.GitBranchOptionsResponse{Data: make([]*gitserver.GitOption, 0)})
		}

		filteredRes := make([]*gitserver.GitOption, 0)
		for _, op := range res {
			if utils.BranchPass(config, op.Value) {
				filteredRes = append(filteredRes, op)
			}
		}

		return proto.Marshal(&gitserver.GitBranchOptionsResponse{Data: filteredRes})
	})
	if err != nil {
		return nil, err
	}
	res := &gitserver.GitBranchOptionsResponse{}
	_ = proto.Unmarshal(remember, res)
	return res, nil
}

func (g *GitServer) CommitOptions(ctx context.Context, request *gitserver.GitCommitOptionsRequest) (*gitserver.GitCommitOptionsResponse, error) {
	remember, err := app.Cache().Remember(fmt.Sprintf("CommitOptions:%s-%s", request.ProjectId, request.Branch), 3, func() ([]byte, error) {
		commits, err := plugins.GetGitServer().ListCommits(request.ProjectId, request.Branch)
		if err != nil {
			return nil, err
		}

		res := make([]*gitserver.GitOption, 0, len(commits))
		for _, commit := range commits {
			res = append(res, &gitserver.GitOption{
				Value:     commit.GetID(),
				IsLeaf:    true,
				Label:     fmt.Sprintf("[%s]: %s", utils.ToHumanizeDatetimeString(commit.GetCommittedDate()), commit.GetTitle()),
				Type:      OptionTypeCommit,
				ProjectId: request.ProjectId,
				Branch:    request.Branch,
			})
		}

		return proto.Marshal(&gitserver.GitCommitOptionsResponse{Data: res})
	})
	if err != nil {
		return nil, err
	}
	res := &gitserver.GitCommitOptionsResponse{}
	_ = proto.Unmarshal(remember, res)
	return res, nil
}

func (g *GitServer) Commit(ctx context.Context, request *gitserver.GitCommitRequest) (*gitserver.GitCommitResponse, error) {
	remember, err := app.Cache().Remember(fmt.Sprintf("Commit:%s-%s", request.ProjectId, request.Commit), 60*60, func() ([]byte, error) {
		commit, err := plugins.GetGitServer().GetCommit(request.ProjectId, request.Commit)
		if err != nil {
			return nil, err
		}
		res := &gitserver.GitCommitResponse{
			Data: &gitserver.GitOption{
				Value:     commit.GetID(),
				IsLeaf:    true,
				Label:     fmt.Sprintf("[%s]: %s", utils.ToHumanizeDatetimeString(commit.GetCommittedDate()), commit.GetTitle()),
				Type:      OptionTypeCommit,
				ProjectId: request.ProjectId,
				Branch:    request.Branch,
			},
		}
		return proto.Marshal(res)
	})
	if err != nil {
		return nil, err
	}
	msg := &gitserver.GitCommitResponse{}
	_ = proto.Unmarshal(remember, msg)
	return msg, nil
}

func (g *GitServer) PipelineInfo(ctx context.Context, request *gitserver.GitPipelineInfoRequest) (*gitserver.GitPipelineInfoResponse, error) {
	pipeline, err := plugins.GetGitServer().GetCommitPipeline(request.ProjectId, request.Commit)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &gitserver.GitPipelineInfoResponse{
		Status: pipeline.GetStatus(),
		WebUrl: pipeline.GetWebURL(),
	}, nil
}

func (g *GitServer) MarsConfigFile(ctx context.Context, request *gitserver.GitConfigFileRequest) (*gitserver.GitConfigFileResponse, error) {
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
		return &gitserver.GitConfigFileResponse{
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
		pid = fmt.Sprintf("%v", request.ProjectId)
		branch = request.Branch
		filename = configFile
	}

	content, err := plugins.GetGitServer().GetFileContentWithBranch(pid, branch, filename)
	if err != nil {
		mlog.Debug(err)
		return &gitserver.GitConfigFileResponse{
			Data:     marsC.ConfigFileValues,
			Type:     marsC.ConfigFileType,
			Elements: marsC.Elements,
		}, nil
	}

	return &gitserver.GitConfigFileResponse{
		Data:     content,
		Type:     marsC.ConfigFileType,
		Elements: marsC.Elements,
	}, nil
}
