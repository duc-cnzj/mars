package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars/api/v4/project"
	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/annotation"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/socket"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/duc-cnzj/mars/v4/internal/util"
	"github.com/duc-cnzj/mars/v4/internal/util/pagination"
	"github.com/duc-cnzj/mars/v4/internal/util/serialize"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
)

var _ project.ProjectServer = (*projectSvc)(nil)

type projectSvc struct {
	project.UnimplementedProjectServer

	jobManager socket.JobManager
	projRepo   repo.ProjectRepo
	gitRepo    repo.GitRepo
	k8sRepo    repo.K8sRepo
	eventRepo  repo.EventRepo
	logger     mlog.Logger
	helmer     repo.HelmerRepo
	nsRepo     repo.NamespaceRepo
	repoRepo   repo.RepoRepo
	plMgr      application.PluginManger
}

func NewProjectSvc(
	repoRepo repo.RepoRepo,
	plMgr application.PluginManger,
	jobManager socket.JobManager,
	projRepo repo.ProjectRepo,
	gitRepo repo.GitRepo,
	k8sRepo repo.K8sRepo,
	eventRepo repo.EventRepo,
	logger mlog.Logger,
	helmer repo.HelmerRepo,
	nsRepo repo.NamespaceRepo,
) project.ProjectServer {
	return &projectSvc{
		jobManager: jobManager,
		projRepo:   projRepo,
		gitRepo:    gitRepo,
		k8sRepo:    k8sRepo,
		eventRepo:  eventRepo,
		logger:     logger.WithModule("services/project"),
		helmer:     helmer,
		nsRepo:     nsRepo,
		repoRepo:   repoRepo,
		plMgr:      plMgr,
	}
}

func (p *projectSvc) List(ctx context.Context, request *project.ListRequest) (*project.ListResponse, error) {
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	list, pag, err := p.projRepo.List(ctx, &repo.ListProjectInput{
		Page:          page,
		PageSize:      size,
		OrderByIDDesc: lo.ToPtr(true),
	})
	if err != nil {
		return nil, err
	}

	return &project.ListResponse{
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Count:    pag.Count,
		Items:    serialize.Serialize(list, transformer.FromProject),
	}, nil
}

func (p *projectSvc) WebApply(ctx context.Context, input *project.WebApplyRequest) (*project.WebApplyResponse, error) {
	show, err := p.repoRepo.Show(ctx, int(input.RepoId))
	if err != nil {
		return nil, err
	}
	if input.Name == "" {
		input.Name = show.Name
	}
	msger := newEmptyMessager()
	if show.NeedGitRepo {
		input.GitBranch, input.GitCommit, err = p.getBranchAndCommitIfMissing(input.GitBranch, input.GitCommit, show, msger)
		if err != nil {
			return nil, err
		}
	}

	p.logger.DebugCtx(ctx, "WebApply..")
	user := MustGetUser(ctx)
	jobInput := &socket.JobInput{
		Type:        websocket.Type_ApplyProject,
		NamespaceId: input.NamespaceId,
		Name:        input.Name,
		RepoID:      int32(show.ID),
		GitBranch:   input.GitBranch,
		GitCommit:   input.GitCommit,
		Config:      input.Config,
		ExtraValues: input.ExtraValues,
		Version:     input.Version,
		User:        user,
		DryRun:      input.DryRun,
		PubSub:      application.NewEmptyPubSub(),
		Messager:    msger,
	}
	job := p.jobManager.NewJob(jobInput)

	ch := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			p.logger.WarningCtx(ctx, "WebApply ctx done", ctx.Err())
			job.Stop(ctx.Err())
		case <-ch:
		}
	}()
	err = socket.InstallProject(ctx, job)
	close(ch)
	if err != nil {
		return nil, err
	}

	var projectModel *types.ProjectModel
	if job.IsNotDryRun() {
		pro, err := p.projRepo.Show(ctx, job.Project().ID)
		if err != nil {
			return nil, err
		}
		projectModel = transformer.FromProject(pro)
	}

	return &project.WebApplyResponse{
		YamlFiles: job.Manifests(),
		Project:   projectModel,
		DryRun:    input.GetDryRun(),
	}, nil
}

func (p *projectSvc) Apply(input *project.ApplyRequest, server project.Project_ApplyServer) error {
	var pubsub application.PubSub = application.NewEmptyPubSub()
	if input.WebsocketSync {
		pubsub = p.plMgr.Ws().New("", "")
	}
	defer pubsub.Close()
	t := websocket.Type_ApplyProject
	show, _ := p.repoRepo.Show(server.Context(), int(input.RepoId))
	if input.Name == "" {
		input.Name = show.Name
	}

	msger := NewMessager(
		input.SendPercent,
		util.GetSlugName(input.NamespaceId, input.Name),
		t,
		server,
	)

	var err error
	if show.NeedGitRepo {
		input.GitBranch, input.GitCommit, err = p.getBranchAndCommitIfMissing(input.GitBranch, input.GitCommit, show, msger)
		if err != nil {
			return err
		}
	}

	user := MustGetUser(server.Context())
	ch := make(chan struct{})

	jobInput := &socket.JobInput{
		Type:           t,
		NamespaceId:    input.NamespaceId,
		Name:           input.Name,
		RepoID:         input.RepoId,
		GitBranch:      input.GitBranch,
		GitCommit:      input.GitCommit,
		Config:         input.Config,
		Atomic:         lo.ToPtr(input.Atomic),
		ExtraValues:    input.ExtraValues,
		Version:        input.Version,
		TimeoutSeconds: input.InstallTimeoutSeconds,
		User:           user,
		PubSub:         pubsub,
		Messager:       msger,
	}
	job := p.jobManager.NewJob(jobInput)

	go func() {
		select {
		case <-server.Context().Done():
			job.Stop(server.Context().Err())
		case <-ch:
		}
	}()
	err = socket.InstallProject(server.Context(), job)
	close(ch)

	return err
}

func (p *projectSvc) getBranchAndCommitIfMissing(inBranch, inCommit string, show *repo.Repo, msger socket.DeployMsger) (branch string, commit string, err error) {
	branch = inBranch
	commit = inCommit
	if branch == "" {
		branch = show.DefaultBranch
		msger.SendMsg(fmt.Sprintf("未传入分支，使用默认分支 %s", branch))
	}
	if commit == "" {
		commits, _ := p.gitRepo.ListCommits(context.TODO(), int(show.GitProjectID), branch)
		if len(commits) < 1 {
			return "", "", errors.New("没有可用的 commit")
		}
		lastCommit := commits[0]
		commit = lastCommit.GetID()
		msger.SendMsg(fmt.Sprintf("未传入commit，使用最新的commit [%s](%s)", lastCommit.GetTitle(), lastCommit.GetWebURL()))
	}
	return
}

func (p *projectSvc) Delete(ctx context.Context, request *project.DeleteRequest) (*project.DeleteResponse, error) {
	//var event = p.eventer
	projectModel, err := p.projRepo.Show(ctx, int(request.Id))
	if err != nil {
		return nil, err
	}
	if err := p.helmer.Uninstall(projectModel.Name, projectModel.Namespace.Name, p.logger.Debugf); err != nil {
		p.logger.Error(err)
	}
	p.projRepo.Delete(ctx, int(request.Id))
	p.eventRepo.Dispatch(repo.EventProjectDeleted, projectModel)

	p.eventRepo.AuditLogWithRequest(
		types.EventActionType_Delete,
		MustGetUser(ctx).Name,
		fmt.Sprintf("删除项目: %d: %s/%s ", projectModel.ID, projectModel.Namespace.Name, projectModel.Name),
		request,
	)

	return &project.DeleteResponse{}, nil
}

func (p *projectSvc) Show(ctx context.Context, request *project.ShowRequest) (*project.ShowResponse, error) {
	projectModel, err := p.projRepo.Show(ctx, int(request.Id))
	if err != nil {
		return nil, err
	}
	cpu, memory := p.k8sRepo.GetCpuAndMemory(p.k8sRepo.GetAllPodMetrics(projectModel))
	nodePortMapping := p.projRepo.GetNodePortMappingByProjects(projectModel.Namespace.Name, projectModel)
	ingMapping := p.projRepo.GetIngressMappingByProjects(projectModel.Namespace.Name, projectModel)
	lbMapping := p.projRepo.GetLoadBalancerMappingByProjects(projectModel.Namespace.Name, projectModel)

	var urls = make([]*types.ServiceEndpoint, 0)
	urls = append(urls, ingMapping.Get(projectModel.Name)...)
	urls = append(urls, nodePortMapping.Get(projectModel.Name)...)
	urls = append(urls, lbMapping.Get(projectModel.Name)...)

	return &project.ShowResponse{
		Item:   transformer.FromProject(projectModel),
		Urls:   urls,
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

func (p *projectSvc) Version(ctx context.Context, req *project.VersionRequest) (*project.VersionResponse, error) {
	show, _ := p.projRepo.Show(ctx, int(req.Id))

	return &project.VersionResponse{Version: int32(show.Version)}, nil
}

func (p *projectSvc) AllContainers(ctx context.Context, request *project.AllContainersRequest) (*project.AllContainersResponse, error) {
	projectModel, err := p.projRepo.Show(ctx, int(request.Id))
	if err != nil {
		return nil, err
	}

	var list = p.projRepo.GetAllPods(projectModel)

	var containerList []*types.StateContainer
	for _, item := range list {
		var ignores = make(map[string]struct{})
		if s, ok := item.Pod.Annotations[annotation.IgnoreContainerNames]; ok {
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

var _ socket.DeployMsger = (*emptyMessager)(nil)

type emptyMessager struct{}

func newEmptyMessager() *emptyMessager {
	return &emptyMessager{}
}

func (e *emptyMessager) Current() int64 {
	return 0
}
func (e *emptyMessager) Add() {
}
func (e *emptyMessager) To(percent int64) {
}
func (e *emptyMessager) SendEndError(err error)                                                {}
func (e *emptyMessager) SendError(err error)                                                   {}
func (e *emptyMessager) SendMsg(s string)                                                      {}
func (e *emptyMessager) SendProtoMsg(message application.WebsocketMessage)                     {}
func (e *emptyMessager) SendProcessPercent(int64)                                              {}
func (e *emptyMessager) SendMsgWithContainerLog(msg string, containers []*websocket.Container) {}
func (e *emptyMessager) SendDeployedResult(resultType websocket.ResultType, s string, p *types.ProjectModel) {
}

var _ socket.DeployMsger = (*messager)(nil)

type messager struct {
	percent     socket.Percentable
	sendPercent bool

	slugName string
	t        websocket.Type
	server   project.Project_ApplyServer
}

func NewMessager(sendPercent bool, slugName string, t websocket.Type, server project.Project_ApplyServer) socket.DeployMsger {
	m := messager{sendPercent: sendPercent, slugName: slugName, t: t, server: server}
	m.percent = socket.NewProcessPercent(&m, socket.NewRealSleeper())
	return &m
}

func (m *messager) Current() int64 {
	return m.percent.Current()
}

func (m *messager) Add() {
	m.percent.Add()
}

func (m *messager) To(percent int64) {
	m.percent.To(percent)
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
			Percent: int32(p),
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

func (m *messager) SendMsgWithContainerLog(msg string, containers []*websocket.Container) {
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
