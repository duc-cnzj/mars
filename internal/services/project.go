package services

import (
	"context"
	"fmt"

	"github.com/duc-cnzj/mars/api/v6/proto/project"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/transformer"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
	"github.com/samber/lo"
)

var _ project.ProjectServer = (*projectSvc)(nil)

// projectSvc 是 project.ProjectServer 的 gRPC 实现：编排项目部署全流程（apply/展示/删除/
// 资源统计/容器列表），聚合部署 job 与各业务仓库，经 access 校验访问权限，由 NewProjectSvc 构造。
type projectSvc struct {
	project.UnimplementedProjectServer

	jobManager deploy.JobManager
	projBiz    biz.ProjectBiz
	gitBiz     biz.GitBiz
	k8sBiz     biz.K8sBiz
	eventBiz   biz.EventBiz
	logger     mlog.Logger
	repoBiz    biz.RepoBiz
	deployBiz  biz.DeployBiz
	plMgr      app.PluginManager
	accessBiz  biz.AccessBiz
}

// ProjectSvcDeps 收口 NewProjectSvc 的构造依赖，由 wire 按字段注入。
type ProjectSvcDeps struct {
	RepoBiz    biz.RepoBiz
	PluginMgr  app.PluginManager
	JobManager deploy.JobManager
	ProjBiz    biz.ProjectBiz
	GitBiz     biz.GitBiz
	K8sBiz     biz.K8sBiz
	EventBiz   biz.EventBiz
	Logger     mlog.Logger
	DeployBiz  biz.DeployBiz
	AccessBiz  biz.AccessBiz
}

// NewProjectSvc 收口项目服务的构造依赖，由 wire 按字段注入。
func NewProjectSvc(deps ProjectSvcDeps) project.ProjectServer {
	logger := deps.Logger.WithModule("services/project")
	return &projectSvc{
		jobManager: deps.JobManager,
		projBiz:    deps.ProjBiz,
		gitBiz:     deps.GitBiz,
		k8sBiz:     deps.K8sBiz,
		eventBiz:   deps.EventBiz,
		logger:     logger,
		repoBiz:    deps.RepoBiz,
		deployBiz:  deps.DeployBiz,
		plMgr:      deps.PluginMgr,
		accessBiz:  deps.AccessBiz,
	}
}

// List 分页列出当前用户可见的项目（按命名空间访问谓词过滤），按 id 倒序返回。
func (p *projectSvc) List(ctx context.Context, request *project.ListRequest) (*project.ListResponse, error) {
	page, size := pagination.InitByDefault(request.Page, request.PageSize)
	user := biz.MustGetUser(ctx)
	list, pag, err := p.projBiz.List(ctx, &biz.ListProjectInput{
		Page:          page,
		PageSize:      size,
		OrderByIDDesc: lo.ToPtr(true),
		// 透传当前用户，data 层按命名空间访问谓词过滤：非 admin 只能看到
		// 可访问命名空间下的项目，与 namespace.List 的可见范围一致。
		Email:   user.Email,
		IsAdmin: user.IsAdmin(),
	})
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}

	return &project.ListResponse{
		Page:     pag.Page,
		PageSize: pag.PageSize,
		Count:    pag.Count,
		Items:    slice.Map(list, transformer.FromProject),
	}, nil
}

// WebApply 单请求发起部署：与 streaming Apply 共享 apply 编排，返回 manifests 与结果项目模型。
func (p *projectSvc) WebApply(ctx context.Context, input *project.WebApplyRequest) (*project.WebApplyResponse, error) {
	p.logger.DebugCtx(ctx, "WebApply..")
	job, err := p.apply(
		ctx,
		biz.MustGetUser(ctx),
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
		// ApplyRequest 没有 DryRun 字段，WebApplyRequest 才有，必须单独透传，
		// 否则 dry_run=true 会真的部署（JobInput.DryRun 恒为 false）。
		input.GetDryRun(),
	)
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}

	var projectModel *types.ProjectModel
	if job.IsNotDryRun() {
		pro, err := p.projBiz.Show(ctx, job.Project().ID)
		if err != nil {
			return nil, logError(ctx, p.logger, err)
		}
		projectModel = transformer.FromProject(pro)
	}

	return &project.WebApplyResponse{
		YamlFiles: job.Manifests(),
		Project:   projectModel,
		DryRun:    input.GetDryRun(),
	}, nil
}

// Apply 以服务端流方式发起部署：进度/日志/结果经 messager 逐帧下发到客户端。
func (p *projectSvc) Apply(input *project.ApplyRequest, server project.Project_ApplyServer) error {
	msger := newMessager(
		input.SendPercent,
		websocket.Type_ApplyProject,
		server,
	)

	ctx := server.Context()
	_, err := p.apply(
		ctx,
		biz.MustGetUser(ctx),
		msger,
		input,
		// streaming Apply 的请求类型没有 dry_run 字段，恒为实际部署。
		false,
	)

	return logError(ctx, p.logger, err)
}

// apply 是 Apply/WebApply 共用的部署编排入口：按需创建 websocket pubsub，
// 组装 JobInput 后委托 deploy.ApplyProject 执行（dryRun 由调用方透传）。
func (p *projectSvc) apply(
	ctx context.Context,
	user *biz.UserInfo,
	msger deploy.DeployMsger,
	input *project.ApplyRequest,
	dryRun bool,
) (deploy.Job, error) {
	// WebsocketSync 的 pubsub 属传输层关注点（来自 proto 请求字段），留在本层创建/关闭；
	// 其余编排（鉴权、仓库取回、git ensure、版本反查、Job 装配、ctx watcher、InstallProject）
	// 已下沉 deploy.ApplyProject，gRPC/WS 共享同一实现。
	var pubsub = deploy.NewEmptyPubSub()
	if input.WebsocketSync {
		pubsub = p.plMgr.Ws().New("", "")
	}
	defer pubsub.Close()

	jobInput := &deploy.JobInput{
		Type:           websocket.Type_ApplyProject,
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
		DryRun:         dryRun,
		PubSub:         pubsub,
		Messager:       msger,
	}

	return deploy.ApplyProject(ctx, deploy.ApplyProjectDeps{
		AccessBiz:  p.accessBiz,
		RepoBiz:    p.repoBiz,
		GitBiz:     p.gitBiz,
		ProjectBiz: p.projBiz,
		JobMgr:     p.jobManager,
		Logger:     p.logger,
	}, &deploy.ApplyProjectInput{JobInput: jobInput})
}

// Show 返回项目详情，响应前做项目级访问控制。
func (p *projectSvc) Show(ctx context.Context, request *project.ShowRequest) (*project.ShowResponse, error) {
	projectModel, err := p.accessBiz.RequireProjectAccess(ctx, int(request.Id))
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}

	return &project.ShowResponse{
		Item: transformer.FromProject(projectModel),
	}, nil
}

// MemoryCpuAndEndpoints 返回项目的 CPU/内存聚合用量与端点列表，响应前做项目级访问控制。
func (p *projectSvc) MemoryCpuAndEndpoints(ctx context.Context, req *project.MemoryCpuAndEndpointsRequest) (*project.MemoryCpuAndEndpointsResponse, error) {
	projectModel, err := p.accessBiz.RequireProjectAccess(ctx, int(req.Id))
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}
	cpu, memory := biz.ProjectCpuMemory(ctx, p.k8sBiz, projectModel)
	urls, err := p.projBiz.GetProjectEndpointsInNamespace(ctx, projectModel.Namespace.Name, projectModel.ID)
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}
	return &project.MemoryCpuAndEndpointsResponse{
		Urls:   urls,
		Cpu:    cpu,
		Memory: memory,
	}, nil
}

// Delete 删除项目（先卸载 release 再删 DB），响应前做项目级访问控制，落删除审计日志。
func (p *projectSvc) Delete(ctx context.Context, request *project.DeleteRequest) (*project.DeleteResponse, error) {
	proj, err := p.accessBiz.RequireProjectAccess(ctx, int(request.Id))
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}

	// 卸载顺序不变式（先卸载 release 成功才删 DB，失败保留记录可重试）已收敛到 biz.DeployBiz。
	if err := p.deployBiz.DeleteProject(ctx, int(request.Id), proj, p.logger.Debugf); err != nil {
		p.logger.ErrorCtx(ctx, "删除项目失败: "+proj.Name+"/"+proj.Namespace.Name, err)
		return nil, err
	}

	p.eventBiz.AuditLogWithRequest(
		types.EventActionType_Delete,
		biz.MustGetUser(ctx).Name,
		biz.MustGetUser(ctx).Email,
		fmt.Sprintf("删除项目: %d: %s/%s ", proj.ID, proj.Namespace.Name, proj.Name),
		request,
	)

	return &project.DeleteResponse{}, nil
}

// Version 返回项目当前部署版本号，响应前做项目级访问控制。
func (p *projectSvc) Version(ctx context.Context, req *project.VersionRequest) (*project.VersionResponse, error) {
	// 与 Show/MemoryCpuAndEndpoints/Delete/AllContainers 对齐：Version 也按 ProjectID
	// 解析资源，私有命名空间的项目版本（反映部署频率/活跃度）不允许被非授权用户探测。
	if _, err := p.accessBiz.RequireProjectAccess(ctx, int(req.Id)); err != nil {
		return nil, logError(ctx, p.logger, err)
	}

	v, err := p.projBiz.Version(ctx, int(req.Id))
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}

	return &project.VersionResponse{Version: int32(v)}, nil
}

// AllContainers 返回项目下全部活跃容器，响应前做项目级访问控制。
func (p *projectSvc) AllContainers(ctx context.Context, request *project.AllContainersRequest) (*project.AllContainersResponse, error) {
	if _, err := p.accessBiz.RequireProjectAccess(ctx, int(request.Id)); err != nil {
		return nil, logError(ctx, p.logger, err)
	}
	pods, err := p.projBiz.GetAllActiveContainers(ctx, int(request.Id))
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}

	return &project.AllContainersResponse{Items: pods}, nil
}

// CheckApplyStatus 判定项目最近一次部署后新版本容器是否正常运行（非原子 WebApply 用），
// 响应前做项目级访问控制，聚合状态/原因/容器明细/失败诊断一并返回。
func (p *projectSvc) CheckApplyStatus(ctx context.Context, request *project.CheckApplyStatusRequest) (*project.CheckApplyStatusResponse, error) {
	if _, err := p.accessBiz.RequireProjectAccess(ctx, int(request.Id)); err != nil {
		return nil, logError(ctx, p.logger, err)
	}
	status, err := p.projBiz.CheckApplyStatus(ctx, int(request.Id))
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}
	return &project.CheckApplyStatusResponse{
		Status:     status.Status,
		Reason:     status.Reason,
		Containers: status.Containers,
		Failures:   mapDomainFailures(status.Failures),
	}, nil
}

// ResourceTree 返回项目资源拓扑树（完整资源列表），响应前做项目级访问控制。
func (p *projectSvc) ResourceTree(ctx context.Context, request *project.ResourceTreeRequest) (*project.ResourceTreeResponse, error) {
	if _, err := p.accessBiz.RequireProjectAccess(ctx, int(request.Id)); err != nil {
		return nil, logError(ctx, p.logger, err)
	}
	tree, err := p.projBiz.ResourceTree(ctx, int(request.Id))
	if err != nil {
		return nil, logError(ctx, p.logger, err)
	}
	return mapDomainTree(tree), nil
}

// mapDomainTree 把领域 ResourceTree 映射为 proto ResourceTreeResponse。
func mapDomainTree(tree *biz.ResourceTree) *project.ResourceTreeResponse {
	nodes := make([]*project.ResourceTreeNode, 0, len(tree.Nodes))
	for _, n := range tree.Nodes {
		nodes = append(nodes, &project.ResourceTreeNode{
			Id:        n.ID,
			Kind:      n.Kind,
			Name:      n.Name,
			Namespace: n.Namespace,
			Status:    n.Status,
			Labels:    n.Labels,
			Old:       n.Old,
		})
	}
	edges := make([]*project.ResourceTreeEdge, 0, len(tree.Edges))
	for _, e := range tree.Edges {
		edges = append(edges, &project.ResourceTreeEdge{
			Id:     e.ID,
			Type:   e.Type,
			Source: e.Source,
			Target: e.Target,
		})
	}
	return &project.ResourceTreeResponse{Status: tree.Status, Nodes: nodes, Edges: edges}
}

// mapDomainFailures 把领域 ContainerFailure 转成 proto ContainerFailure（透传全部诊断字段）。
func mapDomainFailures(failures []*biz.ContainerFailure) []*project.ContainerFailure {
	out := make([]*project.ContainerFailure, 0, len(failures))
	for _, f := range failures {
		out = append(out, &project.ContainerFailure{
			Kind:      f.Kind,
			Workload:  f.Workload,
			Pod:       f.Pod,
			Container: f.Container,
			Reason:    f.Reason,
			Message:   f.Message,
			Logs:      f.Logs,
		})
	}
	return out
}

var _ deploy.DeployMsger = (*emptyMessager)(nil)

// emptyMessager 实现 deploy.DeployMsger 的空操作版本：所有推送方法均 no-op，
// 供 WebApply 等非流式场景使用（丢弃全部部署进度/日志帧）。
type emptyMessager struct{}

// newEmptyMessager 构造一个丢弃全部消息的空 messager。
func newEmptyMessager() *emptyMessager {
	return &emptyMessager{}
}

// Current 返回当前进度（恒为 0）。
func (e *emptyMessager) Current() int64 {
	return 0
}

// Add 进度空操作。
func (e *emptyMessager) Add() {
}

// To 空操作：直接设置进度被忽略。
func (e *emptyMessager) To(percent int64) {
}

// SendEndError 空实现。
func (e *emptyMessager) SendEndError(err error) {}

// SendMsg 空实现。
func (e *emptyMessager) SendMsg(s string) {}

// SendProtoMsg 空实现。
func (e *emptyMessager) SendProtoMsg(message app.WebsocketMessage) {}

// SendProcessPercent 空实现。
func (e *emptyMessager) SendProcessPercent(int64) {}

// SendMsgWithContainerLog 空实现。
func (e *emptyMessager) SendMsgWithContainerLog(msg string, containers []*websocket.Container) {}

// SendDeployedResult 空实现。
func (e *emptyMessager) SendDeployedResult(resultType websocket.ResultType, s string, p *types.ProjectModel) {
}

// SetSlug 空实现：丢弃全部帧的空 messager 无 slug 关联需求。
func (e *emptyMessager) SetSlug(string) {}

var _ deploy.DeployMsger = (*messager)(nil)

// messager 是 deploy.DeployMsger 的流式实现：按 sendPercent 决定是否推送进度帧，
// 将部署进度/日志/结果帧序列化后写入 gRPC 流，由 newMessager 构造。
type messager struct {
	percent     deploy.Percentable
	sendPercent bool

	slugName string
	t        websocket.Type
	server   project.Project_ApplyServer
}

// newMessager 构造带进度上报的 deploy 消息器：按 sendPercent 决定是否推送进度帧。
// slug 不在构造期绑定：部署帧 slug 依赖最终部署名（创建场景客户端不发 name），由共享
// 编排 ApplyProject 名解析后经 DeployMsger.SetSlug 统一回填。
func newMessager(sendPercent bool, t websocket.Type, server project.Project_ApplyServer) deploy.DeployMsger {
	m := messager{sendPercent: sendPercent, t: t, server: server}
	m.percent = deploy.NewProcessPercent(&m, deploy.NewRealSleeper())
	return &m
}

// Current 返回当前进度百分比。
func (m *messager) Current() int64 {
	return m.percent.Current()
}

// Add 进度 +1。
func (m *messager) Add() {
	m.percent.Add()
}

// To 直接设置进度百分比。
func (m *messager) To(percent int64) {
	m.percent.To(percent)
}

// SendDeployedResult 发送最终部署结果帧（成功/失败 + 项目模型）。
func (m *messager) SendDeployedResult(resultType websocket.ResultType, s string, p *types.ProjectModel) {
	m.send(&project.ApplyResponse{
		Metadata: m.metadata(m.t, resultType, true, s),
		Project:  p,
	})
}

// SendEndError 发送结束帧并携带错误信息。
func (m *messager) SendEndError(err error) {
	m.send(&project.ApplyResponse{Metadata: m.metadata(m.t, websocket.ResultType_Error, true, err.Error())})
}

// SendProcessPercent 按 sendPercent 开关推送进度百分比帧。
func (m *messager) SendProcessPercent(p int64) {
	if m.sendPercent {
		// Percent 是类型特有字段，metadata 收敛公共字段后单独覆盖。
		res := m.metadata(websocket.Type_ProcessPercent, websocket.ResultType_Success, false, "")
		res.Percent = int32(p)
		m.send(&project.ApplyResponse{Metadata: res})
	}
}

// SendMsg 发送普通消息帧（非结束）。
func (m *messager) SendMsg(s string) {
	m.send(&project.ApplyResponse{Metadata: m.metadata(m.t, websocket.ResultType_Success, false, s)})
}

// SendProtoMsg 透传原始 websocket 元数据帧。
func (m *messager) SendProtoMsg(message app.WebsocketMessage) {
	m.send(&project.ApplyResponse{Metadata: message.GetMetadata()})
}

// SendMsgWithContainerLog 发送携带容器日志的消息帧。
func (m *messager) SendMsgWithContainerLog(msg string, containers []*websocket.Container) {
	m.send(&project.ApplyResponse{Metadata: m.metadata(m.t, websocket.ResultType_LogWithContainers, false, msg)})
}

// SetSlug 就地更新流帧携带的部署标识 slug：部署名在共享编排 ApplyProject 中被缺省解析后
// 调用（创建部署未显式传名时），使客户端能按最终名关联部署日志。
func (m *messager) SetSlug(slug string) {
	m.slugName = slug
}

// metadata 收敛 5 个发送方法的公共字段（Slug/Type/Result/End/Message），与 websocket 侧
// messageSender.metadata 同一模式，消除逐方法重复拼装；Percent 等类型特有字段由调用方覆盖。
func (m *messager) metadata(t websocket.Type, result websocket.ResultType, end bool, msg string) *websocket.Metadata {
	return &websocket.Metadata{
		Slug:    m.slugName,
		Type:    t,
		Result:  result,
		End:     end,
		Message: msg,
	}
}

// send 把 ApplyResponse 帧下发到 gRPC 流。
func (m *messager) send(res *project.ApplyResponse) {
	m.server.Send(res)
}
