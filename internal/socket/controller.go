package socket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/pkg/cluster"
	websocket_pb "github.com/duc-cnzj/mars/pkg/websocket"
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
}

type WsConn struct {
	id   string
	uid  string
	conn *websocket.Conn

	userMu           sync.RWMutex
	user             contracts.UserInfo
	pubSub           plugins.PubSub
	cancelSignaler   CancelSignaler
	terminalSessions SessionMapper
}

func (wc *WebsocketManager) initConn(r *http.Request, c *websocket.Conn) *WsConn {
	var uid string
	uid = r.URL.Query().Get("uid")
	if uid == "" {
		uid = uuid.New().String()
	}
	id := uuid.New().String()

	ps := plugins.GetWsSender().New(uid, id)
	var wsconn = &WsConn{
		pubSub:         ps,
		id:             id,
		uid:            uid,
		conn:           c,
		cancelSignaler: &CancelSignals{cs: map[string]func(error){}},
	}
	wsconn.terminalSessions = &SessionMap{Sessions: make(map[string]*MyPtyHandler), conn: wsconn}
	app.Metrics().IncWebsocketConn()
	Wait.Inc()

	return wsconn
}

func (c *WsConn) Shutdown() {
	mlog.Debug("[Websocket]: Ws exit ")

	c.cancelSignaler.CancelAll()
	c.terminalSessions.CloseAll()
	c.pubSub.Close()
	c.conn.Close()
	app.Metrics().DecWebsocketConn()
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

func (c *WsConn) GetShellChannel(sessionID string) (chan *websocket_pb.TerminalMessage, error) {
	if handler, ok := c.terminalSessions.Get(sessionID); ok {
		return handler.shellCh, nil
	}

	return nil, fmt.Errorf("%v not found channel", sessionID)
}

type WebsocketManager struct{}

func NewWebsocketManager() *WebsocketManager {
	return &WebsocketManager{}
}

func (*WebsocketManager) TickClusterHealth() {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		sub := plugins.GetWsSender().New("", "")
		for {
			select {
			case <-ticker.C:
				info := utils.ClusterInfo()
				sub.ToAll(&websocket_pb.WsHandleClusterResponse{
					Metadata: &websocket_pb.ResponseMetadata{
						Type: WsClusterInfoSync,
					},
					Info: &cluster.ClusterInfoResponse{
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
			case <-app.App().Done():
				mlog.Info("[Websocket]: app shutdown and stop WsClusterInfoSync")
				ticker.Stop()
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

	go write(wsconn)

	NewMessageSender(wsconn, "", WsSetUid).SendMsg(wsconn.uid)

	ch := make(chan struct{}, 1)
	go func() {
		var err error
		defer func() {
			utils.HandlePanic("[Websocket]: read recovery")
			mlog.Debugf("[Websocket]: go read exit, err: %v", err)
		}()
		err = read(wsconn)
		ch <- struct{}{}
	}()

	select {
	case <-app.App().Done():
		return
	case <-ch:
		return
	}
}

func write(wsconn *WsConn) error {
	defer utils.HandlePanic("Websocket: Write")

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
			mlog.Debugf("[Websocket]: tick ping/pong uid: %s, id: %s", wsconn.uid, wsconn.id)
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
		mlog.Debugf("[Websocket]: 收到心跳 id: %s, uid %s", wsconn.id, wsconn.uid)
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
				handler(wsconn, wsRequest.Type, message)
			}
		}(&wsRequest, message)
	}
}

func HandleWsAuthorize(c *WsConn, t websocket_pb.Type, message []byte) {
	defer utils.HandlePanic("HandleWsAuthorize")

	var input websocket_pb.AuthorizeTokenInput
	if err := proto.Unmarshal(message, &input); err != nil {
		mlog.Error("[Websocket]: " + err.Error())
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}

	if claims, b := app.Auth().VerifyToken(input.Token); b {
		c.SetUser(claims.UserInfo)
	}
}

func HandleWsHandleCloseShell(c *WsConn, t websocket_pb.Type, message []byte) {
	defer utils.HandlePanic("HandleWsHandleCloseShell")

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
	defer utils.HandlePanic("HandleWsHandleExecShellMsg")

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
		NewMessageSender(c, "", WsHandleExecShell).SendEndMsg(ResultError, err.Error())
		return
	}

	mlog.Debugf("[Websocket]: 收到 初始化连接 WsHandleExecShell 消息, id: %v", sessionID)

	NewMessageSender(c, "", WsHandleExecShell).SendProtoMsg(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.ResponseMetadata{
			Id:     c.id,
			Uid:    c.uid,
			Type:   WsHandleExecShell,
			Result: ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			SessionId: sessionID,
		},
		Container: &websocket_pb.Container{
			Namespace: input.Namespace,
			Pod:       input.Pod,
			Container: input.Container,
		},
	})
}

func HandleWsCancel(c *WsConn, t websocket_pb.Type, message []byte) {
	defer utils.HandlePanic("HandleWsCancel")

	var input websocket_pb.CancelInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}

	// cancel
	var slugName = utils.Md5(fmt.Sprintf("%d-%s", input.NamespaceId, input.Name))
	if c.cancelSignaler.Has(slugName) {
		c.cancelSignaler.Cancel(slugName)
	}
}

func HandleWsCreateProject(c *WsConn, t websocket_pb.Type, message []byte) {
	defer utils.HandlePanic("HandleWsCreateProject")

	var input websocket_pb.ProjectInput
	if err := proto.Unmarshal(message, &input); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)

		return
	}
	job := NewJober(&input, t, c)
	if err := c.cancelSignaler.Add(job.ID(), job.Stop); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)
		return
	}
	defer c.cancelSignaler.Remove(job.ID())
	installProject(job)
}

func HandleWsUpdateProject(c *WsConn, t websocket_pb.Type, message []byte) {
	defer utils.HandlePanic("HandleWsUpdateProject")

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

	job := NewJober(&websocket_pb.ProjectInput{
		NamespaceId:     int64(p.NamespaceId),
		Name:            p.Name,
		GitlabProjectId: int64(p.GitlabProjectId),
		GitlabBranch:    input.GitlabBranch,
		GitlabCommit:    input.GitlabCommit,
		Config:          input.Config,
		Atomic:          input.Atomic,
	}, t, c)
	if err := c.cancelSignaler.Add(job.ID(), job.Stop); err != nil {
		NewMessageSender(c, "", t).SendEndError(err)
		return
	}
	defer c.cancelSignaler.Remove(job.ID())
	installProject(job)
}

func installProject(job Job) {
	var err error
	defer func() {
		job.CallDestroyFuncs()
		if err != nil {
			job.Prune()
		}
	}()

	if err = job.Validate(); err != nil {
		job.Messager().SendEndError(err)
		return
	}

	if err = job.LoadConfigs(); err != nil {
		if err := job.GetStoppedErrorIfHas(); err != nil {
			job.Messager().SendEndMsg(ResultDeployCanceled, err.Error())
			return
		}
		job.Messager().SendEndError(err)
		return
	}

	res := &WsResponse{Metadata: &websocket_pb.ResponseMetadata{Type: WsReloadProjects}}
	if err = job.Run(); err != nil {
		job.PubSub().ToAll(res)
		return
	}

	job.PubSub().ToOthers(res)
}
