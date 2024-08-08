package socket

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/metrics"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
)

type Conn interface {
	ID() string
	UID() string
	SetUser(info *auth.UserInfo)
	GetUser() *auth.UserInfo
	PubSub() application.PubSub

	AddTask(id string, fn func(error)) error
	RunTask(id string) error
	RemoveTask(id string)

	GetPtyHandler(sessionID string) (PtyHandler, bool)
	SetPtyHandler(sessionID string, session PtyHandler)
	ClosePty(ctx context.Context, sessionId string, status uint32, reason string)

	SetWriteDeadline(t time.Time) error
	WriteMessage(messageType int, data []byte) error
	SetReadLimit(limit int64)
	SetReadDeadline(t time.Time) error
	SetPongHandler(h func(appData string) error)
	ReadMessage() (messageType int, p []byte, err error)
	NextWriter(messageType int) (io.WriteCloser, error)
	Close(ctx context.Context) error
}

var _ Conn = (*WsConn)(nil)

type WsConn struct {
	ws *websocket.Conn

	// 每个浏览器窗口的 id 是不一样的
	id string
	// 同一个浏览的 uid 是一样的
	uid    string
	pubSub application.PubSub

	userMu sync.RWMutex
	user   *auth.UserInfo

	// taskManager 是任务管理器，用来管理每个 job 的部署和取消
	taskManager TaskManager
	// sm 是 session 的管理器
	// 用来管理这个conn下的每个 session 的状态
	// 一个链接可以有多个 session
	// 一个 session 可以有多个shell
	sm SessionMapper
}

func (wc *WebsocketManager) newWsConn(
	uid, id string,
	c *websocket.Conn,
	taskManager TaskManager,
	sm SessionMapper,
) Conn {
	wc.counter.Inc()
	return &WsConn{
		ws:          c,
		id:          id,
		uid:         uid,
		pubSub:      wc.pl.Ws().New(uid, id),
		taskManager: taskManager,
		sm:          sm,
	}
}

func (c *WsConn) ID() string {
	return c.id
}

func (c *WsConn) SetUser(info *auth.UserInfo) {
	c.userMu.Lock()
	defer c.userMu.Unlock()
	c.user = info
}

func (c *WsConn) PubSub() application.PubSub {
	return c.pubSub
}

func (c *WsConn) GetUser() *auth.UserInfo {
	c.userMu.RLock()
	defer c.userMu.RUnlock()
	return c.user
}

func (c *WsConn) Close(ctx context.Context) error {
	c.taskManager.StopAll()
	c.pubSub.Close()
	c.ws.Close()
	c.sm.CloseAll(ctx)
	var username string
	if c.GetUser() != nil {
		username = c.GetUser().Name
	}
	metrics.WebsocketConnectionsCount.With(prometheus.Labels{"username": username}).Dec()
	return nil
}

func (c *WsConn) SetWriteDeadline(t time.Time) error {
	return c.ws.SetWriteDeadline(t)
}

func (c *WsConn) WriteMessage(messageType int, data []byte) error {
	return c.ws.WriteMessage(messageType, data)
}

func (c *WsConn) SetReadLimit(limit int64) {
	c.ws.SetReadLimit(limit)
}

func (c *WsConn) SetReadDeadline(t time.Time) error {
	return c.ws.SetReadDeadline(t)
}

func (c *WsConn) SetPongHandler(h func(appData string) error) {
	c.ws.SetPongHandler(h)
}

func (c *WsConn) ReadMessage() (messageType int, p []byte, err error) {
	return c.ws.ReadMessage()
}

func (c *WsConn) NextWriter(messageType int) (io.WriteCloser, error) {
	return c.ws.NextWriter(messageType)
}

func (c *WsConn) UID() string {
	return c.uid
}

func (c *WsConn) GetPtyHandler(sessionID string) (PtyHandler, bool) {
	return c.sm.Get(sessionID)
}

func (c *WsConn) SetPtyHandler(sessionID string, session PtyHandler) {
	c.sm.Set(sessionID, session)
}

func (c *WsConn) ClosePty(ctx context.Context, sessionId string, status uint32, reason string) {
	c.sm.Close(ctx, sessionId, status, reason)
}

func (c *WsConn) AddTask(id string, fn func(error)) error {
	return c.taskManager.Register(id, fn)
}

func (c *WsConn) RemoveTask(id string) {
	c.taskManager.Remove(id)
}

func (c *WsConn) RunTask(id string) error {
	if c.taskManager.Has(id) {
		c.taskManager.Stop(id)
		return nil
	}
	return errors.New("task not found")
}
