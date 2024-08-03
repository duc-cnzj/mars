package socket

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/utils/closeable"
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

type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue

	Container() *repo.Container
	SetShell(string)
	Toast(string) error

	Send(ctx context.Context, message *websocket_pb.TerminalMessage) error
	Resize(remotecommand.TerminalSize) error

	Recorder() repo.Recorder

	ResetTerminalRowCol(bool)
	Rows() uint16
	Cols() uint16

	Close(context.Context, string) bool
	IsClosed() bool
}

type myPtyHandler struct {
	logger    mlog.Logger
	sessionID string
	container *repo.Container
	recorder  repo.Recorder
	conn      *WsConn

	doneChan  chan struct{}
	sizeStore sizeStore

	shellMu sync.RWMutex
	shellCh chan *websocket_pb.TerminalMessage

	sizeMu   sync.RWMutex
	sizeChan chan remotecommand.TerminalSize

	closeable.Closeable
}

func (t *myPtyHandler) SetShell(shell string) {
	t.recorder.SetShell(shell)
}

func (t *myPtyHandler) Container() *repo.Container {
	return t.container
}

func (t *myPtyHandler) Rows() uint16 {
	return t.sizeStore.Rows()
}

func (t *myPtyHandler) Cols() uint16 {
	return t.sizeStore.Cols()
}

func (t *myPtyHandler) Read(p []byte) (n int, err error) {
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
		t.logger.Debugf("[Websocket]: resize cols: %v  rows: %v", msg.Cols, msg.Rows)
		t.Resize(remotecommand.TerminalSize{Width: uint16(msg.Cols), Height: uint16(msg.Rows)})
		return 0, nil
	default:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

func (t *myPtyHandler) Write(p []byte) (n int, err error) {
	select {
	case <-t.doneChan:
		return len(p), fmt.Errorf("[Websocket]: %v doneChan closed", t.sessionID)
	default:
	}
	if t.IsClosed() {
		return len(p), fmt.Errorf("[Websocket]: %v ws already closed", t.sessionID)
	}
	t.recorder.Write(string(p))
	if t.sizeStore.TerminalRowColNeedReset() && t.sizeStore.Cols() != 0 {
		t.logger.Debugf("reset shell size rows: %d, cols: %d", t.sizeStore.Rows(), t.sizeStore.Cols())
		t.sizeStore.ResetTerminalRowCol(false)
		t.Resize(remotecommand.TerminalSize{Width: t.sizeStore.Cols(), Height: t.sizeStore.Rows()})
	}
	NewMessageSender(t.conn, t.sessionID, WsHandleExecShellMsg).SendProtoMsg(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     t.conn.id,
			Uid:    t.conn.uid,
			Slug:   t.sessionID,
			Type:   WsHandleExecShellMsg,
			Result: ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			Op:        OpStdout,
			Data:      p,
			SessionId: t.sessionID,
		},
		Container: &types.Container{
			Namespace: t.Container().Namespace,
			Pod:       t.Container().Pod,
			Container: t.Container().Container,
		},
	})

	return len(p), nil
}

func (t *myPtyHandler) ResetTerminalRowCol(reset bool) {
	t.sizeStore.ResetTerminalRowCol(reset)
}

func (t *myPtyHandler) Recorder() repo.Recorder {
	return t.recorder
}

func (t *myPtyHandler) Next() *remotecommand.TerminalSize {
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

func (t *myPtyHandler) Send(ctx context.Context, m *websocket_pb.TerminalMessage) error {
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

func (t *myPtyHandler) Resize(size remotecommand.TerminalSize) error {
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

func (t *myPtyHandler) IsClosed() bool {
	return t.Closeable.IsClosed()
}

func (t *myPtyHandler) CloseDoneChan() bool {
	if !t.Closeable.Close() {
		return false
	}
	close(t.doneChan)
	return true
}

func (t *myPtyHandler) Close(ctx context.Context, reason string) bool {
	if !t.Closeable.IsClosed() {
		return false
	}
	NewMessageSender(t.conn, t.sessionID, WsHandleCloseShell).SendProtoMsg(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     t.conn.id,
			Uid:    t.conn.uid,
			Slug:   t.sessionID,
			Type:   WsHandleCloseShell,
			Result: ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			SessionId: t.sessionID,
			Op:        OpStdout,
			Data:      []byte(reason),
		},
		Container: &types.Container{
			Namespace: t.Container().Namespace,
			Pod:       t.Container().Pod,
			Container: t.Container().Container,
		},
	})

	t.Send(ctx, &websocket_pb.TerminalMessage{
		Op:        OpStdin,
		Data:      ETX,
		SessionId: t.sessionID,
	})
	time.Sleep(200 * time.Millisecond)
	t.Send(ctx, &websocket_pb.TerminalMessage{
		Op:        OpStdin,
		Data:      END_OF_TRANSMISSION,
		SessionId: t.sessionID,
	})
	t.Recorder().Close()
	close(t.doneChan)
	return true
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (t *myPtyHandler) Toast(p string) error {
	NewMessageSender(t.conn, t.sessionID, WsHandleExecShellMsg).SendProtoMsg(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     t.conn.id,
			Uid:    t.conn.uid,
			Slug:   t.sessionID,
			Type:   WsHandleExecShellMsg,
			Result: ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			Op:        OpToast,
			Data:      []byte(p),
			SessionId: t.sessionID,
		},
		Container: &types.Container{
			Container: t.Container().Container,
			Namespace: t.Container().Namespace,
			Pod:       t.Container().Pod,
		},
	})
	return nil
}

type SessionMapper interface {
	Get(sessionId string) (PtyHandler, bool)
	Set(sessionId string, session PtyHandler)
	CloseAll(ctx context.Context)
	Close(ctx context.Context, sessionId string, status uint32, reason string)
}

// sessionMap stores a map of all myPtyHandler objects and a sessLock to avoid concurrent conflict
type sessionMap struct {
	wg     sync.WaitGroup
	logger mlog.Logger

	sessLock sync.RWMutex
	Sessions map[string]PtyHandler
}

func NewSessionMap(logger mlog.Logger) SessionMapper {
	return &sessionMap{Sessions: map[string]PtyHandler{}, logger: logger}
}

// Get return a given terminalSession by sessionId
func (sm *sessionMap) Get(sessionId string) (PtyHandler, bool) {
	sm.sessLock.RLock()
	defer sm.sessLock.RUnlock()
	h, ok := sm.Sessions[sessionId]
	return h, ok
}

// Set store a myPtyHandler to sessionMap
func (sm *sessionMap) Set(sessionId string, session PtyHandler) {
	sm.sessLock.Lock()
	defer sm.sessLock.Unlock()
	sm.Sessions[sessionId] = session
}

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

// Close shuts down the SockJS connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (sm *sessionMap) Close(ctx context.Context, sessionId string, status uint32, reason string) {
	sm.logger.Debugf("[Websocket]: session %v closed, reason: %s, status: %v.", sessionId, reason, status)
	sm.sessLock.Lock()
	defer sm.sessLock.Unlock()
	if s, ok := sm.Sessions[sessionId]; ok {
		delete(sm.Sessions, sessionId)
		sm.wg.Add(1)
		go func() {
			defer sm.wg.Done()
			s.Close(ctx, reason)
		}()
	}
}

// startProcess is called by handleAttach
// Executed cmd in the contracts.Container specified in request and connects it up with the ptyHandler (a session)
func (wc *WebsocketManager) startProcess(ctx context.Context, container *repo.Container, cmd []string, ptyHandler PtyHandler) error {
	return wc.k8sRepo.Execute(ctx, container, &repo.ExecuteInput{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TTY:               true,
		Cmd:               cmd,
		TerminalSizeQueue: ptyHandler,
	})
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
func (wc *WebsocketManager) WaitForTerminal(ctx context.Context, sm SessionMapper, container *repo.Container, shell, sessionId string) {
	defer func() {
		wc.logger.Debugf("[Websocket]: WaitForTerminal EXIT: total go: %v", runtime.NumGoroutine())
	}()
	var err error
	validShells := []string{"bash", "sh", "powershell", "cmd"}
	session, _ := sm.Get(sessionId)
	if isValidShell(validShells, shell) {
		cmd := []string{shell}
		session.SetShell(shell)
		err = wc.startProcess(ctx, container, cmd, session)
	} else {
		// No shell given or it was not valid: try some shells until one succeeds or all fail
		// FIXME: if the first shell fails then the first keyboard event is lost
		for idx, testShell := range validShells {
			wc.logger.Debug("try" + testShell)
			if session.IsClosed() {
				wc.logger.Debugf("session 已关闭，不会继续尝试连接其他 shell: '%s'", strings.Join(validShells[idx:], ", "))
				break
			}
			cmd := []string{testShell}
			session.SetShell(testShell)
			if err = wc.startProcess(ctx, container, cmd, session); err == nil {
				break
			}
			// 当出现 bash 回退的时候，需要注意，resize 不会触发，导致，新的 'sh', cols, rows 和用户端不一致，所以需要重置，
			// 通过 sizeStore 记录上次用户的 rows, cols, 当 bash 回退时，在用户输入时应用到新的 sh 中
			session = wc.resetSession(session)
			sm.Set(sessionId, session)
		}
	}

	if err != nil {
		wc.logger.Debugf("[Websocket]: %v", err.Error())
		if !silence(err) {
			session.Toast(err.Error())
		}
		sm.Close(ctx, sessionId, 2, err.Error())
		return
	}

	sm.Close(ctx, sessionId, 1, "Process exited")
}

func (wc *WebsocketManager) resetSession(session PtyHandler) PtyHandler {
	var cols, rows uint16 = 106, 25
	func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		af := time.NewTimer(3 * time.Second)
		defer ticker.Stop()
		defer af.Stop()
		for session.Cols() == 0 {
			select {
			case <-ticker.C:
				wc.logger.Debug("sleep....")
				break
			case <-af.C:
				wc.logger.Warningf("can't get previous cols,rows, use default rows: 25, cols: 106.")
				return
			}
		}
		cols = session.Cols()
		rows = session.Rows()
	}()
	wc.logger.Debug("done....")

	spty := session.(*myPtyHandler)
	var newSession PtyHandler = session
	if spty.CloseDoneChan() {
		newSession = &myPtyHandler{
			container: spty.container,
			recorder:  spty.recorder,
			sessionID: spty.sessionID,
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
	ID string `json:"sessionID"`
}

func checkSessionID(container *types.Container, id string) bool {
	prefix := fmt.Sprintf("%s-%s-%s:", container.Namespace, container.Pod, container.Container)
	return strings.HasPrefix(id, prefix)
}

func (wc *WebsocketManager) StartShell(ctx context.Context, input *websocket_pb.WsHandleExecShellInput, conn *WsConn) (string, error) {
	var (
		container = &repo.Container{
			Namespace: input.Container.Namespace,
			Pod:       input.Container.Pod,
			Container: input.Container.Container,
		}
		sessionID = input.SessionId
	)

	if !checkSessionID(input.Container, sessionID) {
		return "", fmt.Errorf("invalid session sessionID, must format: '<namespace>-<pod>-<container>:<randomID>', input: '%s'", sessionID)
	}

	pty := &myPtyHandler{
		container: container,
		sessionID: sessionID,
		conn:      conn,
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
		doneChan:  make(chan struct{}),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 100),
		recorder:  wc.fileRepo.NewRecorder(types.EventActionType_Shell, conn.GetUser(), container),
	}
	conn.sm.Set(sessionID, pty)

	go func() {
		defer wc.logger.HandlePanic("Websocket: WaitForTerminal")
		wc.WaitForTerminal(ctx, conn.sm, container, "", sessionID)
	}()

	return sessionID, nil
}
