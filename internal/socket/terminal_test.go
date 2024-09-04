package socket

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v5/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v5/websocket"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"k8s.io/client-go/tools/remotecommand"
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

func TestMyPtyHandler_sizeStore(t *testing.T) {
	pty := &myPtyHandler{
		sizeStore: &sizeStore{},
	}
	pty.ResetTerminalRowCol(true)
	assert.True(t, pty.sizeStore.TerminalRowColNeedReset())
	pty.ResetTerminalRowCol(false)
	assert.False(t, pty.sizeStore.TerminalRowColNeedReset())
	assert.Equal(t, uint16(0), pty.sizeStore.Cols())
	assert.Equal(t, uint16(0), pty.sizeStore.Rows())
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

func TestSessionMap_Get(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	sm := NewSessionMap(logger)
	session := &myPtyHandler{}
	sm.Set("testSession", session)

	retrievedSession, ok := sm.Get("testSession")
	assert.True(t, ok)
	assert.Equal(t, session, retrievedSession)

	_, ok = sm.Get("nonExistentSession")
	assert.False(t, ok)
}

func TestSessionMap_Set(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	sm := NewSessionMap(logger)
	session := &myPtyHandler{}

	sm.Set("testSession", session)
	retrievedSession, ok := sm.Get("testSession")

	assert.True(t, ok)
	assert.Equal(t, session, retrievedSession)
}

type testPtyHandler struct {
	PtyHandler
}

func (*testPtyHandler) Close(context.Context, string) bool {
	return true
}

func (*testPtyHandler) IsClosed() bool {
	return false
}

func TestSessionMap_CloseAll(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	sm := NewSessionMap(logger)
	session1 := &testPtyHandler{}
	session2 := &testPtyHandler{}

	sm.Set("session1", session1)
	sm.Set("session2", session2)

	sm.CloseAll(context.TODO())

	_, ok := sm.Get("session1")
	assert.False(t, ok)

	_, ok = sm.Get("session2")
	assert.False(t, ok)
}

func TestSessionMap_Close(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	sm := NewSessionMap(logger)
	session := &testPtyHandler{}

	sm.Set("testSession", session)
	sm.Close(context.TODO(), "testSession", 0, "testReason")

	_, ok := sm.Get("testSession")
	assert.False(t, ok)
}

func TestWebsocketManager_startProcess(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sRepo := repo.NewMockK8sRepo(m)
	ws := &WebsocketManager{
		k8sRepo: sRepo,
	}
	c := &repo.Container{}
	h := &testPtyHandler{}
	sRepo.EXPECT().Execute(gomock.Any(), c, &repo.ExecuteInput{
		Stdin:             h,
		Stdout:            h,
		Stderr:            h,
		TTY:               true,
		Cmd:               []string{},
		TerminalSizeQueue: h,
	})
	ws.startProcess(context.TODO(), c, []string{}, h)
}

func TestWebsocketManager_WaitForTerminal(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sRepo := repo.NewMockK8sRepo(m)
	conn := NewMockConn(m)
	ws := &WebsocketManager{
		k8sRepo: sRepo,
		logger:  mlog.NewForConfig(nil),
	}
	c := &repo.Container{}
	handler := NewMockPtyHandler(m)
	handler.EXPECT().Toast(gomock.Any())
	conn.EXPECT().ClosePty(gomock.Any(), "sid", uint32(2), "x")
	conn.EXPECT().GetPtyHandler("sid").Return(handler, true)
	handler.EXPECT().SetShell("sh")
	sRepo.EXPECT().Execute(gomock.Any(), c, gomock.Any()).Return(errors.New("x"))
	ws.WaitForTerminal(context.TODO(), conn, c, "sh", "sid")
}

func TestWebsocketManager_WaitForTerminal2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sRepo := repo.NewMockK8sRepo(m)
	conn := NewMockConn(m)
	ws := &WebsocketManager{
		k8sRepo: sRepo,
		logger:  mlog.NewForConfig(nil),
	}
	c := &repo.Container{}
	handler := NewMockPtyHandler(m)
	conn.EXPECT().ClosePty(gomock.Any(), "sid", uint32(1), "Process exited")
	conn.EXPECT().GetPtyHandler("sid").Return(handler, true)
	handler.EXPECT().SetShell("bash")
	handler.EXPECT().IsClosed().Return(false)
	sRepo.EXPECT().Execute(gomock.Any(), c, gomock.Any()).Return(nil)
	ws.WaitForTerminal(context.TODO(), conn, c, "xsh", "sid")
}

func TestWebsocketManager_WaitForTerminal3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sRepo := repo.NewMockK8sRepo(m)
	conn := NewMockConn(m)
	ws := &WebsocketManager{
		k8sRepo: sRepo,
		logger:  mlog.NewForConfig(nil),
	}
	c := &repo.Container{}
	handler := NewMockPtyHandler(m)
	conn.EXPECT().ClosePty(gomock.Any(), "sid", uint32(1), "Process exited")
	conn.EXPECT().GetPtyHandler("sid").Return(handler, true)
	handler.EXPECT().IsClosed().Return(true)
	ws.WaitForTerminal(context.TODO(), conn, c, "xsh", "sid")
}

type testRecorder struct {
	repo.Recorder
}

func Test_resetSession(t *testing.T) {
	t.Parallel()
	old := &myPtyHandler{
		container: &repo.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &testRecorder{},
		sessionID: "id",
		conn:      &WsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan remotecommand.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: &sizeStore{cols: 10, rows: 10},
	}
	session := (&WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}).resetSession(old).(*myPtyHandler)

	assert.Equal(t, old.sessionID, session.sessionID)
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
		container: &repo.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &testRecorder{},
		sessionID: "id",
		conn:      &WsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan remotecommand.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: &sizeStore{cols: 10, rows: 10},
	}
	session := (&WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}).resetSession(old).(*myPtyHandler)
	assert.NotSame(t, session, old)
	old.CloseDoneChan()
	session = (&WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}).resetSession(old).(*myPtyHandler)
	assert.Same(t, session, old)
}

func Test_resetSession1(t *testing.T) {
	t.Parallel()
	old := &myPtyHandler{
		container: &repo.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &testRecorder{},
		sessionID: "id",
		conn:      &WsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan remotecommand.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: &sizeStore{},
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		old.sizeStore.Set(100, 100)
	}()
	session := (&WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}).resetSession(old).(*myPtyHandler)

	assert.Equal(t, uint16(100), session.sizeStore.Cols())
	assert.Equal(t, uint16(100), session.sizeStore.Rows())
}

func Test_resetSession2(t *testing.T) {
	t.Parallel()
	old := &myPtyHandler{
		container: &repo.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &testRecorder{},
		sessionID: "id",
		conn:      &WsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan remotecommand.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: &sizeStore{},
	}
	go func() {
		time.Sleep(4 * time.Second)
		old.sizeStore.Set(100, 100)
	}()
	session := (&WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}).resetSession(old).(*myPtyHandler)

	assert.Equal(t, uint16(106), session.sizeStore.Cols())
	assert.Equal(t, uint16(25), session.sizeStore.Rows())
}

func TestMyPtyHandler_Next_DoneChan(t *testing.T) {
	p := &myPtyHandler{
		recorder: &testRecorder{},
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}),
	}

	close(p.doneChan)
	assert.Nil(t, p.Next())
}

func TestMyPtyHandler_Next(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	r := repo.NewMockRecorder(m)
	p := &myPtyHandler{
		recorder:  r,
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
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
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
	}
	close(p2.doneChan)
	p2.Resize(remotecommand.TerminalSize{})
	assert.Len(t, p2.sizeChan, 0)

	p3 := &myPtyHandler{
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
	}
	assert.Nil(t, p3.Resize(remotecommand.TerminalSize{}))
	assert.Equal(t, "sizeChan chan full", p3.Resize(remotecommand.TerminalSize{}).Error())
	close(p3.doneChan)
	assert.Equal(t, "doneChan closed", p3.Resize(remotecommand.TerminalSize{}).Error())
	assert.Len(t, p3.sizeChan, 1)
}

func TestMyPtyHandler_Read2(t *testing.T) {
	p := &myPtyHandler{
		sessionID: "duc",
		recorder:  &testRecorder{},
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 1),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
		logger:    mlog.NewForConfig(nil),
	}
	b := make([]byte, 1024)
	p.Send(context.TODO(), &websocket_pb.TerminalMessage{
		Op:   OpStdin,
		Data: []byte("hello duc"),
	})
	n, _ := p.Read(b)
	assert.Equal(t, "hello duc", string(b[0:n]))
	p.Send(context.TODO(), &websocket_pb.TerminalMessage{
		Op:   OpResize,
		Rows: 10,
		Cols: 20,
	})
	p.Read(b)
	p.Send(context.TODO(), &websocket_pb.TerminalMessage{
		Op:   OpResize,
		Rows: 10,
		Cols: 20,
	})
	n, _ = p.Read(b)
	assert.Equal(t, 0, n)
	assert.Len(t, p.sizeChan, 1)
	p.Send(context.TODO(), &websocket_pb.TerminalMessage{
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
		sessionID: "duc",
		recorder:  &testRecorder{},
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 1),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
		logger:    mlog.NewForConfig(nil),
	}
	close(p2.doneChan)
	bv := make([]byte, 100)
	i, err := p2.Read(bv)
	assert.Error(t, err)
	assert.Equal(t, END_OF_TRANSMISSION, bv[:i])

	p3 := &myPtyHandler{
		sessionID: "duc",
		recorder:  &testRecorder{},
		shellCh:   make(chan *websocket_pb.TerminalMessage, 1),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
		logger:    mlog.NewForConfig(nil),
	}
	assert.Len(t, p3.shellCh, 0)
	assert.Nil(t, p3.Send(context.TODO(), nil))
	assert.Nil(t, p3.Send(context.TODO(), nil))
	close(p3.doneChan)
	assert.Equal(t, "doneChan closed", p3.Send(context.TODO(), nil).Error())
	assert.Len(t, p3.shellCh, 1)
}

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

func TestMyPtyHandler_Close(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := repo.NewMockRecorder(m)
	ps := application.NewMockPubSub(m)
	eventRepo := repo.NewMockEventRepo(m)
	p := &myPtyHandler{
		sessionID: "duc",
		conn:      &WsConn{pubSub: ps},
		recorder:  recorder,
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 2),
		doneChan:  make(chan struct{}),
		container: &repo.Container{},
		eventRepo: eventRepo,
		logger:    mlog.NewForConfig(nil),
	}
	eventRepo.EXPECT().FileAuditLogWithDuration(
		types.EventActionType_Shell, "duc",
		gomock.Not(nil), 1,
		time.Second)
	recorder.EXPECT().User().Return(&auth.UserInfo{Name: "duc"})
	recorder.EXPECT().Duration().Return(time.Second)
	recorder.EXPECT().Container().Return(&repo.Container{}).AnyTimes()
	recorder.EXPECT().File().Return(&repo.File{ID: 1})
	recorder.EXPECT().Close().Return(errors.New("x"))
	assert.False(t, p.IsClosed())
	assert.Len(t, p.shellCh, 0)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	p.Close(context.TODO(), "aaaa")
	assert.True(t, p.IsClosed())
	p.Close(context.TODO(), "aaaa")
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
	p.Send(context.TODO(), nil)
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

func TestMyPtyHandler_Write(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := repo.NewMockRecorder(m)
	conn := NewMockConn(m)
	pty := &myPtyHandler{
		sessionID: "sid",
		conn:      conn,
		logger:    mlog.NewForConfig(nil),
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
	assert.Same(t, recorder, pty.Recorder())
	conn.EXPECT().PubSub().Return(sub)
	conn.EXPECT().ID().Return("id")
	conn.EXPECT().UID().Return("uid")
	recorder.EXPECT().Write([]byte("data")).Return(0, errors.New("x"))
	n, err := pty.Write([]byte("data"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
}

func TestMyPtyHandler_Write3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ps := application.NewMockPubSub(m)
	r := repo.NewMockRecorder(m)
	p := &myPtyHandler{
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		sizeStore: &sizeStore{
			cols:  106,
			rows:  25,
			reset: true,
		},
		sessionID: "duc",
		container: &repo.Container{},
		logger:    mlog.NewForConfig(nil),
		conn:      &WsConn{pubSub: ps},
		recorder:  r,
		doneChan:  make(chan struct{}),
	}
	r.EXPECT().Write([]byte("aaa")).Times(1)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	n, err := p.Write([]byte("aaa"))
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
	p.Closeable.Close()
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
	ps := application.NewMockPubSub(m)
	r := repo.NewMockRecorder(m)
	p := &myPtyHandler{
		sizeChan: make(chan remotecommand.TerminalSize),
		sizeStore: &sizeStore{
			cols:  106,
			rows:  25,
			reset: true,
		},
		container: &repo.Container{},
		sessionID: "duc",
		logger:    mlog.NewForConfig(nil),
		conn:      &WsConn{pubSub: ps},
		recorder:  r,
		doneChan:  make(chan struct{}),
	}
	r.EXPECT().Write([]byte("aaa")).Times(1)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	// 会走 default，没有 select 会卡住
	n, err := p.Write([]byte("aaa"))
	assert.True(t, true)
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
}

func TestStartShell_WithValidSessionID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	fileRepo := repo.NewMockFileRepo(m)
	fileRepo.EXPECT().NewRecorder(gomock.Any(), gomock.Any())
	conn.EXPECT().GetUser().Return(&auth.UserInfo{})
	ws := &WebsocketManager{
		logger:   mlog.NewForConfig(nil),
		fileRepo: fileRepo,
	}

	input := &websocket_pb.WsHandleExecShellInput{
		Container: &websocket_pb.Container{
			Namespace: "namespace",
			Pod:       "pod",
			Container: "container",
		},
		SessionId: "namespace-pod-container:randomID",
	}

	conn.EXPECT().SetPtyHandler(input.SessionId, gomock.Any())
	conn.EXPECT().GetPtyHandler(input.SessionId).Return(nil, false)

	sessionID, err := ws.StartShell(context.TODO(), input, conn)
	time.Sleep(1 * time.Second)
	assert.NoError(t, err)
	assert.Equal(t, input.SessionId, sessionID)
}

func TestStartShell_WithInvalidSessionID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	ws := &WebsocketManager{
		logger: mlog.NewForConfig(nil),
	}

	input := &websocket_pb.WsHandleExecShellInput{
		Container: &websocket_pb.Container{
			Namespace: "namespace",
			Pod:       "pod",
			Container: "container",
		},
		SessionId: "invalidSessionID",
	}

	_, err := ws.StartShell(context.TODO(), input, conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session sessionID")
}

func Test_myPtyHandler_Send(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Equal(t, context.Canceled, (&myPtyHandler{}).Send(ctx, nil))
}
