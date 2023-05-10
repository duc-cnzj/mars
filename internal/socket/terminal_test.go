package socket

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/duc-cnzj/mars-client/v4/types"
	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mock"
	"github.com/duc-cnzj/mars/v4/internal/plugins"
	"github.com/duc-cnzj/mars/v4/internal/testutil"
)

func TestHandleExecShell(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{
		Client:     nil,
		RestConfig: nil,
	}).AnyTimes()

	sess := mock.NewMockSessionMapper(m)
	c := &WsConn{
		newExecutorFunc: func() contracts.RemoteExecutor {
			return nil
		},
		user:   contracts.UserInfo{Name: "duc"},
		pubSub: &plugins.EmptyPubSub{},
	}
	c.terminalSessions = sess
	pty := mock.NewMockPtyHandler(m)
	sess.EXPECT().Set(gomock.Any(), gomock.Any()).Times(1)
	sess.EXPECT().Get(gomock.Any()).Return(pty, true).AnyTimes()
	pty.EXPECT().IsClosed().Return(true).AnyTimes()
	sess.EXPECT().Close(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	shell, e := HandleExecShell(&websocket_pb.WsHandleExecShellInput{
		Type: 0,
		Container: &types.Container{
			Namespace: "ns",
			Pod:       "pod",
			Container: "c",
		},
		SessionId: "ns-pod-c:xxxx",
	}, c)
	assert.Nil(t, e)
	assert.NotEmpty(t, shell)
	time.Sleep(1 * time.Second)

	_, err := HandleExecShell(&websocket_pb.WsHandleExecShellInput{
		Type: 0,
		Container: &types.Container{
			Namespace: "ns",
			Pod:       "pod",
			Container: "c",
		},
		SessionId: "xxxx",
	}, c)
	assert.Equal(t, "invalid session id, must format: '<namespace>-<pod>-<container>:<randomID>', input: 'xxxx'", err.Error())
}

type closeEqualMatcher struct {
	gomock.Matcher
}

func (c closeEqualMatcher) Matches(x any) bool {
	response := x.(*websocket_pb.WsHandleShellResponse)
	return response.Metadata.Type == WsHandleCloseShell
}

func TestMyPtyHandler_Close(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Uploader().Return(nil).AnyTimes()
	app.EXPECT().LocalUploader().Return(nil).AnyTimes()
	ps := mock.NewMockPubSub(m)
	p := &myPtyHandler{
		id:       "duc",
		conn:     &WsConn{pubSub: ps},
		recorder: &recorder{},
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		shellCh:  make(chan *websocket_pb.TerminalMessage, 2),
		doneChan: make(chan struct{}),
	}
	assert.False(t, p.IsClosed())
	assert.Len(t, p.shellCh, 0)
	ps.EXPECT().ToSelf(&closeEqualMatcher{}).Times(1)
	p.Close("aaaa")
	assert.True(t, p.IsClosed())
	p.Close("aaaa")
	assert.Len(t, p.shellCh, 2)
	a := <-p.shellCh
	assert.Equal(t, ETX, a.Data)
	b := <-p.shellCh
	assert.Equal(t, END_OF_TRANSMISSION, b.Data)
	defaultTimes := 0
	select {
	case <-p.shellCh:
	default:
		defaultTimes++
	}
	select {
	case <-p.sizeChan:
	default:
		defaultTimes++
	}
	assert.Equal(t, 2, defaultTimes)
	_, ok := <-p.doneChan
	assert.False(t, ok)
	p.Send(nil)
	p.Resize(remotecommand.TerminalSize{Width: 1, Height: 1})
	select {
	case <-p.shellCh:
	default:
		defaultTimes++
	}
	select {
	case <-p.sizeChan:
	default:
		defaultTimes++
	}
	assert.Equal(t, 2, defaultTimes)
}

func TestMyPtyHandler_Next(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	r := mock.NewMockRecorderInterface(m)
	p := &myPtyHandler{
		recorder: r,
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}),
	}
	r.EXPECT().Resize(uint16(10), uint16(20)).Times(1)
	p.Resize(remotecommand.TerminalSize{
		Width:  10,
		Height: 20,
	})
	next := p.Next()
	assert.Equal(t, uint16(10), next.Width)
	assert.Equal(t, uint16(20), next.Height)
	p.Resize(remotecommand.TerminalSize{
		Width:  100,
		Height: 200,
	})
	r.EXPECT().Resize(uint16(100), uint16(200)).Times(1)
	next = p.Next()
	assert.Equal(t, uint16(100), next.Width)
	assert.Equal(t, uint16(200), next.Height)
	assert.Equal(t, uint16(100), p.sizeStore.Cols())
	assert.Equal(t, uint16(200), p.sizeStore.Rows())

	close(p.sizeChan)
	assert.Nil(t, p.Next())
	assert.Equal(t, uint16(100), p.sizeStore.Cols())
	assert.Equal(t, uint16(200), p.sizeStore.Rows())

	p2 := &myPtyHandler{
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}),
	}
	close(p2.doneChan)
	p2.Resize(remotecommand.TerminalSize{})
	assert.Len(t, p2.sizeChan, 0)

	p3 := &myPtyHandler{
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}),
	}
	assert.Nil(t, p3.Resize(remotecommand.TerminalSize{}))
	assert.Equal(t, "sizeChan chan full", p3.Resize(remotecommand.TerminalSize{}).Error())
	close(p3.doneChan)
	assert.Equal(t, "doneChan closed", p3.Resize(remotecommand.TerminalSize{}).Error())
	assert.Len(t, p3.sizeChan, 1)
}

func TestMyPtyHandler_Next_DoneChan(t *testing.T) {
	p := &myPtyHandler{
		recorder: &recorder{},
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}),
	}

	close(p.doneChan)
	assert.Nil(t, p.Next())
}

func TestMyPtyHandler_Read(t *testing.T) {
	p := &myPtyHandler{
		id:       "duc",
		recorder: &recorder{},
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		shellCh:  make(chan *websocket_pb.TerminalMessage, 1),
		doneChan: make(chan struct{}),
	}
	b := make([]byte, 1024)
	p.Send(&websocket_pb.TerminalMessage{
		Op:   OpStdin,
		Data: []byte("hello duc"),
	})
	n, _ := p.Read(b)
	assert.Equal(t, "hello duc", string(b[0:n]))
	p.Send(&websocket_pb.TerminalMessage{
		Op:   OpResize,
		Rows: 10,
		Cols: 20,
	})
	p.Read(b)
	p.Send(&websocket_pb.TerminalMessage{
		Op:   OpResize,
		Rows: 10,
		Cols: 20,
	})
	n, _ = p.Read(b)
	assert.Equal(t, 0, n)
	assert.Len(t, p.sizeChan, 1)
	p.Send(&websocket_pb.TerminalMessage{
		Op: "xxxx",
	})
	n, err := p.Read(b)
	assert.Greater(t, n, 0)
	assert.Error(t, err)
	close(p.shellCh)
	_, err = p.Read(b)
	assert.Equal(t, "[Websocket]: duc channel closed", err.Error())
	close(p.doneChan)
	n, err = p.Read(b)
	assert.Error(t, err)
	assert.Greater(t, n, 0)

	p2 := &myPtyHandler{
		id:       "duc",
		recorder: &recorder{},
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		shellCh:  make(chan *websocket_pb.TerminalMessage, 1),
		doneChan: make(chan struct{}),
	}
	close(p2.doneChan)
	bv := make([]byte, 100)
	i, err := p2.Read(bv)
	assert.Error(t, err)
	assert.Equal(t, END_OF_TRANSMISSION, bv[:i])

	p3 := &myPtyHandler{
		id:       "duc",
		recorder: &recorder{},
		shellCh:  make(chan *websocket_pb.TerminalMessage, 1),
		doneChan: make(chan struct{}),
	}
	assert.Len(t, p3.shellCh, 0)
	assert.Nil(t, p3.Send(nil))
	assert.Equal(t, "shellCh chan full", p3.Send(nil).Error())
	close(p3.doneChan)
	assert.Equal(t, "doneChan closed", p3.Send(nil).Error())
	assert.Len(t, p3.shellCh, 1)
}

func TestMyPtyHandler_Recorder(t *testing.T) {
	p := &myPtyHandler{
		recorder: &recorder{},
	}
	assert.Implements(t, (*contracts.RecorderInterface)(nil), p.Recorder())
}

func TestMyPtyHandler_SetShell(t *testing.T) {
	p := &myPtyHandler{
		recorder: &recorder{},
	}
	p.SetShell("xxx")
	assert.Equal(t, "xxx", p.recorder.GetShell())
}

func TestMyPtyHandler_Toast(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ps := mock.NewMockPubSub(m)
	p := &myPtyHandler{
		conn: &WsConn{pubSub: ps},
		id:   "aaa",
	}
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	p.Toast("xxx")
}

func TestMyPtyHandler_Write(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ps := mock.NewMockPubSub(m)
	r := mock.NewMockRecorderInterface(m)
	p := &myPtyHandler{
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		sizeStore: sizeStore{
			cols:  106,
			rows:  25,
			reset: true,
		},
		id:       "duc",
		conn:     &WsConn{pubSub: ps},
		recorder: r,
		doneChan: make(chan struct{}),
	}
	r.EXPECT().Write("aaa").Times(1)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	n, err := p.Write([]byte("aaa"))
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
	p.closeMu.Lock()
	p.closed = true
	p.closeMu.Unlock()
	n, err = p.Write([]byte("aaa"))
	assert.Equal(t, "[Websocket]: duc ws already closed", err.Error())
	assert.Equal(t, 3, n)

	close(p.doneChan)
	n, err = p.Write([]byte("aaa"))
	assert.Equal(t, "[Websocket]: duc doneChan closed", err.Error())
	assert.Equal(t, 3, n)
	sch := <-p.sizeChan
	assert.Equal(t, uint16(106), sch.Width)
	assert.Equal(t, uint16(25), sch.Height)
	assert.False(t, p.sizeStore.TerminalRowColNeedReset())
}

func TestMyPtyHandler_Write_with_chan_full(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ps := mock.NewMockPubSub(m)
	r := mock.NewMockRecorderInterface(m)
	p := &myPtyHandler{
		sizeChan: make(chan remotecommand.TerminalSize),
		sizeStore: sizeStore{
			cols:  106,
			rows:  25,
			reset: true,
		},
		id:       "duc",
		conn:     &WsConn{pubSub: ps},
		recorder: r,
		doneChan: make(chan struct{}),
	}
	r.EXPECT().Write("aaa").Times(1)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	// 会走 default，没有 select 会卡住
	n, err := p.Write([]byte("aaa"))
	assert.True(t, true)
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
}

func TestSessionMap_Close(t *testing.T) {
	t.Parallel()
	m := gomock.NewController(t)
	defer m.Finish()
	p1 := mock.NewMockPtyHandler(m)
	p2 := mock.NewMockPtyHandler(m)
	sm := sessionMap{
		sessLock: sync.RWMutex{},
		Sessions: map[string]contracts.PtyHandler{},
	}
	sm.Set("p1", p1)
	sm.Set("p2", p2)
	p1.EXPECT().Close("err")
	sm.Close("p1", 0, "err")
	time.Sleep(1 * time.Second)
}

func TestSessionMap_CloseAll(t *testing.T) {
	t.Parallel()
	m := gomock.NewController(t)
	defer m.Finish()
	p1 := mock.NewMockPtyHandler(m)
	p2 := mock.NewMockPtyHandler(m)
	sm := sessionMap{
		sessLock: sync.RWMutex{},
		Sessions: map[string]contracts.PtyHandler{},
	}
	sm.Set("p1", p1)
	sm.Set("p2", p2)
	assert.Len(t, sm.Sessions, 2)
	p1.EXPECT().Close("websocket conn closed")
	p2.EXPECT().Close("websocket conn closed")
	sm.CloseAll()
	assert.Len(t, sm.Sessions, 0)
}

func TestSessionMap_Send(t *testing.T) {
	t.Parallel()
	sm := sessionMap{
		sessLock: sync.RWMutex{},
		Sessions: map[string]contracts.PtyHandler{},
	}
	ch := make(chan *websocket_pb.TerminalMessage, 1)
	sm.Set("a", &myPtyHandler{
		shellCh: ch,
	})
	assert.Len(t, ch, 0)
	sm.Send(&websocket_pb.TerminalMessage{
		Data:      []byte("aa"),
		SessionId: "a",
	})
	assert.Len(t, ch, 1)
	sm.Send(&websocket_pb.TerminalMessage{
		Data:      []byte("aa"),
		SessionId: "a",
	})
	assert.Len(t, ch, 1)
}

func TestSessionMap_Set(t *testing.T) {
	t.Parallel()
	sm := sessionMap{
		sessLock: sync.RWMutex{},
		Sessions: map[string]contracts.PtyHandler{},
	}
	p := &myPtyHandler{}
	sm.Set("a", p)
	get, _ := sm.Get("a")
	assert.Same(t, p, get)
}

func TestWaitForTerminal(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	r := mock.NewMockRemoteExecutor(m)
	r.EXPECT().WithContainer(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(r)
	r.EXPECT().WithCommand(gomock.Any()).Times(1).Return(r)
	r.EXPECT().WithMethod(gomock.Any()).Times(1).Return(r)
	r.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	c := &WsConn{
		newExecutorFunc: func() contracts.RemoteExecutor {
			return r
		},
	}
	c.terminalSessions = NewSessionMap(c)
	pty := mock.NewMockPtyHandler(m)
	pty.EXPECT().SetShell("bash").Times(1)
	pty.EXPECT().Close("Process exited").Times(1)
	c.terminalSessions.Set("session_id", pty)
	WaitForTerminal(c, nil, nil, &contracts.Container{}, "bash", "session_id")
	assert.Len(t, c.terminalSessions.(*sessionMap).Sessions, 0)
	c.terminalSessions.CloseAll()
}

func TestWaitForTerminalUnsetShell(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	r := mock.NewMockRemoteExecutor(m)
	c := &WsConn{
		newExecutorFunc: func() contracts.RemoteExecutor {
			return r
		},
	}
	c.terminalSessions = NewSessionMap(c)
	pty := mock.NewMockPtyHandler(m)
	pty.EXPECT().SetShell(gomock.Any()).Times(0)
	pty.EXPECT().IsClosed().Return(true)
	pty.EXPECT().Close("Process exited").Times(1)
	c.terminalSessions.Set("session_id", pty)
	WaitForTerminal(c, nil, nil, &contracts.Container{}, "", "session_id")
	assert.Len(t, c.terminalSessions.(*sessionMap).Sessions, 0)
	c.terminalSessions.CloseAll()
}

func TestWaitForTerminalUnsetShell2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	r := mock.NewMockRemoteExecutor(m)
	r.EXPECT().WithContainer(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(r)
	r.EXPECT().WithCommand(gomock.Any()).Times(1).Return(r)
	r.EXPECT().WithMethod(gomock.Any()).Times(1).Return(r)
	r.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
	c := &WsConn{
		newExecutorFunc: func() contracts.RemoteExecutor {
			return r
		},
	}
	c.terminalSessions = NewSessionMap(c)
	pty := mock.NewMockPtyHandler(m)
	pty.EXPECT().IsClosed().Return(false).Times(1)
	pty.EXPECT().SetShell("bash").Times(1)
	pty.EXPECT().Close("Process exited").Times(1)
	c.terminalSessions.Set("session_id", pty)
	WaitForTerminal(c, nil, nil, &contracts.Container{}, "", "session_id")
	assert.Len(t, c.terminalSessions.(*sessionMap).Sessions, 0)
	c.terminalSessions.CloseAll()
}

func TestWaitForTerminalReturnError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	r := mock.NewMockRemoteExecutor(m)
	r.EXPECT().WithContainer(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(r)
	r.EXPECT().WithCommand(gomock.Any()).Times(1).Return(r)
	r.EXPECT().WithMethod(gomock.Any()).Times(1).Return(r)
	r.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("xxx"))
	c := &WsConn{
		newExecutorFunc: func() contracts.RemoteExecutor {
			return r
		},
	}
	c.terminalSessions = NewSessionMap(c)
	pty := mock.NewMockPtyHandler(m)
	pty.EXPECT().SetShell("bash").Times(1)
	pty.EXPECT().Toast("xxx").Times(1)
	pty.EXPECT().Close("xxx").Times(1)
	c.terminalSessions.Set("session_id", pty)
	WaitForTerminal(c, nil, nil, &contracts.Container{}, "bash", "session_id")
	assert.Len(t, c.terminalSessions.(*sessionMap).Sessions, 0)
	c.terminalSessions.CloseAll()
}

func TestWaitForTerminalReturnErrorSilence(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	r := mock.NewMockRemoteExecutor(m)
	r.EXPECT().WithContainer(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(r)
	r.EXPECT().WithCommand(gomock.Any()).Times(1).Return(r)
	r.EXPECT().WithMethod(gomock.Any()).Times(1).Return(r)
	r.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("command terminated with exit code 126"))
	c := &WsConn{
		newExecutorFunc: func() contracts.RemoteExecutor {
			return r
		},
	}
	c.terminalSessions = NewSessionMap(c)
	pty := mock.NewMockPtyHandler(m)
	pty.EXPECT().SetShell("bash").Times(1)
	pty.EXPECT().Toast(gomock.Any()).Times(0)
	pty.EXPECT().Close("command terminated with exit code 126").Times(1)
	c.terminalSessions.Set("session_id", pty)
	WaitForTerminal(c, nil, nil, &contracts.Container{}, "bash", "session_id")
	assert.Len(t, c.terminalSessions.(*sessionMap).Sessions, 0)
	c.terminalSessions.CloseAll()
}

func Test_isValidShell(t *testing.T) {
	t.Parallel()
	assert.False(t, isValidShell([]string{"bash"}, "sh"))
	assert.True(t, isValidShell([]string{"bash", "sh"}, "sh"))
}

func Test_silence(t *testing.T) {
	t.Parallel()
	assert.True(t, silence(errors.New("command terminated with exit code 126")))
	assert.False(t, silence(errors.New("command terminated with exit code 127")))
}

func Test_startProcess(t *testing.T) {}

func Test_sizeStore_Changed(t *testing.T) {
	t.Parallel()
	ss := sizeStore{
		cols:  0,
		rows:  0,
		reset: false,
	}
	assert.True(t, ss.Changed(100, 100))
	ss.Set(100, 100)
	assert.False(t, ss.Changed(100, 100))
	assert.True(t, ss.Changed(100, 0))
	assert.True(t, ss.Changed(0, 100))
}

func TestMyPtyHandler_Rows(t *testing.T) {
	t.Parallel()
	assert.Equal(t, uint16(0), (&myPtyHandler{}).Rows())
	assert.Equal(t, uint16(100), (&myPtyHandler{sizeStore: sizeStore{rows: 100}}).Rows())
}

func TestMyPtyHandler_Cols(t *testing.T) {
	t.Parallel()
	assert.Equal(t, uint16(0), (&myPtyHandler{}).Cols())
	assert.Equal(t, uint16(100), (&myPtyHandler{sizeStore: sizeStore{cols: 100}}).Cols())
}

func Test_resetSession(t *testing.T) {
	t.Parallel()
	old := &myPtyHandler{
		container: contracts.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &recorder{},
		id:        "id",
		conn:      &WsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan remotecommand.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: sizeStore{cols: 10, rows: 10},
	}
	session := resetSession(old).(*myPtyHandler)

	assert.Equal(t, old.id, session.id)
	assert.Equal(t, old.container, session.container)
	assert.Same(t, old.recorder, session.recorder)
	assert.Same(t, old.conn, session.conn)

	assert.Equal(t, old.sizeStore.Cols(), session.sizeStore.Cols())
	assert.Equal(t, old.sizeStore.Rows(), session.sizeStore.Rows())
	assert.True(t, session.sizeStore.TerminalRowColNeedReset())

	assert.NotEqual(t, old.shellCh, session.shellCh)
	assert.NotEqual(t, old.sizeChan, session.sizeChan)
	assert.NotEqual(t, old.doneChan, session.doneChan)
}
func Test_resetSession4(t *testing.T) {
	t.Parallel()
	old := &myPtyHandler{
		container: contracts.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &recorder{},
		id:        "id",
		conn:      &WsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan remotecommand.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: sizeStore{cols: 10, rows: 10},
	}
	session := resetSession(old).(*myPtyHandler)
	assert.NotSame(t, session, old)
	old.CloseDoneChan()
	session = resetSession(old).(*myPtyHandler)
	assert.Same(t, session, old)
}

func Test_resetSession1(t *testing.T) {
	t.Parallel()
	old := &myPtyHandler{
		container: contracts.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder: &recorder{},
		id:       "id",
		conn:     &WsConn{},
		doneChan: make(chan struct{}, 1),
		sizeChan: make(chan remotecommand.TerminalSize, 10),
		shellCh:  make(chan *websocket_pb.TerminalMessage, 10),
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		old.sizeStore.Set(100, 100)
	}()
	session := resetSession(old).(*myPtyHandler)

	assert.Equal(t, uint16(100), session.sizeStore.Cols())
	assert.Equal(t, uint16(100), session.sizeStore.Rows())
}

func Test_resetSession2(t *testing.T) {
	t.Parallel()
	old := &myPtyHandler{
		container: contracts.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder: &recorder{},
		id:       "id",
		conn:     &WsConn{},
		doneChan: make(chan struct{}, 1),
		sizeChan: make(chan remotecommand.TerminalSize, 10),
		shellCh:  make(chan *websocket_pb.TerminalMessage, 10),
	}
	go func() {
		time.Sleep(4 * time.Second)
		old.sizeStore.Set(100, 100)
	}()
	session := resetSession(old).(*myPtyHandler)

	assert.Equal(t, uint16(106), session.sizeStore.Cols())
	assert.Equal(t, uint16(25), session.sizeStore.Rows())
}

func TestMyPtyHandler_ResetTerminalRowCol(t *testing.T) {
	pty := &myPtyHandler{}
	pty.ResetTerminalRowCol(true)
	assert.True(t, pty.sizeStore.TerminalRowColNeedReset())
	pty.ResetTerminalRowCol(false)
	assert.False(t, pty.sizeStore.TerminalRowColNeedReset())
}

func TestMyPtyHandler_ClosePreviousChannels(t *testing.T) {
	pty := &myPtyHandler{
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}, 1),
		shellCh:  make(chan *websocket_pb.TerminalMessage),
	}
	assert.True(t, pty.CloseDoneChan())
	assert.False(t, pty.CloseDoneChan())
	assert.True(t, pty.IsClosed())
	defaultTimes := 0
	select {
	case <-pty.shellCh:
	default:
		defaultTimes++
	}
	select {
	case <-pty.sizeChan:
	default:
		defaultTimes++
	}
	assert.Equal(t, 2, defaultTimes)
	_, ok := <-pty.doneChan
	assert.False(t, ok)
	pty.Send(nil)
	pty.Resize(remotecommand.TerminalSize{Width: 1, Height: 1})
	select {
	case <-pty.shellCh:
	default:
		defaultTimes++
	}
	select {
	case <-pty.sizeChan:
	default:
		defaultTimes++
	}
	assert.Equal(t, 2, defaultTimes)
}

func Test_startProcess1(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	e := mock.NewMockRemoteExecutor(m)
	pty := &myPtyHandler{}
	cfg := &rest.Config{}
	e.EXPECT().WithMethod("POST").Return(e)
	e.EXPECT().WithCommand([]string{"ls"}).Return(e)
	e.EXPECT().WithContainer("ns", "pod", "c").Return(e)
	e.EXPECT().Execute(context.TODO(), nil, cfg, pty, pty, pty, true, pty).Return(nil)
	assert.Nil(t, startProcess(e, nil, cfg, &contracts.Container{
		Namespace: "ns",
		Pod:       "pod",
		Container: "c",
	}, []string{"ls"}, pty))
}

func Test_checkSessionID(t *testing.T) {
	var tests = []struct {
		c     *types.Container
		id    string
		wants bool
	}{
		{
			c: &types.Container{
				Namespace: "ns",
				Pod:       "pod",
				Container: "c",
			},
			id:    "xxxx",
			wants: false,
		},
		{
			c: &types.Container{
				Namespace: "ns",
				Pod:       "pod",
				Container: "c",
			},
			id:    "ns-pod-c:xx",
			wants: true,
		},
		{
			c: &types.Container{
				Namespace: "ns",
				Pod:       "pod",
				Container: "c",
			},
			id:    "ns-pod-c:",
			wants: true,
		},
		{
			c: &types.Container{
				Namespace: "ns",
				Pod:       "pod",
				Container: "c",
			},
			id:    "ns-pod-c",
			wants: false,
		},
		{
			c: &types.Container{
				Namespace: "ns",
				Pod:       "pod",
				Container: "c",
			},
			id:    "ns-c-pod",
			wants: false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.wants, checkSessionID(tt.c, tt.id))
		})
	}
}
