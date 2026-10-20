package websocket

import (
	"bytes"
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
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
	assert.True(t, shouldSilenceShellError(errors.New("command terminated with exit code 126")))
	assert.True(t, shouldSilenceShellError(errors.New("command terminated with exit code 130")))
	assert.False(t, shouldSilenceShellError(errors.New("command terminated with exit code 131")))
}

func TestSizeStore(t *testing.T) {
	s := &sizeStore{}

	s.Set(10, 20)
	assert.Equal(t, uint16(10), s.Width())
	assert.Equal(t, uint16(20), s.Height())
	assert.True(t, s.Changed(11, 20))
	assert.False(t, s.Changed(10, 20))

	s.ResetTerminalRowCol(true)
	assert.True(t, s.TerminalRowColNeedReset())
	s.ResetTerminalRowCol(false)
	assert.False(t, s.TerminalRowColNeedReset())
}

func TestContainer(t *testing.T) {
	container := &biz.Container{
		Namespace: "namespace",
		Pod:       "pod",
		Container: "container",
	}
	pty := &ptyHandler{
		container: container,
	}
	assert.Equal(t, container, pty.Container())
}

func TestPtyHandler_SetShell(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := data.NewMockRecorder(m)
	pty := &ptyHandler{
		recorder: recorder,
	}
	recorder.EXPECT().SetShell("bash")
	pty.SetShell("bash")
}

func TestPtyHandler_IsClosed(t *testing.T) {
	pty := &ptyHandler{}
	assert.False(t, pty.IsClosed())
	pty.Closeable.Close()
	assert.True(t, pty.IsClosed())
}

func TestPtyHandler_sizeStore(t *testing.T) {
	pty := &ptyHandler{
		sizeStore: &sizeStore{},
	}
	pty.ResetTerminalRowCol(true)
	assert.True(t, pty.sizeStore.TerminalRowColNeedReset())
	pty.ResetTerminalRowCol(false)
	assert.False(t, pty.sizeStore.TerminalRowColNeedReset())
	assert.Equal(t, uint16(0), pty.sizeStore.Width())
	assert.Equal(t, uint16(0), pty.sizeStore.Height())
}

func TestPtyHandler_Read(t *testing.T) {
	pty := &ptyHandler{
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

func TestPtyHandler_Toast(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	conn := NewMockConn(m)
	pty := &ptyHandler{
		conn:      conn,
		container: &biz.Container{},
	}
	conn.EXPECT().ID().Return("id")
	conn.EXPECT().UID().Return("uid")
	sub := app.NewMockPubSub(m)
	conn.EXPECT().PubSub().Return(sub)
	sub.EXPECT().ToSelf(gomock.Any())

	assert.NoError(t, pty.Toast("message"))
}

func TestSessionMap_Get(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	sm := NewSessionMap(logger)
	session := &ptyHandler{}
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
	session := &ptyHandler{}

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

func TestWebsocketManager_execInContainer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sRepo := data.NewMockK8sRepo(m)
	ws := &websocketManager{
		k8sRepo: sRepo,
	}
	c := &biz.Container{}
	h := &testPtyHandler{}
	sRepo.EXPECT().Execute(gomock.Any(), c, &biz.ExecuteInput{
		Stdin:             h,
		Stdout:            h,
		Stderr:            h,
		TTY:               true,
		Cmd:               []string{},
		TerminalSizeQueue: h,
	})
	ws.execInContainer(context.TODO(), c, []string{}, h)
}

func TestWebsocketManager_runTerminal(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sRepo := data.NewMockK8sRepo(m)
	conn := NewMockConn(m)
	ws := &websocketManager{
		k8sRepo: sRepo,
		logger:  mlog.NewForConfig(nil),
	}
	c := &biz.Container{}
	handler := NewMockPtyHandler(m)
	handler.EXPECT().Toast(gomock.Any())
	conn.EXPECT().ClosePty(gomock.Any(), "sid", uint32(2), "x")
	conn.EXPECT().GetPtyHandler("sid").Return(handler, true)
	handler.EXPECT().SetShell("sh")
	sRepo.EXPECT().Execute(gomock.Any(), c, gomock.Any()).Return(errors.New("x"))
	ws.runTerminal(context.TODO(), conn, c, "sh", "sid")
}

func TestWebsocketManager_runTerminal2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sRepo := data.NewMockK8sRepo(m)
	conn := NewMockConn(m)
	ws := &websocketManager{
		k8sRepo: sRepo,
		logger:  mlog.NewForConfig(nil),
	}
	c := &biz.Container{}
	handler := NewMockPtyHandler(m)
	conn.EXPECT().ClosePty(gomock.Any(), "sid", uint32(1), "Process exited")
	conn.EXPECT().GetPtyHandler("sid").Return(handler, true)
	handler.EXPECT().SetShell("bash")
	handler.EXPECT().IsClosed().Return(false)
	sRepo.EXPECT().Execute(gomock.Any(), c, gomock.Any()).Return(nil)
	ws.runTerminal(context.TODO(), conn, c, "xsh", "sid")
}

func TestWebsocketManager_runTerminal3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sRepo := data.NewMockK8sRepo(m)
	conn := NewMockConn(m)
	ws := &websocketManager{
		k8sRepo: sRepo,
		logger:  mlog.NewForConfig(nil),
	}
	c := &biz.Container{}
	handler := NewMockPtyHandler(m)
	conn.EXPECT().ClosePty(gomock.Any(), "sid", uint32(1), "Process exited")
	conn.EXPECT().GetPtyHandler("sid").Return(handler, true)
	handler.EXPECT().IsClosed().Return(true)
	ws.runTerminal(context.TODO(), conn, c, "xsh", "sid")
}

type testRecorder struct {
	biz.Recorder
}

func Test_resetSession(t *testing.T) {
	t.Parallel()
	old := &ptyHandler{
		container: &biz.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &testRecorder{},
		sessionID: "id",
		conn:      &wsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan biz.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: &sizeStore{width: 10, height: 10},
	}
	session := (&websocketManager{
		logger: mlog.NewForConfig(nil),
	}).resetSession(old).(*ptyHandler)

	assert.Equal(t, old.sessionID, session.sessionID)
	assert.Equal(t, old.container, session.container)
	assert.Same(t, old.recorder, session.recorder)
	assert.Same(t, old.conn, session.conn)

	assert.Equal(t, old.sizeStore.Width(), session.sizeStore.Width())
	assert.Equal(t, old.sizeStore.Height(), session.sizeStore.Height())
	assert.True(t, session.sizeStore.TerminalRowColNeedReset())

	assert.NotEqual(t, old.shellCh, session.shellCh)
	assert.NotEqual(t, old.sizeChan, session.sizeChan)
	assert.NotEqual(t, old.doneChan, session.doneChan)
}

func Test_resetSession4(t *testing.T) {
	t.Parallel()
	old := &ptyHandler{
		container: &biz.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &testRecorder{},
		sessionID: "id",
		conn:      &wsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan biz.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: &sizeStore{width: 10, height: 10},
	}
	session := (&websocketManager{
		logger: mlog.NewForConfig(nil),
	}).resetSession(old).(*ptyHandler)
	assert.NotSame(t, session, old)
	old.CloseDoneChan()
	session = (&websocketManager{
		logger: mlog.NewForConfig(nil),
	}).resetSession(old).(*ptyHandler)
	assert.Same(t, session, old)
}

func Test_resetSession1(t *testing.T) {
	t.Parallel()
	old := &ptyHandler{
		container: &biz.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &testRecorder{},
		sessionID: "id",
		conn:      &wsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan biz.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: &sizeStore{},
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		old.sizeStore.Set(100, 100)
	}()
	session := (&websocketManager{
		logger: mlog.NewForConfig(nil),
	}).resetSession(old).(*ptyHandler)

	assert.Equal(t, uint16(100), session.sizeStore.Width())
	assert.Equal(t, uint16(100), session.sizeStore.Height())
}

func Test_resetSession2(t *testing.T) {
	t.Parallel()
	old := &ptyHandler{
		container: &biz.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &testRecorder{},
		sessionID: "id",
		conn:      &wsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan biz.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: &sizeStore{},
	}
	go func() {
		time.Sleep(4 * time.Second)
		old.sizeStore.Set(100, 100)
	}()
	session := (&websocketManager{
		logger: mlog.NewForConfig(nil),
	}).resetSession(old).(*ptyHandler)

	assert.Equal(t, uint16(106), session.sizeStore.Width())
	assert.Equal(t, uint16(25), session.sizeStore.Height())
}

func TestPtyHandler_Next_DoneChan(t *testing.T) {
	p := &ptyHandler{
		recorder: &testRecorder{},
		sizeChan: make(chan biz.TerminalSize, 1),
		doneChan: make(chan struct{}),
	}

	close(p.doneChan)
	assert.Nil(t, p.Next())
}

func TestPtyHandler_Next(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	r := data.NewMockRecorder(m)
	p := &ptyHandler{
		recorder:  r,
		sizeChan:  make(chan biz.TerminalSize, 1),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
	}
	r.EXPECT().Resize(uint16(10), uint16(20)).Times(1)
	p.Resize(biz.TerminalSize{
		Width:  10,
		Height: 20,
	})
	next := p.Next()
	assert.Equal(t, uint16(10), next.Width)
	assert.Equal(t, uint16(20), next.Height)
	p.Resize(biz.TerminalSize{
		Width:  100,
		Height: 200,
	})
	r.EXPECT().Resize(uint16(100), uint16(200)).Times(1)
	next = p.Next()
	assert.Equal(t, uint16(100), next.Width)
	assert.Equal(t, uint16(200), next.Height)
	assert.Equal(t, uint16(100), p.sizeStore.Width())
	assert.Equal(t, uint16(200), p.sizeStore.Height())

	close(p.sizeChan)
	assert.Nil(t, p.Next())
	assert.Equal(t, uint16(100), p.sizeStore.Width())
	assert.Equal(t, uint16(200), p.sizeStore.Height())

	p2 := &ptyHandler{
		sizeChan:  make(chan biz.TerminalSize, 1),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
	}
	close(p2.doneChan)
	p2.Resize(biz.TerminalSize{})
	assert.Len(t, p2.sizeChan, 0)

	p3 := &ptyHandler{
		sizeChan:  make(chan biz.TerminalSize, 1),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
	}
	assert.Nil(t, p3.Resize(biz.TerminalSize{}))
	assert.Equal(t, "sizeChan chan full", p3.Resize(biz.TerminalSize{}).Error())
	close(p3.doneChan)
	assert.Equal(t, "doneChan closed", p3.Resize(biz.TerminalSize{}).Error())
	assert.Len(t, p3.sizeChan, 1)
}

func TestPtyHandler_Read2(t *testing.T) {
	p := &ptyHandler{
		sessionID: "duc",
		recorder:  &testRecorder{},
		sizeChan:  make(chan biz.TerminalSize, 1),
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
		Op:     OpResize,
		Height: 10,
		Width:  20,
	})
	p.Read(b)
	p.Send(context.TODO(), &websocket_pb.TerminalMessage{
		Op:     OpResize,
		Height: 10,
		Width:  20,
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

	p2 := &ptyHandler{
		sessionID: "duc",
		recorder:  &testRecorder{},
		sizeChan:  make(chan biz.TerminalSize, 1),
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

	p3 := &ptyHandler{
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

// TestPtyHandler_SendResize_AfterDoneChanClose 回归：doneChan 关闭后多次 Send/Resize
// 命中 doneChan 分支，shellCh/sizeChan 只应被首个调用方关闭一次，二次调用不得
// "close of closed channel" panic。修复前两次调用即 panic。
func TestPtyHandler_SendResize_AfterDoneChanClose(t *testing.T) {
	p := &ptyHandler{
		sessionID: "duc",
		logger:    mlog.NewForConfig(nil),
		sizeChan:  make(chan biz.TerminalSize, 1),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 1),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
	}
	close(p.doneChan)

	// 多次 Send 命中 doneChan 分支，均返回错误且不 panic（shellCh 只 close 一次）。
	assert.Equal(t, "doneChan closed", p.Send(context.TODO(), nil).Error())
	assert.Equal(t, "doneChan closed", p.Send(context.TODO(), nil).Error())
	assert.Equal(t, "doneChan closed", p.Send(context.TODO(), nil).Error())

	// 多次 Resize 命中 doneChan 分支，均返回错误且不 panic（sizeChan 只 close 一次）。
	assert.Equal(t, "doneChan closed", p.Resize(biz.TerminalSize{}).Error())
	assert.Equal(t, "doneChan closed", p.Resize(biz.TerminalSize{}).Error())
}

func Test_sizeStore_Changed(t *testing.T) {
	t.Parallel()
	ss := sizeStore{
		width:  0,
		height: 0,
		reset:  false,
	}
	assert.True(t, ss.Changed(100, 100))
	ss.Set(100, 100)
	assert.False(t, ss.Changed(100, 100))
	assert.True(t, ss.Changed(100, 0))
	assert.True(t, ss.Changed(0, 100))
}

func TestPtyHandler_Close(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := data.NewMockRecorder(m)
	ps := app.NewMockPubSub(m)
	eventRepo := data.NewMockEventRepo(m)
	p := &ptyHandler{
		sessionID: "duc",
		conn:      &wsConn{pubSub: ps},
		recorder:  recorder,
		sizeChan:  make(chan biz.TerminalSize, 1),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 2),
		doneChan:  make(chan struct{}),
		container: &biz.Container{},
		eventRepo: eventRepo,
		logger:    mlog.NewForConfig(nil),
	}
	eventRepo.EXPECT().FileAuditLogWithDuration(
		types.EventActionType_Shell, "duc", "duc@x.com",
		gomock.Not(nil), 1,
		time.Second)
	recorder.EXPECT().User().Return(&biz.UserInfo{Name: "duc", Email: "duc@x.com"}).AnyTimes()
	recorder.EXPECT().Duration().Return(time.Second)
	recorder.EXPECT().Container().Return(&biz.Container{}).AnyTimes()
	recorder.EXPECT().File().Return(&biz.File{ID: 1})
	recorder.EXPECT().Close().Return(errors.New("x"))
	assert.False(t, p.IsClosed())
	assert.Len(t, p.shellCh, 0)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)

	// 模拟远端 read 循环持续消费 shellCh，使 waitShellDrained 在控制帧送达后尽快返回，
	// 并顺序收集 ETX/EOT 供下方断言（生产上由 remotecommand 的 stdin goroutine 承担）。
	got := make(chan []byte, 2)
	go func() {
		for msg := range p.shellCh {
			got <- msg.Data
		}
	}()

	p.Close(context.TODO(), "aaaa")
	assert.True(t, p.IsClosed())
	// 第二次 Close 幂等：Closeable 已关，直接返回 false，不再入队。
	p.Close(context.TODO(), "aaaa")

	// 控制帧顺序：先 Ctrl-C 中断容器前台进程，再 Ctrl-D 让 shell 读侧收 EOF。
	assert.Equal(t, ETX, <-got)
	assert.Equal(t, END_OF_TRANSMISSION, <-got)

	// 关闭前 shellCh/sizeChan 仍为打开且空（Send/Resize 尚未被调用）。
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

	// doneChan 已被 Close 关闭。
	_, ok := <-p.doneChan
	assert.False(t, ok)

	// 关闭后 Send/Resize 命中 doneChan 分支，分别关闭 shellCh/sizeChan。
	p.Send(context.TODO(), nil)
	p.Resize(biz.TerminalSize{Width: 1, Height: 1})

	// shellCh 已关闭：接收立即得零值（ok=false），select 命中 case 而非 default。
	select {
	case _, more := <-p.shellCh:
		assert.False(t, more, "shellCh 应在 Send 后被关闭")
	default:
		t.Fatal("shellCh 应在 Send 后被关闭")
	}
	// sizeChan 已关闭：同上。
	select {
	case _, more := <-p.sizeChan:
		assert.False(t, more, "sizeChan 应在 Resize 后被关闭")
	default:
		t.Fatal("sizeChan 应在 Resize 后被关闭")
	}
}

func TestPtyHandler_sendControlFrame_canceledCtx(t *testing.T) {
	p := &ptyHandler{
		sessionID: "duc",
		logger:    mlog.NewForConfig(nil),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 2),
		doneChan:  make(chan struct{}),
	}
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	// ctx 已取消：Send 立即返 ctx.Err，控制帧不入队，sendControlFrame 记日志后返回。
	p.sendControlFrame(ctx, ETX)
	assert.Len(t, p.shellCh, 0)
}

func TestPtyHandler_waitShellDrained_timeout(t *testing.T) {
	p := &ptyHandler{
		sessionID: "duc",
		logger:    mlog.NewForConfig(nil),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 2),
	}
	// 队列里有消息且无人消费：排不空，超时后返回（覆盖 af.C 分支，SIGHUP 兜底）。
	p.shellCh <- &websocket_pb.TerminalMessage{Op: OpStdin, Data: ETX}
	start := time.Now()
	p.waitShellDrained(10 * time.Millisecond)
	assert.GreaterOrEqual(t, time.Since(start), 10*time.Millisecond)
	assert.Len(t, p.shellCh, 1)
}

func TestPtyHandler_Write(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := data.NewMockRecorder(m)
	conn := NewMockConn(m)
	pty := &ptyHandler{
		sessionID: "sid",
		conn:      conn,
		logger:    mlog.NewForConfig(nil),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{width: 10, height: 10},
		recorder:  recorder,
		container: &biz.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
	}
	sub := app.NewMockPubSub(m)
	sub.EXPECT().ToSelf(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{
			Id:     "id",
			Uid:    "uid",
			Slug:   "sid",
			Type:   WsHandleExecShellMsg,
			Result: deploy.ResultSuccess,
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
	// Write 只入合批缓冲，flush 后才落录音与 WS 帧（定时器窗口 10ms）。
	pty.flushOutput()
}

// TestPtyHandler_Write_coalesces 验证同一窗口内的多次 PTY 写入合并投递，
// 且数据按写入顺序拼接到一起（流式输出帧数从 ~1000/s 降下来）。
//
// 断言累积后的字节内容而非调用次数：窗口是真实时钟（10ms），调度抖动把
// 一批拆成两帧时内容仍必须完整有序，不能让测试因时序假失败。
func TestPtyHandler_Write_coalesces(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := data.NewMockRecorder(m)
	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	p := &ptyHandler{
		sessionID: "sid",
		conn:      conn,
		logger:    mlog.NewForConfig(nil),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
		recorder:  recorder,
		container: &biz.Container{Namespace: "a", Pod: "b", Container: "c"},
	}
	conn.EXPECT().PubSub().Return(sub).AnyTimes()
	conn.EXPECT().ID().Return("id").AnyTimes()
	conn.EXPECT().UID().Return("uid").AnyTimes()
	var gotRecorder, gotFrame []byte
	recorder.EXPECT().Write(gomock.Any()).AnyTimes().Do(func(b []byte) {
		gotRecorder = append(gotRecorder, b...)
	})
	sub.EXPECT().ToSelf(gomock.Any()).AnyTimes().Do(func(msg app.WebsocketMessage) {
		res := msg.(*websocket_pb.WsHandleShellResponse)
		assert.Equal(t, OpStdout, res.TerminalMessage.Op)
		gotFrame = append(gotFrame, res.TerminalMessage.Data...)
	})

	for _, s := range []string{"a", "b", "c"} {
		n, err := p.Write([]byte(s))
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
	}
	p.flushOutput()

	assert.Equal(t, []byte("abc"), gotRecorder, "录音内容必须完整且有序")
	assert.Equal(t, []byte("abc"), gotFrame, "WS 帧内容必须完整且有序")
}

// TestPtyHandler_Write_flushTimer 验证没有后续写入时，合批窗口到期会自动
// flush，交互式输出的最后一段不会滞留在缓冲区里。
func TestPtyHandler_Write_flushTimer(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := data.NewMockRecorder(m)
	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	p := &ptyHandler{
		sessionID: "sid",
		conn:      conn,
		logger:    mlog.NewForConfig(nil),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
		recorder:  recorder,
		container: &biz.Container{Namespace: "a", Pod: "b", Container: "c"},
	}
	sent := make(chan struct{})
	conn.EXPECT().PubSub().Return(sub)
	conn.EXPECT().ID().Return("id")
	conn.EXPECT().UID().Return("uid")
	recorder.EXPECT().Write([]byte("x")).Times(1)
	sub.EXPECT().ToSelf(gomock.Any()).Times(1).Do(func(app.WebsocketMessage) { close(sent) })

	_, err := p.Write([]byte("x"))
	assert.NoError(t, err)

	select {
	case <-sent:
	case <-time.After(time.Second):
		t.Fatal("合批窗口到期后未 flush")
	}
}

// TestPtyHandler_Write_preservesOrderAndBytes 用不同长度的分块验证合批只改变
// 帧数量，不改变字节内容与顺序（终端输出是字节流，错一个字节就可能是乱码）。
func TestPtyHandler_Write_preservesOrderAndBytes(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := data.NewMockRecorder(m)
	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	p := &ptyHandler{
		sessionID: "sid",
		conn:      conn,
		logger:    mlog.NewForConfig(nil),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
		recorder:  recorder,
		container: &biz.Container{Namespace: "a", Pod: "b", Container: "c"},
	}
	conn.EXPECT().PubSub().Return(sub).AnyTimes()
	conn.EXPECT().ID().Return("id").AnyTimes()
	conn.EXPECT().UID().Return("uid").AnyTimes()

	var want, gotRecorder, gotFrame []byte
	// 累积而非覆盖：窗口到期把一批拆成多帧时，只有拼接后的内容才是完整字节流。
	recorder.EXPECT().Write(gomock.Any()).AnyTimes().Do(func(b []byte) {
		gotRecorder = append(gotRecorder, b...)
	})
	sub.EXPECT().ToSelf(gomock.Any()).AnyTimes().Do(func(msg app.WebsocketMessage) {
		gotFrame = append(gotFrame, msg.(*websocket_pb.WsHandleShellResponse).TerminalMessage.Data...)
	})

	for i := 0; i < 50; i++ {
		chunk := bytes.Repeat([]byte{byte('a' + i%26)}, i%7+1)
		want = append(want, chunk...)
		_, err := p.Write(chunk)
		assert.NoError(t, err)
	}
	p.flushOutput()

	assert.Equal(t, want, gotRecorder, "录音内容必须与写入顺序一致")
	assert.Equal(t, want, gotFrame, "WS 帧内容必须与写入顺序一致")
}

// TestPtyHandler_Write_flushAtLimit 验证积压达到 ptyFlushBytes 时由 Write 内
// 同步 flush（不等定时器），并覆盖临界值两侧：单次写入正好达标、已有
// 上限-1 字节时再写 1 字节达标；达标后不得遗留窗口定时器。
func TestPtyHandler_Write_flushAtLimit(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := data.NewMockRecorder(m)
	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	p := &ptyHandler{
		sessionID: "sid",
		conn:      conn,
		logger:    mlog.NewForConfig(nil),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
		recorder:  recorder,
		container: &biz.Container{Namespace: "a", Pod: "b", Container: "c"},
	}
	conn.EXPECT().PubSub().Return(sub).AnyTimes()
	conn.EXPECT().ID().Return("id").AnyTimes()
	conn.EXPECT().UID().Return("uid").AnyTimes()
	var gotRecorder, gotFrame []byte
	recorder.EXPECT().Write(gomock.Any()).AnyTimes().Do(func(b []byte) {
		gotRecorder = append(gotRecorder, b...)
	})
	sub.EXPECT().ToSelf(gomock.Any()).AnyTimes().Do(func(msg app.WebsocketMessage) {
		gotFrame = append(gotFrame, msg.(*websocket_pb.WsHandleShellResponse).TerminalMessage.Data...)
	})

	// 边界一：单次写入正好等于上限，Write 内立即 flush，不启动定时器。
	atLimit := bytes.Repeat([]byte("x"), ptyFlushBytes)
	_, err := p.Write(atLimit)
	assert.NoError(t, err)
	p.outMu.Lock()
	assert.Nil(t, p.outTimer, "达到上限应立即 flush，不应遗留定时器")
	p.outMu.Unlock()
	assert.Equal(t, atLimit, gotRecorder, "达到上限应立即落录音")
	assert.Equal(t, atLimit, gotFrame, "达到上限应立即推 WS 帧")

	// 边界二：缓冲已有上限-1 字节，再写 1 字节正好达标，同样立即 flush。
	gotRecorder, gotFrame = nil, nil
	p.outMu.Lock()
	p.outBuf = bytes.Repeat([]byte("a"), ptyFlushBytes-1)
	p.outMu.Unlock()
	_, err = p.Write([]byte("b"))
	assert.NoError(t, err)
	p.outMu.Lock()
	assert.Nil(t, p.outTimer, "补足上限后不应遗留定时器")
	p.outMu.Unlock()
	want := append(bytes.Repeat([]byte("a"), ptyFlushBytes-1), 'b')
	assert.Equal(t, want, gotRecorder, "补足上限应立即落录音")
	assert.Equal(t, want, gotFrame, "补足上限应立即推 WS 帧")
}

// TestPtyHandler_flushOutput_discardsAfterDoneChanClose 验证 doneChan 关闭后
// 迟到的 flush 不再投递：Close/CloseDoneChan 都在关通道前先 flush，这里只
// 兜住"Write 已通过 IsClosed 检查、随后会话才关闭"的竞态尾部。
func TestPtyHandler_flushOutput_discardsAfterDoneChanClose(t *testing.T) {
	// 最小依赖构造：若 doneChan 分支没有提前返回，nil 的 recorder/conn 会直接
	// panic，等价于"关闭后绝不再投递"的断言。
	p := &ptyHandler{doneChan: make(chan struct{})}
	p.outBuf = []byte("late")

	close(p.doneChan)
	p.flushOutput()

	assert.Nil(t, p.outBuf, "关闭后的待发缓冲应被丢弃，不再滞留")
}

// TestPtyHandler_CloseDoneChan_flushesPending 验证 shell 回退路径（resetSession
// → CloseDoneChan）也会 flush 待发输出，不会丢掉上一个 shell 的最后一段错误信息。
func TestPtyHandler_CloseDoneChan_flushesPending(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := data.NewMockRecorder(m)
	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	p := &ptyHandler{
		sessionID: "sid",
		conn:      conn,
		logger:    mlog.NewForConfig(nil),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
		recorder:  recorder,
		container: &biz.Container{Namespace: "a", Pod: "b", Container: "c"},
	}
	conn.EXPECT().PubSub().Return(sub)
	conn.EXPECT().ID().Return("id")
	conn.EXPECT().UID().Return("uid")
	recorder.EXPECT().Write([]byte("tail")).Times(1)
	sub.EXPECT().ToSelf(gomock.Any()).Times(1).Do(func(msg app.WebsocketMessage) {
		assert.Equal(t, []byte("tail"), msg.(*websocket_pb.WsHandleShellResponse).TerminalMessage.Data)
	})

	_, err := p.Write([]byte("tail"))
	assert.NoError(t, err)
	assert.True(t, p.CloseDoneChan())

	_, err = p.Write([]byte("x"))
	assert.Error(t, err, "doneChan 关闭后不再接受写入")
}

// TestPtyHandler_Write_streamingFrames 用接近实测的节奏（1KB / 1ms，约等于
// opencode 流式输出的 1000 次/秒）驱动 Write，验证合批把帧数压到窗口数量级。
//
// 只断言上界而不锁死次数：窗口是真实时钟，负载高时 Sleep 被拉长、写入间隔
// 可能超过 10ms，窗口会在两次写入之间到期从而多出几帧；不合批则恒为 100 帧，
// 上界足以区分两种实现。字节总数必须严格相等，合批不允许丢数据。
func TestPtyHandler_Write_streamingFrames(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	recorder := data.NewMockRecorder(m)
	conn := NewMockConn(m)
	sub := app.NewMockPubSub(m)
	p := &ptyHandler{
		sessionID: "sid",
		conn:      conn,
		logger:    mlog.NewForConfig(nil),
		doneChan:  make(chan struct{}),
		sizeStore: &sizeStore{},
		recorder:  recorder,
		container: &biz.Container{Namespace: "a", Pod: "b", Container: "c"},
	}
	var frames, sentBytes int32
	conn.EXPECT().PubSub().Return(sub).AnyTimes()
	conn.EXPECT().ID().Return("id").AnyTimes()
	conn.EXPECT().UID().Return("uid").AnyTimes()
	recorder.EXPECT().Write(gomock.Any()).AnyTimes()
	sub.EXPECT().ToSelf(gomock.Any()).AnyTimes().Do(func(msg app.WebsocketMessage) {
		atomic.AddInt32(&frames, 1)
		res := msg.(*websocket_pb.WsHandleShellResponse)
		atomic.AddInt32(&sentBytes, int32(len(res.TerminalMessage.Data)))
	})

	chunk := bytes.Repeat([]byte("x"), 1024)
	for i := 0; i < 100; i++ {
		_, err := p.Write(chunk)
		assert.NoError(t, err)
		time.Sleep(time.Millisecond)
	}
	p.flushOutput()

	t.Logf("100 次 1KB 写入（1ms 间隔）→ %d 帧", atomic.LoadInt32(&frames))
	assert.Less(t, int(atomic.LoadInt32(&frames)), 40, "100 次 1KB 写入应合批到几十帧")
	assert.Equal(t, int32(100*1024), atomic.LoadInt32(&sentBytes), "合批不能丢数据")
}

func TestPtyHandler_Write3(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ps := app.NewMockPubSub(m)
	r := data.NewMockRecorder(m)
	p := &ptyHandler{
		sizeChan: make(chan biz.TerminalSize, 1),
		sizeStore: &sizeStore{
			width:  106,
			height: 25,
			reset:  true,
		},
		sessionID: "duc",
		container: &biz.Container{},
		logger:    mlog.NewForConfig(nil),
		conn:      &wsConn{pubSub: ps},
		recorder:  r,
		doneChan:  make(chan struct{}),
	}
	r.EXPECT().Write([]byte("aaa")).Times(1)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	n, err := p.Write([]byte("aaa"))
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
	p.flushOutput()
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

func TestPtyHandler_Write_with_chan_full(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ps := app.NewMockPubSub(m)
	r := data.NewMockRecorder(m)
	p := &ptyHandler{
		sizeChan: make(chan biz.TerminalSize),
		sizeStore: &sizeStore{
			width:  106,
			height: 25,
			reset:  true,
		},
		container: &biz.Container{},
		sessionID: "duc",
		logger:    mlog.NewForConfig(nil),
		conn:      &wsConn{pubSub: ps},
		recorder:  r,
		doneChan:  make(chan struct{}),
	}
	r.EXPECT().Write([]byte("aaa")).Times(1)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	// 会走 default，没有 select 会卡住
	n, err := p.Write([]byte("aaa"))
	assert.Nil(t, err)
	assert.Equal(t, 3, n)
	p.flushOutput()
}

func TestStartShell_WithValidSessionID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	fileRepo := data.NewMockFileRepo(m)
	fileRepo.EXPECT().NewRecorder(gomock.Any(), gomock.Any())
	conn.EXPECT().GetUser().Return(&biz.UserInfo{})
	ws := &websocketManager{
		logger:   mlog.NewForConfig(nil),
		fileRepo: fileRepo,
	}

	input := &websocket_pb.WsHandleExecShellInput{
		Container: &websocket_pb.Container{
			Namespace: "namespace",
			Pod:       "pod",
			Container: "container",
		},
		SessionId: "abc123",
	}

	conn.EXPECT().SetPtyHandler(input.SessionId, gomock.Any())
	conn.EXPECT().GetPtyHandler(input.SessionId).Return(nil, false)

	sessionID, err := ws.StartShell(context.TODO(), input, conn)
	time.Sleep(1 * time.Second)
	assert.NoError(t, err)
	assert.Equal(t, input.SessionId, sessionID)
}

func TestStartShell_EmptySessionID(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	conn := NewMockConn(m)
	ws := &websocketManager{
		logger: mlog.NewForConfig(nil),
	}

	input := &websocket_pb.WsHandleExecShellInput{
		Container: &websocket_pb.Container{
			Namespace: "namespace",
			Pod:       "pod",
			Container: "container",
		},
		SessionId: "",
	}

	_, err := ws.StartShell(context.TODO(), input, conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session sessionID")
}

func Test_ptyHandler_Send(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	assert.Equal(t, context.Canceled, (&ptyHandler{}).Send(ctx, nil))
}

// TestWebsocketManager_runTerminal_shellRetry 覆盖 shell 启动失败后的重试分支：
// bash 失败 → resetSession 重建会话 → 重试 sh 成功。resetSession 对 session 做
// *ptyHandler 类型断言，故用真实 handler（带非零 sizeStore 跳过轮询），recorder 用 mock。
func TestWebsocketManager_runTerminal_shellRetry(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	sRepo := data.NewMockK8sRepo(m)
	sRepo.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("bash not found"))
	sRepo.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	conn := NewMockConn(m)
	ws := &websocketManager{k8sRepo: sRepo, logger: mlog.NewForConfig(nil)}

	recorder := data.NewMockRecorder(m)
	recorder.EXPECT().SetShell(gomock.Any()).AnyTimes()

	old := &ptyHandler{
		container: &biz.Container{},
		recorder:  recorder,
		sessionID: "sid",
		conn:      &wsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan biz.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: &sizeStore{width: 100, height: 25},
	}
	conn.EXPECT().GetPtyHandler("sid").Return(old, true)
	conn.EXPECT().SetPtyHandler("sid", gomock.Any())
	conn.EXPECT().ClosePty(gomock.Any(), "sid", uint32(1), "Process exited")

	ws.runTerminal(context.TODO(), conn, &biz.Container{}, "xsh", "sid")
}
