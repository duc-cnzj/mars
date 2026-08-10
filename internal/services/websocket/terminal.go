package websocket

// terminal.go 定义终端编排逻辑：StartShell 拉起容器内 shell，
// runTerminal/execInContainer/resetSession 驱动 pty 的完整生命周期，
// 以及 shell 白名单、sessionID 格式校验等辅助函数。

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"k8s.io/client-go/tools/remotecommand"
)

// execInContainer 在指定容器内执行 cmd，并把 stdin/stdout/stderr 及终端 resize
// 全部接到 pty 上（TTY 模式），由 runTerminal 在用户发起 shell 时调用。
func (wc *websocketManager) execInContainer(ctx context.Context, container *biz.Container, cmd []string, pty PtyHandler) error {
	return wc.k8sRepo.Execute(ctx, container, &biz.ExecuteInput{
		Stdin:             pty,
		Stdout:            pty,
		Stderr:            pty,
		TTY:               true,
		Cmd:               cmd,
		TerminalSizeQueue: pty,
	})
}

// isValidShell 判断 shell 是否在白名单内。
func isValidShell(validShells []string, shell string) bool {
	for _, validShell := range validShells {
		if validShell == shell {
			return true
		}
	}
	return false
}

// silenceShellExitMessages 是无需向用户弹出的 shell 退出信息（如用户按 Ctrl-C/Ctrl-D 退出）。
var silenceShellExitMessages = []string{
	"command terminated with exit code 126",
	"command terminated with exit code 130",
}

// shouldSilenceShellError 判断 shell 退出错误是否需要静默（命中上述列表则不弹 toast）。
func shouldSilenceShellError(err error) bool {
	for _, message := range silenceShellExitMessages {
		if strings.Contains(err.Error(), message) {
			return true
		}
	}
	return false
}

// runTerminal 由 StartShell 以 goroutine 启动：按用户请求的 shell（缺省时依次
// 回退 bash/sh/powershell/cmd）在容器内拉起终端进程，成功后以 status 1 收尾，
// 失败则向会话弹 toast 并以 status 2 关闭 pty。
func (wc *websocketManager) runTerminal(ctx context.Context, conn Conn, container *biz.Container, shell, sessionID string) {
	defer func() {
		wc.logger.Debugf("[Websocket]: runTerminal EXIT: total go: %v", runtime.NumGoroutine())
	}()
	var err error
	validShells := []string{"bash", "sh", "powershell", "cmd"}
	session, got := conn.GetPtyHandler(sessionID)
	if !got {
		return
	}
	if isValidShell(validShells, shell) {
		cmd := []string{shell}
		session.SetShell(shell)
		err = wc.execInContainer(ctx, container, cmd, session)
	} else {
		// 未指定 shell 或不在白名单：依次尝试 bash/sh/powershell/cmd，直到有一个成功或全部失败。
		// FIXME: 若第一个 shell 启动失败，第一次键盘输入会丢失。
		for idx, testShell := range validShells {
			wc.logger.Debug("try: " + testShell)
			if session.IsClosed() {
				wc.logger.Debugf("session 已关闭，不会继续尝试连接其他 shell: '%s'", strings.Join(validShells[idx:], ", "))
				break
			}
			cmd := []string{testShell}
			session.SetShell(testShell)
			if err = wc.execInContainer(ctx, container, cmd, session); err == nil {
				break
			}
			// 当出现 bash 回退的时候，需要注意，resize 不会触发，导致，新的 'sh', width, height 和用户端不一致，所以需要重置，
			// 通过 sizeStore 记录上次用户的 height, width, 当 bash 回退时，在用户输入时应用到新的 sh 中
			session = wc.resetSession(session)
			conn.SetPtyHandler(sessionID, session)
		}
	}

	if err != nil {
		wc.logger.Debugf("[Websocket]: %v", err.Error())
		if !shouldSilenceShellError(err) {
			session.Toast(err.Error())
		}
		conn.ClosePty(ctx, sessionID, 2, err.Error())
		return
	}

	conn.ClosePty(ctx, sessionID, 1, "Process exited")
}

// resetSession 在 shell 回退后重建 ptyHandler：先等上一会话拿到用户尺寸
// （超时用默认 106x25），再以新 channel + 重置标志克隆一个尺寸正确的会话。
func (wc *websocketManager) resetSession(session PtyHandler) PtyHandler {
	var width, height uint16 = 106, 25
	func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		af := time.NewTimer(3 * time.Second)
		defer ticker.Stop()
		defer af.Stop()
		for session.Width() == 0 {
			// 轮询等待 resize 生效：select 内的 break 只会跳出 select，
			// 去掉它，由 for 条件重新评估是否已拿到 width/height。
			select {
			case <-ticker.C:
				wc.logger.Debug("sleep....")
			case <-af.C:
				wc.logger.Warningf("can't get previous width,height, use default height: 25, width: 106.")
				return
			}
		}
		width = session.Width()
		height = session.Height()
	}()
	wc.logger.Debug("done....")

	spty := session.(*ptyHandler)
	var newSession PtyHandler = session
	if spty.CloseDoneChan() {
		newSession = &ptyHandler{
			logger:    spty.logger,
			sessionID: spty.sessionID,
			container: spty.container,
			recorder:  spty.recorder,
			eventRepo: spty.eventRepo,
			conn:      spty.conn,
			doneChan:  make(chan struct{}),
			sizeChan:  make(chan remotecommand.TerminalSize, 1),
			shellCh:   make(chan *websocket_pb.TerminalMessage, 500),
			sizeStore: &sizeStore{
				width:  width,
				height: height,
				reset:  true,
			},
		}
	}
	return newSession
}

// isValidSessionID 校验 sessionID 是否符合 '<namespace>-<pod>-<container>:' 前缀。
func isValidSessionID(container *websocket_pb.Container, id string) bool {
	prefix := fmt.Sprintf("%s-%s-%s:", container.Namespace, container.Pod, container.Container)
	return strings.HasPrefix(id, prefix)
}

// StartShell 为请求的容器创建 ptyHandler 会话并登记到连接，
// 再以 goroutine 启动 runTerminal 拉 shell；返回会话 sessionID。
func (wc *websocketManager) StartShell(ctx context.Context, input *websocket_pb.WsHandleExecShellInput, conn Conn) (string, error) {
	var (
		container = &biz.Container{
			Namespace: input.Container.Namespace,
			Pod:       input.Container.Pod,
			Container: input.Container.Container,
		}
		sessionID = input.SessionId
	)

	if !isValidSessionID(input.Container, sessionID) {
		return "", fmt.Errorf("invalid session sessionID, must format: '<namespace>-<pod>-<container>:<randomID>', input: '%s'", sessionID)
	}

	pty := &ptyHandler{
		logger:    wc.logger,
		sessionID: sessionID,
		eventRepo: wc.eventRepo,
		container: container,
		recorder:  wc.fileRepo.NewRecorder(conn.GetUser(), container),
		conn:      conn,
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
		shellCh:   make(chan *websocket_pb.TerminalMessage, 500),
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
	}
	conn.SetPtyHandler(sessionID, pty)

	go func() {
		defer wc.logger.HandlePanic("Websocket: runTerminal")
		wc.runTerminal(ctx, conn, container, "", sessionID)
	}()

	return sessionID, nil
}
