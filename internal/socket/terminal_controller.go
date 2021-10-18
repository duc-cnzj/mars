package controllers

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

	app "github.com/duc-cnzj/mars/internal/app/helper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

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
	bound    chan error
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}

	shellCh chan TerminalMessage
}

func (t *MyPtyHandler) Read(p []byte) (n int, err error) {
	select {
	case <-t.doneChan:
		return 0, fmt.Errorf("%v doneChan closed!", t.id)
	default:
	}
	ch, err := t.conn.GetShellChannel(t.id)
	if err != nil {
		return 0, err
	}
	msg, ok := <-ch
	if !ok {
		return 0, fmt.Errorf("%v channel closed", t.id)
	}
	mlog.Debugf("[Websocket] %v %v %v 从终端读取消息：%v", t.Namespace, t.Pod, t.Container.Container, msg)
	switch msg.Op {
	case "stdin":
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
	SendMsg(t.conn, t.id, WsHandleExecShellMsg, res)

	return len(p), nil
}

func (t *MyPtyHandler) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
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

	SendMsg(t.conn, t.id, WsHandleExecShellMsg, string(msg))
	return nil
}

// SessionMap stores a map of all MyPtyHandler objects and a lock to avoid concurrent conflict
type SessionMap struct {
	conn     *WsConn
	Sessions map[string]*MyPtyHandler
	Lock     sync.RWMutex
}

// Get return a given terminalSession by sessionId
func (sm *SessionMap) Get(sessionId string) *MyPtyHandler {
	sm.Lock.RLock()
	defer sm.Lock.RUnlock()
	return sm.Sessions[sessionId]
}

// Set store a MyPtyHandler to SessionMap
func (sm *SessionMap) Set(sessionId string, session *MyPtyHandler) {
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	sm.Sessions[sessionId] = session
}

func (sm *SessionMap) CloseAll() {
	mlog.Debug("[Websocket] close all.")
	for _, s := range sm.Sessions {
		sm.Close(s.id, 1, "websocket conn closed")
	}
}

// Close shuts down the SockJS connection and sends the status code and reason to the client
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (sm *SessionMap) Close(sessionId string, status uint32, reason string) {
	mlog.Debugf("[Websocket] session %v closed.", sessionId)
	sm.Lock.Lock()
	defer sm.Lock.Unlock()
	if _, ok := sm.Sessions[sessionId]; !ok {
		return
	}
	msg, _ := json.Marshal(struct {
		Container
		TerminalMessage
	}{
		TerminalMessage: TerminalMessage{
			SessionID: sessionId,
			Op:        "stdout",
			Data:      reason,
		},
		Container: sm.Sessions[sessionId].Container,
	})

	SendMsg(sm.Sessions[sessionId].conn, sm.Sessions[sessionId].id, WsHandleExecShellMsg, string(msg))
	sm.conn.DeleteShellChannel(sessionId)
	close(sm.Sessions[sessionId].doneChan)
	delete(sm.Sessions, sessionId)
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

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
	if err != nil {
		mlog.Error(err)

		return err
	}

	return nil
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
		mlog.Debugf("[Websocket] WaitForTerminal EXIT: total go: %v", runtime.NumGoroutine())
	}()
	var err error
	validShells := []string{"bash", "sh", "powershell", "cmd"}

	if isValidShell(validShells, shell) {
		cmd := []string{shell}
		err = startProcess(k8sClient, cfg, container, cmd, conn.terminalSessions.Get(sessionId))
	} else {
		// No shell given or it was not valid: try some shells until one succeeds or all fail
		// FIXME: if the first shell fails then the first keyboard event is lost
		for _, testShell := range validShells {
			cmd := []string{testShell}
			if err = startProcess(k8sClient, cfg, container, cmd, conn.terminalSessions.Get(sessionId)); err == nil {
				break
			}
		}
	}

	if err != nil {
		mlog.Error(err)
		s := conn.terminalSessions.Get(sessionId)
		if strings.Contains(err.Error(), "unable to upgrade connection") {
			if pod, e := app.K8sClientSet().CoreV1().Pods(container.Namespace).Get(context.Background(), container.Pod, metav1.GetOptions{}); e == nil && pod.Status.Phase == metav1.StatusFailure && pod.Status.Reason == "Evicted" {
				app.K8sClientSet().CoreV1().Pods(container.Namespace).Delete(context.TODO(), container.Pod, metav1.DeleteOptions{})
				s.Toast(fmt.Sprintf("delete po %s when evicted in namespace %s!", container.Pod, container.Namespace))
			}
		} else {
			s.Toast(err.Error())
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
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
		doneChan: make(chan struct{}),
		shellCh:  make(chan TerminalMessage, 100),
	})

	go WaitForTerminal(conn, app.K8sClientSet(), app.K8sClient().RestConfig, &Container{
		Namespace: input.Namespace,
		Pod:       input.Pod,
		Container: input.Container,
	}, "", sessionID)

	return sessionID, nil
}
