package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v4/cluster"
	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/locker"
	"github.com/duc-cnzj/mars/v4/internal/metrics"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

type HandleRequestFunc func(ctx context.Context, c *WsConn, t websocket_pb.Type, message []byte)

type WsConn struct {
	id     string
	uid    string
	conn   contracts.WebsocketConn
	pubSub application.PubSub

	userMu sync.RWMutex
	user   *auth.UserInfo

	cs CancelSignaler
	sm SessionMapper
}

func (wc *WebsocketManager) initConn(uid, id string, c *websocket.Conn) *WsConn {
	return &WsConn{
		id:     id,
		uid:    uid,
		conn:   c,
		pubSub: wc.pl.Ws().New(uid, id),
		cs:     NewCancelSignal(),
		sm:     NewSessionMap(wc.logger),
	}
}

func (c *WsConn) Shutdown() {
	//c.cancelSignaler.CancelAll()
	//c.terminalSessions.CloseAll()
	c.pubSub.Close()
	c.conn.Close()
	metrics.WebsocketConnectionsCount.With(prometheus.Labels{"username": c.GetUser().Name}).Dec()
	Wait.Dec()
}

func (c *WsConn) SetUser(info *auth.UserInfo) {
	c.userMu.Lock()
	defer c.userMu.Unlock()
	c.user = info
}

func (c *WsConn) GetUser() *auth.UserInfo {
	c.userMu.RLock()
	defer c.userMu.RUnlock()
	return c.user
}

type WebsocketManager struct {
	healthTickDuration time.Duration

	logger     mlog.Logger
	k8sCli     *data.K8sClient
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

	executor repo.ExecutorManager
	db       *ent.Client

	handlers map[websocket_pb.Type]HandleRequestFunc
}

func NewWebsocketManager(
	logger mlog.Logger,
	jobManager JobManager,
	data *data.Data,
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
		jobManager:         jobManager,
		fileRepo:           fileRepo,
		healthTickDuration: 15 * time.Second,
		logger:             logger,
		k8sCli:             data.K8sClient,
		pl:                 pl,
		auth:               auth,
		uploader:           uploader,
		locker:             locker,
		k8sRepo:            clusterRepo,
		eventRepo:          eventRepo,
		executor:           executor,
		db:                 data.DB,
	}
	mgr.handlers = map[websocket_pb.Type]HandleRequestFunc{
		WsAuthorize:          mgr.HandleWsAuthorize,
		WsHandleExecShell:    mgr.HandleStartShell,
		WsHandleExecShellMsg: mgr.HandleWsShellMsg,
		WsHandleCloseShell:   mgr.HandleWsCloseShell,
		WsCancel:             mgr.HandleWsCancelDeploy,
		ProjectPodEvent:      mgr.HandleWsProjectPodEvent,
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
						Info: &cluster.InfoResponse{
							Status:            info.Status,
							FreeMemory:        info.FreeMemory,
							FreeCpu:           info.FreeCpu,
							FreeRequestMemory: info.FreeRequestMemory,
							FreeRequestCpu:    info.FreeRequestCpu,
							TotalMemory:       info.TotalMemory,
							TotalCpu:          info.TotalCpu,
							UsageMemoryRate:   info.UsageMemoryRate,
							UsageCpuRate:      info.UsageCpuRate,
							RequestMemoryRate: info.RequestMemoryRate,
							RequestCpuRate:    info.RequestCpuRate,
						},
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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
	if inputUid == "" {
		uid = inputUid
	}

	wsConn := wc.initConn(uid, id, c)
	g, ctx := errgroup.WithContext(r.Context())

	defer func() {
		wc.logger.Debugf("[Websocket]: Serve exit")
		wsConn.Shutdown()
	}()

	g.Go(func() error {
		defer wc.logger.HandlePanic("[ProjectPodEventSubscriber]: Run")
		return wsConn.pubSub.Run(ctx)
	})

	g.Go(func() error {
		var err error
		defer func() {
			wc.logger.Debugf("[Websocket]: go write exit, err: %v", err)
		}()
		defer wc.logger.HandlePanic("Websocket: Write, err %v")

		return wc.write(ctx, wsConn)
	})

	NewMessageSender(wsConn, "", WsSetUid).SendMsg(wsConn.uid)

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

func (wc *WebsocketManager) write(ctx context.Context, wsconn *WsConn) (err error) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		wc.logger.Debugf("[Websocket]: go write exit, %v", err)
		ticker.Stop()
		wsconn.conn.Close()
	}()
	ch := wsconn.pubSub.Subscribe()
	var w io.WriteCloser
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message, ok := <-ch:
			if !ok {
				return wsconn.conn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			wsconn.conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err = wsconn.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return err
			}
			w.Write(message)

			if err = w.Close(); err != nil {
				return err
			}
		case <-ticker.C:
			wsconn.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = wsconn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}
		}
	}
}

func (wc *WebsocketManager) read(ctx context.Context, wsconn *WsConn) error {
	wsconn.conn.SetReadLimit(maxMessageSize)
	wsconn.conn.SetReadDeadline(time.Now().Add(pongWait))
	wsconn.conn.SetPongHandler(func(string) error {
		wsconn.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var wsRequest websocket_pb.WsRequestMetadata
		_, message, err := wsconn.conn.ReadMessage()
		if err != nil {
			wc.logger.Debugf("[Websocket]: read error: %v", err)
			return err
		}
		if err := proto.Unmarshal(message, &wsRequest); err != nil {
			NewMessageSender(wsconn, "", WsInternalError).SendEndError(err)

			continue
		}

		go wc.handleWsRead(ctx, wsconn, &wsRequest, message)
	}
}

func (wc *WebsocketManager) handleWsRead(ctx context.Context, wsconn *WsConn, wsRequest *websocket_pb.WsRequestMetadata, message []byte) {
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

		// websocket.onopen 事件不一定是最早发出来的，所以要等 onopen 的认证结束后才能进行后面的操作
		if wsconn.GetUser() == nil || (wsconn.GetUser() != nil && wsconn.GetUser().GetID() == "" && wsRequest.Type != websocket_pb.Type_HandleAuthorize) {
			NewMessageSender(wsconn, "", WsAuthorize).SendMsg("认证中，请稍等~")
			return
		}
		handler(ctx, wsconn, wsRequest.Type, message)
	}
}

func (wc *WebsocketManager) HandleWsAuthorize(ctx context.Context, c *WsConn, t websocket_pb.Type, message []byte) {
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

func (wc *WebsocketManager) HandleWsProjectPodEvent(ctx context.Context, c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.ProjectPodEventJoinInput
	if err := proto.Unmarshal(message, &input); err != nil {
		wc.logger.Error("[Websocket]: " + err.Error())
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}
	if input.Join {
		c.pubSub.(application.ProjectPodEventSubscriber).Join(input.GetProjectId())
	} else {
		c.pubSub.(application.ProjectPodEventSubscriber).Leave(input.GetNamespaceId(), input.GetProjectId())
	}
}

func (wc *WebsocketManager) HandleWsCloseShell(ctx context.Context, c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.TerminalMessageInput
	if err := proto.Unmarshal(message, &input); err != nil {
		wc.logger.Error(err.Error())
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}
	msg := fmt.Sprintf("[Websocket]: %v 收到客户端主动断开的消息", input.Message.SessionId)
	wc.logger.Debugf(msg)
	c.sm.Close(ctx, input.Message.SessionId, 0, msg)
}

func (wc *WebsocketManager) HandleWsShellMsg(ctx context.Context, c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.TerminalMessageInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}

	if pty, ok := c.sm.Get(input.Message.SessionId); ok {
		if err := pty.Send(ctx, input.Message); err != nil {
			pty.Close(ctx, err.Error())
		}
	}
}

func (wc *WebsocketManager) HandleStartShell(ctx context.Context, c *WsConn, t websocket_pb.Type, message []byte) {
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
			Id:     c.id,
			Uid:    c.uid,
			Type:   WsHandleExecShell,
			Result: ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			SessionId: sessionID,
		},
		Container: &types.Container{
			Namespace: input.Container.Namespace,
			Pod:       input.Container.Pod,
			Container: input.Container.Container,
		},
	})
}

func (wc *WebsocketManager) HandleWsCancelDeploy(ctx context.Context, c *WsConn, t websocket_pb.Type, message []byte) {
	var (
		input websocket_pb.CancelInput
		cs    = c.cs
	)
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}

	var slugName = util.GetSlugName(input.NamespaceId, input.Name)
	if cs.Has(slugName) {
		ns, _ := wc.nsRepo.Show(ctx, int(input.NamespaceId))
		wc.eventRepo.AuditLogWithChange(types.EventActionType_CancelDeploy, c.GetUser().Name, fmt.Sprintf("用户取消部署 namespace: %s, 服务 %s.", ns.Name, input.Name), nil, nil)
		cs.Cancel(slugName)
	}
}

func (wc *WebsocketManager) HandleWsCreateProject(ctx context.Context, c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.CreateProjectInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}
	wc.upgradeOrInstall(ctx, c, &JobInput{
		Type:         input.Type,
		NamespaceId:  input.NamespaceId,
		Name:         input.Name,
		GitProjectId: input.GitProjectId,
		GitBranch:    input.GitBranch,
		GitCommit:    input.GitCommit,
		Config:       input.Config,
		Atomic:       input.Atomic,
		ExtraValues:  input.ExtraValues,
		User:         c.GetUser(),
		PubSub:       c.pubSub,
	})
}

func (wc *WebsocketManager) HandleWsUpdateProject(ctx context.Context, c *WsConn, t websocket_pb.Type, message []byte) {
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
		Type:           t,
		NamespaceId:    int64(p.NamespaceID),
		Name:           p.Name,
		GitProjectId:   int64(p.GitProjectID),
		GitBranch:      input.GitBranch,
		GitCommit:      input.GitCommit,
		Config:         input.Config,
		Atomic:         input.Atomic,
		ExtraValues:    input.ExtraValues,
		Version:        input.Version,
		TimeoutSeconds: 0,
		User:           c.GetUser(),
		PubSub:         c.pubSub,
	})
}

func (wc *WebsocketManager) upgradeOrInstall(ctx context.Context, c *WsConn, input *JobInput) error {
	slug := util.GetSlugName(input.NamespaceId, input.Name)
	job := wc.jobManager.NewJob(input)
	var cs = c.cs

	if input.IsNotDryRun() {
		if err := cs.Add(job.ID(), job.Stop); err != nil {
			NewMessageSender(c, slug, input.Type).SendDeployedResult(ResultDeployFailed, "正在清理中，请稍后再试。", nil)
			return nil
		}
		job.OnFinally(1000, func(err error, base func()) {
			cs.Remove(job.ID())
			base()
		})
	}
	return InstallProject(ctx, job)
}

func InstallProject(ctx context.Context, job Job) (err error) {
	return job.GlobalLock().Validate().LoadConfigs().Run(ctx).Finish().Error()
}
