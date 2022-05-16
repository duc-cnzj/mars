package socket

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	ETX                 = "\u0003"
	END_OF_TRANSMISSION = "\u0004"
)

const (
	OpStdout = "stdout"
	OpStdin  = "stdin"
	OpResize = "resize"
	OpToast  = "toast"
)

type Container struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}

type MyPtyHandler struct {
	Container
	recorder RecorderInterface
	id       string
	conn     *WsConn
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}

	shellCh chan *websocket_pb.TerminalMessage

	closeLock sync.Mutex
	isClosed  bool
}

func (t *MyPtyHandler) SetShell(shell string) {
	t.recorder.SetShell(shell)
}

func (t *MyPtyHandler) TerminalMessageChan() chan *websocket_pb.TerminalMessage {
	return t.shellCh
}

func (t *MyPtyHandler) Recorder() RecorderInterface {
	return t.recorder
}

func (t *MyPtyHandler) Close(reason string) {
	t.closeLock.Lock()
	if t.isClosed {
		t.closeLock.Unlock()
		return
	}
	t.isClosed = true
	t.closeLock.Unlock()

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
			Data:      reason,
		},
		Container: &types.Container{
			Namespace: t.Container.Namespace,
			Pod:       t.Container.Pod,
			Container: t.Container.Container,
		},
	})

	t.shellCh <- &websocket_pb.TerminalMessage{
		Op:        OpStdin,
		Data:      ETX,
		SessionId: t.id,
	}
	time.Sleep(200 * time.Millisecond)
	t.shellCh <- &websocket_pb.TerminalMessage{
		Op:        OpStdin,
		Data:      END_OF_TRANSMISSION,
		SessionId: t.id,
	}
	t.recorder.Close()
	close(t.shellCh)
	close(t.sizeChan)
	close(t.doneChan)
}

func (t *MyPtyHandler) Read(p []byte) (n int, err error) {
	select {
	case <-t.doneChan:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("[Websocket]: %v doneChan closed", t.id)
	default:
	}
	msg, ok := <-t.shellCh
	if !ok {
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("[Websocket]: %v channel closed", t.id)
	}
	switch msg.Op {
	case OpStdin:
		return copy(p, msg.Data), nil
	case OpResize:
		mlog.Debugf("[Websocket]: resize cols: %v  rows: %v", msg.Cols, msg.Rows)
		t.sizeChan <- remotecommand.TerminalSize{Width: uint16(msg.Cols), Height: uint16(msg.Rows)}
		return 0, nil
	default:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

func (t *MyPtyHandler) Write(p []byte) (n int, err error) {
	select {
	case <-t.doneChan:
		return 0, fmt.Errorf("[Websocket]: %v doneChan closed", t.id)
	default:
	}
	send := true
	t.closeLock.Lock()
	if t.isClosed {
		send = false
	}
	t.closeLock.Unlock()
	if send {
		t.recorder.Write(string(p))
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
				Data:      string(p),
				SessionId: t.id,
			},
			Container: &types.Container{
				Namespace: t.Container.Namespace,
				Pod:       t.Container.Pod,
				Container: t.Container.Container,
			},
		})
	}

	return len(p), nil
}

func (t *MyPtyHandler) Next() *remotecommand.TerminalSize {
	select {
	case size, ok := <-t.sizeChan:
		if !ok {
			return nil
		}
		return &size
	case <-t.doneChan:
		return nil
	}
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
			Data:      p,
			SessionId: t.id,
		},
		Container: &types.Container{
			Container: t.Container.Container,
			Namespace: t.Container.Namespace,
			Pod:       t.Container.Pod,
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
		select {
		case h.TerminalMessageChan() <- m:
		default:
			mlog.Warningf("[Websocket]: sessionId %v 的 shellCh 满了: %d", m.SessionId, len(h.TerminalMessageChan()))
		}
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
// Executed cmd in the Container specified in request and connects it up with the ptyHandler (a session)
func startProcess(k8sClient kubernetes.Interface, cfg *rest.Config, container *Container, cmd []string, ptyHandler contracts.PtyHandler) error {
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
func WaitForTerminal(conn *WsConn, k8sClient kubernetes.Interface, cfg *rest.Config, container *Container, shell, sessionId string) {
	defer func() {
		utils.HandlePanic("Websocket: WaitForTerminal")
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
		for _, testShell := range validShells {
			cmd := []string{testShell}
			session.SetShell(testShell)
			if err = startProcess(k8sClient, cfg, container, cmd, session); err == nil {
				break
			}
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

type TerminalResponse struct {
	ID string `json:"id"`
}

func HandleExecShell(input *websocket_pb.WsHandleExecShellInput, conn *WsConn) (string, error) {
	var c = Container{
		Namespace: input.Container.Namespace,
		Pod:       input.Container.Pod,
		Container: input.Container.Container,
	}

	sessionID := GenMyPtyHandlerId()
	r := &Recorder{
		container: c,
	}
	pty := &MyPtyHandler{
		Container: c,
		id:        sessionID,
		conn:      conn,
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
		doneChan:  make(chan struct{}, 1),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 100),
		recorder:  r,
	}
	r.t = pty
	conn.terminalSessions.Set(sessionID, pty)

	go WaitForTerminal(conn, app.K8sClientSet(), app.K8sClient().RestConfig, &Container{
		Namespace: input.Container.Namespace,
		Pod:       input.Container.Pod,
		Container: input.Container.Container,
	}, "", sessionID)

	return sessionID, nil
}
