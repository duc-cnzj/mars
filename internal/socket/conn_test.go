package socket

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/counter"
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

//func TestWsConn_WriteMessage(t *testing.T) {
//	err := (&WsConn{}).WriteMessage(websocket.TextMessage, []byte("test message"))
//	assert.Nil(t, err)
//}
//
//func TestWsConn_ReadMessage(t *testing.T) {
//	conn := &WsConn{}
//	messageType, p, err := conn.ReadMessage()
//	assert.Nil(t, err)
//	assert.Equal(t, websocket.TextMessage, messageType)
//	assert.Equal(t, []byte("test message"), p)
//}
//
//func TestWsConn_SetReadDeadline(t *testing.T) {
//	conn := &WsConn{}
//	err := conn.SetReadDeadline(time.Now().Add(10 * time.Second))
//	assert.Nil(t, err)
//}
//
//func TestWsConn_SetWriteDeadline(t *testing.T) {
//	conn := &WsConn{}
//	err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
//	assert.Nil(t, err)
//}
//
//func TestWsConn_Close(t *testing.T) {
//	conn := &WsConn{}
//	err := conn.Close(context.Background())
//	assert.Nil(t, err)
////}

func TestWsConn_AddTask_RunTask_RemoveTask(t *testing.T) {
	conn := &WsConn{taskManager: NewTaskManager(mlog.NewLogger(nil))}
	err := conn.AddTask("task1", func(err error) {})
	assert.Nil(t, err)

	err = conn.RunTask("task1")
	assert.Nil(t, err)

	conn.RemoveTask("task1")
	err = conn.RunTask("task1")
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
