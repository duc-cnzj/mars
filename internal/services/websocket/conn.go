package websocket

// conn.go 定义 ws 连接抽象：
//   - Conn：一条 websocket 连接的完整接口（用户/发布订阅/部署取消/pty 会话）
//   - GorillaWs：gorilla/websocket 底层传输的最小接口（测试缝）
//   - wsConn：Conn 的生产实现，聚合发布订阅、部署取消任务注册表与终端会话注册表

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Ws* 是 websocket 协议类型标签（websocket_pb.Type 的语义别名）。
// 与部署结果状态 Result*（internal/deploy）不同，这些只属于 ws 传输层。
const (
	WsSetUid             = websocket_pb.Type_SetUid
	WsCancel             = websocket_pb.Type_CancelProject
	WsCreateProject      = websocket_pb.Type_CreateProject
	WsUpdateProject      = websocket_pb.Type_UpdateProject
	WsProcessPercent     = websocket_pb.Type_ProcessPercent
	WsClusterInfoSync    = websocket_pb.Type_ClusterInfoSync
	WsInternalError      = websocket_pb.Type_InternalError
	WsHandleExecShell    = websocket_pb.Type_HandleExecShell
	WsHandleExecShellMsg = websocket_pb.Type_HandleExecShellMsg
	WsHandleCloseShell   = websocket_pb.Type_HandleCloseShell
	WsAuthorize          = websocket_pb.Type_HandleAuthorize
	ProjectPodEvent      = websocket_pb.Type_ProjectPodEvent
)

// 连接调优参数：仅 ws 传输层使用。
// 以包级变量承载（而非 const），测试可临时调小 pingPeriod 覆盖 write 循环的
// ping 分支，与 metrics 包 tickDuration、namespace 包 deleteTimeout 同款手法。
var (
	// 允许对端发送的最大消息字节数。
	maxMessageSize int64 = 1024 * 1024 * 20 // 20MB
	// 允许向对端写入一条消息的时限。
	writeWait = 10 * time.Second
	// 允许读取对端下一个 pong 消息的时限。
	pongWait = 60 * time.Second
	// 向对端发送 ping 的周期，必须小于 pongWait。
	pingPeriod = (pongWait * 8) / 10
)

// Conn 是一条 websocket 连接的抽象：除底层传输（GorillaWs）外，
// 还聚合了连接级状态——用户信息（SetUser/GetUser）、发布订阅（PubSub）、
// 部署取消任务（AddCancelDeployTask/RunCancelDeployTask）与终端会话
// （GetPtyHandler/SetPtyHandler/ClosePty），并负责连接收尾（CloseAndClean）。
type Conn interface {
	// ID 返回连接的唯一 id（每个浏览器窗口不同）。
	ID() string
	// UID 返回浏览器的 uid（同一浏览器的多个窗口相同）。
	UID() string
	// SetUser 以写锁保存认证后的用户信息。
	SetUser(info *biz.UserInfo)
	// GetUser 以读锁返回当前用户信息（未认证时为 nil）。
	GetUser() *biz.UserInfo
	// PubSub 返回该连接的发布订阅实例。
	PubSub() application.PubSub

	// AddCancelDeployTask 登记一个部署任务的取消回调。
	AddCancelDeployTask(id string, fn func(error)) error
	// RunCancelDeployTask 触发指定部署任务的取消；任务不存在时返回错误。
	RunCancelDeployTask(id string) error
	// RemoveCancelDeployTask 移除一个部署任务的取消回调。
	RemoveCancelDeployTask(id string)

	// GetPtyHandler 按 sessionID 取回终端会话处理器。
	GetPtyHandler(sessionID string) (PtyHandler, bool)
	// SetPtyHandler 登记一个终端会话处理器。
	SetPtyHandler(sessionID string, session PtyHandler)
	// ClosePty 关闭指定终端会话，并向客户端发送状态码与原因。
	ClosePty(ctx context.Context, sessionID string, status uint32, reason string)
	// CloseAndClean 是连接收尾：停止在途部署任务、关闭发布订阅与底层连接、
	// 关闭全部终端会话，并递减连接计数。
	CloseAndClean(ctx context.Context) error
	GorillaWs
}

// GorillaWs 是 gorilla/websocket 连接的最小接口，将底层传输与业务解耦，
// 使 wsConn 可在测试中注入 MockGorillaWs。
type GorillaWs interface {
	// SetWriteDeadline 设置写入对端消息的时限（到期写失败）。
	SetWriteDeadline(t time.Time) error
	// WriteMessage 直接写入一条完整消息（用于 ping 帧）。
	WriteMessage(messageType int, data []byte) error
	// SetReadLimit 限制对端可发送的最大消息字节数。
	SetReadLimit(limit int64)
	// SetReadDeadline 设置读取对端消息的时限（到期读失败）。
	SetReadDeadline(t time.Time) error
	// SetPongHandler 注册对端 pong 帧的回调（用于刷新读时限）。
	SetPongHandler(h func(appData string) error)
	// ReadMessage 读取一条完整消息。
	ReadMessage() (messageType int, p []byte, err error)
	// NextWriter 返回下一条消息的写入器（逐条写入后必须 Close）。
	NextWriter(messageType int) (io.WriteCloser, error)
	// Close 关闭底层连接。
	Close() error
}

var _ Conn = (*wsConn)(nil)

// wsConn 是 Conn 的生产实现：组合底层 gorilla 连接与三个连接级组件
// （发布订阅、部署取消任务注册表、终端会话注册表）。
type wsConn struct {
	GorillaWs

	// 每个浏览器窗口的 id 是不一样的
	id string
	// 同一个浏览的 uid 是一样的
	uid    string
	pubSub application.PubSub

	userMu sync.RWMutex
	user   *biz.UserInfo

	// taskManager 是任务管理器，用来管理每个 job 的部署和取消
	taskManager TaskManager
	// sessions 是该连接下的终端会话管理器（sessionMap）：
	// 一个连接可以同时打开多个 session，每个 session 对应一个 shell。
	sessions SessionMapper

	// closeOnce 保证 CloseAndClean 只执行一次完整清理：Serve 的 defer 与 write 循环
	// 的 defer 会对同一连接各调一次 CloseAndClean（write 先返回触发第一次，Serve 在
	// errgroup.Wait 后触发第二次），不加幂等会让 prometheus 连接数 gauge 每条连接
	// 被双递减，长期趋负。
	closeOnce sync.Once
}

// newWsConn 为一条新连接组装 wsConn，并递增连接计数（counter）。
func (wc *websocketManager) newWsConn(
	uid, id string,
	c GorillaWs,
	taskManager TaskManager,
	sm SessionMapper,
) Conn {
	wc.counter.Inc()
	return &wsConn{
		GorillaWs:   c,
		id:          id,
		uid:         uid,
		pubSub:      wc.pluginManager.Ws().New(uid, id),
		taskManager: taskManager,
		sessions:    sm,
	}
}

// ID 返回连接的唯一 id（每个浏览器窗口不同）。
func (c *wsConn) ID() string {
	return c.id
}

// SetUser 以写锁保存认证后的用户信息。
func (c *wsConn) SetUser(info *biz.UserInfo) {
	c.userMu.Lock()
	defer c.userMu.Unlock()
	c.user = info
}

// PubSub 返回该连接的发布订阅实例。
func (c *wsConn) PubSub() application.PubSub {
	return c.pubSub
}

// GetUser 以读锁返回当前用户信息（未认证时为 nil）。
func (c *wsConn) GetUser() *biz.UserInfo {
	c.userMu.RLock()
	defer c.userMu.RUnlock()
	return c.user
}

// CloseAndClean 是连接收尾：停止在途部署任务、关闭发布订阅与底层连接、
// 关闭全部终端会话，并递减连接计数。以 sync.Once 保证幂等——Serve 与 write 循环
// 都会触发本方法，只有第一次真正执行清理（见 closeOnce 字段注释）。
func (c *wsConn) CloseAndClean(ctx context.Context) error {
	c.closeOnce.Do(func() {
		c.taskManager.StopAll()
		c.pubSub.Close()
		c.Close()
		c.sessions.CloseAll(ctx)
		var username string
		if c.GetUser() != nil {
			username = c.GetUser().Name
		}
		metrics.WebsocketConnectionsCount.With(prometheus.Labels{"username": username}).Dec()
	})
	return nil
}

// UID 返回浏览器的 uid（同一浏览器的多个窗口相同）。
func (c *wsConn) UID() string {
	return c.uid
}

// GetPtyHandler 按 sessionID 取回终端会话处理器。
func (c *wsConn) GetPtyHandler(sessionID string) (PtyHandler, bool) {
	return c.sessions.Get(sessionID)
}

// SetPtyHandler 登记一个终端会话处理器。
func (c *wsConn) SetPtyHandler(sessionID string, session PtyHandler) {
	c.sessions.Set(sessionID, session)
}

// ClosePty 关闭指定终端会话，并向客户端发送状态码与原因。
func (c *wsConn) ClosePty(ctx context.Context, sessionID string, status uint32, reason string) {
	c.sessions.Close(ctx, sessionID, status, reason)
}

// AddCancelDeployTask 登记一个部署任务的取消回调。
func (c *wsConn) AddCancelDeployTask(id string, fn func(error)) error {
	return c.taskManager.Register(id, fn)
}

// RemoveCancelDeployTask 移除一个部署任务的取消回调。
func (c *wsConn) RemoveCancelDeployTask(id string) {
	c.taskManager.Remove(id)
}

// RunCancelDeployTask 触发指定部署任务的取消；任务不存在时返回错误。
func (c *wsConn) RunCancelDeployTask(id string) error {
	if c.taskManager.Has(id) {
		c.taskManager.Stop(id)
		return nil
	}
	return errors.New("task not found")
}
