package socket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/grpc/services"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/plugins"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/gosimple/slug"
)

type HandleRequestFunc func(c *WsConn, wsRequest WsRequest)

var handlers map[string]HandleRequestFunc = map[string]HandleRequestFunc{
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
	user             services.UserInfo
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

func (c *WsConn) SetUser(info services.UserInfo) {
	c.userMu.Lock()
	defer c.userMu.Unlock()
	c.user = info
}

func (c *WsConn) GetUser() services.UserInfo {
	c.userMu.RLock()
	defer c.userMu.RUnlock()
	return c.user
}

func (c *WsConn) GetShellChannel(sessionID string) (chan TerminalMessage, error) {
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
				marshal, _ := json.Marshal(utils.ClusterInfo())

				sub.ToAll(&WsResponse{
					Type: WsClusterInfoSync,
					Data: string(marshal),
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

			w, err := wsconn.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return err
			}
			w.Write([]byte(message))

			if err := w.Close(); err != nil {
				return err
			}
		case <-ticker.C:
			mlog.Debugf("[Websocket] tick ping/pong uid: %s, id: %s", wsconn.uid, wsconn.id)
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
		mlog.Debugf("[Websocket] 收到心跳 id: %s, uid %s", wsconn.id, wsconn.uid)
		return nil
	})
	for {
		var wsRequest WsRequest
		_, message, err := wsconn.conn.ReadMessage()
		if err != nil {
			mlog.Debugf("[Websocket] read error: %v %v", err, message)
			return err
		}
		if err := json.Unmarshal(message, &wsRequest); err != nil {
			NewMessageSender(wsconn, "", WsInternalError).SendEndError(err)

			continue
		}

		go func(wsRequest WsRequest) {
			mlog.Debugf("[Websocket]: user: %v, type: %v", wsconn.GetUser().Name, wsRequest.Type)

			if handler, ok := handlers[wsRequest.Type]; ok {
				handler(wsconn, wsRequest)
			}
		}(wsRequest)
	}
}

type Token struct {
	Token string `json:"token"`
}

func HandleWsAuthorize(c *WsConn, wsRequest WsRequest) {
	defer utils.HandlePanic("HandleWsAuthorize")

	var input Token
	if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
		mlog.Error(wsRequest.Data, &input)
		NewMessageSender(c, "", wsRequest.Type).SendEndError(err)

		return
	}
	var token = strings.TrimSpace(strings.TrimLeft(input.Token, "Bearer"))
	parse, err := jwt.ParseWithClaims(token, &services.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return app.Config().Pubkey(), nil
	})
	if err == nil && parse.Valid {
		claims, _ := parse.Claims.(*services.JwtClaims)
		c.SetUser(claims.UserInfo)
	}
}

func HandleWsHandleCloseShell(c *WsConn, wsRequest WsRequest) {
	defer utils.HandlePanic("HandleWsHandleCloseShell")

	var input TerminalMessage
	if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
		mlog.Error(wsRequest.Data, &input)
		NewMessageSender(c, "", wsRequest.Type).SendEndError(err)

		return
	}
	mlog.Debugf("[Websocket] %v 收到客户端主动断开的消息", input.SessionID)
	c.terminalSessions.Close(input.SessionID, 0, "")
}

func HandleWsHandleExecShellMsg(c *WsConn, wsRequest WsRequest) {
	defer utils.HandlePanic("HandleWsHandleExecShellMsg")

	var input TerminalMessage
	if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
		mlog.Error(wsRequest.Data, &input)
		NewMessageSender(c, "", wsRequest.Type).SendEndError(err)

		return
	}
	if input.SessionID != "" {
		c.terminalSessions.Send(input)
	}
}

func HandleWsHandleExecShell(c *WsConn, wsRequest WsRequest) {
	var input WsHandleExecShellInput
	if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
		mlog.Error(wsRequest.Data, &input)
		NewMessageSender(c, "", wsRequest.Type).SendEndError(err)
		return
	}

	sessionID, err := HandleExecShell(input, c)
	if err != nil {
		mlog.Error(err)
		NewMessageSender(c, "", WsHandleExecShell).SendEndMsg(ResultError, err.Error())
		return
	}

	mlog.Debugf("[Websocket] 收到 初始化连接 WsHandleExecShell 消息, id: %v", sessionID)

	res := struct {
		WsHandleExecShellInput
		SessionID string `json:"session_id"`
	}{
		WsHandleExecShellInput: input,
		SessionID:              sessionID,
	}
	marshal, _ := json.Marshal(res)
	NewMessageSender(c, "", WsHandleExecShell).SendEndMsg(ResultSuccess, string(marshal))
}

type CancelInput struct {
	NamespaceId int    `uri:"namespace_id" json:"namespace_id"`
	Name        string `json:"name"`
}

func HandleWsCancel(c *WsConn, wsRequest WsRequest) {
	defer utils.HandlePanic("HandleWsCancel")

	var input CancelInput
	if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
		mlog.Error(wsRequest.Data, &input)
		NewMessageSender(c, "", wsRequest.Type).SendEndError(err)

		return
	}

	// cancel
	var slugName = utils.Md5(fmt.Sprintf("%d-%s", input.NamespaceId, input.Name))
	if c.cancelSignaler.Has(slugName) {
		input.Name = slug.Make(input.Name)
		c.cancelSignaler.Cancel(slugName)
	}
}

type ProjectInput struct {
	NamespaceId int `uri:"namespace_id" json:"namespace_id"`

	Name            string `json:"name"`
	GitlabProjectId int    `json:"gitlab_project_id"`
	GitlabBranch    string `json:"gitlab_branch"`
	GitlabCommit    string `json:"gitlab_commit"`
	Config          string `json:"config"`
	Atomic          bool   `json:"atomic"`
}

func HandleWsCreateProject(c *WsConn, wsRequest WsRequest) {
	defer utils.HandlePanic("HandleWsCreateProject")

	var input ProjectInput
	if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
		mlog.Error(wsRequest.Data, &input)
		NewMessageSender(c, "", wsRequest.Type).SendEndError(err)

		return
	}

	job := NewJober(input, wsRequest.Type, wsRequest, c)
	if err := c.cancelSignaler.Add(job.ID(), job.Stop); err != nil {
		NewMessageSender(c, "", wsRequest.Type).SendEndError(err)
		return
	}
	defer c.cancelSignaler.Remove(job.ID())
	installProject(job)
}

type UpdateProject struct {
	ProjectId int `json:"project_id"`

	GitlabBranch string `json:"gitlab_branch"`
	GitlabCommit string `json:"gitlab_commit"`
	Config       string `json:"config"`
	Atomic       bool   `json:"atomic"`
}

func HandleWsUpdateProject(c *WsConn, wsRequest WsRequest) {
	defer utils.HandlePanic("HandleWsUpdateProject")

	var input UpdateProject
	if err := json.Unmarshal([]byte(wsRequest.Data), &input); err != nil {
		mlog.Error(wsRequest.Data, &input)
		NewMessageSender(c, "", wsRequest.Type).SendEndError(err)
		return
	}
	var p models.Project
	if err := app.DB().Where("`id` = ?", input.ProjectId).First(&p).Error; err != nil {
		mlog.Error(wsRequest.Data, &input)
		NewMessageSender(c, "", wsRequest.Type).SendEndError(err)
		return
	}

	job := NewJober(ProjectInput{
		NamespaceId:     p.NamespaceId,
		Name:            p.Name,
		GitlabProjectId: p.GitlabProjectId,
		GitlabBranch:    input.GitlabBranch,
		GitlabCommit:    input.GitlabCommit,
		Config:          input.Config,
		Atomic:          input.Atomic,
	}, wsRequest.Type, wsRequest, c)
	if err := c.cancelSignaler.Add(job.ID(), job.Stop); err != nil {
		NewMessageSender(c, "", wsRequest.Type).SendEndError(err)
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
		if job.IsStopped() {
			job.Messager().SendEndMsg(ResultDeployCanceled, err.Error())
			return
		}
		job.Messager().SendEndError(err)
		return
	}

	if err = job.Run(); err != nil {
		job.PubSub().ToAll(&WsResponse{Type: WsReloadProjects})
		return
	}

	job.PubSub().ToOthers(&WsResponse{Type: WsReloadProjects})
}
