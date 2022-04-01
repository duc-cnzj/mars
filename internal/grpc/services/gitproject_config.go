package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	gopath "path"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars-client/v4/event"
	"github.com/duc-cnzj/mars-client/v4/gitproject"
	"github.com/duc-cnzj/mars-client/v4/mars"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
)

type GitProjectConfigSvc struct {
	gitproject.UnimplementedGitProjectConfigServer
}

func (m *GitProjectConfigSvc) GetDefaultChartValues(ctx context.Context, request *gitproject.GitProjectConfigDefaultChartValuesRequest) (*gitproject.GitProjectConfigDefaultChartValuesResponse, error) {
	marsC, err := GetProjectMarsConfig(fmt.Sprintf("%v", request.GitProjectId), request.Branch)
	if err != nil {
		return nil, err
	}
	var pid, branch, path string
	if marsC.LocalChartPath == "" {
		return &gitproject.GitProjectConfigDefaultChartValuesResponse{Value: ""}, nil
	}

	if utils.IsRemoteChart(marsC) {
		split := strings.Split(marsC.LocalChartPath, "|")
		pid = split[0]
		branch = split[1]
		path = split[2]
	} else {
		pid = fmt.Sprintf("%v", request.GitProjectId)
		branch = request.Branch
		path = marsC.LocalChartPath
	}

	filename := gopath.Join(path, "values.yaml")
	if branch == "" {
		branch = "master"
	}
	f, err := plugins.GetGitServer().GetFileContentWithBranch(pid, branch, filename)
	if err != nil {
		return nil, err
	}

	return &gitproject.GitProjectConfigDefaultChartValuesResponse{Value: f}, nil
}

var GetProjectMarsConfig = utils.GetProjectMarsConfig

func getDefaultBranch(projectId int) (string, error) {
	p, err := plugins.GetGitServer().GetProject(fmt.Sprintf("%d", projectId))
	if err != nil {
		mlog.Error(err)
		return "", err
	}
	return p.GetDefaultBranch(), nil
}

func (m *GitProjectConfigSvc) Show(ctx context.Context, request *gitproject.GitProjectConfigShowRequest) (*gitproject.GitProjectConfigShowResponse, error) {
	var branch string = request.Branch
	if branch == "" {
		branch, _ = getDefaultBranch(int(request.GitProjectId))
	}
	config, err := GetProjectMarsConfig(int(request.GitProjectId), branch)
	if err != nil {
		return &gitproject.GitProjectConfigShowResponse{
			Branch: branch,
			Config: &mars.MarsConfig{},
		}, nil
	}

	return &gitproject.GitProjectConfigShowResponse{
		Branch: branch,
		Config: config,
	}, nil
}

func (m *GitProjectConfigSvc) GlobalConfig(ctx context.Context, request *gitproject.GitProjectConfigGlobalConfigRequest) (*gitproject.GitProjectConfigGlobalConfigResponse, error) {
	var project models.GitProject
	if app.DB().Where("`git_project_id` = ?", request.GitProjectId).First(&project).Error != nil {
		return &gitproject.GitProjectConfigGlobalConfigResponse{
			Enabled: false,
			Config:  project.GlobalMarsConfig(),
		}, nil
	}

	return &gitproject.GitProjectConfigGlobalConfigResponse{
		Enabled: project.GlobalEnabled,
		Config:  project.GlobalMarsConfig(),
	}, nil
}

func (m *GitProjectConfigSvc) ToggleGlobalStatus(ctx context.Context, request *gitproject.GitProjectConfigToggleGlobalStatusRequest) (*gitproject.GitProjectConfigToggleGlobalStatusResponse, error) {
	var project models.GitProject
	if err := app.DB().Where("`git_project_id` = ?", request.GitProjectId).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			app.DB().Create(&models.GitProject{
				GitProjectId:  int(request.GitProjectId),
				Enabled:       false,
				GlobalEnabled: request.Enabled,
			})
		}
		return &gitproject.GitProjectConfigToggleGlobalStatusResponse{}, nil
	}
	app.DB().Model(&project).UpdateColumn("global_enabled", request.Enabled)
	AuditLogWithChange(MustGetUser(ctx).Name, event.ActionType_Update, fmt.Sprintf("打开/关闭 %s 的全局配置: %t", project.Name, request.Enabled), nil, nil)

	return &gitproject.GitProjectConfigToggleGlobalStatusResponse{}, nil
}

func (m *GitProjectConfigSvc) Update(ctx context.Context, request *gitproject.GitProjectConfigUpdateRequest) (*gitproject.GitProjectConfigUpdateResponse, error) {
	var project models.GitProject
	if err := app.DB().Where("`git_project_id` = ?", request.GitProjectId).First(&project).Error; err != nil {
		return nil, err
	}

	if request.Config != nil && len(request.Config.Branches) == 0 {
		request.Config.Branches = []string{"*"}
	}
	marshal, err := json.Marshal(request.Config)
	if err != nil {
		return nil, err
	}

	var oldConf models.GitProject = project

	app.DB().Model(&project).UpdateColumn("global_config", string(marshal))

	AuditLogWithChange(MustGetUser(ctx).Name, event.ActionType_Update,
		fmt.Sprintf("更新项目 %s (id: %d) 全局配置", project.Name, project.ID), oldConf, project)

	return &gitproject.GitProjectConfigUpdateResponse{Config: project.GlobalMarsConfig()}, nil
}

func (m *GitProjectConfigSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}
