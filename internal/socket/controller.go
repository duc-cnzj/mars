package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/internal/utils/recovery"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"

	"github.com/duc-cnzj/mars-client/v4/cluster"
	"github.com/duc-cnzj/mars-client/v4/types"
	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/metrics"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
)

type HandleRequestFunc func(c *WsConn, t websocket_pb.Type, message []byte)

var handlers map[websocket_pb.Type]HandleRequestFunc = map[websocket_pb.Type]HandleRequestFunc{
	WsAuthorize:          HandleWsAuthorize,
	WsHandleCloseShell:   HandleWsHandleCloseShell,
	WsHandleExecShellMsg: HandleWsHandleExecShellMsg,
	WsHandleExecShell:    HandleWsHandleExecShell,
	WsCancel:             HandleWsCancel,
	WsCreateProject:      HandleWsCreateProject,
	WsUpdateProject:      HandleWsUpdateProject,
	ProjectPodEvent:      HandleWsProjectPodEvent,
}

type WsConn struct {
	doneFunc func()

	id   string
	uid  string
	conn contracts.WebsocketConn

	NewJobFunc NewJobFunc

	userMu           sync.RWMutex
	user             contracts.UserInfo
	pubSub           contracts.PubSub
	cancelSignaler   contracts.CancelSignaler
	terminalSessions contracts.SessionMapper
	doneCtx          context.Context
}

func (wc *WebsocketManager) initConn(r *http.Request, c *websocket.Conn) *WsConn {
	var uid string
	uid = r.URL.Query().Get("uid")
	if uid == "" {
		uid = uuid.New().String()
	}
	id := uuid.New().String()

	ps := plugins.GetWsSender().New(uid, id)
	ctx, cancelFunc := context.WithCancel(context.TODO())
	var wsconn = &WsConn{
		doneCtx:        ctx,
		doneFunc:       cancelFunc,
		pubSub:         ps,
		id:             id,
		uid:            uid,
		NewJobFunc:     NewJober,
		conn:           c,
		cancelSignaler: &CancelSignals{cs: map[string]func(error){}},
	}
	wsconn.terminalSessions = &SessionMap{Sessions: make(map[string]contracts.PtyHandler), conn: wsconn}

	Wait.Inc()

	return wsconn
}

func (c *WsConn) Shutdown() {
	mlog.Debug("[Websocket]: Ws exit ")

	c.doneFunc()
	c.cancelSignaler.CancelAll()
	c.terminalSessions.CloseAll()
	c.pubSub.Close()
	c.conn.Close()
	metrics.WebsocketConnectionsCount.With(prometheus.Labels{"username": c.GetUser().Name}).Dec()
	Wait.Dec()
}

func (c *WsConn) SetUser(info contracts.UserInfo) {
	c.userMu.Lock()
	defer c.userMu.Unlock()
	c.user = info
}

func (c *WsConn) GetUser() contracts.UserInfo {
	c.userMu.RLock()
	defer c.userMu.RUnlock()
	return c.user
}

type WebsocketManager struct {
	healthTickDuration time.Duration
}

func NewWebsocketManager(healthTickDuration time.Duration) *WebsocketManager {
	return &WebsocketManager{healthTickDuration: healthTickDuration}
}

func (wc *WebsocketManager) TickClusterHealth() {
	ticker := time.NewTicker(wc.healthTickDuration)
	done := app.App().Done()
	sub := plugins.GetWsSender().New("", "")
	lock := app.CacheLock()

	go func() {
		defer recovery.HandlePanic("TickClusterHealth")
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if lock.Acquire("TickClusterHealth", 5) {
					func() {
						defer lock.Release("TickClusterHealth")
						info := utils.ClusterInfo()
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
				mlog.Info("[Websocket]: app shutdown and stop WsClusterInfoSync")
				return
			}
		}
	}()
}

func (wc *WebsocketManager) Info(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	marshal, _ := json.Marshal(plugins.GetWsSender().New("", "").Info())
	writer.Write(marshal)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (wc *WebsocketManager) Ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		mlog.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("cant Upgrade websocket connection"))
		return
	}

	wsconn := wc.initConn(r, c)
	defer wsconn.Shutdown()

	go func() {
		defer recovery.HandlePanic("[ProjectPodEventSubscriber]: Run")
		wsconn.pubSub.Run(wsconn.doneCtx)
	}()

	go func() {
		defer recovery.HandlePanic("Websocket: Write")

		write(wsconn)
	}()

	NewMessageSender(wsconn, "", WsSetUid).SendMsg(wsconn.uid)

	ch := make(chan struct{})
	go func() {
		var err error
		defer func() {
			mlog.Debugf("[Websocket]: go read exit, err: %v", err)
		}()
		defer recovery.HandlePanic("[Websocket]: read recovery")
		err = read(wsconn)
		close(ch)
	}()

	select {
	case <-app.App().Done():
		return
	case <-ch:
		return
	}
}

func write(wsconn *WsConn) error {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		mlog.Debugf("[Websocket]: go write exit")
		ticker.Stop()
		wsconn.conn.Close()
	}()
	ch := wsconn.pubSub.Subscribe()
	for {
		select {
		case message, ok := <-ch:
			wsconn.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return wsconn.conn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			w, err := wsconn.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return err
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return err
			}
		case <-ticker.C:
			wsconn.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsconn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}
		}
	}
}

func read(wsconn *WsConn) error {
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
			mlog.Debugf("[Websocket]: read error: %v", err)
			return err
		}
		if err := proto.Unmarshal(message, &wsRequest); err != nil {
			NewMessageSender(wsconn, "", WsInternalError).SendEndError(err)

			continue
		}

		go func(wsRequest *websocket_pb.WsRequestMetadata, message []byte) {
			if handler, ok := handlers[wsRequest.Type]; ok {
				defer recovery.HandlePanicWithCallback(wsRequest.Type.String(), func(err error) {
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
				if wsconn.GetUser().Sub == "" && wsRequest.Type != websocket_pb.Type_HandleAuthorize {
					NewMessageSender(wsconn, "", WsAuthorize).SendMsg("认证中，请稍等~")
					return
				}
				handler(wsconn, wsRequest.Type, message)
			}
		}(&wsRequest, message)
	}
}

func HandleWsAuthorize(c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.AuthorizeTokenInput
	if err := proto.Unmarshal(message, &input); err != nil {
		mlog.Error("[Websocket]: " + err.Error())
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}

	if claims, b := app.Auth().VerifyToken(input.Token); b {
		c.SetUser(claims.UserInfo)
		metrics.WebsocketConnectionsCount.With(prometheus.Labels{"username": claims.UserInfo.Name}).Inc()
	}
}

func HandleWsProjectPodEvent(c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.ProjectPodEventJoinInput
	if err := proto.Unmarshal(message, &input); err != nil {
		mlog.Error("[Websocket]: " + err.Error())
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}
	if input.Join {
		c.pubSub.(contracts.ProjectPodEventSubscriber).Join(input.GetProjectId())
	} else {
		c.pubSub.(contracts.ProjectPodEventSubscriber).Leave(input.GetNamespaceId(), input.GetProjectId())
	}
}

func HandleWsHandleCloseShell(c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.TerminalMessageInput
	if err := proto.Unmarshal(message, &input); err != nil {
		mlog.Error(err.Error())
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}
	mlog.Debugf("[Websocket]: %v 收到客户端主动断开的消息", input.Message.SessionId)
	c.terminalSessions.Close(input.Message.SessionId, 0, "")
}

func HandleWsHandleExecShellMsg(c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.TerminalMessageInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}
	if input.Message.SessionId != "" {
		c.terminalSessions.Send(input.Message)
	}
}

func HandleWsHandleExecShell(c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.WsHandleExecShellInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)
		return
	}

	sessionID, err := HandleExecShell(&input, c)
	if err != nil {
		mlog.Error(err)
		NewMessageSender(c, "", WsHandleExecShell).SendEndError(err)
		return
	}

	mlog.Debugf("[Websocket]: 收到 初始化连接 WsHandleExecShell 消息, id: %v", sessionID)

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

func HandleWsCancel(c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.CancelInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}

	// cancel
	var slugName = utils.GetSlugName(input.NamespaceId, input.Name)
	if c.cancelSignaler.Has(slugName) {
		var ns models.Namespace
		app.DB().Select("name").First(&ns, input.NamespaceId)
		AuditLogWithChange(c.GetUser().Name, types.EventActionType_CancelDeploy, fmt.Sprintf("用户取消部署 namespace: %s, 服务 %s.", ns.Name, input.Name), nil, nil)
		c.cancelSignaler.Cancel(slugName)
	}
}

func HandleWsCreateProject(c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.CreateProjectInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}
	slug := utils.GetSlugName(input.NamespaceId, input.Name)
	job := c.NewJobFunc(&input, c.GetUser(), slug, NewMessageSender(c, slug, t), c.pubSub, 0)
	if err := c.cancelSignaler.Add(job.ID(), job.Stop); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)
		return
	}
	defer c.cancelSignaler.Remove(job.ID())
	InstallProject(job)
}

func HandleWsUpdateProject(c *WsConn, t websocket_pb.Type, message []byte) {
	var input websocket_pb.UpdateProjectInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)
		return
	}
	var p models.Project
	if err := app.DB().Where("`id` = ?", input.ProjectId).First(&p).Error; err != nil {
		NewMessageSender(c, "", t).SendEndError(err)
		return
	}

	slug := utils.GetSlugName(p.NamespaceId, p.Name)
	job := c.NewJobFunc(&websocket_pb.CreateProjectInput{
		Type:         t,
		NamespaceId:  int64(p.NamespaceId),
		Name:         p.Name,
		GitProjectId: int64(p.GitProjectId),
		GitBranch:    input.GitBranch,
		GitCommit:    input.GitCommit,
		Config:       input.Config,
		Atomic:       input.Atomic,
		ExtraValues:  input.ExtraValues,
	}, c.GetUser(), slug, NewMessageSender(c, slug, t), c.pubSub, 0)
	if err := c.cancelSignaler.Add(job.ID(), job.Stop); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)
		return
	}
	defer c.cancelSignaler.Remove(job.ID())
	InstallProject(job)
}

func InstallProject(job contracts.Job) (err error) {
	defer func() {
		job.CallDestroyFuncs()
		if err != nil && !job.IsDryRun() {
			job.Prune()
		}
		job.Finish()
	}()

	handleStopErr := func(e error) {
		job.Messager().SendDeployedResult(websocket_pb.ResultType_DeployedCanceled, e.Error(), job.ProjectModel())
		job.Messager().Stop(e)
		err = e
	}

	if err = job.Validate(); err != nil {
		if e := job.GetStoppedErrorIfHas(); e != nil {
			handleStopErr(e)
			return
		}
		job.Messager().SendEndError(err)
		return
	}

	if err = job.LoadConfigs(); err != nil {
		if e := job.GetStoppedErrorIfHas(); e != nil {
			handleStopErr(e)
			return
		}
		job.Messager().SendEndError(err)
		return
	}

	err = job.Run()
	return
}
