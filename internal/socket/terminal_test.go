package socket

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/internal/testutil"

	"k8s.io/client-go/tools/remotecommand"

	websocket_pb "github.com/duc-cnzj/mars-client/v4/websocket"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGenMyPtyHandlerId(t *testing.T) {
	t.Parallel()
	id := GenMyPtyHandlerId()
	assert.Len(t, id, 36)
}

func TestHandleExecShell(t *testing.T) {}

func TestMyPtyHandler_Close(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Uploader().Return(nil).AnyTimes()
	ps := mock.NewMockPubSub(m)
	p := &MyPtyHandler{
		id:       "duc",
		conn:     &WsConn{pubSub: ps},
		recorder: &Recorder{},
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		shellCh:  make(chan *websocket_pb.TerminalMessage, 2),
		doneChan: make(chan struct{}),
	}
	assert.False(t, p.IsClosed())
	assert.Len(t, p.shellCh, 0)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
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
	p := &MyPtyHandler{
		recorder: r,
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}),
	}
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

	p2 := &MyPtyHandler{
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}),
	}
	close(p2.doneChan)
	p2.Resize(remotecommand.TerminalSize{})
	assert.Len(t, p2.sizeChan, 0)

	p3 := &MyPtyHandler{
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
	p := &MyPtyHandler{
		recorder: &Recorder{},
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}),
	}

	close(p.doneChan)
	assert.Nil(t, p.Next())
}

func TestMyPtyHandler_Read(t *testing.T) {
	p := &MyPtyHandler{
		id:       "duc",
		recorder: &Recorder{},
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		shellCh:  make(chan *websocket_pb.TerminalMessage, 1),
		doneChan: make(chan struct{}),
	}
	b := make([]byte, 1024)
	p.Send(&websocket_pb.TerminalMessage{
		Op:   OpStdin,
		Data: "hello duc",
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

	p2 := &MyPtyHandler{
		id:       "duc",
		recorder: &Recorder{},
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		shellCh:  make(chan *websocket_pb.TerminalMessage, 1),
		doneChan: make(chan struct{}),
	}
	close(p2.doneChan)
	bv := make([]byte, 100)
	i, err := p2.Read(bv)
	assert.Error(t, err)
	assert.Equal(t, END_OF_TRANSMISSION, string(bv[:i]))

	p3 := &MyPtyHandler{
		id:       "duc",
		recorder: &Recorder{},
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
	p := &MyPtyHandler{
		recorder: &Recorder{},
	}
	assert.Implements(t, (*contracts.RecorderInterface)(nil), p.Recorder())
}

func TestMyPtyHandler_SetShell(t *testing.T) {
	p := &MyPtyHandler{
		recorder: &Recorder{},
	}
	p.SetShell("xxx")
	assert.Equal(t, "xxx", p.recorder.GetShell())
}

func TestMyPtyHandler_Toast(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ps := mock.NewMockPubSub(m)
	p := &MyPtyHandler{
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
	p := &MyPtyHandler{
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
	p := &MyPtyHandler{
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
	sm := SessionMap{
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
	sm := SessionMap{
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
	sm := SessionMap{
		sessLock: sync.RWMutex{},
		Sessions: map[string]contracts.PtyHandler{},
	}
	ch := make(chan *websocket_pb.TerminalMessage, 1)
	sm.Set("a", &MyPtyHandler{
		shellCh: ch,
	})
	assert.Len(t, ch, 0)
	sm.Send(&websocket_pb.TerminalMessage{
		Data:      "aa",
		SessionId: "a",
	})
	assert.Len(t, ch, 1)
	sm.Send(&websocket_pb.TerminalMessage{
		Data:      "aa",
		SessionId: "a",
	})
	assert.Len(t, ch, 1)
}

func TestSessionMap_Set(t *testing.T) {
	t.Parallel()
	sm := SessionMap{
		sessLock: sync.RWMutex{},
		Sessions: map[string]contracts.PtyHandler{},
	}
	p := &MyPtyHandler{}
	sm.Set("a", p)
	get, _ := sm.Get("a")
	assert.Same(t, p, get)
}

func TestWaitForTerminal(t *testing.T) {}

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
	assert.False(t, ss.Changed(100, 100))
	ss.Set(100, 100)
	assert.False(t, ss.Changed(100, 100))
	assert.True(t, ss.Changed(100, 0))
	assert.True(t, ss.Changed(0, 100))
}

func TestMyPtyHandler_Rows(t *testing.T) {
	t.Parallel()
	assert.Equal(t, uint16(0), (&MyPtyHandler{}).Rows())
	assert.Equal(t, uint16(100), (&MyPtyHandler{sizeStore: sizeStore{rows: 100}}).Rows())
}

func TestMyPtyHandler_Cols(t *testing.T) {
	t.Parallel()
	assert.Equal(t, uint16(0), (&MyPtyHandler{}).Cols())
	assert.Equal(t, uint16(100), (&MyPtyHandler{sizeStore: sizeStore{cols: 100}}).Cols())
}

func Test_resetSession(t *testing.T) {
	t.Parallel()
	old := &MyPtyHandler{
		container: contracts.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &Recorder{},
		id:        "id",
		conn:      &WsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan remotecommand.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: sizeStore{cols: 10, rows: 10},
	}
	session := resetSession(old).(*MyPtyHandler)

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
	old := &MyPtyHandler{
		container: contracts.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder:  &Recorder{},
		id:        "id",
		conn:      &WsConn{},
		doneChan:  make(chan struct{}, 1),
		sizeChan:  make(chan remotecommand.TerminalSize, 10),
		shellCh:   make(chan *websocket_pb.TerminalMessage, 10),
		sizeStore: sizeStore{cols: 10, rows: 10},
	}
	session := resetSession(old).(*MyPtyHandler)
	assert.NotSame(t, session, old)
	old.CloseDoneChan()
	session = resetSession(old).(*MyPtyHandler)
	assert.Same(t, session, old)
}

func Test_resetSession1(t *testing.T) {
	t.Parallel()
	old := &MyPtyHandler{
		container: contracts.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder: &Recorder{},
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
	session := resetSession(old).(*MyPtyHandler)

	assert.Equal(t, uint16(100), session.sizeStore.Cols())
	assert.Equal(t, uint16(100), session.sizeStore.Rows())
}

func Test_resetSession2(t *testing.T) {
	t.Parallel()
	old := &MyPtyHandler{
		container: contracts.Container{
			Namespace: "a",
			Pod:       "b",
			Container: "c",
		},
		recorder: &Recorder{},
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
	session := resetSession(old).(*MyPtyHandler)

	assert.Equal(t, uint16(106), session.sizeStore.Cols())
	assert.Equal(t, uint16(25), session.sizeStore.Rows())
}

func TestMyPtyHandler_ResetTerminalRowCol(t *testing.T) {
	pty := &MyPtyHandler{}
	pty.ResetTerminalRowCol(true)
	assert.True(t, pty.sizeStore.TerminalRowColNeedReset())
	pty.ResetTerminalRowCol(false)
	assert.False(t, pty.sizeStore.TerminalRowColNeedReset())
}

func TestMyPtyHandler_ClosePreviousChannels(t *testing.T) {
	pty := &MyPtyHandler{
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
