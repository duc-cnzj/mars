package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
)

// 本文件覆盖 ws 协议的消息处理：
//   - 入站 handler：HandleAuthorize/HandleJoinRoom/HandleStartShell/HandleShellMessage/
//     HandleCloseShell/HandleCreateProject/HandleUpdateProject/HandleCancelDeploy/installProject
//   - 出站消息适配器：messageSender（deploy.DeployMsger 实现，向客户端推送协议帧）

// ---- 入站协议 handler ----

// HandleAuthorize 处理认证帧：校验 token，成功则写入连接用户信息并递增连接计数。
// 复用 biz.AuthBiz 统一鉴权核心（与 gRPC/HTTP 一致），失败时与 master 行为对齐：
// 静默忽略、不发送任何帧。
func (wc *websocketManager) HandleAuthorize(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.AuthorizeTokenInput
	if err := proto.Unmarshal(message, &input); err != nil {
		wc.logger.Error("[Websocket]: " + err.Error())
		newMessageSender(c, "", t).SendEndError(err)

		return
	}

	if user, err := wc.authBiz.VerifyToken(ctx, input.Token); err == nil {
		c.SetUser(user)
		metrics.WebsocketConnectionsCount.With(prometheus.Labels{"username": user.Name}).Inc()
	}
}

// HandleJoinRoom 处理项目事件订阅帧：按 Join 标志加入或离开项目的订阅房间。
// Join（订阅 Pod 事件流）前必须做项目级访问控制：事件流含部署动态
// （创建/删除/更新/版本），与 project.Show/AllContainers 对齐，防止任意已认证
// 用户枚举 project_id 订阅私有项目事件（IDOR）。Leave（退订）无泄露面不拦截。
func (wc *websocketManager) HandleJoinRoom(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.ProjectPodEventJoinInput
	if err := proto.Unmarshal(message, &input); err != nil {
		wc.logger.Error("[Websocket]: " + err.Error())
		newMessageSender(c, "", t).SendEndError(err)

		return
	}
	// PubSub 接口本身就嵌入了 ProjectPodEventSubscriber（application.Plugins.go），
	// 无需运行时断言即可直接调 Join/Leave。
	wc.logger.Debug("HandleJoinRoom: ", input.String())
	if input.Join {
		// 订阅前鉴权：RequireProjectAccess 加载项目并校验所属命名空间可访问性，
		// 无权访问回错误帧且不触达 PubSub.Join，防止私有项目事件流泄露。
		if _, err := wc.accessBiz.RequireProjectAccess(ctx, int(input.ProjectId)); err != nil {
			wc.logger.Error("[Websocket]: join room permission denied: ", err)
			newMessageSender(c, "", t).SendEndError(err)
			return
		}
		if err := c.PubSub().Join(int64(input.GetProjectId())); err != nil {
			wc.logger.Error("join: ", err, input.String())
		}
		return
	}
	if err := c.PubSub().Leave(int64(input.GetNamespaceId()), int64(input.GetProjectId())); err != nil {
		wc.logger.Error("leave: ", err, input.String())
	}
}

// HandleStartShell 处理开启终端帧：拉起容器内 shell，成功则回带 sessionID 的响应帧。
func (wc *websocketManager) HandleStartShell(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.WsHandleExecShellInput
	if err := proto.Unmarshal(message, &input); err != nil {
		newMessageSender(c, "", t).SendEndError(err)
		return
	}
	// 拉起容器内 shell 前必须做命名空间级访问控制：交互终端可执行任意命令，
	// 与 gRPC container.Exec 对齐（containerSvc.Exec 首行 RequireNamespaceAccessByName），
	// 防止任意已认证用户枚举 namespace/pod/container 进入私有命名空间的容器 shell（RCE）。
	if _, err := wc.accessBiz.RequireNamespaceAccessByName(ctx, input.Container.GetNamespace()); err != nil {
		wc.logger.Error("[Websocket]: start shell permission denied: ", err)
		newMessageSender(c, "", t).SendEndError(err)
		return
	}
	sessionID, err := wc.StartShell(ctx, &input, c)
	if err != nil {
		wc.logger.Error(err)
		newMessageSender(c, "", WsHandleExecShell).SendEndError(err)
		return
	}

	wc.logger.Debugf("[Websocket]: 收到 初始化连接 WsHandleExecShell 消息, sessionID: %v", sessionID)

	newMessageSender(c, "", WsHandleExecShell).SendProtoMsg(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     c.ID(),
			Uid:    c.UID(),
			Type:   WsHandleExecShell,
			Result: deploy.ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			SessionId: sessionID,
		},
		Container: &websocket_pb.Container{
			Namespace: input.Container.Namespace,
			Pod:       input.Container.Pod,
			Container: input.Container.Container,
		},
	})
}

// HandleShellMessage 处理终端输入/resize 帧：转发给对应会话的 pty 处理器。
func (wc *websocketManager) HandleShellMessage(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.TerminalMessageInput
	if err := proto.Unmarshal(message, &input); err != nil {
		newMessageSender(c, "", t).SendEndError(err)

		return
	}

	if pty, ok := c.GetPtyHandler(input.Message.SessionId); ok {
		if err := pty.Send(ctx, input.Message); err != nil {
			wc.logger.Error("[Websocket]: " + err.Error())
		}
	}
}

// HandleCloseShell 处理客户端主动断开终端帧：关闭对应 pty 会话。
func (wc *websocketManager) HandleCloseShell(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.TerminalMessageInput
	if err := proto.Unmarshal(message, &input); err != nil {
		wc.logger.Error(err.Error())
		newMessageSender(c, "", t).SendEndError(err)

		return
	}
	msg := fmt.Sprintf("[Websocket]: %v 收到客户端主动断开的消息", input.Message.SessionId)
	wc.logger.Debugf(msg)
	c.ClosePty(ctx, input.Message.SessionId, 0, msg)
}

// HandleCreateProject 处理创建项目帧：组装 JobInput 并执行部署流水线。
func (wc *websocketManager) HandleCreateProject(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.CreateProjectInput
	if err := proto.Unmarshal(message, &input); err != nil {
		newMessageSender(c, "", t).SendEndError(err)

		return
	}

	// 空 name 的"以仓库名缺省"解析由 deploy.ApplyProject 统一收敛，
	// 这里不再重复取 repo，直接透传原始 name——与 gRPC project.Apply 的 slug 行为对齐。
	// 仓库取回失败（含空 name + 无效 repoID）由 ApplyProject 兜底返回错误，不 panic。
	name := input.GetName()

	// JobInput.Type 取路由帧类型 t（而非消息体 input.Type）：客户端发送的是裸消息体，
	// 服务端 read() 把同一份字节 unmarshal 成 WsRequestMetadata 取 field1 做路由，故
	// t 与 input.Type 恒等；以 t 为权威可防"帧 Create + 体 Update"绕过 typeValidated
	// 的语义，且与 HandleUpdateProject 的 Type: t 对齐。
	if err := wc.installProject(ctx, c, &deploy.JobInput{
		Type:        t,
		NamespaceId: input.NamespaceId,
		Name:        name,
		RepoID:      input.RepoId,
		GitBranch:   input.GitBranch,
		GitCommit:   input.GitCommit,
		Config:      input.Config,
		Atomic:      input.Atomic,
		ExtraValues: input.ExtraValues,
		User:        c.GetUser(),
		PubSub:      c.PubSub(),
		Messager:    newMessageSender(c, deploy.GetSlugName(input.NamespaceId, name), t),
	}); err != nil {
		wc.logger.Error(err)
	}
}

// HandleUpdateProject 处理更新项目帧：按现有项目信息组装 JobInput 执行部署。
func (wc *websocketManager) HandleUpdateProject(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.UpdateProjectInput
	if err := proto.Unmarshal(message, &input); err != nil {
		newMessageSender(c, "", t).SendEndError(err)
		return
	}

	// 更新前先做项目级访问控制（RequireProjectAccess = Show + 所属命名空间校验），
	// 与 HandleJoinRoom 的订阅门卫对齐：私有命名空间的项目携带完整部署配置，
	// 未授权用户不得通过"更新失败帧 vs 部署静默拒绝"的可观测差异探测项目 ID 存在性（IDOR）。
	p, err := wc.accessBiz.RequireProjectAccess(ctx, int(input.ProjectId))
	if err != nil {
		wc.logger.Error("[Websocket]: update project permission denied: ", err)
		newMessageSender(c, "", t).SendEndError(err)
		return
	}

	wc.logger.Warning("update project", input.String())

	if err := wc.installProject(ctx, c, &deploy.JobInput{
		Type:           t,
		NamespaceId:    int32(p.NamespaceID),
		Name:           p.Name,
		RepoID:         int32(p.RepoID),
		GitBranch:      input.GitBranch,
		GitCommit:      input.GitCommit,
		Config:         input.Config,
		Atomic:         input.Atomic,
		ExtraValues:    input.ExtraValues,
		Version:        lo.ToPtr(input.Version),
		ProjectID:      input.ProjectId,
		TimeoutSeconds: int32(wc.config.InstallTimeout.Seconds()),
		User:           c.GetUser(),
		PubSub:         c.PubSub(),
		Messager:       newMessageSender(c, deploy.GetSlugName(p.NamespaceID, p.Name), t),
	}); err != nil {
		// 与 HandleCreateProject 一致：部署结果会经 SendDeployedResult 推给客户端，
		// 这里补服务端日志，避免更新路径的失败在可观测性上成为盲区。
		wc.logger.Error(err)
	}
}

// HandleCancelDeploy 处理取消部署帧：触发对应部署任务的取消回调并记录审计日志。
func (wc *websocketManager) HandleCancelDeploy(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.CancelInput
	if err := proto.Unmarshal(message, &input); err != nil {
		newMessageSender(c, "", t).SendEndError(err)

		return
	}

	var slugName = deploy.GetSlugName(input.NamespaceId, input.Name)

	if err := c.RunCancelDeployTask(slugName); err == nil {
		ns, err := wc.nsRepo.Show(ctx, int(input.NamespaceId))
		if err != nil {
			wc.logger.Error(err, input.NamespaceId)
			return
		}
		wc.eventRepo.AuditLog(
			types.EventActionType_CancelDeploy,
			c.GetUser().Name,
			fmt.Sprintf("用户取消部署 namespace: %s, 服务 %s.", ns.Name, input.Name))
	}
}

// installProject 是创建/更新项目的公共入口：复用 deploy.ApplyProject 的共享编排
// （鉴权、仓库取回、git ensure、版本反查、Job 装配、ctx watcher、InstallProject）。
// 本层仅以 OnJob 钩子登记取消回调（任务已存在则回失败帧，ApplyProject 据此跳过流水线），
// 并在部署结束后移除回调。相比旧实现补齐了命名空间访问校验与连接断开的部署取消。
func (wc *websocketManager) installProject(ctx context.Context, c Conn, input *deploy.JobInput) error {
	indent, _ := json.MarshalIndent(input, "", "  ")
	wc.logger.Debug(string(indent))

	_, err := deploy.ApplyProject(ctx, deploy.ApplyProjectDeps{
		// 复用注入的 wc.accessBiz（用户提取内部走 MustGetUser）：ApplyProject 入口已把
		// JobInput.User 物化进 ctx，WS 的 user（Conn）与 gRPC 一致从 ctx 取值。
		AccessBiz:  wc.accessBiz,
		RepoBiz:    wc.repoBiz,
		GitBiz:     wc.gitBiz,
		ProjectBiz: wc.projBiz,
		JobMgr:     wc.jobManager,
		Logger:     wc.logger,
	}, &deploy.ApplyProjectInput{
		JobInput: input,
		OnJob: func(job deploy.Job) error {
			if input.IsNotDryRun() {
				if err := c.AddCancelDeployTask(job.ID(), job.Stop); err != nil {
					newMessageSender(c, deploy.GetSlugName(input.NamespaceId, input.Name), input.Type).
						SendDeployedResult(deploy.ResultDeployFailed, "正在清理中，请稍后再试。", nil)
					// 传输层已发失败帧 → ApplyProject 跳过 watcher 与 InstallProject。
					return err
				}
				job.OnFinally(1000, func(err error, base func()) {
					c.RemoveCancelDeployTask(job.ID())
					base()
				})
			}
			return nil
		},
	})
	return err
}

// ---- 出站消息适配器（deploy.DeployMsger） ----

// wsResponse 是 ws 元数据响应的语义别名（省去长包名前缀）。
type wsResponse = websocket_pb.WsMetadataResponse

var _ deploy.DeployMsger = (*messageSender)(nil)

// messageSender 实现 deploy.DeployMsger：把部署进度/结果/日志以协议帧推给
// 指定连接的发布订阅，是 ws 侧对接部署流水线的出站消息适配器。
type messageSender struct {
	conn     Conn
	slugName string
	wsType   websocket_pb.Type

	percent deploy.Percentable
}

// newMessageSender 构造面向某连接的出站消息适配器，并绑定进度上报器。
func newMessageSender(
	conn Conn,
	slugName string,
	wsType websocket_pb.Type,
) deploy.DeployMsger {
	m := &messageSender{
		conn:     conn,
		slugName: slugName,
		wsType:   wsType,
	}
	m.percent = deploy.NewProcessPercent(m, deploy.NewRealSleeper())
	return m
}

// SendDeployedResult 推送部署最终结果帧（End=true）。
func (ms *messageSender) SendDeployedResult(result websocket_pb.ResultType, msg string, p *types.ProjectModel) {
	ms.send(&wsResponse{Metadata: ms.metadata(ms.wsType, result, true, msg)})
}

// SendEndError 推送部署错误帧（Result=Error, End=true）。
func (ms *messageSender) SendEndError(err error) {
	ms.send(&wsResponse{Metadata: ms.metadata(ms.wsType, deploy.ResultError, true, err.Error())})
}

// SendProcessPercent 推送部署进度百分比帧。
func (ms *messageSender) SendProcessPercent(percent int64) {
	meta := ms.metadata(WsProcessPercent, deploy.ResultSuccess, false, "")
	meta.Percent = int32(percent)
	ms.send(&wsResponse{Metadata: meta})
}

// SendMsg 推送普通消息帧。
func (ms *messageSender) SendMsg(msg string) {
	ms.send(&wsResponse{Metadata: ms.metadata(ms.wsType, deploy.ResultSuccess, false, msg)})
}

// SendMsgWithContainerLog 推送带容器列表的日志帧（供跳转容器日志）。
func (ms *messageSender) SendMsgWithContainerLog(msg string, containers []*websocket_pb.Container) {
	ms.send(&websocket_pb.WsWithContainerMessageResponse{
		Metadata:   ms.metadata(ms.wsType, deploy.ResultLogWithContainers, false, msg),
		Containers: containers,
	})
}

// metadata 收敛 5 个发送方法的公共字段（Slug/Type/Result/End/Uid/Id/Message），
// 消除逐方法重复拼装；类型特有字段（Percent 等）由调用方覆盖。
func (ms *messageSender) metadata(wsType websocket_pb.Type, result websocket_pb.ResultType, end bool, msg string) *websocket_pb.Metadata {
	return &websocket_pb.Metadata{
		Slug:    ms.slugName,
		Type:    wsType,
		Result:  result,
		End:     end,
		Uid:     ms.conn.UID(),
		Id:      ms.conn.ID(),
		Message: msg,
	}
}

// SendProtoMsg 推送任意协议消息（WebsocketMessage）。
func (ms *messageSender) SendProtoMsg(msg application.WebsocketMessage) {
	ms.send(msg)
}

// send 把消息写入连接的发布订阅（ToSelf）。
func (ms *messageSender) send(res application.WebsocketMessage) {
	ms.conn.PubSub().ToSelf(res)
}

// Current 返回当前部署进度百分比。
func (ms *messageSender) Current() int64 {
	return ms.percent.Current()
}

// Add 步进一次进度（推进一个处理阶段）。
func (ms *messageSender) Add() {
	ms.percent.Add()
}

// To 把进度直接置为指定百分比。
func (ms *messageSender) To(percent int64) {
	ms.percent.To(percent)
}
