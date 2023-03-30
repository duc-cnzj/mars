package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars-client/v4/project"
	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/annotations"
	app "github.com/duc-cnzj/mars/v4/internal/app/helper"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/event/events"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/models"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
	"github.com/duc-cnzj/mars/v4/internal/scopes"
	"github.com/duc-cnzj/mars/v4/internal/socket"
	"github.com/duc-cnzj/mars/v4/internal/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func init() {
	RegisterServer(func(s grpc.ServiceRegistrar, app contracts.ApplicationInterface) {
		project.RegisterProjectServer(s, &projectSvc{
			helmer:     &socket.DefaultHelmer{},
			NewJobFunc: socket.NewJober,
		})
	})
	RegisterEndpoint(project.RegisterProjectHandlerFromEndpoint)
}

type projectSvc struct {
	helmer     contracts.Helmer
	NewJobFunc socket.NewJobFunc
	project.UnimplementedProjectServer
}

func (p *projectSvc) List(ctx context.Context, request *project.ListRequest) (*project.ListResponse, error) {
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

func (p *projectSvc) ApplyDryRun(ctx context.Context, input *project.ApplyRequest) (*project.DryRunApplyResponse, error) {
	var pubsub contracts.PubSub = &plugins.EmptyPubSub{}
	t := websocket.Type_ApplyProject
	msger := newEmptyMessager()
	if err := p.completeInput(input, msger); err != nil {
		return nil, err
	}
	mlog.Debug("ApplyDryRun..")
	user := MustGetUser(ctx)
	job := p.NewJobFunc(&socket.JobInput{
		Type:         t,
		NamespaceId:  input.NamespaceId,
		Name:         input.Name,
		GitProjectId: input.GitProjectId,
		GitBranch:    input.GitBranch,
		GitCommit:    input.GitCommit,
		Config:       input.Config,
		Atomic:       input.Atomic,
		ExtraValues:  input.ExtraValues,
		Version:      input.Version,
	}, *user, "", msger, pubsub, input.InstallTimeoutSeconds, socket.WithDryRun())

	ch := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			job.Stop(ctx.Err())
		case <-ch:
		}
	}()
	err := socket.InstallProject(job)
	close(ch)
	if err != nil {
		return nil, err
	}

	return &project.DryRunApplyResponse{Results: job.Manifests()}, nil
}

func (p *projectSvc) Apply(input *project.ApplyRequest, server project.Project_ApplyServer) error {
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
	ch := make(chan struct{})

	job := p.NewJobFunc(&socket.JobInput{
		Type:         t,
		NamespaceId:  input.NamespaceId,
		Name:         input.Name,
		GitProjectId: input.GitProjectId,
		GitBranch:    input.GitBranch,
		GitCommit:    input.GitCommit,
		Config:       input.Config,
		Atomic:       input.Atomic,
		ExtraValues:  input.ExtraValues,
		Version:      input.Version,
	}, *user, "", msger, pubsub, input.InstallTimeoutSeconds)

	go func() {
		select {
		case <-server.Context().Done():
			job.Stop(server.Context().Err())
		case <-ch:
		}
	}()
	err := socket.InstallProject(job)
	close(ch)

	return err
}

func (p *projectSvc) completeInput(input *project.ApplyRequest, msger contracts.Msger) error {
	if input.GitCommit == "" {
		commits, _ := plugins.GetGitServer().ListCommits(fmt.Sprintf("%d", input.GitProjectId), input.GitBranch)
		if len(commits) < 1 {
			return errors.New("没有可用的 commit")
		}
		lastCommit := commits[0]
		input.GitCommit = lastCommit.GetID()
		msger.SendMsg(fmt.Sprintf("未传入commit，使用最新的commit [%s](%s)", lastCommit.GetTitle(), lastCommit.GetWebURL()))
	}
	return nil
}

func (p *projectSvc) Delete(ctx context.Context, request *project.DeleteRequest) (*project.DeleteResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		return nil, err
	}
	if err := p.helmer.Uninstall(projectModel.Name, projectModel.Namespace.Name, mlog.Debugf); err != nil {
		mlog.Error(err)
	}
	app.DB().Delete(&projectModel)
	app.Event().Dispatch(events.EventProjectDeleted, &projectModel)

	AuditLog(MustGetUser(ctx).Name, types.EventActionType_Delete,
		fmt.Sprintf("删除项目: %d: %s/%s ", projectModel.ID, projectModel.Namespace.Name, projectModel.Name))

	return &project.DeleteResponse{}, nil
}

func (p *projectSvc) Show(ctx context.Context, request *project.ShowRequest) (*project.ShowResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	marsC, _ := GetProjectMarsConfig(projectModel.GitProjectId, projectModel.GitBranch)
	cpu, memory := utils.GetCpuAndMemory(projectModel.GetAllPodMetrics())

	nodePortMapping := utils.GetNodePortMappingByProjects(projectModel.Namespace.Name, projectModel)
	ingMapping := utils.GetIngressMappingByProjects(projectModel.Namespace.Name, projectModel)
	lbMapping := utils.GetLoadBalancerMappingByProjects(projectModel.Namespace.Name, projectModel)

	var urls = make([]*types.ServiceEndpoint, 0)
	urls = append(urls, ingMapping.Get(projectModel.Name)...)
	urls = append(urls, nodePortMapping.Get(projectModel.Name)...)
	urls = append(urls, lbMapping.Get(projectModel.Name)...)

	return &project.ShowResponse{
		Project:  projectModel.ProtoTransform(),
		Urls:     urls,
		Cpu:      cpu,
		Memory:   memory,
		Elements: marsC.Elements,
	}, nil
}

func (p *projectSvc) Version(ctx context.Context, req *project.VersionRequest) (*project.VersionResponse, error) {
	var pm = models.Project{ID: int(req.ProjectId)}
	app.DB().Select("id", "version").First(&pm)

	return &project.VersionResponse{Version: int64(pm.Version)}, nil
}

func (p *projectSvc) AllContainers(ctx context.Context, request *project.AllContainersRequest) (*project.AllContainersResponse, error) {
	var projectModel models.Project
	if err := app.DB().Preload("Namespace").Where("`id` = ?", request.ProjectId).First(&projectModel).Error; err != nil {
		return nil, err
	}

	var list = projectModel.GetAllPods()

	var containerList []*types.StateContainer
	for _, item := range list {
		var ignores = make(map[string]struct{})
		if s, ok := item.Pod.Annotations[annotations.IgnoreContainerNames]; ok {
			split := strings.Split(s, ",")
			for _, sp := range split {
				ignores[strings.TrimSpace(sp)] = struct{}{}
			}
		}
		for _, c := range item.Pod.Spec.Containers {
			if _, found := ignores[c.Name]; found {
				continue
			}
			containerList = append(containerList,
				&types.StateContainer{
					Namespace:   projectModel.Namespace.Name,
					Pod:         item.Pod.Name,
					Container:   c.Name,
					IsOld:       item.IsOld,
					Terminating: item.Terminating,
					Pending:     item.Pending,
				},
			)
		}
	}

	return &project.AllContainersResponse{Items: containerList}, nil
}

func (p *projectSvc) HostVariables(ctx context.Context, req *project.HostVariablesRequest) (*project.HostVariablesResponse, error) {
	marsC, err := utils.GetProjectMarsConfig(req.GitProjectId, req.GitBranch)
	if err != nil {
		return nil, err
	}
	if req.ProjectName == "" {
		req.ProjectName = utils.GetProjectName(fmt.Sprintf("%d", req.GitProjectId), marsC)
	}

	sub := utils.GetPreOccupiedLenByValuesYaml(marsC.ValuesYaml)
	hosts := make(map[string]string)
	for i := 1; i <= 10; i++ {
		hosts[fmt.Sprintf("%s%d", socket.VarHost, i)] = plugins.GetDomainManager().GetDomainByIndex(req.ProjectName, utils.GetMarsNamespace(req.Namespace), i, sub)
	}

	return &project.HostVariablesResponse{Hosts: hosts}, nil
}

type emptyMessager struct {
}

func newEmptyMessager() *emptyMessager {
	return &emptyMessager{}
}

func (e *emptyMessager) SendEndError(err error)                                            {}
func (e *emptyMessager) SendError(err error)                                               {}
func (e *emptyMessager) SendMsg(s string)                                                  {}
func (e *emptyMessager) SendProtoMsg(message contracts.WebsocketMessage)                   {}
func (e *emptyMessager) SendProcessPercent(int64)                                          {}
func (e *emptyMessager) SendMsgWithContainerLog(msg string, containers []*types.Container) {}
func (e *emptyMessager) SendDeployedResult(resultType websocket.ResultType, s string, p *types.ProjectModel) {
}

type messager struct {
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

func (m *messager) SendProcessPercent(p int64) {
	if m.sendPercent {
		res := &websocket.Metadata{
			Slug:    m.slugName,
			Type:    websocket.Type_ProcessPercent,
			Result:  websocket.ResultType_Success,
			End:     false,
			Percent: p,
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

func (m *messager) SendMsgWithContainerLog(msg string, containers []*types.Container) {
	m.send(&project.ApplyResponse{Metadata: &websocket.Metadata{
		Slug:    m.slugName,
		Type:    m.t,
		Result:  websocket.ResultType_LogWithContainers,
		End:     false,
		Message: msg,
	}})
}

func (m *messager) send(res *project.ApplyResponse) {
	m.server.Send(res)
}
