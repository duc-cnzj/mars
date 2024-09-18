package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/duc-cnzj/mars/api/v5/project"
	"github.com/duc-cnzj/mars/api/v5/types"
	"github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/socket"
	"github.com/duc-cnzj/mars/v5/internal/transformer"
	"github.com/duc-cnzj/mars/v5/internal/util/pagination"
	"github.com/duc-cnzj/mars/v5/internal/util/serialize"
	"github.com/samber/lo"
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
	p.logger.DebugCtx(ctx, "WebApply..")
	job, err := p.apply(
		ctx,
		MustGetUser(ctx),
		newEmptyMessager(),
		&project.ApplyRequest{
			NamespaceId: input.NamespaceId,
			Name:        input.Name,
			RepoId:      input.RepoId,
			GitBranch:   input.GitBranch,
			GitCommit:   input.GitCommit,
			Config:      input.Config,
			ExtraValues: input.ExtraValues,
			Version:     input.Version,
		},
	)
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
	msger := NewMessager(
		input.SendPercent,
		socket.GetSlugName(input.NamespaceId, input.Name),
		websocket.Type_ApplyProject,
		server,
	)

	ctx := server.Context()
	_, err := p.apply(
		ctx,
		MustGetUser(ctx),
		msger,
		input,
	)

	return err
}

func (p *projectSvc) apply(
	ctx context.Context,
	user *auth.UserInfo,
	msger socket.DeployMsger,
	input *project.ApplyRequest,
) (socket.Job, error) {
	var err error

	if can := p.nsRepo.CanAccess(ctx, int(input.NamespaceId), MustGetUser(ctx)); !can {
		return nil, ErrorPermissionDenied
	}

	var pubsub application.PubSub = socket.NewEmptyPubSub()
	if input.WebsocketSync {
		pubsub = p.plMgr.Ws().New("", "")
	}
	defer pubsub.Close()
	t := websocket.Type_ApplyProject
	show, err := p.repoRepo.Show(ctx, int(input.RepoId))
	if err != nil {
		return nil, err
	}
	if input.Name == "" {
		input.Name = show.Name
	}

	if show.NeedGitRepo {
		input.GitBranch, input.GitCommit, err = p.getBranchAndCommitIfMissing(input.GitBranch, input.GitCommit, show, msger)
		if err != nil {
			return nil, err
		}
	}
	ch := make(chan struct{})
	var projectID int32
	if lo.FromPtr(input.Version) > 0 {
		proj, err := p.projRepo.FindByName(ctx, input.Name, int(input.NamespaceId))
		if err == nil {
			projectID = int32(proj.ID)
		}
	}

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
		ProjectID:      projectID,
		TimeoutSeconds: input.InstallTimeoutSeconds,
		User:           user,
		DryRun:         false,
		PubSub:         pubsub,
		Messager:       msger,
	}
	job := p.jobManager.NewJob(jobInput)

	go func() {
		select {
		case <-ctx.Done():
			job.Stop(ctx.Err())
		case <-ch:
		}
	}()
	err = socket.InstallProject(ctx, job)
	close(ch)

	return job, err
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
		commit = lastCommit.ID
		msger.SendMsg(fmt.Sprintf("未传入commit，使用最新的commit [%s](%s)", lastCommit.Title, lastCommit.WebURL))
	}
	return
}

func (p *projectSvc) Show(ctx context.Context, request *project.ShowRequest) (*project.ShowResponse, error) {
	projectModel, err := p.projRepo.Show(ctx, int(request.Id))
	if err != nil {
		return nil, err
	}
	if can := p.nsRepo.CanAccess(ctx, projectModel.NamespaceID, MustGetUser(ctx)); !can {
		return nil, ErrorPermissionDenied
	}

	return &project.ShowResponse{
		Item: transformer.FromProject(projectModel),
	}, nil
}

func (p *projectSvc) MemoryCpuAndEndpoints(ctx context.Context, req *project.MemoryCpuAndEndpointsRequest) (*project.MemoryCpuAndEndpointsResponse, error) {
	projectModel, err := p.projRepo.Show(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}
	cpu, memory := p.k8sRepo.GetCpuAndMemory(ctx, p.k8sRepo.GetAllPodMetrics(ctx, projectModel))
	urls, err := p.projRepo.GetProjectEndpointsInNamespace(ctx, projectModel.Namespace.Name, projectModel.ID)
	if err != nil {
		return nil, err
	}
	return &project.MemoryCpuAndEndpointsResponse{
		Urls:   urls,
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

func (p *projectSvc) Delete(ctx context.Context, request *project.DeleteRequest) (*project.DeleteResponse, error) {
	proj, err := p.projRepo.Show(ctx, int(request.Id))
	if err != nil {
		return nil, err
	}
	if can := p.nsRepo.CanAccess(ctx, proj.NamespaceID, MustGetUser(ctx)); !can {
		return nil, ErrorPermissionDenied
	}

	if err := p.projRepo.Delete(ctx, int(request.Id)); err != nil {
		return nil, err
	}
	if err := p.helmer.Uninstall(proj.Name, proj.Namespace.Name, p.logger.Debugf); err != nil {
		p.logger.Error(err)
	}
	p.eventRepo.Dispatch(repo.EventProjectDeleted, &repo.ProjectDeletedPayload{
		NamespaceID: proj.NamespaceID,
		ProjectID:   proj.ID,
	})

	p.eventRepo.AuditLogWithRequest(
		types.EventActionType_Delete,
		MustGetUser(ctx).Name,
		fmt.Sprintf("删除项目: %d: %s/%s ", proj.ID, proj.Namespace.Name, proj.Name),
		request,
	)

	return &project.DeleteResponse{}, nil
}

func (p *projectSvc) Version(ctx context.Context, req *project.VersionRequest) (*project.VersionResponse, error) {
	v, err := p.projRepo.Version(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return &project.VersionResponse{Version: int32(v)}, nil
}

func (p *projectSvc) AllContainers(ctx context.Context, request *project.AllContainersRequest) (*project.AllContainersResponse, error) {
	pods, err := p.projRepo.GetAllActiveContainers(ctx, int(request.Id))
	if err != nil {
		return nil, err
	}

	return &project.AllContainersResponse{Items: pods}, nil
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
