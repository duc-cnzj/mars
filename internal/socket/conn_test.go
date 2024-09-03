package socket

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/util/counter"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWsConn_ID(t *testing.T) {
	conn := &WsConn{
		id: "id",
	}
	assert.Equal(t, "id", conn.ID())
}

func TestWsConn_UID(t *testing.T) {
	conn := &WsConn{uid: "uid"}
	assert.Equal(t, "uid", conn.UID())
}

func TestWsConn_SetUser_GetUser(t *testing.T) {
	conn := &WsConn{}
	userInfo := &auth.UserInfo{Name: "testUser"}
	conn.SetUser(userInfo)
	assert.Equal(t, userInfo, conn.GetUser())
}

func TestWsConn_AddTask_RunTask_RemoveTask(t *testing.T) {
	conn := &WsConn{taskManager: NewTaskManager(mlog.NewForConfig(nil))}
	err := conn.AddCancelDeployTask("task1", func(err error) {})
	assert.Nil(t, err)

	err = conn.RunCancelDeployTask("task1")
	assert.Nil(t, err)

	conn.RemoveCancelDeployTask("task1")
	err = conn.RunCancelDeployTask("task1")
	assert.NotNil(t, err)
	assert.Equal(t, "task not found", err.Error())
}

func TestWebsocketManager_newWsConn(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	c := counter.NewCounter()
	pl := application.NewMockPluginManger(m)
	ws := application.NewMockWsSender(m)
	ws.EXPECT().New("uid", "id")
	pl.EXPECT().Ws().Return(ws).AnyTimes()
	(&WebsocketManager{counter: c, pl: pl}).newWsConn("uid", "id", nil, nil, nil)
	assert.Equal(t, 1, c.Count())
}

func TestWsConn_GetPtyHandler(t *testing.T) {
	_, b := (&WsConn{
		sm: NewSessionMap(mlog.NewForConfig(nil)),
	}).GetPtyHandler("sessionID")
	assert.False(t, b)
}

func TestWsConn_SetPtyHandler(t *testing.T) {
	w := &WsConn{
		sm: NewSessionMap(mlog.NewForConfig(nil)),
	}
	w.SetPtyHandler("sessionID", &testPtyHandler{})
	h, b := w.GetPtyHandler("sessionID")
	assert.True(t, b)
	assert.NotNil(t, h)
}

func TestWsConn_ClosePty(t *testing.T) {
	w := &WsConn{
		sm: NewSessionMap(mlog.NewForConfig(nil)),
	}
	w.SetPtyHandler("sessionID", &testPtyHandler{})
	w.ClosePty(context.TODO(), "sessionID", uint32(2), "")
}

func TestWsConn_CloseAndClean(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	ws := NewMockGorillaWs(m)
	sub := application.NewMockPubSub(m)
	tm := NewMockTaskManager(m)
	mapper := NewMockSessionMapper(m)
	w := &WsConn{
		GorillaWs:   ws,
		pubSub:      sub,
		user:        &auth.UserInfo{},
		taskManager: tm,
		sm:          mapper,
	}

	ws.EXPECT().Close()
	tm.EXPECT().StopAll()
	ctx := context.TODO()
	mapper.EXPECT().CloseAll(ctx)
	sub.EXPECT().Close()
	assert.Nil(t, w.CloseAndClean(ctx))
}
