package socket

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"

	app "github.com/duc-cnzj/mars/internal/app/helper"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const ETX = "\u0003"
const END_OF_TRANSMISSION = "\u0004"

type Container struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}

type MyPtyHandler struct {
	Container
	id       string
	conn     *WsConn
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}

	shellCh chan TerminalMessage

	closeLock sync.Mutex
	isClosed  bool

	cacheLock sync.RWMutex
	cache     []byte
}

func (t *MyPtyHandler) Close(reason string) {
	t.closeLock.Lock()
	if t.isClosed {
		t.closeLock.Unlock()
		return
	}
	t.isClosed = true
	t.closeLock.Unlock()
	msg, _ := json.Marshal(struct {
		Container
		TerminalMessage
	}{
		TerminalMessage: TerminalMessage{
			SessionID: t.id,
			Op:        "stdout",
			Data:      reason,
		},
		Container: t.Container,
	})

	NewMessageSender(t.conn, t.id, WsHandleExecShellMsg).SendMsg(string(msg))

	t.shellCh <- TerminalMessage{
		Op:        "stdin",
		Data:      ETX,
		SessionID: t.id,
	}
	time.Sleep(200 * time.Millisecond)
	t.shellCh <- TerminalMessage{
		Op:        "stdin",
		Data:      END_OF_TRANSMISSION,
		SessionID: t.id,
	}
	close(t.shellCh)
	close(t.sizeChan)
	close(t.doneChan)
	t.cacheLock.RLock()
	if len(t.cache) > 0 {
		mlog.Infof("[Websocket]: user %v, send: '%v', namespace: %v, pod: %v.", t.conn.GetUser().Name, string(t.cache), t.Namespace, t.Pod)
	}
	t.cacheLock.RUnlock()
}

func (t *MyPtyHandler) Read(p []byte) (n int, err error) {
	select {
	case <-t.doneChan:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("[Websocket]: %v doneChan closed", t.id)
	default:
	}
	msg, ok := <-t.shellCh
	if !ok {
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("%v channel closed", t.id)
	}
	switch msg.Op {
	case "stdin":
		t.cacheLock.Lock()

		switch {
		case msg.Data == "\r":
			mlog.Infof("[Websocket]: user %v, send: '%v', namespace: %v, pod: %v.", t.conn.GetUser().Name, string(t.cache), t.Namespace, t.Pod)
			t.cache = make([]byte, 0, 20)
		case strings.ContainsRune(msg.Data, rune(byte(''))):
		default:
			t.cache = append(t.cache, []byte(msg.Data)...)
		}

		t.cacheLock.Unlock()
		return copy(p, msg.Data), nil
	case "resize":
		t.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

func (t *MyPtyHandler) Write(p []byte) (n int, err error) {
	select {
	case <-t.doneChan:
		return 0, fmt.Errorf("%v doneChan closed!", t.id)
	default:
	}
	msg, err := json.Marshal(struct {
		Container
		TerminalMessage
	}{
		TerminalMessage: TerminalMessage{
			SessionID: t.id,
			Op:        "stdout",
			Data:      string(p),
		},
		Container: t.Container,
	})
	if err != nil {
		mlog.Error(err)
		return 0, err
	}
	res := string(msg)
	if !strings.HasSuffix(res, "\r\n") {
		res = res + "\r\n"
	}

	send := true
	t.closeLock.Lock()
	if t.isClosed {
		send = false
	}
	t.closeLock.Unlock()
	if send {
		NewMessageSender(t.conn, t.id, WsHandleExecShellMsg).SendMsg(res)
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

// TerminalMessage is the messaging protocol between ShellController and MyPtyHandler.
//
// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// bind    fe->be     SessionID      Id sent back from TerminalResponse
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type TerminalMessage struct {
	Op        string `json:"op"`
	Data      string `json:"data"`
	SessionID string `json:"session_id"`
	Rows      uint16 `json:"rows"`
	Cols      uint16 `json:"cols"`
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (t *MyPtyHandler) Toast(p string) error {
	msg, err := json.Marshal(TerminalMessage{
		SessionID: t.id,
		Op:        "toast",
		Data:      p,
	})
	if err != nil {
		mlog.Error(err)

		return err
	}

	NewMessageSender(t.conn, t.id, WsHandleExecShellMsg).SendMsg(string(msg))
	return nil
}

type SessionMapper interface {
	Send(message TerminalMessage)
	Get(sessionId string) (*MyPtyHandler, bool)
	Set(sessionId string, session *MyPtyHandler)
	CloseAll()
	Close(sessionId string, status uint32, reason string)
}

// SessionMap stores a map of all MyPtyHandler objects and a lock to avoid concurrent conflict
type SessionMap struct {
	conn     *WsConn
	Sessions map[string]*MyPtyHandler
	Lock     sync.RWMutex
}

func (sm *SessionMap) Send(m TerminalMessage) {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	if h, ok := sm.Sessions[m.SessionID]; ok {
		select {
		case h.shellCh <- m:
		default:
			mlog.Warningf("[Websocket]: sessionId %v 的 shellCh 满了: %d", m.SessionID, len(h.shellCh))
		}
	}
}

// Get return a given terminalSession by sessionId
func (sm *SessionMap) Get(sessionId string) (*MyPtyHandler, bool) {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	h, ok := sm.Sessions[sessionId]
	return h, ok
}

// Set store a MyPtyHandler to SessionMap
func (sm *SessionMap) Set(sessionId string, session *MyPtyHandler) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	sm.Sessions[sessionId] = session
}

func (sm *SessionMap) CloseAll() {
	mlog.Debug("[Websocket] close all.")
	sm.Lock.Lock()
	defer sm.Lock.Unlock()

	for _, s := range sm.Sessions {
		s.Close("websocket conn closed")
	}
	sm.Sessions = map[string]*MyPtyHandler{}
}

// Close shuts down the SockJS connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (sm *SessionMap) Close(sessionId string, status uint32, reason string) {
	mlog.Debugf("[Websocket] session %v closed, reason: %s.", sessionId, reason)
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	if s, ok := sm.Sessions[sessionId]; ok {
		delete(sm.Sessions, sessionId)
		go s.Close(reason)
	}
}

// startProcess is called by handleAttach
// Executed cmd in the Container specified in request and connects it up with the ptyHandler (a session)
func startProcess(k8sClient kubernetes.Interface, cfg *rest.Config, container *Container, cmd []string, ptyHandler PtyHandler) error {
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

// GenMyPtyHandlerId generates a random session ID string. The format is not really interesting.
// This ID is used to identify the session when the client opens the SockJS connection.
// Not the same as the SockJS session id! We can't use that as that is generated
// on the client side and we don't have it yet at this point.
func GenMyPtyHandlerId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		mlog.Error(err)

		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
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

// WaitForTerminal is called from apihandler.handleAttach as a goroutine
// Waits for the SockJS connection to be opened by the client the session to be bound in handleMyPtyHandler
func WaitForTerminal(conn *WsConn, k8sClient kubernetes.Interface, cfg *rest.Config, container *Container, shell, sessionId string) {
	defer func() {
		utils.HandlePanic("Websocket: WaitForTerminal")
		mlog.Debugf("[Websocket] WaitForTerminal EXIT: total go: %v", runtime.NumGoroutine())
	}()
	var err error
	validShells := []string{"bash", "sh", "powershell", "cmd"}

	session, _ := conn.terminalSessions.Get(sessionId)
	if isValidShell(validShells, shell) {
		cmd := []string{shell}
		err = startProcess(k8sClient, cfg, container, cmd, session)
	} else {
		// No shell given or it was not valid: try some shells until one succeeds or all fail
		// FIXME: if the first shell fails then the first keyboard event is lost
		for _, testShell := range validShells {
			cmd := []string{testShell}
			if err = startProcess(k8sClient, cfg, container, cmd, session); err == nil {
				break
			}
		}
	}

	if err != nil {
		mlog.Errorf("[Websocket]: %v", err.Error())
		if strings.Contains(err.Error(), "unable to upgrade connection") {
			if pod, e := app.K8sClientSet().CoreV1().Pods(container.Namespace).Get(context.Background(), container.Pod, metav1.GetOptions{}); e == nil && pod.Status.Phase == metav1.StatusFailure && pod.Status.Reason == "Evicted" {
				app.K8sClientSet().CoreV1().Pods(container.Namespace).Delete(context.TODO(), container.Pod, metav1.DeleteOptions{})
				session.Toast(fmt.Sprintf("delete po %session when evicted in namespace %session!", container.Pod, container.Namespace))
			}
		} else {
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

// HandleExecShell Handles execute shell API call
type WsHandleExecShellInput struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}

func HandleExecShell(input WsHandleExecShellInput, conn *WsConn) (string, error) {
	if running, reason := utils.IsPodRunning(input.Namespace, input.Pod); !running {
		return "", errors.New(reason)
	}

	sessionID, err := GenMyPtyHandlerId()
	if err != nil {
		return "", err
	}

	conn.terminalSessions.Set(sessionID, &MyPtyHandler{
		Container: Container{
			Namespace: input.Namespace,
			Pod:       input.Pod,
			Container: input.Container,
		},
		id:       sessionID,
		conn:     conn,
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}, 1),
		shellCh:  make(chan TerminalMessage, 100),
		cache:    make([]byte, 0, 20),
	})

	go WaitForTerminal(conn, app.K8sClientSet(), app.K8sClient().RestConfig, &Container{
		Namespace: input.Namespace,
		Pod:       input.Pod,
		Container: input.Container,
	}, "", sessionID)

	return sessionID, nil
}
