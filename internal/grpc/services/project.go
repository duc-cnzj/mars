package services

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/duc-cnzj/mars-client/v4/endpoint"

	"github.com/duc-cnzj/mars-client/v4/event"
	"github.com/duc-cnzj/mars-client/v4/model"
	"github.com/duc-cnzj/mars-client/v4/project"
	"github.com/duc-cnzj/mars-client/v4/websocket"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/scopes"
	"github.com/duc-cnzj/mars/internal/socket"
	"github.com/duc-cnzj/mars/internal/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ProjectSvc struct {
	project.UnimplementedProjectServer
}

func (p *ProjectSvc) List(ctx context.Context, request *project.ProjectListRequest) (*project.ProjectListResponse, error) {
	var (
		page     = int(request.Page)
		pageSize = int(request.PageSize)
		projects []models.Project
		count    int64
	)
	if err := app.DB().Preload("Namespace").Scopes(scopes.Paginate(&page, &pageSize)).Order("`id` DESC").Find(&projects).Error; err != nil {
		return nil, err
	}
	app.DB().Model(&models.Project{}).Count(&count)
	res := make([]*model.ProjectModel, 0, len(projects))
	for _, p := range projects {
		var ns *model.NamespaceModel
		if p.Namespace.ID != 0 {
			ns = &model.NamespaceModel{
				Id:               int64(p.Namespace.ID),
				Name:             p.Namespace.Name,
				ImagePullSecrets: p.Namespace.ImagePullSecretsArray(),
				CreatedAt:        utils.ToRFC3339DatetimeString(&p.Namespace.CreatedAt),
				UpdatedAt:        utils.ToRFC3339DatetimeString(&p.Namespace.UpdatedAt),
			}
		}
		res = append(res, &model.ProjectModel{
			Id:             int64(p.ID),
			Name:           p.Name,
			GitProjectId:   int64(p.GitProjectId),
			GitBranch:      p.GitBranch,
			GitCommit:      p.GitCommit,
			Config:         p.Config,
			OverrideValues: p.OverrideValues,
			DockerImage:    p.DockerImage,
			PodSelectors:   p.PodSelectors,
			NamespaceId:    int64(p.NamespaceId),
			Atomic:         p.Atomic,
			CreatedAt:      utils.ToRFC3339DatetimeString(&p.CreatedAt),
			UpdatedAt:      utils.ToRFC3339DatetimeString(&p.UpdatedAt),
			Namespace:      ns,
		})
	}

	return &project.ProjectListResponse{
		Page:     request.Page,
		PageSize: request.PageSize,
		Count:    count,
		Items:    res,
	}, nil
}

func (p *ProjectSvc) ApplyDryRun(ctx context.Context, input *project.ProjectApplyRequest) (*project.ProjectDryRunApplyResponse, error) {
	var pubsub plugins.PubSub = &plugins.EmptyPubSub{}
	t := websocket.Type_ApplyProject
	if input.GitCommit == "" {
		commits, _ := plugins.GetGitServer().ListCommits(fmt.Sprintf("%d", input.GitProjectId), input.GitBranch)
		if len(commits) < 1 {
			return nil, errors.New("没有可用的 commit")
		}
		lastCommit := commits[0]
		input.GitCommit = lastCommit.GetID()
	}
	mlog.Debug("ApplyDryRun..")
	user := MustGetUser(ctx)
	errMsger := newErrorMessager()
	job := socket.NewJober(&websocket.ProjectInput{
		Type:         t,
		NamespaceId:  input.NamespaceId,
		Name:         input.Name,
		GitProjectId: input.GitProjectId,
		GitBranch:    input.GitBranch,
		GitCommit:    input.GitCommit,
		Config:       input.Config,
		Atomic:       input.Atomic,
		ExtraValues:  input.ExtraValues,
	}, *user, "", errMsger, pubsub, input.InstallTimeoutSeconds, socket.WithDryRun())

	ch := make(chan struct{}, 1)
	go func() {
		select {
		case <-ctx.Done():
			job.Stop(ctx.Err())
		case <-ch:
		}
	}()
	socket.InstallProject(job)
	ch <- struct{}{}

	if errMsger.HasErrors() {
		return nil, errMsger
	}

	return &project.ProjectDryRunApplyResponse{Results: job.Manifests()}, nil
}

func (p *ProjectSvc) Apply(input *project.ProjectApplyRequest, server project.Project_ApplyServer) error {
	var pubsub plugins.PubSub = &plugins.EmptyPubSub{}
	if input.WebsocketSync {
		pubsub = plugins.GetWsSender().New("", "")
	}
	t := websocket.Type_ApplyProject
	msger := &messager{
		slugName: utils.GetSlugName(input.NamespaceId, input.Name),
		t:        t,
		server:   server,
	}
	if input.GitCommit == "" {
		commits, _ := plugins.GetGitServer().ListCommits(fmt.Sprintf("%d", input.GitProjectId), input.GitBranch)
		if len(commits) < 1 {
			return errors.New("没有可用的 commit")
		}
		lastCommit := commits[0]
		input.GitCommit = lastCommit.GetID()
		msger.SendMsg(fmt.Sprintf("未传入commit，使用最新的commit [%s](%s)", lastCommit.GetTitle(), lastCommit.GetWebURL()))
	}
	user := MustGetUser(server.Context())
	ch := make(chan struct{}, 1)

	job := socket.NewJober(&websocket.ProjectInput{
		Type:         t,
		NamespaceId:  input.NamespaceId,
		Name:         input.Name,
		GitProjectId: input.GitProjectId,
		GitBranch:    input.GitBranch,
		GitCommit:    input.GitCommit,
		Config:       input.Config,
		Atomic:       input.Atomic,
		ExtraValues:  input.ExtraValues,
	}, *user, "", msger, pubsub, input.InstallTimeoutSeconds)

	go func() {
		select {
		case <-server.Context().Done():
			job.Stop(server.Context().Err())
		case <-ch:
			return
		}
	}()
	socket.InstallProject(job)
	ch <- struct{}{}

	return nil
}

func (p *ProjectSvc) Delete(ctx context.Context, request *project.ProjectDeleteRequest) (*project.ProjectDeleteResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		return nil, err
	}
	if err := utils.UninstallRelease(projectModel.Name, projectModel.Namespace.Name, mlog.Debugf); err != nil {
		mlog.Error(err)
	}
	app.DB().Delete(&projectModel)
	app.Event().Dispatch(events.EventProjectDeleted, map[string]any{"data": &projectModel})

	AuditLog(MustGetUser(ctx).Name, event.ActionType_Delete,
		fmt.Sprintf("删除项目: %d: %s/%s ", projectModel.ID, projectModel.Namespace.Name, projectModel.Name))

	return &project.ProjectDeleteResponse{}, nil
}

func (p *ProjectSvc) Show(ctx context.Context, request *project.ProjectShowRequest) (*project.ProjectShowResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	marsC, _ := GetProjectMarsConfig(projectModel.GitProjectId, projectModel.GitBranch)
	cpu, memory := utils.GetCpuAndMemory(projectModel.GetAllPodMetrics())
	commit, err := plugins.GetGitServer().GetCommit(fmt.Sprintf("%d", projectModel.GitProjectId), projectModel.GitCommit)
	if err != nil {
		mlog.Error(err)
		return nil, err
	}

	nodePortMapping := utils.GetNodePortMappingByNamespace(projectModel.Namespace.Name)
	ingMapping := utils.GetIngressMappingByNamespace(projectModel.Namespace.Name)

	var urls = make([]*endpoint.ServiceEndpoint, 0)
	for key, values := range ingMapping {
		if projectModel.Name == key {
			urls = append(urls, values...)
		}
	}
	for key, values := range nodePortMapping {
		if projectModel.Name == key {
			urls = append(urls, values...)
		}
	}

	return &project.ProjectShowResponse{
		Id:              int64(projectModel.ID),
		Name:            projectModel.Name,
		GitProjectId:    int64(projectModel.GitProjectId),
		GitBranch:       projectModel.GitBranch,
		GitCommit:       projectModel.GitCommit,
		Config:          projectModel.Config,
		DockerImage:     projectModel.DockerImage,
		Atomic:          projectModel.Atomic,
		GitCommitWebUrl: commit.GetWebURL(),
		GitCommitTitle:  commit.GetTitle(),
		GitCommitAuthor: commit.GetAuthorName(),
		GitCommitDate:   utils.ToHumanizeDatetimeString(commit.GetCreatedAt()),
		Urls:            urls,
		Namespace: &project.ProjectShowResponse_Namespace{
			Id:   int64(projectModel.NamespaceId),
			Name: projectModel.Namespace.Name,
		},
		Cpu:               cpu,
		Memory:            memory,
		OverrideValues:    projectModel.OverrideValues,
		CreatedAt:         utils.ToRFC3339DatetimeString(&projectModel.CreatedAt),
		UpdatedAt:         utils.ToRFC3339DatetimeString(&projectModel.UpdatedAt),
		HumanizeCreatedAt: utils.ToHumanizeDatetimeString(&projectModel.CreatedAt),
		HumanizeUpdatedAt: utils.ToHumanizeDatetimeString(&projectModel.CreatedAt),
		ExtraValues:       projectModel.GetExtraValues(),
		Elements:          marsC.GetElements(),
		ConfigType:        marsC.GetConfigFileType(),
	}, nil
}

func (p *ProjectSvc) AllContainers(ctx context.Context, request *project.ProjectAllContainersRequest) (*project.ProjectAllContainersResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		return nil, err
	}

	var list = projectModel.GetAllPods()

	var containerList []*project.ProjectPod
	for _, item := range list {
		for _, c := range item.Spec.Containers {
			containerList = append(containerList,
				&project.ProjectPod{
					Namespace:     projectModel.Namespace.Name,
					PodName:       item.Name,
					ContainerName: c.Name,
				},
			)
		}
	}

	return &project.ProjectAllContainersResponse{Items: containerList}, nil
}

type errorMessager struct {
	sync.RWMutex
	errors []error
}

func newErrorMessager() *errorMessager {
	return &errorMessager{errors: make([]error, 0)}
}

func (e *errorMessager) HasErrors() bool {
	e.RLock()
	defer e.RUnlock()
	return len(e.errors) > 0
}

func (e *errorMessager) Error() string {
	e.RLock()
	defer e.RUnlock()
	var line string
	for _, err := range e.errors {
		line += err.Error() + "\n"
	}
	return line
}

func (e *errorMessager) addError(err error) {
	e.Lock()
	defer e.Unlock()
	e.errors = append(e.errors, err)
}

func (e *errorMessager) SendEndError(err error) {
	e.addError(err)
}

func (e *errorMessager) SendError(err error) {
	e.addError(err)
}

func (e *errorMessager) SendMsg(s string) {
	mlog.Debug(s)
}

func (e *errorMessager) SendProtoMsg(message plugins.WebsocketMessage) {
}

func (e *errorMessager) SendProcessPercent(s string) {
}

func (e *errorMessager) Stop(err error) {
	e.addError(err)
}

func (e *errorMessager) SendDeployedResult(resultType websocket.ResultType, s string, m *models.Project) {
}

type messager struct {
	mu        sync.RWMutex
	isStopped bool
	stoperr   error

	slugName string
	t        websocket.Type
	server   project.Project_ApplyServer
}

func (m *messager) SendDeployedResult(resultType websocket.ResultType, s string, p *models.Project) {
	var ns model.NamespaceModel
	if p.Namespace.ID != 0 {
		ns = model.NamespaceModel{
			Id:               int64(p.Namespace.ID),
			Name:             p.Namespace.Name,
			ImagePullSecrets: p.Namespace.ImagePullSecretsArray(),
			CreatedAt:        utils.ToRFC3339DatetimeString(&p.Namespace.CreatedAt),
			UpdatedAt:        utils.ToRFC3339DatetimeString(&p.Namespace.UpdatedAt),
		}
	}
	m.send(&project.ProjectApplyResponse{
		Metadata: &websocket.Metadata{
			Slug:   m.slugName,
			Type:   m.t,
			Result: resultType,
			End:    true,
			Data:   s,
		},
		Project: &model.ProjectModel{
			Id:             int64(p.ID),
			Name:           p.Name,
			GitProjectId:   int64(p.GitProjectId),
			GitBranch:      p.GitBranch,
			GitCommit:      p.GitCommit,
			Config:         p.Config,
			OverrideValues: p.OverrideValues,
			DockerImage:    p.DockerImage,
			PodSelectors:   p.PodSelectors,
			NamespaceId:    int64(p.NamespaceId),
			Atomic:         p.Atomic,
			CreatedAt:      utils.ToRFC3339DatetimeString(&p.CreatedAt),
			UpdatedAt:      utils.ToRFC3339DatetimeString(&p.UpdatedAt),
			ExtraValues:    p.ExtraValues,
			Namespace:      &ns,
		},
	})
}

func (m *messager) SendEndError(err error) {
	m.send(&project.ProjectApplyResponse{Metadata: &websocket.Metadata{
		Slug:   m.slugName,
		Type:   m.t,
		Result: websocket.ResultType_Error,
		End:    true,
		Data:   err.Error(),
	}})
}

func (m *messager) SendError(err error) {
	m.send(&project.ProjectApplyResponse{Metadata: &websocket.Metadata{
		Slug:   m.slugName,
		Type:   m.t,
		Result: websocket.ResultType_Error,
		End:    false,
		Data:   err.Error(),
	}})
}

func (m *messager) SendProcessPercent(s string) {
	// 不需要
}

func (m *messager) SendMsg(s string) {
	m.send(&project.ProjectApplyResponse{Metadata: &websocket.Metadata{
		Slug:   m.slugName,
		Type:   m.t,
		Result: websocket.ResultType_Success,
		End:    false,
		Data:   s,
	}})
}

func (m *messager) SendProtoMsg(message plugins.WebsocketMessage) {
	m.send(&project.ProjectApplyResponse{Metadata: message.GetMetadata()})
}

func (m *messager) send(res *project.ProjectApplyResponse) {
	if m.IsStopped() {
		return
	}
	m.server.Send(res)
}

func (m *messager) Stop(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isStopped = true
	m.stoperr = err
}

func (m *messager) IsStopped() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.isStopped
}
