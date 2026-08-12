package websocket

import (
	"context"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/counter"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWsConn_ID(t *testing.T) {
	conn := &wsConn{
		id: "id",
	}
	assert.Equal(t, "id", conn.ID())
}

func TestWsConn_UID(t *testing.T) {
	conn := &wsConn{uid: "uid"}
	assert.Equal(t, "uid", conn.UID())
}

func TestWsConn_SetUser_GetUser(t *testing.T) {
	conn := &wsConn{}
	userInfo := &biz.UserInfo{Name: "testUser"}
	conn.SetUser(userInfo)
	assert.Equal(t, userInfo, conn.GetUser())
}

func TestWsConn_AddTask_RunTask_RemoveTask(t *testing.T) {
	conn := &wsConn{taskManager: NewTaskManager(mlog.NewForConfig(nil))}
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
	pl := app.NewMockPluginManager(m)
	ws := app.NewMockWsSender(m)
	ws.EXPECT().New("uid", "id")
	pl.EXPECT().Ws().Return(ws).AnyTimes()
	(&websocketManager{counter: c, pluginManager: pl}).newWsConn("uid", "id", nil, nil, nil)
	assert.Equal(t, 1, c.Count())
}

func TestWsConn_GetPtyHandler(t *testing.T) {
	_, b := (&wsConn{
		sessions: NewSessionMap(mlog.NewForConfig(nil)),
	}).GetPtyHandler("sessionID")
	assert.False(t, b)
}

func TestWsConn_SetPtyHandler(t *testing.T) {
	w := &wsConn{
		sessions: NewSessionMap(mlog.NewForConfig(nil)),
	}
	w.SetPtyHandler("sessionID", &testPtyHandler{})
	h, b := w.GetPtyHandler("sessionID")
	assert.True(t, b)
	assert.NotNil(t, h)
}

func TestWsConn_ClosePty(t *testing.T) {
	w := &wsConn{
		sessions: NewSessionMap(mlog.NewForConfig(nil)),
	}
	w.SetPtyHandler("sessionID", &testPtyHandler{})
	w.ClosePty(context.TODO(), "sessionID", uint32(2), "")
	_, ok := w.GetPtyHandler("sessionID")
	assert.False(t, ok, "ClosePty 后会话应从注册表移除")
}

func TestWsConn_CloseAndClean(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	ws := NewMockGorillaWs(m)
	sub := app.NewMockPubSub(m)
	tm := NewMockTaskManager(m)
	mapper := NewMockSessionMapper(m)
	w := &wsConn{
		GorillaWs:   ws,
		pubSub:      sub,
		user:        &biz.UserInfo{},
		taskManager: tm,
		sessions:    mapper,
	}

	ws.EXPECT().Close()
	tm.EXPECT().StopAll()
	ctx := context.TODO()
	mapper.EXPECT().CloseAll(ctx)
	sub.EXPECT().Close()
	assert.Nil(t, w.CloseAndClean(ctx))
}

// TestWsConn_CloseAndClean_Idempotent 是 P1 幂等回归测试：Serve 的 defer 与 write 循环
// 的 defer 会对同一连接各调一次 CloseAndClean，若不加 sync.Once 保护，
// prometheus 连接数 gauge 会被每条连接双递减、长期趋负。
// 本测试把 gauge 打到已知值 10，连续调用两次 CloseAndClean，
// 断言：四个 mock 各只被调一次（gomock 默认 Times(1)），gauge 只递减到 9（而非 8）。
func TestWsConn_CloseAndClean_Idempotent(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	ws := NewMockGorillaWs(m)
	sub := app.NewMockPubSub(m)
	tm := NewMockTaskManager(m)
	mapper := NewMockSessionMapper(m)
	const username = "idempotent-user"
	w := &wsConn{
		GorillaWs:   ws,
		pubSub:      sub,
		user:        &biz.UserInfo{Name: username},
		taskManager: tm,
		sessions:    mapper,
	}

	// gomock 默认 Times(1)：CloseAndClean 被调两次，四个 mock 仍各只能被调一次。
	ws.EXPECT().Close()
	tm.EXPECT().StopAll()
	ctx := context.TODO()
	mapper.EXPECT().CloseAll(ctx)
	sub.EXPECT().Close()

	gauge := metrics.WebsocketConnectionsCount.WithLabelValues(username)
	gauge.Set(10)
	assert.Equal(t, float64(10), testutil.ToFloat64(gauge))

	assert.Nil(t, w.CloseAndClean(ctx))
	assert.Nil(t, w.CloseAndClean(ctx))

	// 只递减一次：10 → 9。若 sync.Once 失效会变成 8。
	assert.Equal(t, float64(9), testutil.ToFloat64(gauge))
}
