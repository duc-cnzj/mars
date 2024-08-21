package socket

import (
	"errors"
	"testing"

	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestIsValidShell(t *testing.T) {
	validShells := []string{"bash", "sh", "powershell", "cmd"}

	assert.True(t, isValidShell(validShells, "bash"))
	assert.True(t, isValidShell(validShells, "sh"))
	assert.True(t, isValidShell(validShells, "powershell"))
	assert.True(t, isValidShell(validShells, "cmd"))
	assert.False(t, isValidShell(validShells, "invalidShell"))
}

func TestSilence(t *testing.T) {
	assert.True(t, silence(errors.New("command terminated with exit code 126")))
	assert.True(t, silence(errors.New("command terminated with exit code 130")))
	assert.False(t, silence(errors.New("command terminated with exit code 131")))
}

func TestCheckSessionID(t *testing.T) {
	container := &websocket_pb.Container{
		Namespace: "namespace",
		Pod:       "pod",
		Container: "container",
	}

	assert.True(t, checkSessionID(container, "namespace-pod-container:randomID"))
	assert.False(t, checkSessionID(container, "invalidSessionID"))
}

func TestSizeStore(t *testing.T) {
	s := &sizeStore{}

	s.Set(10, 20)
	assert.Equal(t, uint16(10), s.Cols())
	assert.Equal(t, uint16(20), s.Rows())
	assert.True(t, s.Changed(11, 20))
	assert.False(t, s.Changed(10, 20))

	s.ResetTerminalRowCol(true)
	assert.True(t, s.TerminalRowColNeedReset())
	s.ResetTerminalRowCol(false)
	assert.False(t, s.TerminalRowColNeedReset())
}

func TestContainer(t *testing.T) {
	container := &repo.Container{
		Namespace: "namespace",
		Pod:       "pod",
		Container: "container",
	}
	pty := &myPtyHandler{
		container: container,
	}
	assert.Equal(t, container, pty.Container())
}

func TestMyPtyHandler_SetShell(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := repo.NewMockRecorder(m)
	pty := &myPtyHandler{
		recorder: recorder,
	}
	recorder.EXPECT().SetShell("bash")
	pty.SetShell("bash")
}

func TestMyPtyHandler_IsClosed(t *testing.T) {
	pty := &myPtyHandler{}
	assert.False(t, pty.IsClosed())
	pty.Closeable.Close()
	assert.True(t, pty.IsClosed())
}

func TestMyPtyHandler_ResetTerminalRowCol(t *testing.T) {
	pty := &myPtyHandler{
		sizeStore: &sizeStore{},
	}
	pty.ResetTerminalRowCol(true)
	assert.True(t, pty.sizeStore.TerminalRowColNeedReset())
	pty.ResetTerminalRowCol(false)
	assert.False(t, pty.sizeStore.TerminalRowColNeedReset())
}

func TestMyPtyHandler_Read(t *testing.T) {
	pty := &myPtyHandler{
		doneChan: make(chan struct{}),
		shellCh:  make(chan *websocket_pb.TerminalMessage, 1),
	}
	pty.shellCh <- &websocket_pb.TerminalMessage{Op: OpStdin, Data: []byte("data")}
	p := make([]byte, 4)
	n, err := pty.Read(p)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "data", string(p))
}

func TestMyPtyHandler_Write(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := repo.NewMockRecorder(m)
	conn := NewMockConn(m)
	pty := &myPtyHandler{
		sessionID: "sid",
		conn:      conn,
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{cols: 10, rows: 10},
		recorder:  recorder,
		container: &repo.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
	}
	sub := application.NewMockPubSub(m)
	sub.EXPECT().ToSelf(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     "id",
			Uid:    "uid",
			Slug:   "sid",
			Type:   WsHandleExecShellMsg,
			Result: ResultSuccess,
		},
		TerminalMessage: &websocket_pb.TerminalMessage{
			Op:        OpStdout,
			Data:      []byte("data"),
			SessionId: "sid",
		},
		Container: &websocket_pb.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
	})
	conn.EXPECT().PubSub().Return(sub)
	conn.EXPECT().ID().Return("id")
	conn.EXPECT().UID().Return("uid")
	recorder.EXPECT().Write([]byte("data")).Return(4, nil)
	n, err := pty.Write([]byte("data"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
}

func TestMyPtyHandler_Toast(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	conn := NewMockConn(m)
	pty := &myPtyHandler{
		conn:      conn,
		container: &repo.Container{},
	}
	conn.EXPECT().ID().Return("id")
	conn.EXPECT().UID().Return("uid")
	sub := application.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub)
	sub.EXPECT().ToSelf(gomock.Any())

	assert.NoError(t, pty.Toast("message"))
}
