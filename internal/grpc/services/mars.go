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

	"github.com/duc-cnzj/mars/client/event"
	"github.com/duc-cnzj/mars/client/mars"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	"gorm.io/gorm"
)

type Mars struct {
	mars.UnimplementedMarsServer
}

func (m *Mars) GetDefaultChartValues(ctx context.Context, request *mars.MarsDefaultChartValuesRequest) (*mars.MarsDefaultChartValuesResponse, error) {
	marsC, err := GetProjectMarsConfig(fmt.Sprintf("%v", request.ProjectId), request.Branch)
	if err != nil {
		return nil, err
	}
	var pid, branch, path string
	if marsC.LocalChartPath == "" {
		return &mars.MarsDefaultChartValuesResponse{Value: ""}, nil
	}

	if utils.IsRemoteChart(marsC) {
		split := strings.Split(marsC.LocalChartPath, "|")
		pid = split[0]
		branch = split[1]
		path = split[2]
	} else {
		pid = fmt.Sprintf("%v", request.ProjectId)
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

	return &mars.MarsDefaultChartValuesResponse{Value: f}, nil
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

func (m *Mars) Show(ctx context.Context, request *mars.MarsShowRequest) (*mars.MarsShowResponse, error) {
	var branch string = request.Branch
	if branch == "" {
		branch, _ = getDefaultBranch(int(request.ProjectId))
	}
	config, err := GetProjectMarsConfig(int(request.ProjectId), branch)
	if err != nil {
		return &mars.MarsShowResponse{
			Branch: branch,
			Config: &mars.MarsConfig{},
		}, nil
	}

	return &mars.MarsShowResponse{
		Branch: branch,
		Config: config,
	}, nil
}

func (m *Mars) GlobalConfig(ctx context.Context, request *mars.MarsGlobalConfigRequest) (*mars.MarsGlobalConfigResponse, error) {
	var project models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", request.ProjectId).First(&project).Error != nil {
		return &mars.MarsGlobalConfigResponse{
			Enabled: false,
			Config:  project.GlobalMarsConfig(),
		}, nil
	}

	return &mars.MarsGlobalConfigResponse{
		Enabled: project.GlobalEnabled,
		Config:  project.GlobalMarsConfig(),
	}, nil
}

func (m *Mars) ToggleEnabled(ctx context.Context, request *mars.MarsToggleEnabledRequest) (*mars.MarsToggleEnabledResponse, error) {
	var project models.GitlabProject
	if err := app.DB().Where("`gitlab_project_id` = ?", request.ProjectId).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			app.DB().Create(&models.GitlabProject{
				GitlabProjectId: int(request.ProjectId),
				Enabled:         false,
				GlobalEnabled:   request.Enabled,
			})
		}
		return &mars.MarsToggleEnabledResponse{}, nil
	}
	app.DB().Model(&project).UpdateColumn("global_enabled", request.Enabled)
	AuditLogWithChange(MustGetUser(ctx).Name, event.ActionType_Update, fmt.Sprintf("打开/关闭 %s 的全局配置: %t", project.Name, request.Enabled), nil, nil)

	return &mars.MarsToggleEnabledResponse{}, nil
}

func (m *Mars) Update(ctx context.Context, request *mars.MarsUpdateRequest) (*mars.MarsUpdateResponse, error) {
	var project models.GitlabProject
	if err := app.DB().Where("`gitlab_project_id` = ?", request.ProjectId).First(&project).Error; err != nil {
		return nil, err
	}

	if request.Config != nil && len(request.Config.Branches) == 0 {
		request.Config.Branches = []string{"*"}
	}
	marshal, err := json.Marshal(request.Config)
	if err != nil {
		return nil, err
	}

	var oldConf models.GitlabProject = project

	app.DB().Model(&project).UpdateColumn("global_config", string(marshal))

	AuditLogWithChange(MustGetUser(ctx).Name, event.ActionType_Update,
		fmt.Sprintf("更新项目 %s (id: %d) 全局配置", project.Name, project.ID), oldConf, project)

	return &mars.MarsUpdateResponse{Config: project.GlobalMarsConfig()}, nil
}

func (m *Mars) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}
