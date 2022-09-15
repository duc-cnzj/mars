package socket

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/internal/utils/recovery"

	"github.com/duc-cnzj/mars-client/v4/types"
	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

var (
	ETX                 = []byte("\u0003")
	END_OF_TRANSMISSION = []byte("\u0004")
)

const (
	OpStdout = "stdout"
	OpStdin  = "stdin"
	OpResize = "resize"
	OpToast  = "toast"
)

type sizeStore struct {
	rwMu       sync.RWMutex
	cols, rows uint16

	resetMu sync.RWMutex
	reset   bool
}

func (s *sizeStore) ResetTerminalRowCol(reset bool) {
	s.resetMu.Lock()
	defer s.resetMu.Unlock()
	s.reset = reset
}

func (s *sizeStore) TerminalRowColNeedReset() bool {
	s.resetMu.RLock()
	defer s.resetMu.RUnlock()
	return s.reset
}

func (s *sizeStore) Set(cols, rows uint16) {
	s.rwMu.Lock()
	defer s.rwMu.Unlock()
	s.rows = rows
	s.cols = cols
}

func (s *sizeStore) Changed(cols, rows uint16) bool {
	s.rwMu.RLock()
	defer s.rwMu.RUnlock()
	if s.rows == 0 || s.cols == 0 {
		return false
	}
	if s.rows != rows {
		return true
	}
	if s.cols != cols {
		return true
	}

	return false
}

func (s *sizeStore) Cols() uint16 {
	s.rwMu.RLock()
	defer s.rwMu.RUnlock()
	return s.cols
}

func (s *sizeStore) Rows() uint16 {
	s.rwMu.RLock()
	defer s.rwMu.RUnlock()
	return s.rows
}

type MyPtyHandler struct {
	container contracts.Container
	recorder  contracts.RecorderInterface
	id        string
	conn      *WsConn
	doneChan  chan struct{}
	sizeStore sizeStore

	shellMu sync.RWMutex
	shellCh chan *websocket_pb.TerminalMessage

	sizeMu   sync.RWMutex
	sizeChan chan remotecommand.TerminalSize

	closeMu sync.RWMutex
	closed  bool
}

func (t *MyPtyHandler) SetShell(shell string) {
	t.recorder.SetShell(shell)
}

func (t *MyPtyHandler) Container() contracts.Container {
	return t.container
}

func (t *MyPtyHandler) Rows() uint16 {
	return t.sizeStore.Rows()
}

func (t *MyPtyHandler) Cols() uint16 {
	return t.sizeStore.Cols()
}

func (t *MyPtyHandler) Read(p []byte) (n int, err error) {
	var (
		msg *websocket_pb.TerminalMessage
		ok  bool
	)
	select {
	case <-t.doneChan:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("[Websocket]: %v doneChan closed", t.id)
	case msg, ok = <-t.shellCh:
		if !ok {
			return copy(p, END_OF_TRANSMISSION), fmt.Errorf("[Websocket]: %v channel closed", t.id)
		}
	}

	switch msg.Op {
	case OpStdin:
		return copy(p, msg.Data), nil
	case OpResize:
		mlog.Debugf("[Websocket]: resize cols: %v  rows: %v", msg.Cols, msg.Rows)
		t.Resize(remotecommand.TerminalSize{Width: uint16(msg.Cols), Height: uint16(msg.Rows)})
		return 0, nil
	default:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

func (t *MyPtyHandler) Write(p []byte) (n int, err error) {
	select {
	case <-t.doneChan:
		return len(p), fmt.Errorf("[Websocket]: %v doneChan closed", t.id)
	default:
	}
	if t.IsClosed() {
		return len(p), fmt.Errorf("[Websocket]: %v ws already closed", t.id)
	}
	t.recorder.Write(string(p))
	if t.sizeStore.TerminalRowColNeedReset() && t.sizeStore.Cols() != 0 {
		mlog.Debugf("reset shell size rows: %d, cols: %d", t.sizeStore.Rows(), t.sizeStore.Cols())
		t.sizeStore.ResetTerminalRowCol(false)
		t.Resize(remotecommand.TerminalSize{Width: t.sizeStore.Cols(), Height: t.sizeStore.Rows()})
	}
	NewMessageSender(t.conn, t.id, WsHandleExecShellMsg).SendProtoMsg(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     t.conn.id,
			Uid:    t.conn.uid,
			Slug:   t.id,
			Type:   WsHandleExecShellMsg,
			Result: ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			Op:        OpStdout,
			Data:      p,
			SessionId: t.id,
		},
		Container: &types.Container{
			Namespace: t.Container().Namespace,
			Pod:       t.Container().Pod,
			Container: t.Container().Container,
		},
	})

	return len(p), nil
}

func (t *MyPtyHandler) ResetTerminalRowCol(reset bool) {
	t.sizeStore.ResetTerminalRowCol(reset)
}

func (t *MyPtyHandler) Recorder() contracts.RecorderInterface {
	return t.recorder
}

func (t *MyPtyHandler) Next() *remotecommand.TerminalSize {
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

func (t *MyPtyHandler) Send(m *websocket_pb.TerminalMessage) error {
	t.shellMu.Lock()
	defer t.shellMu.Unlock()

	select {
	case <-t.doneChan:
		close(t.shellCh)
		return errors.New("doneChan closed")
	default:
	}

	select {
	case t.shellCh <- m:
	default:
		return errors.New("shellCh chan full")
	}
	return nil
}

func (t *MyPtyHandler) Resize(size remotecommand.TerminalSize) error {
	t.sizeMu.Lock()
	defer t.sizeMu.Unlock()
	select {
	case <-t.doneChan:
		close(t.sizeChan)
		return errors.New("doneChan closed")
	default:
	}

	select {
	case t.sizeChan <- size:
	default:
		return errors.New("sizeChan chan full")
	}
	return nil
}

func (t *MyPtyHandler) IsClosed() bool {
	t.closeMu.RLock()
	defer t.closeMu.RUnlock()
	return t.closed
}

func (t *MyPtyHandler) CloseDoneChan() bool {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()
	if t.closed {
		return false
	}
	t.closed = true
	mlog.Debug("close prev chan")
	close(t.doneChan)
	return true

}

func (t *MyPtyHandler) Close(reason string) bool {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()
	if t.closed {
		return false
	}
	t.closed = true
	NewMessageSender(t.conn, t.id, WsHandleExecShellMsg).SendProtoMsg(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     t.conn.id,
			Uid:    t.conn.uid,
			Slug:   t.id,
			Type:   WsHandleExecShellMsg,
			Result: ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			SessionId: t.id,
			Op:        OpStdout,
			Data:      []byte(reason),
		},
		Container: &types.Container{
			Namespace: t.Container().Namespace,
			Pod:       t.Container().Pod,
			Container: t.Container().Container,
		},
	})

	t.Send(&websocket_pb.TerminalMessage{
		Op:        OpStdin,
		Data:      ETX,
		SessionId: t.id,
	})
	time.Sleep(200 * time.Millisecond)
	t.Send(&websocket_pb.TerminalMessage{
		Op:        OpStdin,
		Data:      END_OF_TRANSMISSION,
		SessionId: t.id,
	})
	t.Recorder().Close()
	close(t.doneChan)
	return true
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (t *MyPtyHandler) Toast(p string) error {
	NewMessageSender(t.conn, t.id, WsHandleExecShellMsg).SendProtoMsg(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     t.conn.id,
			Uid:    t.conn.uid,
			Slug:   t.id,
			Type:   WsHandleExecShellMsg,
			Result: ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			Op:        OpToast,
			Data:      []byte(p),
			SessionId: t.id,
		},
		Container: &types.Container{
			Container: t.Container().Container,
			Namespace: t.Container().Namespace,
			Pod:       t.Container().Pod,
		},
	})
	return nil
}

// SessionMap stores a map of all MyPtyHandler objects and a sessLock to avoid concurrent conflict
type SessionMap struct {
	conn *WsConn

	sessLock sync.RWMutex
	Sessions map[string]contracts.PtyHandler
}

func (sm *SessionMap) Send(m *websocket_pb.TerminalMessage) {
	sm.sessLock.RLock()
	defer sm.sessLock.RUnlock()
	if h, ok := sm.Sessions[m.SessionId]; ok {
		h.Send(m)
	}
}

// Get return a given terminalSession by sessionId
func (sm *SessionMap) Get(sessionId string) (contracts.PtyHandler, bool) {
	sm.sessLock.RLock()
	defer sm.sessLock.RUnlock()
	h, ok := sm.Sessions[sessionId]
	return h, ok
}

// Set store a MyPtyHandler to SessionMap
func (sm *SessionMap) Set(sessionId string, session contracts.PtyHandler) {
	sm.sessLock.Lock()
	defer sm.sessLock.Unlock()
	sm.Sessions[sessionId] = session
}

func (sm *SessionMap) CloseAll() {
	mlog.Debug("[Websocket]: close all.")
	sm.sessLock.Lock()
	defer sm.sessLock.Unlock()

	for _, s := range sm.Sessions {
		s.Close("websocket conn closed")
	}
	sm.Sessions = map[string]contracts.PtyHandler{}
}

// Close shuts down the SockJS connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (sm *SessionMap) Close(sessionId string, status uint32, reason string) {
	mlog.Debugf("[Websocket]: session %v closed, reason: %s.", sessionId, reason)
	sm.sessLock.Lock()
	defer sm.sessLock.Unlock()
	if s, ok := sm.Sessions[sessionId]; ok {
		delete(sm.Sessions, sessionId)
		go s.Close(reason)
	}
}

// startProcess is called by handleAttach
// Executed cmd in the contracts.Container specified in request and connects it up with the ptyHandler (a session)
func startProcess(k8sClient kubernetes.Interface, cfg *rest.Config, container *contracts.Container, cmd []string, ptyHandler contracts.PtyHandler) error {
	namespace := container.Namespace
	podName := container.Pod
	containerName := container.Container

	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
}

func GenMyPtyHandlerId() string {
	return uuid.New().String()
}

// isValidShell checks if the shell is an allowed one
func isValidShell(validShells []string, shell string) bool {
	for _, validShell := range validShells {
		if validShell == shell {
			return true
		}
	}
	return false
}

var silenceShellExitMessages = []string{
	"command terminated with exit code 126",
	"command terminated with exit code 130",
}

func silence(err error) bool {
	for _, message := range silenceShellExitMessages {
		if strings.Contains(err.Error(), message) {
			return true
		}
	}
	return false
}

// WaitForTerminal is called from apihandler.handleAttach as a goroutine
// Waits for the SockJS connection to be opened by the client the session to be bound in handleMyPtyHandler
func WaitForTerminal(conn *WsConn, k8sClient kubernetes.Interface, cfg *rest.Config, container *contracts.Container, shell, sessionId string) {
	defer func() {
		mlog.Debugf("[Websocket]: WaitForTerminal EXIT: total go: %v", runtime.NumGoroutine())
	}()
	var err error
	validShells := []string{"bash", "sh", "powershell", "cmd"}

	session, _ := conn.terminalSessions.Get(sessionId)
	if isValidShell(validShells, shell) {
		cmd := []string{shell}
		session.SetShell(shell)
		err = startProcess(k8sClient, cfg, container, cmd, session)
	} else {
		// No shell given or it was not valid: try some shells until one succeeds or all fail
		// FIXME: if the first shell fails then the first keyboard event is lost
		for idx, testShell := range validShells {
			mlog.Debug("try" + testShell)
			if session.IsClosed() {
				mlog.Debugf("session 已关闭，不会继续尝试连接其他 shell: '%s'", strings.Join(validShells[idx:], ", "))
				break
			}
			cmd := []string{testShell}
			session.SetShell(testShell)
			if err = startProcess(k8sClient, cfg, container, cmd, session); err == nil {
				break
			}
			// 当出现 bash 回退的时候，需要注意，resize 不会触发，导致，新的 'sh', cols, rows 和用户端不一致，所以需要重置，
			// 通过 sizeStore 记录上次用户的 rows, cols, 当 bash 回退时，在用户输入时应用到新的 sh 中
			session = resetSession(session)
			conn.terminalSessions.Set(sessionId, session)
		}
	}

	if err != nil {
		mlog.Debugf("[Websocket]: %v", err.Error())
		if !silence(err) {
			session.Toast(err.Error())
		}
		conn.terminalSessions.Close(sessionId, 2, err.Error())
		return
	}

	conn.terminalSessions.Close(sessionId, 1, "Process exited")
}

func resetSession(session contracts.PtyHandler) contracts.PtyHandler {
	var cols, rows uint16 = 106, 25
	func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		af := time.NewTimer(3 * time.Second)
		defer ticker.Stop()
		defer af.Stop()
		for session.Cols() == 0 {
			select {
			case <-ticker.C:
				mlog.Debug("sleep....")
				break
			case <-af.C:
				mlog.Warningf("can't get previous cols,rows, use default rows: 25, cols: 106.")
				return
			}
		}
		cols = session.Cols()
		rows = session.Rows()
	}()
	mlog.Debug("done....")

	spty := session.(*MyPtyHandler)
	var newSession contracts.PtyHandler = session
	if spty.CloseDoneChan() {
		newSession = &MyPtyHandler{
			container: spty.container,
			recorder:  spty.recorder,
			id:        spty.id,
			conn:      spty.conn,
			doneChan:  make(chan struct{}),
			sizeChan:  make(chan remotecommand.TerminalSize, 1),
			shellCh:   make(chan *websocket_pb.TerminalMessage, 100),
			sizeStore: sizeStore{
				cols:  cols,
				rows:  rows,
				reset: true,
			},
		}
	}
	return newSession
}

type TerminalResponse struct {
	ID string `json:"id"`
}

func HandleExecShell(input *websocket_pb.WsHandleExecShellInput, conn *WsConn) (string, error) {
	var c = contracts.Container{
		Namespace: input.Container.Namespace,
		Pod:       input.Container.Pod,
		Container: input.Container.Container,
	}

	sessionID := GenMyPtyHandlerId()
	r := &Recorder{
		timer:     realTimer{},
		container: c,
	}
	pty := &MyPtyHandler{
		container: c,
		id:        sessionID,
		conn:      conn,
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
		doneChan:  make(chan struct{}),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 100),
		recorder:  r,
	}
	r.t = pty
	conn.terminalSessions.Set(sessionID, pty)

	go func() {
		defer recovery.HandlePanic("Websocket: WaitForTerminal")
		WaitForTerminal(conn, app.K8sClientSet(), app.K8sClient().RestConfig, &contracts.Container{
			Namespace: input.Container.Namespace,
			Pod:       input.Container.Pod,
			Container: input.Container.Container,
		}, "", sessionID)
	}()

	return sessionID, nil
}
