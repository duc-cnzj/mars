package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	gopath "path"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	config "github.com/duc-cnzj/mars/internal/mars"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/pkg/mars"
	"github.com/duc-cnzj/mars/pkg/modal"
	"github.com/xanzy/go-gitlab"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

type Mars struct {
	mars.UnimplementedMarsServer
}

func (m *Mars) GetDefaultChartValues(ctx context.Context, request *mars.DefaultChartValuesRequest) (*mars.DefaultChartValues, error) {
	marsC, err := GetProjectMarsConfig(fmt.Sprintf("%v", request.ProjectId), request.Branch)
	if err != nil {
		return nil, err
	}
	var pid, branch, path string
	if marsC.LocalChartPath == "" {
		return &mars.DefaultChartValues{Value: ""}, nil
	}

	if marsC.IsRemoteChart() {
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
	f, _, err := app.GitlabClient().RepositoryFiles.GetFile(pid, filename, &gitlab.GetFileOptions{Ref: gitlab.String(branch)})
	if err != nil {
		return nil, err
	}
	fdata, _ := base64.StdEncoding.DecodeString(f.Content)

	return &mars.DefaultChartValues{Value: string(fdata)}, nil
}

func GetProjectMarsConfig(projectId interface{}, branch string) (*config.Config, error) {
	var marsC config.Config

	var gp models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", projectId).First(&gp).Error == nil {
		if gp.GlobalEnabled {
			return gp.GlobalMarsConfig(), nil
		}
	}

	// 获取 .mars.yaml
	opt := &gitlab.GetFileOptions{}
	if branch != "" {
		opt.Ref = gitlab.String(branch)
	}
	file, _, err := app.GitlabClient().RepositoryFiles.GetFile(projectId, ".mars.yaml", opt)
	if err != nil {
		return nil, err
	}
	data, _ := base64.StdEncoding.DecodeString(file.Content)
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&marsC); err != nil {
		return nil, err
	}

	return &marsC, nil
}

func getDefaultBranch(projectId int) (string, error) {
	p, _, err := app.GitlabClient().Projects.GetProject(projectId, &gitlab.GetProjectOptions{})
	if err != nil {
		mlog.Error(err)
		return "", err
	}
	return p.DefaultBranch, nil
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
			Config: "",
		}, nil
	}
	bf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(bf)
	if err := encoder.Encode(config); err != nil {
		return nil, err
	}

	return &mars.MarsShowResponse{
		Branch: branch,
		Config: bf.String(),
	}, nil
}

func (m *Mars) GlobalConfig(ctx context.Context, request *mars.GlobalConfigRequest) (*mars.GlobalConfigResponse, error) {
	var project models.GitlabProject
	if app.DB().Where("`gitlab_project_id` = ?", request.ProjectId).First(&project).Error != nil {
		return &mars.GlobalConfigResponse{
			Enabled: false,
			Config:  project.GlobalConfigString(),
		}, nil
	}

	return &mars.GlobalConfigResponse{
		Enabled: project.GlobalEnabled,
		Config:  project.GlobalConfigString(),
	}, nil
}

func (m *Mars) ToggleEnabled(ctx context.Context, request *mars.ToggleEnabledRequest) (*emptypb.Empty, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}
	var project models.GitlabProject
	if err := app.DB().Where("`gitlab_project_id` = ?", request.ProjectId).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			app.DB().Create(&models.GitlabProject{
				GitlabProjectId: int(request.ProjectId),
				Enabled:         false,
				GlobalEnabled:   request.Enabled,
			})
		}
		return &emptypb.Empty{}, nil
	}
	app.DB().Model(&project).UpdateColumn("global_enabled", request.Enabled)

	return &emptypb.Empty{}, nil
}

func (m *Mars) Update(ctx context.Context, request *mars.MarsUpdateRequest) (*mars.MarsUpdateResponse, error) {
	if !MustGetUser(ctx).IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, ErrorPermissionDenied.Error())
	}
	var project models.GitlabProject
	if err := app.DB().Where("`gitlab_project_id` = ?", request.ProjectId).First(&project).Error; err != nil {
		return nil, err
	}

	mc := config.Config{}
	if err := yaml.Unmarshal([]byte(request.Config), &mc); err != nil {
		return nil, err
	}

	app.DB().Model(&project).UpdateColumn("global_config", request.Config)
	return &mars.MarsUpdateResponse{
		Data: &modal.GitlabProjectModal{
			Id:              int64(project.ID),
			DefaultBranch:   project.DefaultBranch,
			Name:            project.Name,
			GitlabProjectId: int64(project.GitlabProjectId),
			Enabled:         project.Enabled,
			GlobalEnabled:   project.GlobalEnabled,
			GlobalConfig:    project.GlobalConfig,
			CreatedAt:       timestamppb.New(project.CreatedAt),
			UpdatedAt:       timestamppb.New(project.UpdatedAt),
			DeletedAt:       timestamppb.New(project.DeletedAt.Time),
		},
	}, nil
}
