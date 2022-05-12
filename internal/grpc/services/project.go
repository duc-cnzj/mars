package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/duc-cnzj/mars-client/v4/project"
	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars-client/v4/websocket"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/event/events"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/scopes"
	"github.com/duc-cnzj/mars/internal/socket"
	"github.com/duc-cnzj/mars/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		project.RegisterProjectServer(s, new(ProjectSvc))
	})
	RegisterEndpoint(project.RegisterProjectHandlerFromEndpoint)
}

type ProjectSvc struct {
	project.UnimplementedProjectServer
}

func (p *ProjectSvc) List(ctx context.Context, request *project.ListRequest) (*project.ListResponse, error) {
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
	res := make([]*types.ProjectModel, 0, len(projects))
	for _, p := range projects {
		res = append(res, p.ProtoTransform())
	}

	return &project.ListResponse{
		Page:     request.Page,
		PageSize: request.PageSize,
		Count:    count,
		Items:    res,
	}, nil
}

func (p *ProjectSvc) ApplyDryRun(ctx context.Context, input *project.ApplyRequest) (*project.DryRunApplyResponse, error) {
	var pubsub contracts.PubSub = &plugins.EmptyPubSub{}
	t := websocket.Type_ApplyProject
	errMsger := newErrorMessager()
	if err := p.completeInput(input, errMsger); err != nil {
		return nil, err
	}
	mlog.Debug("ApplyDryRun..")
	user := MustGetUser(ctx)
	job := socket.NewJober(&websocket.CreateProjectInput{
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

	return &project.DryRunApplyResponse{Results: job.Manifests()}, nil
}

func (p *ProjectSvc) Apply(input *project.ApplyRequest, server project.Project_ApplyServer) error {
	var pubsub contracts.PubSub = &plugins.EmptyPubSub{}
	if input.WebsocketSync {
		pubsub = plugins.GetWsSender().New("", "")
	}
	t := websocket.Type_ApplyProject
	msger := &messager{
		slugName:    utils.GetSlugName(input.NamespaceId, input.Name),
		t:           t,
		server:      server,
		sendPercent: input.SendPercent,
	}
	if err := p.completeInput(input, msger); err != nil {
		return err
	}
	user := MustGetUser(server.Context())
	ch := make(chan struct{}, 1)

	job := socket.NewJober(&websocket.CreateProjectInput{
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

func (p *ProjectSvc) completeInput(input *project.ApplyRequest, msger contracts.Msger) error {
	if input.GitCommit == "" {
		commits, _ := plugins.GetGitServer().ListCommits(fmt.Sprintf("%d", input.GitProjectId), input.GitBranch)
		if len(commits) < 1 {
			return errors.New("没有可用的 commit")
		}
		lastCommit := commits[0]
		input.GitCommit = lastCommit.GetID()
		msger.SendMsg(fmt.Sprintf("未传入commit，使用最新的commit [%s](%s)", lastCommit.GetTitle(), lastCommit.GetWebURL()))
	}
	if input.Name == "" {
		gitProject, _ := plugins.GetGitServer().GetProject(strconv.Itoa(int(input.GitProjectId)))
		input.Name = gitProject.GetName()
	}
	return nil
}

func (p *ProjectSvc) Delete(ctx context.Context, request *project.DeleteRequest) (*project.DeleteResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		return nil, err
	}
	if err := utils.UninstallRelease(projectModel.Name, projectModel.Namespace.Name, mlog.Debugf); err != nil {
		mlog.Error(err)
	}
	app.DB().Delete(&projectModel)
	app.Event().Dispatch(events.EventProjectDeleted, map[string]any{"data": &projectModel})

	AuditLog(MustGetUser(ctx).Name, types.EventActionType_Delete,
		fmt.Sprintf("删除项目: %d: %s/%s ", projectModel.ID, projectModel.Namespace.Name, projectModel.Name))

	return &project.DeleteResponse{}, nil
}

func (p *ProjectSvc) Show(ctx context.Context, request *project.ShowRequest) (*project.ShowResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	marsC, _ := GetProjectMarsConfig(projectModel.GitProjectId, projectModel.GitBranch)
	cpu, memory := utils.GetCpuAndMemory(projectModel.GetAllPodMetrics())

	nodePortMapping := utils.GetNodePortMappingByNamespace(projectModel.Namespace.Name)
	ingMapping := utils.GetIngressMappingByNamespace(projectModel.Namespace.Name)

	var urls = make([]*types.ServiceEndpoint, 0)
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

	return &project.ShowResponse{
		Project:  projectModel.ProtoTransform(),
		Urls:     urls,
		Cpu:      cpu,
		Memory:   memory,
		Elements: marsC.Elements,
	}, nil
}

func (p *ProjectSvc) AllContainers(ctx context.Context, request *project.AllContainersRequest) (*project.AllContainersResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		return nil, err
	}

	var list = projectModel.GetAllPods()

	var containerList []*types.Container
	for _, item := range list {
		for _, c := range item.Spec.Containers {
			containerList = append(containerList,
				&types.Container{
					Namespace: projectModel.Namespace.Name,
					Pod:       item.Name,
					Container: c.Name,
				},
			)
		}
	}

	return &project.AllContainersResponse{Items: containerList}, nil
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

func (e *errorMessager) SendProtoMsg(message contracts.WebsocketMessage) {
}

func (e *errorMessager) SendProcessPercent(s string) {
}

func (e *errorMessager) Stop(err error) {
	e.addError(err)
}

func (e *errorMessager) SendDeployedResult(resultType websocket.ResultType, s string, p *types.ProjectModel) {
}

type messager struct {
	mu        sync.RWMutex
	isStopped bool
	stoperr   error

	sendPercent bool

	slugName string
	t        websocket.Type
	server   project.Project_ApplyServer
}

func (m *messager) SendDeployedResult(resultType websocket.ResultType, s string, p *types.ProjectModel) {
	m.send(&project.ApplyResponse{
		Metadata: &websocket.Metadata{
			Slug:    m.slugName,
			Type:    m.t,
			Result:  resultType,
			End:     true,
			Message: s,
		},
		Project: p,
	})
}

func (m *messager) SendEndError(err error) {
	m.send(&project.ApplyResponse{Metadata: &websocket.Metadata{
		Slug:    m.slugName,
		Type:    m.t,
		Result:  websocket.ResultType_Error,
		End:     true,
		Message: err.Error(),
	}})
}

func (m *messager) SendError(err error) {
	m.send(&project.ApplyResponse{Metadata: &websocket.Metadata{
		Slug:    m.slugName,
		Type:    m.t,
		Result:  websocket.ResultType_Error,
		End:     false,
		Message: err.Error(),
	}})
}

func (m *messager) SendProcessPercent(s string) {
	if m.sendPercent {
		res := &websocket.Metadata{
			Slug:    m.slugName,
			Type:    websocket.Type_ProcessPercent,
			Result:  websocket.ResultType_Success,
			End:     false,
			Message: s,
		}
		m.send(&project.ApplyResponse{Metadata: res})
	}
}

func (m *messager) SendMsg(s string) {
	m.send(&project.ApplyResponse{Metadata: &websocket.Metadata{
		Slug:    m.slugName,
		Type:    m.t,
		Result:  websocket.ResultType_Success,
		End:     false,
		Message: s,
	}})
}

func (m *messager) SendProtoMsg(message contracts.WebsocketMessage) {
	m.send(&project.ApplyResponse{Metadata: message.GetMetadata()})
}

func (m *messager) send(res *project.ApplyResponse) {
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
