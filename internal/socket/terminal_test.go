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
	_, ok := <-p.shellCh
	assert.False(t, ok)
	_, ok = <-p.sizeChan
	assert.False(t, ok)
	_, ok = <-p.doneChan
	assert.False(t, ok)
}

func TestMyPtyHandler_Next(t *testing.T) {
	p := &MyPtyHandler{
		sizeStore: &sizeStore{},
		recorder:  &Recorder{},
		sizeChan:  make(chan remotecommand.TerminalSize, 1),
		doneChan:  make(chan struct{}),
	}
	p.sizeChan <- remotecommand.TerminalSize{
		Width:  10,
		Height: 20,
	}
	next := p.Next()
	assert.Equal(t, uint16(10), next.Width)
	assert.Equal(t, uint16(20), next.Height)

	close(p.sizeChan)
	assert.Nil(t, p.Next())
	assert.Equal(t, uint16(10), p.sizeStore.Cols())
	assert.Equal(t, uint16(20), p.sizeStore.Rows())
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
	p.shellCh <- &websocket_pb.TerminalMessage{
		Op:   OpStdin,
		Data: "hello duc",
	}
	n, _ := p.Read(b)
	assert.Equal(t, "hello duc", string(b[0:n]))
	p.shellCh <- &websocket_pb.TerminalMessage{
		Op:   OpResize,
		Rows: 10,
		Cols: 20,
	}
	n, _ = p.Read(b)
	assert.Equal(t, 0, n)
	assert.Len(t, p.sizeChan, 1)
	p.shellCh <- &websocket_pb.TerminalMessage{
		Op: "xxxx",
	}
	n, err := p.Read(b)
	assert.Greater(t, n, 0)
	assert.Error(t, err)
	close(p.shellCh)
	_, err = p.Read(b)
	assert.Equal(t, "[Websocket]: duc channel closed", err.Error())
	close(p.doneChan)
	n, err = p.Read(b)
	assert.Equal(t, "[Websocket]: duc doneChan closed", err.Error())
	assert.Greater(t, n, 0)
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

func TestMyPtyHandler_TerminalMessageChan(t *testing.T) {
	p := &MyPtyHandler{
		shellCh: make(chan *websocket_pb.TerminalMessage),
	}
	assert.Equal(t, p.shellCh, p.TerminalMessageChan())
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
		sizeStore: &sizeStore{
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
	p.closeable.Close()
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
