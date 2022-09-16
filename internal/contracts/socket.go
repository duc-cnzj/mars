package contracts

//go:generate mockgen -destination ../mock/mock_pty_handler.go -package mock github.com/duc-cnzj/mars/internal/contracts PtyHandler
//go:generate mockgen -destination ../mock/mock_socket_conn.go -package mock github.com/duc-cnzj/mars/internal/contracts WebsocketConn
//go:generate mockgen -destination ../mock/mock_socket.go -package mock github.com/duc-cnzj/mars/internal/contracts CancelSignaler
//go:generate mockgen -destination ../mock/mock_socket_deploy_msger.go -package mock github.com/duc-cnzj/mars/internal/contracts DeployMsger
//go:generate mockgen -destination ../mock/mock_socket_job.go -package mock github.com/duc-cnzj/mars/internal/contracts Job
//go:generate mockgen -destination ../mock/mock_socket_session_mapper.go -package mock github.com/duc-cnzj/mars/internal/contracts SessionMapper
//go:generate mockgen -destination ../mock/mock_release_installer.go -package mock github.com/duc-cnzj/mars/internal/contracts ReleaseInstaller
//go:generate mockgen -destination ../mock/mock_recorder.go -package mock github.com/duc-cnzj/mars/internal/contracts RecorderInterface

import (
	"context"
	"io"
	"time"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars-client/v4/websocket"
)

type MessageItem struct {
	Msg  string
	Type MessageType

	Containers []*types.Container
}

type MessageType uint8

const (
	_ MessageType = iota
	MessageSuccess
	MessageError
	MessageText
)

type Container struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}

type RecorderInterface interface {
	Resize(cols, rows uint16)
	Write(data string) (err error)
	Close() error
	SetShell(string)
	GetShell() string
}

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue

	Container() Container
	SetShell(string)
	Toast(string) error

	Send(*websocket.TerminalMessage) error
	Resize(remotecommand.TerminalSize) error

	Recorder() RecorderInterface

	ResetTerminalRowCol(bool)
	Rows() uint16
	Cols() uint16

	Close(string) bool
	IsClosed() bool
}

type WebsocketConn interface {
	Close() error
	SetWriteDeadline(t time.Time) error
	WriteMessage(messageType int, data []byte) error
	SetReadLimit(limit int64)
	SetReadDeadline(t time.Time) error
	SetPongHandler(h func(appData string) error)
	ReadMessage() (messageType int, p []byte, err error)
	NextWriter(messageType int) (io.WriteCloser, error)
}

type Percentable interface {
	Current() int64
	Add()
	To(percent int64)
}

type ReleaseInstaller interface {
	Chart() *chart.Chart
	Run(stopCtx context.Context, messageCh SafeWriteMessageChInterface, percenter Percentable, isNew bool) (*release.Release, error)

	Logs() []string
}

type DeployMsger interface {
	Msger
	ProcessPercentMsger

	Stop(error)
	SendDeployedResult(t websocket.ResultType, msg string, p *types.ProjectModel)
}

type Msger interface {
	SendEndError(error)
	SendError(error)
	SendMsg(string)
	SendProtoMsg(WebsocketMessage)
	SendMsgWithContainerLog(msg string, containers []*types.Container)
}

type ProcessPercentMsger interface {
	SendProcessPercent(int64)
}

type SafeWriteMessageChInterface interface {
	Closed()
	Chan() <-chan MessageItem
	Send(m MessageItem)
}

type CancelSignaler interface {
	Remove(id string)
	Has(id string) bool
	Cancel(id string)
	Add(id string, fn func(error)) error
	CancelAll()
}

type Job interface {
	Done() <-chan struct{}
	Finish()
	Prune()

	Stop(error)
	IsStopped() bool
	GetStoppedErrorIfHas() error

	Run() error
	Logs() []string
	IsDryRun() bool
	Manifests() []string

	User() UserInfo
	IsNew() bool
	ProjectModel() *types.ProjectModel
	Commit() CommitInterface

	ID() string
	Validate() error
	LoadConfigs() error
	HandleMessage()
	AddDestroyFunc(fn func())
	CallDestroyFuncs()

	ReleaseInstaller() ReleaseInstaller
	Messager() DeployMsger
	PubSub() PubSub
	Percenter() Percentable
}

type SessionMapper interface {
	Send(message *websocket.TerminalMessage)
	Get(sessionId string) (PtyHandler, bool)
	Set(sessionId string, session PtyHandler)
	CloseAll()
	Close(sessionId string, status uint32, reason string)
}
