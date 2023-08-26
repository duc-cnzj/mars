package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	gopath "path"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/duc-cnzj/mars-client/v4/gitconfig"
	"github.com/duc-cnzj/mars-client/v4/mars"
	"github.com/duc-cnzj/mars-client/v4/types"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
	"github.com/duc-cnzj/mars/v4/internal/utils"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		gitconfig.RegisterGitConfigServer(s, new(gitConfigSvc))
	})
	RegisterEndpoint(gitconfig.RegisterGitConfigHandlerFromEndpoint)
}

type gitConfigSvc struct {
	gitconfig.UnimplementedGitConfigServer
}

func (m *gitConfigSvc) GetDefaultChartValues(ctx context.Context, request *gitconfig.DefaultChartValuesRequest) (*gitconfig.DefaultChartValuesResponse, error) {
	marsC, err := GetProjectMarsConfig(fmt.Sprintf("%v", request.GitProjectId), request.Branch)
	if err != nil {
		return nil, err
	}
	var pid, branch, path string
	if marsC.LocalChartPath == "" {
		return &gitconfig.DefaultChartValuesResponse{Value: ""}, nil
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
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &gitconfig.DefaultChartValuesResponse{Value: f}, nil
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

func (m *gitConfigSvc) Show(ctx context.Context, request *gitconfig.ShowRequest) (*gitconfig.ShowResponse, error) {
	var branch string = request.Branch
	if branch == "" {
		branch, _ = getDefaultBranch(int(request.GitProjectId))
	}
	config, err := GetProjectMarsConfig(int(request.GitProjectId), branch)
	if err != nil {
		return &gitconfig.ShowResponse{
			Branch: branch,
			Config: &mars.Config{},
		}, nil
	}

	return &gitconfig.ShowResponse{
		Branch: branch,
		Config: config,
	}, nil
}

func (m *gitConfigSvc) GlobalConfig(ctx context.Context, request *gitconfig.GlobalConfigRequest) (*gitconfig.GlobalConfigResponse, error) {
	var project models.GitProject
	if app.DB().Where("`git_project_id` = ?", request.GitProjectId).First(&project).Error != nil {
		return &gitconfig.GlobalConfigResponse{
			Enabled: false,
			Config:  project.GlobalMarsConfig(),
		}, nil
	}

	return &gitconfig.GlobalConfigResponse{
		Enabled: project.GlobalEnabled,
		Config:  project.GlobalMarsConfig(),
	}, nil
}

func (m *gitConfigSvc) ToggleGlobalStatus(ctx context.Context, request *gitconfig.ToggleGlobalStatusRequest) (*gitconfig.ToggleGlobalStatusResponse, error) {
	var project models.GitProject
	if err := app.DB().Where("`git_project_id` = ?", request.GitProjectId).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			app.DB().Create(&models.GitProject{
				GitProjectId:  int(request.GitProjectId),
				Enabled:       false,
				GlobalEnabled: request.Enabled,
			})
		}
		return &gitconfig.ToggleGlobalStatusResponse{}, nil
	}
	app.DB().Model(&project).Update("global_enabled", request.Enabled)
	AuditLogWithChange(MustGetUser(ctx).Name, types.EventActionType_Update, fmt.Sprintf("打开/关闭 %s 的全局配置: %t", project.Name, request.Enabled), nil, nil)

	return &gitconfig.ToggleGlobalStatusResponse{}, nil
}

func (m *gitConfigSvc) Update(ctx context.Context, request *gitconfig.UpdateRequest) (*gitconfig.UpdateResponse, error) {
	var project models.GitProject
	if err := app.DB().Where("`git_project_id` = ?", request.GitProjectId).First(&project).Error; err != nil {
		return nil, err
	}

	if request.Config != nil && len(request.Config.Branches) == 0 {
		request.Config.Branches = []string{"*"}
	}
	if request.Config != nil && request.Config.DisplayName != project.GlobalMarsConfig().DisplayName {
		app.Cache().Clear(cache.NewKey(ProjectOptionsCacheKey))
	}
	if request.Config.ConfigField == "" {
		request.Config.IsSimpleEnv = true
	}
	request.Config.ConfigFileValues = strings.TrimRight(request.Config.ConfigFileValues, " ")
	request.Config.ValuesYaml = strings.TrimRight(request.Config.ValuesYaml, " ")
	marshal, _ := json.Marshal(request.Config)

	var oldConf models.GitProject = project

	app.DB().Model(&project).Update("global_config", string(marshal))

	AuditLogWithChange(MustGetUser(ctx).Name, types.EventActionType_Update,
		fmt.Sprintf("更新项目 %s (id: %d) 全局配置", project.Name, project.ID), oldConf, project)

	return &gitconfig.UpdateResponse{Config: project.GlobalMarsConfig()}, nil
}

func (m *gitConfigSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}
