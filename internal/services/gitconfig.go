package services

import (
	"context"
	"fmt"
	gopath "path"
	"strings"

	"github.com/duc-cnzj/mars/api/v4/gitconfig"
	"github.com/duc-cnzj/mars/api/v4/mars"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	mars2 "github.com/duc-cnzj/mars/v4/internal/util/mars"
	"github.com/spf13/cast"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ gitconfig.GitConfigServer = (*gitConfigSvc)(nil)

type gitConfigSvc struct {
	gitconfig.UnimplementedGitConfigServer

	cache       cache.Cache
	gitRepo     repo.GitRepo
	gitProjRepo repo.GitProjectRepo
	logger      mlog.Logger
	eventRepo   repo.EventRepo
}

func NewGitConfigSvc(eventRepo repo.EventRepo, cache cache.Cache, gitRepo repo.GitRepo, gitProjRepo repo.GitProjectRepo, logger mlog.Logger) gitconfig.GitConfigServer {
	return &gitConfigSvc{eventRepo: eventRepo, cache: cache, gitRepo: gitRepo, gitProjRepo: gitProjRepo, logger: logger}
}

func (svc *gitConfigSvc) GetDefaultChartValues(ctx context.Context, request *gitconfig.DefaultChartValuesRequest) (*gitconfig.DefaultChartValuesResponse, error) {
	marsC, err := GetProjectMarsConfig(fmt.Sprintf("%v", request.GitProjectId), request.Branch)
	if err != nil {
		return nil, err
	}
	var pid, branch, path string
	if marsC.LocalChartPath == "" {
		return &gitconfig.DefaultChartValuesResponse{Value: ""}, nil
	}

	if mars2.IsRemoteChart(marsC) {
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
	f, err := svc.gitRepo.GetFileContentWithBranch(ctx, cast.ToInt(pid), branch, filename)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &gitconfig.DefaultChartValuesResponse{Value: f}, nil
}

var GetProjectMarsConfig = mars2.GetProjectConfig

func (svc *gitConfigSvc) getDefaultBranch(projectId int) (string, error) {
	p, err := svc.gitRepo.GetProject(context.TODO(), projectId)
	if err != nil {
		svc.logger.Error(err)
		return "", err
	}
	return p.GetDefaultBranch(), nil
}

func (svc *gitConfigSvc) Show(ctx context.Context, request *gitconfig.ShowRequest) (*gitconfig.ShowResponse, error) {
	var branch string = request.Branch
	if branch == "" {
		branch, _ = svc.getDefaultBranch(int(request.GitProjectId))
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

func (svc *gitConfigSvc) GlobalConfig(ctx context.Context, request *gitconfig.GlobalConfigRequest) (*gitconfig.GlobalConfigResponse, error) {
	proj, err := svc.gitProjRepo.GetByProjectID(ctx, int(request.GitProjectId))
	if err != nil {
		return &gitconfig.GlobalConfigResponse{
			Enabled: false,
			Config:  &mars.Config{},
		}, nil
	}

	return &gitconfig.GlobalConfigResponse{
		Enabled: proj.GlobalEnabled,
		Config:  proj.GlobalMarsConfig(),
	}, nil
}

func (svc *gitConfigSvc) ToggleGlobalStatus(ctx context.Context, request *gitconfig.ToggleGlobalStatusRequest) (*gitconfig.ToggleGlobalStatusResponse, error) {
	project, err := svc.gitProjRepo.GetByProjectID(ctx, int(request.GitProjectId))
	if err != nil {
		return nil, err
	}
	_, err = svc.gitProjRepo.ToggleEnabled(ctx, int(request.GitProjectId))
	if err != nil {
		if ent.IsNotFound(err) {
			svc.gitProjRepo.Upsert(ctx, &repo.UpsertGitProjectInput{
				DefaultBranch: project.DefaultBranch,
				Name:          project.Name,
				GitProjectId:  int(request.GitProjectId),
				Enabled:       request.Enabled,
			})
		}
	}

	svc.eventRepo.AuditLog(
		types.EventActionType_Update,
		MustGetUser(ctx).Name,
		fmt.Sprintf("打开/关闭 %s 的全局配置: %t", project.Name, request.Enabled),
	)

	return &gitconfig.ToggleGlobalStatusResponse{}, nil
}

func (svc *gitConfigSvc) Update(ctx context.Context, request *gitconfig.UpdateRequest) (*gitconfig.UpdateResponse, error) {
	project, err := svc.gitProjRepo.GetByProjectID(ctx, int(request.GitProjectId))
	if err != nil {
		return nil, err
	}

	if request.Config != nil && len(request.Config.Branches) == 0 {
		request.Config.Branches = []string{"*"}
	}
	if request.Config != nil && request.Config.DisplayName != project.GlobalMarsConfig().DisplayName {
		svc.cache.Clear(cache.NewKey(ProjectOptionsCacheKey))
	}
	if request.Config.ConfigField == "" {
		request.Config.IsSimpleEnv = true
	}
	request.Config.ConfigFileValues = strings.TrimRight(request.Config.ConfigFileValues, " ")
	request.Config.ValuesYaml = strings.TrimRight(request.Config.ValuesYaml, " ")

	var oldConf *repo.GitProject = project
	svc.gitProjRepo.UpdateGlobalConfig(ctx, int(request.GitProjectId), request.Config)

	svc.eventRepo.AuditLogWithChange(types.EventActionType_Update, MustGetUser(ctx).Name, fmt.Sprintf("更新项目 %s (id: %d) 全局配置", project.Name, project.ID), oldConf, project)

	return &gitconfig.UpdateResponse{Config: project.GlobalMarsConfig()}, nil
}

func (svc *gitConfigSvc) Authorize(ctx context.Context, fullMethodName string) (context.Context, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}

	return ctx, nil
}
