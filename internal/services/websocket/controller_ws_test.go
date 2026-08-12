package websocket

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/counter"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

var errBoom = errors.New("boom")

// fakeWriteCloser 是 write 循环里 NextWriter 返回的 io.WriteCloser 桩。
type fakeWriteCloser struct {
	writeErr error
	closeErr error
	written  []byte
	closed   bool
}

func (w *fakeWriteCloser) Write(p []byte) (int, error) {
	w.written = append(w.written, p...)
	return len(p), w.writeErr
}

func (w *fakeWriteCloser) Close() error {
	w.closed = true
	return w.closeErr
}

func newWsConnForLoop(m *gomock.Controller, mockWs *MockGorillaWs, sub app.PubSub) *wsConn {
	return &wsConn{
		GorillaWs:   mockWs,
		pubSub:      sub,
		taskManager: NewTaskManager(mlog.NewForConfig(nil)),
		sessions:    NewSessionMap(mlog.NewForConfig(nil)),
	}
}

func TestWebsocketManager_read_readError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	mockWs := NewMockGorillaWs(m)
	mockWs.EXPECT().SetReadLimit(maxMessageSize)
	mockWs.EXPECT().SetReadDeadline(gomock.Any()).AnyTimes()
	mockWs.EXPECT().SetPongHandler(gomock.Any())
	mockWs.EXPECT().ReadMessage().Return(0, nil, errBoom)

	wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
	err := wm.read(context.TODO(), &wsConn{GorillaWs: mockWs})
	assert.ErrorIs(t, err, errBoom)
}

func TestWebsocketManager_read_unmarshalError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	mockWs := NewMockGorillaWs(m)
	mockWs.EXPECT().SetReadLimit(maxMessageSize)
	mockWs.EXPECT().SetReadDeadline(gomock.Any()).AnyTimes()
	mockWs.EXPECT().SetPongHandler(gomock.Any())
	// 非 proto 字节 → unmarshal 失败 → SendEndError（WsInternalError）
	mockWs.EXPECT().ReadMessage().Return(1, []byte("not-a-proto"), nil)
	mockWs.EXPECT().ReadMessage().Return(0, nil, errBoom)

	sub := app.NewMockPubSub(m)
	sub.EXPECT().ToSelf(gomock.Any())

	wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
	err := wm.read(context.TODO(), &wsConn{GorillaWs: mockWs, pubSub: sub})
	assert.ErrorIs(t, err, errBoom)
}

func TestWebsocketManager_read_validMessage(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	// wire 消息自带 Type 字段（field 1），同一字节流可同时被 WsRequestMetadata 反序列化。
	msg, err := proto.Marshal(&websocket_pb.WsRequestMetadata{Type: websocket_pb.Type_HandleAuthorize})
	assert.NoError(t, err)

	var pongHandler func(string) error
	mockWs := NewMockGorillaWs(m)
	mockWs.EXPECT().SetReadLimit(maxMessageSize)
	mockWs.EXPECT().SetReadDeadline(gomock.Any()).AnyTimes()
	mockWs.EXPECT().SetPongHandler(gomock.Any()).Do(func(h func(string) error) { pongHandler = h })
	mockWs.EXPECT().ReadMessage().Return(1, msg, nil)     // 有效消息 → go dispatchEvent
	mockWs.EXPECT().ReadMessage().Return(0, nil, errBoom) // 退出循环

	wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
	called := make(chan struct{}, 1)
	wm.handlers = map[websocket_pb.Type]HandleRequestFunc{
		websocket_pb.Type_HandleAuthorize: func(ctx context.Context, c Conn, ty websocket_pb.Type, message []byte) {
			called <- struct{}{}
		},
	}

	conn := &wsConn{GorillaWs: mockWs}
	err = wm.read(context.TODO(), conn)
	assert.Error(t, err)
	<-called // 等异步 dispatchEvent 完成

	// 触发 pong handler 内部再设读超时
	assert.NoError(t, pongHandler(""))
}

func TestWebsocketManager_write(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	ch := make(chan []byte, 1)
	sub := app.NewMockPubSub(m)
	sub.EXPECT().ID().Return("id").AnyTimes()
	sub.EXPECT().Uid().Return("uid").AnyTimes()
	sub.EXPECT().Subscribe().Return(ch)
	sub.EXPECT().Close().AnyTimes()

	fake := &fakeWriteCloser{}
	mockWs := NewMockGorillaWs(m)
	mockWs.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()
	mockWs.EXPECT().NextWriter(websocket.BinaryMessage).Return(fake, nil)
	// channel 关闭 → 发 CloseMessage 退出
	mockWs.EXPECT().WriteMessage(websocket.CloseMessage, []byte{}).Return(nil)
	mockWs.EXPECT().Close().AnyTimes()

	wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
	conn := newWsConnForLoop(m, mockWs, sub)

	done := make(chan error, 1)
	go func() { done <- wm.write(context.TODO(), conn) }()

	ch <- []byte("hello") // 触发 NextWriter 写帧
	close(ch)             // 触发 !ok → CloseMessage
	assert.NoError(t, <-done)
	assert.Equal(t, []byte("hello"), fake.written)
	assert.True(t, fake.closed)
}

func TestWebsocketManager_write_ctxDone(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	sub := app.NewMockPubSub(m)
	sub.EXPECT().ID().Return("id").AnyTimes()
	sub.EXPECT().Uid().Return("uid").AnyTimes()
	sub.EXPECT().Subscribe().Return(make(chan []byte))
	sub.EXPECT().Close().AnyTimes()

	mockWs := NewMockGorillaWs(m)
	mockWs.EXPECT().Close().AnyTimes()

	wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
	conn := newWsConnForLoop(m, mockWs, sub)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- wm.write(ctx, conn) }()
	cancel()
	assert.ErrorIs(t, <-done, context.Canceled)
}

func TestWebsocketManager_write_nextWriterError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	ch := make(chan []byte, 1)
	sub := app.NewMockPubSub(m)
	sub.EXPECT().ID().Return("id").AnyTimes()
	sub.EXPECT().Uid().Return("uid").AnyTimes()
	sub.EXPECT().Subscribe().Return(ch)
	sub.EXPECT().Close().AnyTimes()

	mockWs := NewMockGorillaWs(m)
	mockWs.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()
	mockWs.EXPECT().NextWriter(websocket.BinaryMessage).Return(nil, errBoom)
	mockWs.EXPECT().Close().AnyTimes()

	wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
	conn := newWsConnForLoop(m, mockWs, sub)

	done := make(chan error, 1)
	go func() { done <- wm.write(context.TODO(), conn) }()
	ch <- []byte("x")
	assert.ErrorIs(t, <-done, errBoom)
}

func TestWebsocketManager_write_closeWriterError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	ch := make(chan []byte, 1)
	sub := app.NewMockPubSub(m)
	sub.EXPECT().ID().Return("id").AnyTimes()
	sub.EXPECT().Uid().Return("uid").AnyTimes()
	sub.EXPECT().Subscribe().Return(ch)
	sub.EXPECT().Close().AnyTimes()

	fake := &fakeWriteCloser{closeErr: errBoom}
	mockWs := NewMockGorillaWs(m)
	mockWs.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()
	mockWs.EXPECT().NextWriter(websocket.BinaryMessage).Return(fake, nil)
	mockWs.EXPECT().Close().AnyTimes()

	wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
	conn := newWsConnForLoop(m, mockWs, sub)

	done := make(chan error, 1)
	go func() { done <- wm.write(context.TODO(), conn) }()
	ch <- []byte("x")
	assert.ErrorIs(t, <-done, errBoom)
}

func TestWebsocketManager_write_ping(t *testing.T) {
	old := pingPeriod
	pingPeriod = time.Millisecond
	defer func() { pingPeriod = old }()

	m := gomock.NewController(t)
	defer m.Finish()

	pingCh := make(chan []byte)
	sub := app.NewMockPubSub(m)
	sub.EXPECT().ID().Return("id").AnyTimes()
	sub.EXPECT().Uid().Return("uid").AnyTimes()
	sub.EXPECT().Subscribe().Return(pingCh).AnyTimes()
	sub.EXPECT().Close().AnyTimes()

	mockWs := NewMockGorillaWs(m)
	mockWs.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()
	mockWs.EXPECT().WriteMessage(websocket.PingMessage, nil).Return(nil).AnyTimes()
	mockWs.EXPECT().WriteMessage(websocket.CloseMessage, []byte{}).Return(nil).Times(1)
	mockWs.EXPECT().Close().AnyTimes()

	wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
	conn := newWsConnForLoop(m, mockWs, sub)

	done := make(chan error, 1)
	go func() { done <- wm.write(context.TODO(), conn) }()

	time.Sleep(10 * time.Millisecond) // 等至少一个 ping 周期
	close(pingCh)                     // 触发 !ok → CloseMessage 退出
	assert.NoError(t, <-done)
}

func TestWebsocketManager_write_pingWriteError(t *testing.T) {
	old := pingPeriod
	pingPeriod = time.Millisecond
	defer func() { pingPeriod = old }()

	m := gomock.NewController(t)
	defer m.Finish()

	pingCh := make(chan []byte)
	sub := app.NewMockPubSub(m)
	sub.EXPECT().ID().Return("id").AnyTimes()
	sub.EXPECT().Uid().Return("uid").AnyTimes()
	sub.EXPECT().Subscribe().Return(pingCh).AnyTimes()
	sub.EXPECT().Close().AnyTimes()

	mockWs := NewMockGorillaWs(m)
	mockWs.EXPECT().SetWriteDeadline(gomock.Any()).AnyTimes()
	mockWs.EXPECT().WriteMessage(websocket.PingMessage, nil).Return(errBoom).AnyTimes()
	mockWs.EXPECT().Close().AnyTimes()

	wm := &websocketManager{logger: mlog.NewForConfig(nil), timer: timer.NewReal()}
	conn := newWsConnForLoop(m, mockWs, sub)

	done := make(chan error, 1)
	go func() { done <- wm.write(context.TODO(), conn) }()
	assert.ErrorIs(t, <-done, errBoom)
}

// TestWebsocketManager_Serve 走真实 gorilla 握手：httptest 服务端 + 客户端 ws 连接，
// 覆盖 Serve/read/write/dispatchEvent 主链路与 WsSetUid 握手帧、counter 增减。
func TestWebsocketManager_Serve(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	ch := make(chan []byte, 100)
	var closeOnce sync.Once
	// stop 与 mock Close 共用同一 Once：避免测试关 ch 后 mock.Close 再关导致 double-close。
	stop := func() { closeOnce.Do(func() { close(ch) }) }
	sub := app.NewMockPubSub(m)
	sub.EXPECT().Subscribe().Return(ch).AnyTimes()
	sub.EXPECT().ToSelf(gomock.Any()).DoAndReturn(func(msg app.WebsocketMessage) error {
		b, err := proto.Marshal(msg.(proto.Message))
		if err != nil {
			return err
		}
		ch <- b
		return nil
	}).AnyTimes()
	sub.EXPECT().ID().Return("cid").AnyTimes()
	sub.EXPECT().Uid().Return("cuid").AnyTimes()
	sub.EXPECT().Run(gomock.Any()).Return(nil).AnyTimes()
	sub.EXPECT().Close().Do(stop).AnyTimes()
	sub.EXPECT().Join(int64(2)).Return(nil).AnyTimes()

	wsSender := app.NewMockWsSender(m)
	wsSender.EXPECT().New(gomock.Any(), gomock.Any()).Return(sub).AnyTimes()
	pl := app.NewMockPluginManager(m)
	pl.EXPECT().Ws().Return(wsSender).AnyTimes()

	authMock := biz.NewMockAuthBiz(m)
	authMock.EXPECT().VerifyToken(gomock.Any(), "token").Return(&biz.UserInfo{Name: "user"}, nil).AnyTimes()

	// JoinRoom 前经 RequireProjectAccess（Show + 所属命名空间校验）：公开命名空间放行。
	projBiz := biz.NewMockProjectBiz(m)
	projBiz.EXPECT().Show(gomock.Any(), 2).Return(&biz.Project{ID: 2, NamespaceID: 1}, nil).AnyTimes()
	nsBiz := biz.NewMockNamespaceBiz(m)
	nsBiz.EXPECT().Show(gomock.Any(), 1).Return(&biz.Namespace{Name: "ns", Private: false}, nil).AnyTimes()

	wm := &websocketManager{
		timer:         timer.NewReal(),
		logger:        mlog.NewForConfig(nil),
		pluginManager: pl,
		authBiz:       authMock,
		counter:       counter.NewCounter(),
		accessBiz:     biz.NewAccessBiz(nsBiz, projBiz),
	}
	wm.handlers = map[websocket_pb.Type]HandleRequestFunc{
		WsAuthorize:     wm.HandleAuthorize,
		ProjectPodEvent: wm.HandleJoinRoom,
	}

	server := httptest.NewServer(http.HandlerFunc(wm.Serve))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?uid=test-uid"
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)

	// 1. 握手帧 WsSetUid
	_, data, err := c.ReadMessage()
	assert.NoError(t, err)
	var resp websocket_pb.WsMetadataResponse
	assert.NoError(t, proto.Unmarshal(data, &resp))
	assert.Equal(t, WsSetUid, resp.Metadata.Type)
	assert.Equal(t, "test-uid", resp.Metadata.Message)

	// 2. 认证
	authMsg, _ := proto.Marshal(&websocket_pb.AuthorizeTokenInput{Type: WsAuthorize, Token: "token"})
	assert.NoError(t, c.WriteMessage(websocket.BinaryMessage, authMsg))

	// 3. 加入房间
	joinMsg, _ := proto.Marshal(&websocket_pb.ProjectPodEventJoinInput{Type: ProjectPodEvent, Join: true, ProjectId: 2})
	assert.NoError(t, c.WriteMessage(websocket.BinaryMessage, joinMsg))

	// 4. 断开：客户端关连接 + 关 Subscribe channel 让 write 循环退出
	c.Close()
	stop()
	assert.Eventually(t, func() bool { return wm.counter.Count() == 0 }, 3*time.Second, 10*time.Millisecond)
}

func TestWebsocketManager_Serve_upgradeError(t *testing.T) {
	wm := &websocketManager{logger: mlog.NewForConfig(nil)}
	server := httptest.NewServer(http.HandlerFunc(wm.Serve))
	defer server.Close()

	// 普通 GET（无 upgrade 头）→ upgrader 失败 → 记录日志返回
	resp, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.NoError(t, resp.Body.Close())
}
