package websocket

import (
	"context"
	"errors"
	"sync"

	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

// 本文件承载 wsConn 的两个并发注册表：
//   - sessionMap（SessionMapper）：一个连接下的终端会话（sessionID → PtyHandler）
//   - taskManager（TaskManager）：在途部署的取消回调（任务 slug → func(error)）
//
// 两者都是「互斥锁保护的 map + 生命周期管理」，是 wsConn 的内部状态。

// ---- 终端会话注册表（SessionMapper） ----

// SessionMapper 是一个连接下的终端会话注册表：按 sessionID 存取 ptyHandler，
// 并提供连接级关闭（CloseAll）与单会话关闭（Close），由 wsConn.sessions 持有。
type SessionMapper interface {
	// Get 按 sessionID 取回终端会话处理器。
	Get(sessionID string) (PtyHandler, bool)
	// Set 把一个 ptyHandler 会话登记到注册表。
	Set(sessionID string, session PtyHandler)
	// CloseAll 并发关闭注册表内的全部会话，等待完成后清空映射（连接断开时调用）。
	CloseAll(ctx context.Context)
	// Close 关闭指定会话：从注册表移除并异步调用会话的 Close（reason 透传给客户端）。
	Close(ctx context.Context, sessionID string, status uint32, reason string)
}

// sessionMap 是 SessionMapper 的生产实现：以 RWMutex（sessLock）保护的
// sessionID → PtyHandler 映射 + WaitGroup（wg）跟踪在途关闭的会话 goroutine。
type sessionMap struct {
	wg     sync.WaitGroup
	logger mlog.Logger

	sessLock sync.RWMutex
	Sessions map[string]PtyHandler
}

// NewSessionMap 构造会话注册表（空映射 + 注入 logger）。
func NewSessionMap(logger mlog.Logger) SessionMapper {
	return &sessionMap{Sessions: map[string]PtyHandler{}, logger: logger}
}

// Get 按 sessionID 取回终端会话处理器。
func (sm *sessionMap) Get(sessionID string) (PtyHandler, bool) {
	sm.sessLock.RLock()
	defer sm.sessLock.RUnlock()
	h, ok := sm.Sessions[sessionID]
	return h, ok
}

// Set 把一个 ptyHandler 会话登记到注册表。
func (sm *sessionMap) Set(sessionID string, session PtyHandler) {
	sm.sessLock.Lock()
	defer sm.sessLock.Unlock()
	sm.Sessions[sessionID] = session
}

// CloseAll 并发关闭注册表内的全部会话，等待完成后清空映射（连接断开时调用）。
func (sm *sessionMap) CloseAll(ctx context.Context) {
	sm.logger.Debug("[Websocket]: close all.")
	sm.sessLock.Lock()
	defer sm.sessLock.Unlock()

	for _, s := range sm.Sessions {
		sm.wg.Add(1)
		go func(s PtyHandler) {
			defer sm.wg.Done()
			s.Close(ctx, "websocket conn closed")
		}(s)
	}
	sm.wg.Wait()
	sm.Sessions = map[string]PtyHandler{}
}

// Close 关闭指定会话：从注册表移除并异步调用会话的 Close（reason 透传给客户端）。
// 会话不存在时为空操作；status 供调用方标注关闭来源（客户端断开/进程退出/异常）。
func (sm *sessionMap) Close(ctx context.Context, sessionID string, status uint32, reason string) {
	sm.logger.Debugf("[Websocket]: session %v closed, reason: %s, status: %v.", sessionID, reason, status)
	sm.sessLock.Lock()
	defer sm.sessLock.Unlock()
	if s, ok := sm.Sessions[sessionID]; ok {
		delete(sm.Sessions, sessionID)
		sm.wg.Add(1)
		go func() {
			defer sm.wg.Done()
			s.Close(ctx, reason)
		}()
	}
}

// ---- 部署取消任务注册表（TaskManager） ----

var errSignalExists = errors.New("项目已经存在")

// TaskManager 是在途部署的取消回调注册表：以部署任务 slug 为键登记/移除/触发
// 取消回调，供取消部署、连接断开时统一清理，由 wsConn.taskManager 持有。
type TaskManager interface {
	// Remove 删除指定部署任务的取消回调。
	Remove(id string)
	// Has 判断指定部署任务是否已登记取消回调。
	Has(id string) bool
	// Stop 触发指定部署任务的取消回调（回调收到 deploy.ErrCancel）。
	Stop(id string)
	// Register 登记一个部署任务的取消回调；同 id 已存在时返回 errSignalExists。
	Register(id string, fn func(error)) error
	// StopAll 触发全部在途部署的取消回调（连接断开时的统一清理）。
	StopAll()
}

var _ TaskManager = (*taskManager)(nil)

// taskManager 以 id（部署任务 slug）→ 取消回调 的映射登记每个在途部署，
// 供取消部署、连接断开（CloseAndClean → StopAll）时统一触发。
type taskManager struct {
	tasks map[string]func(error)
	sync.RWMutex
	logger mlog.Logger
}

// NewTaskManager 构造部署取消任务注册表（空映射 + 注入 logger）。
func NewTaskManager(logger mlog.Logger) TaskManager {
	return &taskManager{tasks: map[string]func(error){}, logger: logger}
}

// Remove 删除指定部署任务的取消回调。
func (tm *taskManager) Remove(id string) {
	tm.Lock()
	defer tm.Unlock()
	delete(tm.tasks, id)
}

// Has 判断指定部署任务是否已登记取消回调。
func (tm *taskManager) Has(id string) bool {
	tm.RLock()
	defer tm.RUnlock()

	_, ok := tm.tasks[id]

	return ok
}

// Stop 触发指定部署任务的取消回调（回调收到 deploy.ErrCancel）。
func (tm *taskManager) Stop(id string) {
	tm.Lock()
	defer tm.Unlock()
	if fn, ok := tm.tasks[id]; ok {
		tm.logger.Debugf("stop task\t%v\t%v", id, deploy.ErrCancel)
		fn(deploy.ErrCancel)
	}
}

// Register 登记一个部署任务的取消回调；同 id 已存在时返回 errSignalExists。
func (tm *taskManager) Register(id string, fn func(error)) error {
	tm.Lock()
	defer tm.Unlock()
	if _, ok := tm.tasks[id]; ok {
		return errSignalExists
	}
	tm.tasks[id] = fn
	return nil
}

// StopAll 触发全部在途部署的取消回调（连接断开时的统一清理）。
// 与手动取消（Stop → deploy.ErrCancel）区分，传 ErrCancelConnClosed 供 Finish 判定取消来源并留痕。
func (tm *taskManager) StopAll() {
	tm.Lock()
	defer tm.Unlock()
	for _, f := range tm.tasks {
		f(deploy.ErrCancelConnClosed)
	}
}
