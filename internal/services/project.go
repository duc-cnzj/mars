package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars/api/v4/project"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/annotations"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/socket"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/duc-cnzj/mars/v4/internal/utils"
	"github.com/duc-cnzj/mars/v4/internal/utils/mars"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
)

var _ project.ProjectServer = (*projectSvc)(nil)

type projectSvc struct {
	project.UnimplementedProjectServer

	jobManager socket.JobManager
	projRepo   repo.ProjectRepo
	wsRepo     repo.WsRepo
	gitRepo    repo.GitRepo
	k8sRepo    repo.K8sRepo
	dm         repo.DomainRepo
	eventRepo  repo.EventRepo
	logger     mlog.Logger
	helmer     repo.HelmerRepo
	nsRepo     repo.NamespaceRepo
	toolRepo   repo.ToolRepo
}

func NewProjectSvc(
	jobManager socket.JobManager,
	projRepo repo.ProjectRepo,
	wsRepo repo.WsRepo,
	gitRepo repo.GitRepo,
	k8sRepo repo.K8sRepo,
	pl application.PluginManger,
	eventRepo repo.EventRepo,
	logger mlog.Logger,
	helmer repo.HelmerRepo,
	nsRepo repo.NamespaceRepo,
) project.ProjectServer {
	return &projectSvc{jobManager: jobManager, projRepo: projRepo, wsRepo: wsRepo, gitRepo: gitRepo, k8sRepo: k8sRepo, dm: pl.Domain(), eventRepo: eventRepo, logger: logger, helmer: helmer, nsRepo: nsRepo}
}

func (p *projectSvc) List(ctx context.Context, request *project.ListRequest) (*project.ListResponse, error) {
	list, pag, err := p.projRepo.List(ctx, &repo.ListProjectInput{
		Page:          request.Page,
		PageSize:      request.PageSize,
		OrderByIDDesc: lo.ToPtr(true),
	})
	if err != nil {
		return nil, err
	}
	res := make([]*types.ProjectModel, 0, len(list))
	for _, p := range list {
		res = append(res, transformer.FromProject(p))
	}

	return &project.ListResponse{
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Count:    pag.Count,
		Items:    res,
	}, nil
}

func (p *projectSvc) ApplyDryRun(ctx context.Context, input *project.ApplyRequest) (*project.DryRunApplyResponse, error) {
	var pubsub application.PubSub = &application.EmptyPubSub{}
	t := websocket.Type_ApplyProject
	msger := newEmptyMessager()
	if err := p.completeInput(input, msger); err != nil {
		return nil, err
	}
	p.logger.Debug("ApplyDryRun..")
	user := MustGetUser(ctx)
	job := p.jobManager.NewJob(&socket.JobInput{
		Type:           t,
		NamespaceId:    input.NamespaceId,
		Name:           input.Name,
		GitProjectId:   input.GitProjectId,
		GitBranch:      input.GitBranch,
		GitCommit:      input.GitCommit,
		Config:         input.Config,
		Atomic:         input.Atomic,
		ExtraValues:    input.ExtraValues,
		Version:        input.Version,
		TimeoutSeconds: input.InstallTimeoutSeconds,
		User:           user,
		DryRun:         true,
		PubSub:         pubsub,
	})

	ch := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			job.Stop(ctx.Err())
		case <-ch:
		}
	}()
	err := socket.InstallProject(ctx, job)
	close(ch)
	if err != nil {
		return nil, err
	}

	return &project.DryRunApplyResponse{Results: job.Manifests()}, nil
}

func (p *projectSvc) Apply(input *project.ApplyRequest, server project.Project_ApplyServer) error {
	var pubsub application.PubSub = &application.EmptyPubSub{}
	if input.WebsocketSync {
		pubsub = p.wsRepo.New("", "")
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

	job := p.jobManager.NewJob(&socket.JobInput{
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
		//, *user, "", msger, pubsub, input.InstallTimeoutSeconds
		TimeoutSeconds: input.InstallTimeoutSeconds,
		User:           user,
		PubSub:         pubsub,
	})

	go func() {
		select {
		case <-server.Context().Done():
			job.Stop(server.Context().Err())
		case <-ch:
		}
	}()
	err := socket.InstallProject(server.Context(), job)
	close(ch)

	return err
}

func (p *projectSvc) completeInput(input *project.ApplyRequest, msger contracts.Msger) error {
	if input.GitCommit == "" {
		commits, _ := p.gitRepo.ListCommits(context.TODO(), int(input.GitProjectId), input.GitBranch)
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
	//var event = p.eventer
	projectModel, err := p.projRepo.Show(ctx, int(request.ProjectId))
	if err != nil {
		return nil, err
	}
	if err := p.helmer.Uninstall(projectModel.Name, projectModel.Edges.Namespace.Name, p.logger.Debugf); err != nil {
		p.logger.Error(err)
	}
	p.projRepo.Delete(ctx, int(request.ProjectId))
	p.eventRepo.Dispatch(repo.EventProjectDeleted, &projectModel)

	p.eventRepo.AuditLog(
		types.EventActionType_Delete,
		MustGetUser(ctx).Name,
		fmt.Sprintf("删除项目: %d: %s/%s ", projectModel.ID, projectModel.Edges.Namespace.Name, projectModel.Name),
	)

	return &project.DeleteResponse{}, nil
}

func (p *projectSvc) Show(ctx context.Context, request *project.ShowRequest) (*project.ShowResponse, error) {
	projectModel, err := p.projRepo.Show(ctx, int(request.ProjectId))
	if err != nil {
		return nil, err
	}
	marsC, _ := GetProjectMarsConfig(projectModel.GitProjectID, projectModel.GitBranch)
	cpu, memory := p.k8sRepo.GetCpuAndMemory(p.k8sRepo.GetAllPodMetrics(projectModel))

	nodePortMapping := p.projRepo.GetNodePortMappingByProjects(projectModel.Edges.Namespace.Name, projectModel)
	ingMapping := p.projRepo.GetIngressMappingByProjects(projectModel.Edges.Namespace.Name, projectModel)
	lbMapping := p.projRepo.GetLoadBalancerMappingByProjects(projectModel.Edges.Namespace.Name, projectModel)

	var urls = make([]*types.ServiceEndpoint, 0)
	urls = append(urls, ingMapping.Get(projectModel.Name)...)
	urls = append(urls, nodePortMapping.Get(projectModel.Name)...)
	urls = append(urls, lbMapping.Get(projectModel.Name)...)

	return &project.ShowResponse{
		Project:  transformer.FromProject(projectModel),
		Urls:     urls,
		Cpu:      cpu,
		Memory:   memory,
		Elements: marsC.Elements,
	}, nil
}

func (p *projectSvc) Version(ctx context.Context, req *project.VersionRequest) (*project.VersionResponse, error) {
	show, _ := p.projRepo.Show(ctx, int(req.ProjectId))

	return &project.VersionResponse{Version: int64(show.Version)}, nil
}

func (p *projectSvc) AllContainers(ctx context.Context, request *project.AllContainersRequest) (*project.AllContainersResponse, error) {
	projectModel, err := p.projRepo.Show(ctx, int(request.ProjectId))
	if err != nil {
		return nil, err
	}

	var list = p.projRepo.GetAllPods(projectModel)

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
					Namespace:   projectModel.Edges.Namespace.Name,
					Pod:         item.Pod.Name,
					Container:   c.Name,
					IsOld:       item.IsOld,
					Terminating: item.Terminating,
					Pending:     item.Pending,
					Ready:       isContainerReady(item.Pod, c.Name),
				},
			)
		}
	}

	return &project.AllContainersResponse{Items: containerList}, nil
}

func isContainerReady(pod *v1.Pod, containerName string) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == containerName {
			return containerStatus.Ready
		}
	}
	return false
}

func (p *projectSvc) HostVariables(ctx context.Context, req *project.HostVariablesRequest) (*project.HostVariablesResponse, error) {
	marsC, err := GetProjectMarsConfig(req.GitProjectId, req.GitBranch)
	if err != nil {
		return nil, err
	}
	if req.ProjectName == "" {
		req.ProjectName = mars.GetProjectName(fmt.Sprintf("%d", req.GitProjectId), marsC)
	}

	sub := p.toolRepo.GetPreOccupiedLenByValuesYaml(marsC.ValuesYaml)
	hosts := make(map[string]string)
	for i := 1; i <= 10; i++ {
		hosts[fmt.Sprintf("%s%d", socket.VarHost, i)] = p.dm.GetDomainByIndex(req.ProjectName, p.nsRepo.GetMarsNamespace(req.Namespace), i, sub)
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
func (e *emptyMessager) SendProtoMsg(message application.WebsocketMessage)                 {}
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

func (m *messager) SendProtoMsg(message application.WebsocketMessage) {
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
