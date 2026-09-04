package ws

// terminal.go 定义终端会话高层抽象（Terminal）与唯一入口 OpenTerminal：
// 单调用即可连到指定容器拉起 shell 并拿到可读写的 Terminal，内部自动处理
// 连接/鉴权等待、sessionID 生成、shell 开启与鉴权竞态兜底。调用方只需
// Write(stdin)/读 Stdout()/Resize/Close，不接触任何底层帧。

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	termio "golang.org/x/term"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"google.golang.org/protobuf/proto"
)

// 终端帧操作类型（对齐服务端 pty.go 的 Op* 常量）。
const (
	opStdout = "stdout"
	opToast  = "toast"
)

// maxOpenRetries 是鉴权竞态兜底的最大重试次数：服务器在鉴权完成前会把
// ExecShell 回"认证中，请稍等~"丢弃，OpenTerminal 借此自动重发直到 shell 开启。
const maxOpenRetries = 20

// Terminal 是一条 ws 终端会话的抽象：
//   - Write 走 stdin，Resize 调整尺寸；
//   - Stdout()/Toast() 接收容器输出与 OOB 提示；
//   - 服务端推送 WsHandleCloseShell 帧或调用 Close 后，Done 通道关闭。
type Terminal struct {
	client    *Client
	sessionID string
	container *websocket_pb.Container

	stdout chan []byte
	toast  chan []byte
	done   chan struct{}
	once   sync.Once

	// cancel 解绑 onSession 订阅；retryCancel 解绑鉴权竞态兜底的 onType 订阅。
	cancel func()
	// opened 在收到 shell-open(type=50) 帧后关闭，用于停止鉴权竞态兜底重试。
	opened     chan struct{}
	openedOnce sync.Once
	// retries 是已发起的竞态重试次数（仅 readLoop goroutine 内访问，无需锁）。
	retries     int
	retryCancel func()
}

// OpenTerminal 在指定容器内拉起一个交互 shell，返回可读写的 Terminal。
// 单调用完成全部交互：内部等待连接/鉴权就绪、自动生成 sessionID、发起
// ExecShell 并订阅该会话的 stdout/toast/关闭帧。ctx 用于控制等待鉴权的时限。
//
// 内置鉴权竞态兜底：协议没有鉴权成功的 ack，服务器可能在鉴权完成前把 ExecShell
// 回"认证中，请稍等~"丢弃（表现为"连上了但没输出"）。OpenTerminal 收到该 gate 帧
// 且 shell 尚未开启时，会隔 150ms 自动重发 ExecShell，直到 shell 真正开启或达到
// maxOpenRetries 次，调用方无感知。
func (c *Client) OpenTerminal(ctx context.Context, container *websocket_pb.Container) (*Terminal, error) {
	if container == nil {
		return nil, errors.New("ws: container 不能为空")
	}
	if err := c.waitReady(ctx); err != nil {
		return nil, err
	}
	sessionID := newSessionID()
	if err := c.execShell(container, sessionID); err != nil {
		return nil, err
	}
	t := &Terminal{
		client:    c,
		sessionID: sessionID,
		container: container,
		stdout:    make(chan []byte, 64),
		toast:     make(chan []byte, 8),
		done:      make(chan struct{}),
		opened:    make(chan struct{}),
	}
	t.cancel = c.onSession(sessionID, t.handle)
	t.retryCancel = c.onType(websocket_pb.Type_HandleAuthorize, t.retryGate)
	return t, nil
}

// handle 是终端会话的订阅回调：把帧翻译成 stdout/toast 通道、结束信号或
// shell 开启确认。仅 readLoop goroutine 调用（onSession 分发），无需加锁。
func (t *Terminal) handle(ev *Event) {
	var resp websocket_pb.WsHandleShellResponse
	if err := proto.Unmarshal(ev.Raw, &resp); err != nil || resp.TerminalMessage == nil {
		return
	}
	// shell-open(type=50) 帧 = 服务器确认 shell 已拉起 → 停止鉴权竞态兜底重试。
	if ev.Type == websocket_pb.Type_HandleExecShell {
		t.openedOnce.Do(func() { close(t.opened) })
	}
	switch resp.TerminalMessage.Op {
	case opStdout:
		select {
		case t.stdout <- resp.TerminalMessage.Data:
		default: // 缓冲满丢弃，避免阻塞读循环
		}
	case opToast:
		select {
		case t.toast <- resp.TerminalMessage.Data:
		default:
		}
	}
	// 服务端主动关闭会话（进程退出/被踢/异常）时通知结束。
	if ev.Type == websocket_pb.Type_HandleCloseShell {
		t.once.Do(func() { close(t.done) })
	}
}

// retryGate 是鉴权竞态兜底：服务器把 ExecShell 回"认证中，请稍等~"gate 帧时，
// 若 shell 尚未开启就隔 150ms 重发 ExecShell。仅 readLoop goroutine 内调用
// （onType 分发），与 handle 串行；重发本身在独立 goroutine 里做，不阻塞读循环。
func (t *Terminal) retryGate(ev *Event) {
	if ev.Metadata == nil || !strings.Contains(ev.Metadata.Message, "认证中") {
		return
	}
	select {
	case <-t.opened:
		return
	default:
	}
	if t.retries >= maxOpenRetries {
		return
	}
	t.retries++
	go func() {
		time.Sleep(150 * time.Millisecond) // 给服务器留出处理 Authorize 的时间
		_ = t.client.execShell(t.container, t.sessionID)
	}()
}

// ID 返回终端会话 id（即 sessionID）。
func (t *Terminal) ID() string {
	return t.sessionID
}

// Write 把 p 作为 stdin 发送给容器进程。
func (t *Terminal) Write(p []byte) (int, error) {
	if err := t.client.sendStdin(t.sessionID, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Resize 调整终端窗口尺寸。
func (t *Terminal) Resize(height, width uint32) error {
	return t.client.resize(t.sessionID, height, width)
}

// Stdout 返回容器输出流（stdout 帧 data）。
func (t *Terminal) Stdout() <-chan []byte {
	return t.stdout
}

// Toast 返回 OOB 提示信息流（toast 帧 data）。
func (t *Terminal) Toast() <-chan []byte {
	return t.toast
}

// Done 返回会话结束信号通道：服务端推送关闭帧或调用 Close 后关闭。
func (t *Terminal) Done() <-chan struct{} {
	return t.done
}

// Close 请求服务端关闭会话并解绑订阅。
func (t *Terminal) Close() error {
	t.once.Do(func() { close(t.done) })
	if t.cancel != nil {
		t.cancel()
	}
	if t.retryCancel != nil {
		t.retryCancel()
	}
	return t.client.closeShell(t.sessionID)
}

// Pump 编排终端会话的数据面：接好三条 I/O 通道并启动转发 goroutine，
// 返回的 stop 用于调用方主动终止转发（幂等，可安全多次调用）。
// 尺寸同步（SIGWINCH）与退出信号（SIGINT/SIGTERM）属于控制面，由调用方自行处理。
//   - in（字节来源，如 os.Stdin）→ 远端转发（canonical 模式，回车后整行转发）；
//   - stdout / toast 消费闭包（nil 表示丢弃该通道）；
//   - opts 配置行为，默认开启 raw 模式（见 WithRawMode）。
func (t *Terminal) Pump(in io.Reader, stdout, toast func([]byte), opts ...PumpOption) (stop func()) {
	cfg := pumpConfig{rawMode: true}
	for _, o := range opts {
		o(&cfg)
	}

	quit := make(chan struct{})
	var once sync.Once // 保证 stop 幂等，二次调用不重复 close/restore

	// raw 模式：把 in（若为 tty 文件）切到 raw 模式，每个按键字节即时透传远端 shell，
	// 从而获得 tab 补全/方向键/clear 等交互；stop 时自动恢复本地终端设置。
	// 注：MakeRaw 成功分支为 tty 集成边界（无 pty helper 可造真实 tty），单测不覆盖。
	restoreRaw := func() {}
	if cfg.rawMode {
		if f, ok := in.(*os.File); ok && termio.IsTerminal(int(f.Fd())) {
			if st, err := termio.MakeRaw(int(f.Fd())); err == nil {
				restoreRaw = func() { _ = termio.Restore(int(f.Fd()), st) }
			}
		}
	}

	// 远端标准输出 → 消费闭包。
	// 注：t.stdout 为未导出字段且库内从不 close，故无需 ok 判关闭。
	if stdout != nil {
		go func() {
			for {
				select {
				case data := <-t.stdout:
					stdout(data)
				case <-quit:
					return
				}
			}
		}()
	}

	// 远端 OOB 提示 → 消费闭包。
	if toast != nil {
		go func() {
			for {
				select {
				case data := <-t.toast:
					toast(data)
				case <-quit:
					return
				}
			}
		}()
	}

	// in → 远端。写失败多为远端已关，静默停止转发即可。
	go func() {
		buf := make([]byte, 4096)
		for {
			n, rerr := in.Read(buf)
			if n > 0 {
				if _, werr := t.Write(buf[:n]); werr != nil {
					return
				}
			}
			if rerr != nil {
				return
			}
		}
	}()

	return func() {
		once.Do(func() {
			close(quit)
			restoreRaw()
		})
	}
}

// PumpOption 配置 Terminal.Pump 的行为。
type PumpOption func(*pumpConfig)

// pumpConfig 承载 Pump 的可配置项。
type pumpConfig struct {
	// rawMode 是否在启动时把本地终端切到 raw 模式（默认开启）。
	rawMode bool
}

// WithRawMode 控制 Pump 是否把 in（若为 tty）切到 raw 模式。
// raw 模式下本地不缓冲、不回显，每个按键字节即时透传远端 shell，从而获得
// tab 补全/方向键/clear 等交互；stop 时自动恢复本地终端设置。
// 传 false 退回 canonical 模式（本地回显开启、Ctrl+C 走本地 SIGINT）。
func WithRawMode(enabled bool) PumpOption {
	return func(c *pumpConfig) { c.rawMode = enabled }
}

// AutoHandleWindowSize 自动跟随本地终端窗口尺寸变化（初始尺寸 + 每次变化）同步远端 pty。
// 依赖底层平台能力（unix 用 ioctl 读尺寸 + SIGWINCH 监听），不支持的平台为 no-op。
// 返回的 stop 用于停止监听，幂等可安全多次调用。
func (t *Terminal) AutoHandleWindowSize() (stop func()) {
	current, changes, stopSrc := windowSizeSource()
	return autoHandleWindowSize(current, changes, stopSrc, func(h, w uint32) { _ = t.Resize(h, w) })
}

// autoHandleWindowSize 是 AutoHandleWindowSize 的平台无关内核：current 读尺寸、
// changes 收变化通知、stopSrc 停止监听、resize 应用尺寸（注入便于测试）。
func autoHandleWindowSize(
	current func() (uint32, uint32, bool),
	changes <-chan os.Signal,
	stopSrc func(),
	resize func(h, w uint32),
) (stop func()) {
	if current == nil || changes == nil {
		return func() {}
	}
	// 初始尺寸立即同步。
	if h, w, ok := current(); ok {
		resize(h, w)
	}
	quit := make(chan struct{})
	var once sync.Once
	go func() {
		for {
			select {
			case <-changes:
				if h, w, ok := current(); ok {
					resize(h, w)
				}
			case <-quit:
				stopSrc()
				return
			}
		}
	}()
	return func() { once.Do(func() { close(quit) }) }
}
