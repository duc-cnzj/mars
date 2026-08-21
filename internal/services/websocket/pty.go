package websocket

// pty.go 定义容器内终端会话：
//   - PtyHandler/ptyHandler：连接容器终端 stdin/stdout 与 ws 帧的桥
//   - sizeStore：终端尺寸与"尺寸重置"标志
//   - Op* 常量与 ETX/传输结束符：终端消息的操作类型与控制字符

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/closeable"
)

// ETX 是 Ctrl-C 控制字符（\x03），用于中断容器内的前台进程。
var ETX = []byte("\u0003")

// END_OF_TRANSMISSION 是传输结束字符（\x04），用于结束终端输入流。
var END_OF_TRANSMISSION = []byte("\u0004")

// Op* 是终端消息（websocket_pb.TerminalMessage.Op）的操作类型标签。
const (
	OpStdout = "stdout" // 容器输出 → 客户端
	OpStdin  = "stdin"  // 客户端输入 → 容器
	OpResize = "resize" // 客户端终端尺寸变化
	OpToast  = "toast"  // OOB 提示信息（终端中央展示）
)

// shellDrainTimeout 是 Close 等待输入队列排空（ETX/EOT 已被 Read 消费）的上限；
// 超过即放弃优雅退出，交由 close(doneChan) 后的 SIGHUP 兜底。
const shellDrainTimeout = 200 * time.Millisecond

// sizeStore 记录终端尺寸（width/height）与"尺寸重置"标志，供 pty 在
// shell 回退后按用户上次尺寸重建终端时使用；内部以 RWMutex 保护并发访问。
type sizeStore struct {
	rwMu          sync.RWMutex
	width, height uint16
	reset         bool
}

// ResetTerminalRowCol 设置是否需要重置终端行列尺寸。
func (s *sizeStore) ResetTerminalRowCol(reset bool) {
	s.rwMu.Lock()
	defer s.rwMu.Unlock()
	s.reset = reset
}

// TerminalRowColNeedReset 返回是否需要重置终端行列尺寸。
func (s *sizeStore) TerminalRowColNeedReset() bool {
	s.rwMu.RLock()
	defer s.rwMu.RUnlock()
	return s.reset
}

// Set 保存终端尺寸。
func (s *sizeStore) Set(width, height uint16) {
	s.rwMu.Lock()
	defer s.rwMu.Unlock()
	s.height = height
	s.width = width
}

// Changed 判断给定尺寸与已存尺寸是否不同（用于触发 resize 上报）。
func (s *sizeStore) Changed(width, height uint16) bool {
	s.rwMu.RLock()
	defer s.rwMu.RUnlock()
	if s.height != height {
		return true
	}
	if s.width != width {
		return true
	}

	return false
}

// Width 返回终端宽度。
func (s *sizeStore) Width() uint16 {
	s.rwMu.RLock()
	defer s.rwMu.RUnlock()
	return s.width
}

// Height 返回终端高度。
func (s *sizeStore) Height() uint16 {
	s.rwMu.RLock()
	defer s.rwMu.RUnlock()
	return s.height
}

// PtyHandler 是容器内终端会话的处理器接口：既是远端命令的输入输出
// （io.Reader/io.Writer/TerminalSizeQueue），也承载尺寸、录音、关闭等生命周期。
// TerminalSizeQueue 直接实现 biz 领域端口，数据层再适配回 client-go，传输层不触碰基础设施类型。
type PtyHandler interface {
	io.Reader
	io.Writer
	biz.TerminalSizeQueue

	// Container 返回该终端会话绑定的容器。
	Container() *biz.Container
	// SetShell 设置会话对应的 shell 命令路径（如 /bin/sh）。
	SetShell(string)
	// Toast 向客户端发送 OOB 提示信息（终端中央展示）。
	Toast(string) error

	// Send 处理一条终端消息（stdin 输入 / resize 尺寸变化）。
	Send(ctx context.Context, message *websocket_pb.TerminalMessage) error
	// Resize 同步远端终端尺寸。
	Resize(biz.TerminalSize) error

	// Recorder 返回该会话的输入录制器（回放审计）。
	Recorder() biz.Recorder

	// ResetTerminalRowCol 设置 shell 回退后是否需要重置终端行列尺寸。
	ResetTerminalRowCol(bool)
	// Height 返回终端高度。
	Height() uint16
	// Width 返回终端宽度。
	Width() uint16

	// Close 关闭会话并向客户端发送退出状态；返回是否已关闭。
	Close(context.Context, string) bool
	// IsClosed 返回会话是否已关闭。
	IsClosed() bool
}

// ptyHandler 是 PtyHandler 的生产实现：连接容器终端的 stdin/stdout 与 ws 帧
// 之间的桥，内部用三个 channel（doneChan/shellCh/sizeChan）解耦读写与关闭。
type ptyHandler struct {
	logger    mlog.Logger
	sessionID string
	container *biz.Container
	recorder  biz.Recorder
	eventRepo biz.EventRepo
	conn      Conn

	doneChan  chan struct{}
	sizeStore *sizeStore

	shellMu sync.RWMutex
	shellCh chan *websocket_pb.TerminalMessage

	sizeMu   sync.RWMutex
	sizeChan chan biz.TerminalSize

	closeable.Closeable
}

// SetShell 记录本次使用的 shell 名（供录音器标识）。
func (t *ptyHandler) SetShell(shell string) {
	t.recorder.SetShell(shell)
}

// Container 返回终端所在的容器信息。
func (t *ptyHandler) Container() *biz.Container {
	return t.container
}

// Height 返回终端高度。
func (t *ptyHandler) Height() uint16 {
	return t.sizeStore.Height()
}

// Width 返回终端宽度。
func (t *ptyHandler) Width() uint16 {
	return t.sizeStore.Width()
}

// Read 从输入队列取一条终端消息：stdin 数据直接返回，resize 触发尺寸更新，
// 队列关闭或会话结束返回传输结束符。
func (t *ptyHandler) Read(p []byte) (n int, err error) {
	var (
		msg *websocket_pb.TerminalMessage
		ok  bool
	)
	select {
	case <-t.doneChan:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("[Websocket]: %v doneChan closed", t.sessionID)
	case msg, ok = <-t.shellCh:
		if !ok {
			return copy(p, END_OF_TRANSMISSION), fmt.Errorf("[Websocket]: %v channel closed", t.sessionID)
		}
	}

	switch msg.Op {
	case OpStdin:
		return copy(p, msg.Data), nil
	case OpResize:
		t.logger.Debugf("[Websocket]: resize width: %v  height: %v", msg.Width, msg.Height)
		t.Resize(biz.TerminalSize{Width: uint16(msg.Width), Height: uint16(msg.Height)})
		return 0, nil
	default:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// shellResponse 收敛 Write/Close/Toast 三处相同的 WsHandleShellResponse 组装
// （Metadata 的 Id/Uid/Slug/Result 与 Container 恒同），仅 Type/Op/Data 变化。
func (t *ptyHandler) shellResponse(wsType websocket_pb.Type, op string, data []byte) *websocket_pb.WsHandleShellResponse {
	return &websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     t.conn.ID(),
			Uid:    t.conn.UID(),
			Slug:   t.sessionID,
			Type:   wsType,
			Result: deploy.ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			Op:        op,
			Data:      data,
			SessionId: t.sessionID,
		},
		Container: &websocket_pb.Container{
			Namespace: t.Container().Namespace,
			Pod:       t.Container().Pod,
			Container: t.Container().Container,
		},
	}
}

// Write 把容器 stdout 转发给客户端，并在需要时按上次尺寸重置终端。
func (t *ptyHandler) Write(p []byte) (n int, err error) {
	select {
	case <-t.doneChan:
		return len(p), fmt.Errorf("[Websocket]: %v doneChan closed", t.sessionID)
	default:
	}
	if t.IsClosed() {
		return len(p), fmt.Errorf("[Websocket]: %v ws already closed", t.sessionID)
	}

	if _, err = t.recorder.Write(p); err != nil {
		t.logger.Debugf("[Websocket]: %v recorder write failed: %v", t.sessionID, err)
	}
	if t.sizeStore.TerminalRowColNeedReset() && t.sizeStore.Width() != 0 {
		t.logger.Debugf("reset shell size height: %d, width: %d", t.sizeStore.Height(), t.sizeStore.Width())
		t.sizeStore.ResetTerminalRowCol(false)
		if err = t.Resize(biz.TerminalSize{Width: t.sizeStore.Width(), Height: t.sizeStore.Height()}); err != nil {
			t.logger.Debugf("resize shell size failed: %v", err)
		}
	}
	newMessageSender(t.conn, WsHandleExecShellMsg).SendProtoMsg(t.shellResponse(WsHandleExecShellMsg, OpStdout, p))

	return len(p), nil
}

// ResetTerminalRowCol 设置是否需要重置终端行列尺寸。
func (t *ptyHandler) ResetTerminalRowCol(reset bool) {
	t.sizeStore.ResetTerminalRowCol(reset)
}

// Recorder 返回会话录音器。
func (t *ptyHandler) Recorder() biz.Recorder {
	return t.recorder
}

// Next 从尺寸队列取下一个终端尺寸（供数据层适配回 remotecommand 轮询）。
func (t *ptyHandler) Next() *biz.TerminalSize {
	select {
	case size, ok := <-t.sizeChan:
		if !ok {
			return nil
		}
		if size.Width != 0 && size.Height != 0 {
			if t.sizeStore.Changed(size.Width, size.Height) {
				t.recorder.Resize(size.Width, size.Height)
			}
			t.sizeStore.Set(size.Width, size.Height)
		}
		return &size
	case <-t.doneChan:
		return nil
	}
}

// Send 把客户端终端消息放入输入队列（满则告警丢弃）。
func (t *ptyHandler) Send(ctx context.Context, m *websocket_pb.TerminalMessage) error {
	t.shellMu.Lock()
	defer t.shellMu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.doneChan:
		close(t.shellCh)
		return errors.New("doneChan closed")
	default:
	}

	select {
	case t.shellCh <- m:
	default:
		t.logger.Warning("[Websocket]: shellCh chan full")
	}
	return nil
}

// Resize 把终端尺寸放入尺寸队列（满则报错）。
func (t *ptyHandler) Resize(size biz.TerminalSize) error {
	select {
	case <-t.doneChan:
		close(t.sizeChan)
		return errors.New("doneChan closed")
	default:
	}

	t.sizeMu.Lock()
	defer t.sizeMu.Unlock()
	select {
	case t.sizeChan <- size:
	default:
		return errors.New("sizeChan chan full")
	}
	return nil
}

// IsClosed 返回会话是否已关闭。
func (t *ptyHandler) IsClosed() bool {
	return t.Closeable.IsClosed()
}

// CloseDoneChan 关闭会话并关闭 doneChan（幂等）。
func (t *ptyHandler) CloseDoneChan() bool {
	if !t.Closeable.Close() {
		return false
	}
	close(t.doneChan)
	return true
}

// sendControlFrame 尽力把一条 stdin 控制帧（ETX/EOT）投递到输入队列。
// k8s exec 没有对端确认通道，投递失败（ctx 取消/队列满）只记日志，不阻断关闭流程。
func (t *ptyHandler) sendControlFrame(ctx context.Context, data []byte) {
	if err := t.Send(ctx, &websocket_pb.TerminalMessage{
		Op:        OpStdin,
		Data:      data,
		SessionId: t.sessionID,
	}); err != nil {
		t.logger.Debugf("[Websocket]: %v 投递 stdin 控制帧失败: %v", t.sessionID, err)
	}
}

// waitShellDrained 在时限内等输入队列排空（ETX/EOT 已被 Read 消费出队），
// 保证 close(doneChan) 前控制帧已送达，避免 Read 的 select 随机丢帧；
// 超时（read 循环被 exec 背压卡死）则放弃优雅退出，交给 SIGHUP 兜底。
func (t *ptyHandler) waitShellDrained(timeout time.Duration) {
	ticker := time.NewTicker(5 * time.Millisecond)
	af := time.NewTimer(timeout)
	defer ticker.Stop()
	defer af.Stop()
	for {
		t.shellMu.RLock()
		empty := len(t.shellCh) == 0
		t.shellMu.RUnlock()
		if empty {
			return
		}
		select {
		case <-ticker.C:
		case <-af.C:
			t.logger.Debugf("[Websocket]: %v 输入队列未排空，放弃优雅退出", t.sessionID)
			return
		}
	}
}

// Close 关闭会话：推送关闭帧、注入 Ctrl-C/Ctrl-D 控制符、写录音审计日志并关闭 doneChan。
// 控制帧是"尽力而为"的优雅退出：先 Ctrl-C（\x03）中断容器前台进程，再 Ctrl-D（\x04）
// 让 shell 读侧收 EOF，送达由 waitShellDrained 保证；真正的强关由 close(doneChan)
// → Read 返错 → exec 断流 → 容器 shell 收 SIGHUP 兜底。
func (t *ptyHandler) Close(ctx context.Context, reason string) bool {
	if !t.Closeable.Close() {
		return false
	}
	newMessageSender(t.conn, WsHandleCloseShell).SendProtoMsg(t.shellResponse(WsHandleCloseShell, OpStdout, []byte(reason)))

	t.sendControlFrame(ctx, ETX)
	t.sendControlFrame(ctx, END_OF_TRANSMISSION)
	t.waitShellDrained(shellDrainTimeout)

	t.logger.Debug("[Websocket]: close shell.")
	if err := t.Recorder().Close(); err != nil {
		t.logger.Error(err)
	}
	recoder := t.Recorder()
	var fid int
	rf := recoder.File()
	if rf != nil {
		fid = rf.ID
	}
	t.eventRepo.FileAuditLogWithDuration(
		types.EventActionType_Shell,
		recoder.User().Name,
		recoder.User().Email,
		fmt.Sprintf("用户进入容器执行命令，container: '%s', namespace: '%s', pod： '%s'", recoder.Container().Container, recoder.Container().Namespace, recoder.Container().Pod),
		fid,
		recoder.Duration(),
	)
	close(t.doneChan)
	return true
}

// Toast 向客户端终端推送 OOB 提示信息（hterm 在终端中央展示）。
func (t *ptyHandler) Toast(p string) error {
	newMessageSender(t.conn, WsHandleExecShellMsg).SendProtoMsg(t.shellResponse(WsHandleExecShellMsg, OpToast, []byte(p)))
	return nil
}
