package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/samber/lo"

	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/locker"
	"github.com/duc-cnzj/mars/v4/internal/metrics"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/transformer"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var _ application.WsServer = (*WebsocketManager)(nil)

type HandleRequestFunc func(ctx context.Context, c Conn, t websocket_pb.Type, message []byte)

type WebsocketManager struct {
	healthTickDuration time.Duration

	data       data.Data
	logger     mlog.Logger
	pl         application.PluginManger
	auth       auth.Auth
	uploader   uploader.Uploader
	locker     locker.Locker
	jobManager JobManager

	projRepo  repo.ProjectRepo
	k8sRepo   repo.K8sRepo
	fileRepo  repo.FileRepo
	nsRepo    repo.NamespaceRepo
	eventRepo repo.EventRepo
	repoRepo  repo.RepoImp

	executor repo.ExecutorManager

	handlers map[websocket_pb.Type]HandleRequestFunc
}

func NewWebsocketManager(
	logger mlog.Logger,
	repoRepo repo.RepoImp,
	jobManager JobManager,
	data data.Data,
	pl application.PluginManger,
	auth auth.Auth,
	uploader uploader.Uploader,
	locker locker.Locker,
	clusterRepo repo.K8sRepo,
	eventRepo repo.EventRepo,
	executor repo.ExecutorManager,
	fileRepo repo.FileRepo,
) application.WsServer {
	mgr := &WebsocketManager{
		repoRepo:           repoRepo,
		jobManager:         jobManager,
		fileRepo:           fileRepo,
		healthTickDuration: 15 * time.Second,
		logger:             logger,
		pl:                 pl,
		auth:               auth,
		uploader:           uploader,
		locker:             locker,
		k8sRepo:            clusterRepo,
		eventRepo:          eventRepo,
		executor:           executor,
		data:               data,
	}
	mgr.handlers = map[websocket_pb.Type]HandleRequestFunc{
		WsAuthorize:          mgr.HandleAuthorize,
		WsHandleExecShell:    mgr.HandleStartShell,
		WsHandleExecShellMsg: mgr.HandleShellMessage,
		WsHandleCloseShell:   mgr.HandleCloseShell,
		WsCancel:             mgr.HandleWsCancelDeploy,
		ProjectPodEvent:      mgr.HandleJoinRoom,
		WsCreateProject:      mgr.HandleWsCreateProject,
		WsUpdateProject:      mgr.HandleWsUpdateProject,
	}
	return mgr
}

func (wc *WebsocketManager) TickClusterHealth(done <-chan struct{}) {
	ticker := time.NewTicker(wc.healthTickDuration)
	lock := wc.locker
	sub := wc.pl.Ws().New("", "")
	defer sub.Close()

	defer wc.logger.HandlePanic("TickClusterHealth")
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if lock.Acquire("TickClusterHealth", 5) {
				func() {
					defer lock.Release("TickClusterHealth")
					info := wc.k8sRepo.ClusterInfo()
					sub.ToAll(&websocket_pb.WsHandleClusterResponse{
						Metadata: &websocket_pb.Metadata{
							Type: WsClusterInfoSync,
						},
						Info: transformer.FromClusterInfo(info),
					})
				}()
			}
		case <-done:
			wc.logger.Info("[Websocket]: app shutdown and stop WsClusterInfoSync")
			return
		}
	}
}

func (wc *WebsocketManager) Info(writer http.ResponseWriter, request *http.Request) {
	sub := wc.pl.Ws().New("", "")
	defer sub.Close()
	marshal, _ := json.Marshal(sub.Info())
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(marshal)
}

func (wc *WebsocketManager) Serve(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wc.logger.Error(err)
		return
	}

	var (
		id  string = uuid.New().String()
		uid string = uuid.New().String()

		inputUid = r.URL.Query().Get("uid")
	)
	if inputUid != "" {
		uid = inputUid
	}

	wsConn := wc.newWsConn(uid, id, c, NewTaskManager(wc.logger), NewSessionMap(wc.logger))
	g, ctx := errgroup.WithContext(r.Context())

	defer func() {
		wc.logger.Debugf("[Websocket]: Serve exit")
		wsConn.Close(r.Context())
	}()

	g.Go(func() error {
		defer wc.logger.HandlePanic("[ProjectPodEventSubscriber]: Run")
		return wsConn.PubSub().Run(ctx)
	})

	g.Go(func() error {
		var err error
		defer func() {
			wc.logger.Debugf("[Websocket]: go write exit, err: %v", err)
		}()
		defer wc.logger.HandlePanic("Websocket: Write, err %v")

		return wc.write(ctx, wsConn)
	})

	NewMessageSender(wsConn, "", WsSetUid).SendMsg(wsConn.UID())

	g.Go(func() error {
		var err error
		defer func() {
			wc.logger.Debugf("[Websocket]: go read exit, err: %v", err)
		}()
		defer wc.logger.HandlePanic("[Websocket]: read recovery")
		return wc.read(ctx, wsConn)
	})

	if err = g.Wait(); err != nil {
		return
	}
}

func (wc *WebsocketManager) read(ctx context.Context, wsconn Conn) error {
	wsconn.SetReadLimit(maxMessageSize)
	wsconn.SetReadDeadline(time.Now().Add(pongWait))
	wsconn.SetPongHandler(func(string) error {
		wsconn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var wsRequest websocket_pb.WsRequestMetadata
		_, message, err := wsconn.ReadMessage()
		if err != nil {
			wc.logger.Debugf("[Websocket]: read error: %v", err)
			return err
		}
		if err := proto.Unmarshal(message, &wsRequest); err != nil {
			NewMessageSender(wsconn, "", WsInternalError).SendEndError(err)

			continue
		}

		go wc.dispatchEvent(ctx, wsconn, &wsRequest, message)
	}
}

func (wc *WebsocketManager) write(ctx context.Context, wsconn Conn) (err error) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		wc.logger.Debugf("[Websocket]: go write exit, %v", err)
		ticker.Stop()
		wsconn.Close(ctx)
	}()
	wc.logger.Debug(wsconn.PubSub().ID(), wsconn.PubSub().Uid())
	ch := wsconn.PubSub().Subscribe()
	var w io.WriteCloser
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message, ok := <-ch:
			if !ok {
				return wsconn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			wsconn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err = wsconn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return err
			}
			w.Write(message)

			if err = w.Close(); err != nil {
				return err
			}
		case <-ticker.C:
			wsconn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = wsconn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}
		}
	}
}

func (wc *WebsocketManager) dispatchEvent(ctx context.Context, wsconn Conn, wsRequest *websocket_pb.WsRequestMetadata, message []byte) {
	if handler, ok := wc.handlers[wsRequest.Type]; ok {
		defer wc.logger.HandlePanicWithCallback(wsRequest.Type.String(), func(err error) {
			metrics.WebsocketPanicCount.With(prometheus.Labels{"method": wsRequest.Type.String()}).Inc()
		})
		defer func(t time.Time) {
			metrics.WebsocketRequestLatency.With(prometheus.Labels{"method": wsRequest.Type.String()}).Observe(time.Since(t).Seconds())
			e := recover()
			if e == nil {
				metrics.WebsocketRequestTotalSuccess.With(prometheus.Labels{"method": wsRequest.Type.String()}).Inc()
			} else {
				metrics.WebsocketRequestTotalFail.With(prometheus.Labels{"method": wsRequest.Type.String()}).Inc()
				panic(e)
			}
		}(time.Now())

		wc.logger.Warningf("wsType: %v, message: %v", wsRequest.Type.String(), string(message))

		// websocket.onopen 事件不一定是最早发出来的，所以要等 onopen 的认证结束后才能进行后面的操作
		if wsconn.GetUser() == nil && wsRequest.Type != websocket_pb.Type_HandleAuthorize {
			NewMessageSender(wsconn, "", WsAuthorize).SendMsg("认证中，请稍等~")
			return
		}
		handler(ctx, wsconn, wsRequest.Type, message)
	}
}

func (wc *WebsocketManager) HandleAuthorize(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.AuthorizeTokenInput
	if err := proto.Unmarshal(message, &input); err != nil {
		wc.logger.Error("[Websocket]: " + err.Error())
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}

	if claims, b := wc.auth.VerifyToken(input.Token); b {
		c.SetUser(claims.UserInfo)
		metrics.WebsocketConnectionsCount.With(prometheus.Labels{"username": claims.UserInfo.Name}).Inc()
	}
}

func (wc *WebsocketManager) HandleJoinRoom(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.ProjectPodEventJoinInput
	if err := proto.Unmarshal(message, &input); err != nil {
		wc.logger.Error("[Websocket]: " + err.Error())
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}
	if input.Join {
		c.PubSub().(application.ProjectPodEventSubscriber).Join(int64(input.GetProjectId()))
	} else {
		c.PubSub().(application.ProjectPodEventSubscriber).Leave(int64(input.GetNamespaceId()), int64(input.GetProjectId()))
	}
}

func (wc *WebsocketManager) HandleStartShell(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.WsHandleExecShellInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)
		return
	}
	sessionID, err := wc.StartShell(ctx, &input, c)
	if err != nil {
		wc.logger.Error(err)
		NewMessageSender(c, "", WsHandleExecShell).SendEndError(err)
		return
	}

	wc.logger.Debugf("[Websocket]: 收到 初始化连接 WsHandleExecShell 消息, sessionID: %v", sessionID)

	NewMessageSender(c, "", WsHandleExecShell).SendProtoMsg(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     c.ID(),
			Uid:    c.UID(),
			Type:   WsHandleExecShell,
			Result: ResultSuccess,
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

func (wc *WebsocketManager) HandleShellMessage(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.TerminalMessageInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}

	if pty, ok := c.GetPtyHandler(input.Message.SessionId); ok {
		if err := pty.Send(ctx, input.Message); err != nil {
			wc.logger.Error("[Websocket]: " + err.Error())
		}
	}
}

func (wc *WebsocketManager) HandleCloseShell(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.TerminalMessageInput
	if err := proto.Unmarshal(message, &input); err != nil {
		wc.logger.Error(err.Error())
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}
	msg := fmt.Sprintf("[Websocket]: %v 收到客户端主动断开的消息", input.Message.SessionId)
	wc.logger.Debugf(msg)
	c.ClosePty(ctx, input.Message.SessionId, 0, msg)
}

func (wc *WebsocketManager) HandleWsCreateProject(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.CreateProjectInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}

	appName := lo.FromPtr(input.Name)
	if err := wc.upgradeOrInstall(ctx, c, &JobInput{
		Type:        input.Type,
		NamespaceId: input.NamespaceId,
		Name:        appName,
		RepoID:      input.RepoId,
		// TODO
		//GitProjectId: input.GitProjectId,
		GitBranch:   input.GitBranch,
		GitCommit:   input.GitCommit,
		Config:      input.Config,
		Atomic:      input.Atomic,
		ExtraValues: input.ExtraValues,
		User:        c.GetUser(),
		PubSub:      c.PubSub(),
		Messager:    NewMessageSender(c, util.GetSlugName(input.NamespaceId, appName), t),
	}); err != nil {
		wc.logger.Error(err)
	}
}

func (wc *WebsocketManager) HandleWsUpdateProject(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.UpdateProjectInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)
		return
	}

	p, err := wc.projRepo.Show(ctx, int(input.ProjectId))
	if err != nil {
		NewMessageSender(c, "", t).SendEndError(err)
		return
	}

	wc.upgradeOrInstall(ctx, c, &JobInput{
		Type:        t,
		NamespaceId: int32(p.NamespaceID),
		Name:        p.Name,
		//TODO
		//GitProjectId:   int32(p.GitProjectID),
		GitBranch:      input.GitBranch,
		GitCommit:      input.GitCommit,
		Config:         input.Config,
		Atomic:         input.Atomic,
		ExtraValues:    input.ExtraValues,
		Version:        &input.Version,
		TimeoutSeconds: int32(wc.data.Config().InstallTimeout.Seconds()),
		User:           c.GetUser(),
		PubSub:         c.PubSub(),
	})
}

func (wc *WebsocketManager) HandleWsCancelDeploy(ctx context.Context, c Conn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.CancelInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}

	var slugName = util.GetSlugName(input.NamespaceId, input.Name)

	if err := c.CancelTask(slugName); err == nil {
		ns, _ := wc.nsRepo.Show(ctx, int(input.NamespaceId))
		wc.eventRepo.AuditLog(
			types.EventActionType_CancelDeploy,
			c.GetUser().Name,
			fmt.Sprintf("用户取消部署 namespace: %s, 服务 %s.", ns.Name, input.Name))
	}
}

func (wc *WebsocketManager) upgradeOrInstall(ctx context.Context, c Conn, input *JobInput) error {
	slug := util.GetSlugName(input.NamespaceId, input.Name)
	job := wc.jobManager.NewJob(input)

	if input.IsNotDryRun() {
		if err := c.AddTask(job.ID(), job.Stop); err != nil {
			NewMessageSender(c, slug, input.Type).SendDeployedResult(ResultDeployFailed, "正在清理中，请稍后再试。", nil)
			return nil
		}
		job.OnFinally(1000, func(err error, base func()) {
			c.StopTask(job.ID())
			base()
		})
	}
	return InstallProject(ctx, job)
}

func InstallProject(ctx context.Context, job Job) (err error) {
	return job.GlobalLock().Validate().LoadConfigs().Run(ctx).Finish().Error()
}
