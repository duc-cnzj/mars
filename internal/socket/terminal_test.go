package socket

import (
	"errors"
	"sync"
	"testing"
	"time"

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

func TestMyPtyHandler_Close(t1 *testing.T) {
}

func TestMyPtyHandler_Next(t1 *testing.T) {}

func TestMyPtyHandler_Read(t1 *testing.T) {}

func TestMyPtyHandler_Recorder(t1 *testing.T) {}

func TestMyPtyHandler_SetShell(t1 *testing.T) {}

func TestMyPtyHandler_TerminalMessageChan(t1 *testing.T) {}

func TestMyPtyHandler_Toast(t1 *testing.T) {}

func TestMyPtyHandler_Write(t1 *testing.T) {}

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
