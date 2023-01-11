package socket

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"k8s.io/client-go/kubernetes/fake"
	fake2 "k8s.io/metrics/pkg/client/clientset/versioned/fake"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars-client/v4/websocket"
	auth2 "github.com/duc-cnzj/mars/internal/auth"
	"github.com/duc-cnzj/mars/internal/cache_lock"
	"github.com/duc-cnzj/mars/internal/config"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mock"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/testutil"
	"github.com/duc-cnzj/mars/internal/utils"
)

func TestHandleWsAuthorize(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	privateKey, _ := x509.MarshalPKCS8PrivateKey(key)
	bf := bytes.Buffer{}
	pem.Encode(&bf, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKey})
	authSvc := auth2.NewJwtAuth(key, key.Public().(*rsa.PublicKey))
	app := testutil.MockApp(m)
	app.EXPECT().Auth().Return(authSvc).AnyTimes()
	sign, _ := authSvc.Sign(contracts.UserInfo{
		Name: "duc",
	})
	marshal, _ := proto.Marshal(&websocket.AuthorizeTokenInput{Token: sign.Token})
	pubsub := mock.NewMockPubSub(m)
	pubsub.EXPECT().ToSelf(gomock.Any()).Times(1)
	conn := &WsConn{
		pubSub: pubsub,
	}
	HandleWsAuthorize(conn, websocket.Type_HandleAuthorize, []byte("1:"))
	HandleWsAuthorize(conn, websocket.Type_HandleAuthorize, marshal)
	assert.Equal(t, "duc", conn.GetUser().Name)
}

func TestHandleWsCancel(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	pubsub := mock.NewMockPubSub(m)
	pubsub.EXPECT().ToSelf(gomock.Any()).Times(1)

	app := testutil.MockApp(m)
	db, fn := testutil.SetGormDB(m, app)
	defer fn()
	ns := &models.Namespace{
		Name: "ns",
	}
	db.Create(ns)
	marshal, _ := proto.Marshal(&websocket.CancelInput{
		Type:        websocket.Type_CancelProject,
		NamespaceId: int64(ns.ID),
		Name:        "app",
	})
	cs := mock.NewMockCancelSignaler(m)
	slug := utils.GetSlugName(ns.ID, "app")
	cs.EXPECT().Has(slug).Return(true).Times(1)
	cs.EXPECT().Cancel(slug).Times(1)
	testutil.AssertAuditLogFired(m, app)
	conn := &WsConn{pubSub: pubsub, cancelSignaler: cs}
	HandleWsCancel(conn, websocket.Type_CancelProject, []byte("1:"))
	HandleWsCancel(conn, websocket.Type_CancelProject, marshal)
}

func TestHandleWsHandleCloseShell(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sm := mock.NewMockSessionMapper(m)
	pubsub := mock.NewMockPubSub(m)
	pubsub.EXPECT().ToSelf(gomock.Any()).Times(1)
	conn := &WsConn{
		pubSub:           pubsub,
		terminalSessions: sm,
	}
	marshal, _ := proto.Marshal(&websocket.TerminalMessageInput{
		Type: websocket.Type_HandleCloseShell,
		Message: &websocket.TerminalMessage{
			SessionId: "sid",
		},
	})
	sm.EXPECT().Close("sid", uint32(0), "").Times(1)
	HandleWsHandleCloseShell(conn, websocket.Type_HandleCloseShell, []byte("1234:"))
	HandleWsHandleCloseShell(conn, websocket.Type_HandleCloseShell, marshal)
}

func TestHandleWsHandleExecShell(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	pubsub := mock.NewMockPubSub(m)
	pubsub.EXPECT().ToSelf(&websocket.WsHandleShellResponse{
		Metadata: &websocket.Metadata{
			Id:     "id-x",
			Uid:    "uid-x",
			Type:   WsHandleExecShell,
			Result: ResultSuccess,
		},
		TerminalMessage: &websocket.TerminalMessage{
			SessionId: "id",
		},
		Container: &types.Container{
			Namespace: "ns",
			Pod:       "pod",
			Container: "app",
		},
	}).Times(1)
	conn := &WsConn{
		id:     "id-x",
		uid:    "uid-x",
		pubSub: pubsub,
		NewShellFunc: func(input *websocket.WsHandleExecShellInput, conn *WsConn) (string, error) {
			return "id", nil
		},
	}
	marshal, _ := proto.Marshal(&websocket.WsHandleExecShellInput{
		Type: websocket.Type_HandleExecShell,
		Container: &types.Container{
			Namespace: "ns",
			Pod:       "pod",
			Container: "app",
		},
	})
	HandleWsHandleExecShell(conn, websocket.Type_HandleExecShell, marshal)
	conn.NewShellFunc = func(input *websocket.WsHandleExecShellInput, conn *WsConn) (string, error) {
		return "", errors.New("xxx")
	}
	pubsub.EXPECT().ToSelf(&websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			Slug:    "",
			Type:    websocket.Type_HandleExecShell,
			Result:  ResultError,
			End:     true,
			Uid:     "uid-x",
			Id:      "id-x",
			Message: "xxx",
		},
	}).Times(1)
	HandleWsHandleExecShell(conn, websocket.Type_HandleExecShell, marshal)
	pubsub.EXPECT().ToSelf(gomock.Any()).Times(1)
	HandleWsHandleExecShell(conn, websocket.Type_HandleExecShell, []byte("1234:"))
}

type protoMatcher struct {
	wants proto.Message
}

func (m *protoMatcher) Matches(x any) bool {
	return proto.Equal(x.(proto.Message), m.wants)
}

func (m *protoMatcher) String() string {
	return ""
}

func TestHandleWsHandleExecShellMsg(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	sm := mock.NewMockSessionMapper(m)
	ps := mock.NewMockPubSub(m)
	conn := &WsConn{
		terminalSessions: sm,
		pubSub:           ps,
	}
	message := &websocket.TerminalMessage{
		Data:      []byte("data"),
		SessionId: "sid",
	}
	marshal, _ := proto.Marshal(&websocket.TerminalMessageInput{
		Type:    websocket.Type_HandleExecShellMsg,
		Message: message,
	})
	sm.EXPECT().Send(&protoMatcher{wants: proto.Clone(message)}).Times(1)
	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	HandleWsHandleExecShellMsg(conn, websocket.Type_HandleExecShellMsg, []byte("1:"))
	HandleWsHandleExecShellMsg(conn, websocket.Type_HandleExecShellMsg, marshal)
}

func TestNewWebsocketManager(t *testing.T) {
	assert.NotNil(t, NewWebsocketManager(1*time.Second))
}

type rw struct {
	h http.Header
	w []byte
	http.ResponseWriter
}

func (receiver *rw) Header() http.Header {
	return receiver.h
}

func (receiver *rw) Write(in []byte) (int, error) {
	receiver.w = in
	return len(in), nil
}

func TestWebsocketManager_Info(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{
		WsSenderPlugin: config.Plugin{
			Name: "test_ws",
		},
	})
	ws := mock.NewMockWsSender(m)
	ps := mock.NewMockPubSub(m)
	ws.EXPECT().New("", "").Return(ps)
	ws.EXPECT().Initialize(gomock.Any()).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).AnyTimes()
	ps.EXPECT().Info().Return("info...")
	app.EXPECT().GetPluginByName("test_ws").Return(ws)
	rwer := &rw{
		h: map[string][]string{},
		w: nil,
	}
	NewWebsocketManager(1*time.Second).Info(rwer, nil)
	assert.Equal(t, "application/json", rwer.h["Content-Type"][0])
	marshal, _ := json.Marshal("info...")
	assert.Equal(t, marshal, rwer.w)
}

type testWait struct {
	c int
}

func (t *testWait) Inc() {
	t.c++
}

func (t *testWait) Dec() {
	t.c--
}

func (t *testWait) Wait() {

}
func (t *testWait) Count() int {
	return t.c
}

func TestWebsocketManager_initConn(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{
		WsSenderPlugin: config.Plugin{
			Name: "test_ws",
		},
	})
	ps := mock.NewMockPubSub(m)

	ws := mock.NewMockWsSender(m)
	ws.EXPECT().New("xxx", gomock.Any()).Return(ps)
	ws.EXPECT().Initialize(gomock.Any()).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).AnyTimes()
	app.EXPECT().GetPluginByName("test_ws").Return(ws)

	parse, _ := url.Parse("https://mars.local/ws?uid=xxx")
	r := &http.Request{
		URL: parse,
	}
	Wait = &testWait{}
	defer func() {
		Wait = NewWaitSocketExit()
	}()
	conn := NewWebsocketManager(1*time.Second).initConn(r, nil)
	assert.Equal(t, ps, conn.pubSub)
	assert.NotEmpty(t, conn.id)
	assert.Equal(t, "xxx", conn.uid)
	assert.Nil(t, conn.conn)
	assert.NotNil(t, conn.terminalSessions)
	assert.Equal(t, 1, Wait.Count())
}

func TestWebsocketManager_initConn2(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	app.EXPECT().Config().Return(&config.Config{
		WsSenderPlugin: config.Plugin{
			Name: "test_ws",
		},
	})
	ps := mock.NewMockPubSub(m)

	ws := mock.NewMockWsSender(m)
	ws.EXPECT().New(gomock.Any(), gomock.Any()).Return(ps)
	ws.EXPECT().Initialize(gomock.Any()).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).AnyTimes()
	app.EXPECT().GetPluginByName("test_ws").Return(ws)

	parse, _ := url.Parse("https://mars.local/ws")
	r := &http.Request{
		URL: parse,
	}
	Wait = &testWait{}
	defer func() {
		Wait = NewWaitSocketExit()
	}()
	conn := NewWebsocketManager(1*time.Second).initConn(r, nil)
	assert.NotEmpty(t, conn.uid)
}

func TestWsConn_GetUser(t *testing.T) {
	c := &WsConn{}
	assert.IsType(t, contracts.UserInfo{}, c.GetUser())
}

func TestWsConn_SetUser(t *testing.T) {
	c := &WsConn{}
	u := contracts.UserInfo{
		Name: "aaa",
	}
	c.SetUser(u)
	assert.Equal(t, "aaa", c.user.Name)
}

func TestWsConn_Shutdown(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cs := mock.NewMockCancelSignaler(m)
	cs.EXPECT().CancelAll().Times(1)
	tm := mock.NewMockSessionMapper(m)
	tm.EXPECT().CloseAll().Times(1)
	ps := mock.NewMockPubSub(m)
	ps.EXPECT().Close().Times(1)
	conn := mock.NewMockWebsocketConn(m)
	conn.EXPECT().Close().Times(1)
	var doneCalled bool
	c := &WsConn{
		doneFunc: func() {
			doneCalled = true
		},
		conn:             conn,
		cancelSignaler:   cs,
		pubSub:           ps,
		terminalSessions: tm,
	}
	Wait = &testWait{}
	defer func() {
		Wait = NewWaitSocketExit()
	}()
	c.Shutdown()
	assert.True(t, doneCalled)
	assert.Equal(t, -1, Wait.Count())
}

func TestWebsocketManager_TickClusterHealth(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	ch := make(chan struct{})
	l := cache_lock.NewMemoryLock([2]int{-1, 100}, cache_lock.NewMemStore())
	app.EXPECT().CacheLock().Return(l).AnyTimes()
	app.EXPECT().Done().Return(ch).Times(1)
	go func() {
		time.Sleep(1500 * time.Millisecond)
		close(ch)
	}()

	app.EXPECT().Config().Return(&config.Config{
		WsSenderPlugin: config.Plugin{
			Name: "test_ws",
		},
	})
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fake.NewSimpleClientset(), MetricsClient: fake2.NewSimpleClientset()}).AnyTimes()
	ps := mock.NewMockPubSub(m)
	ps.EXPECT().ToAll(gomock.Any()).Times(1)

	ws := mock.NewMockWsSender(m)
	ws.EXPECT().New("", "").Return(ps)
	ws.EXPECT().Initialize(gomock.Any()).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).AnyTimes()
	app.EXPECT().GetPluginByName("test_ws").Return(ws)

	NewWebsocketManager(1 * time.Second).TickClusterHealth()
	time.Sleep(2 * time.Second)
}

type slowLocker struct {
	contracts.Locker
}

func (s *slowLocker) Acquire(key string, seconds int64) bool {
	if s.Locker.Acquire(key, seconds) {
		time.Sleep(200 * time.Millisecond)
		return true
	}
	return false
}

func TestWebsocketManager_TickClusterHealth_Parallel(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	app := testutil.MockApp(m)
	ch := make(chan struct{})
	l := cache_lock.NewMemStore()
	app.EXPECT().CacheLock().Return(&slowLocker{Locker: cache_lock.NewMemoryLock([2]int{-1, 100}, l)}).AnyTimes()
	app.EXPECT().Done().Return(ch).AnyTimes()
	go func() {
		time.Sleep(1500 * time.Millisecond)
		close(ch)
	}()

	app.EXPECT().Config().Return(&config.Config{
		WsSenderPlugin: config.Plugin{
			Name: "test_ws",
		},
	}).AnyTimes()
	app.EXPECT().K8sClient().Return(&contracts.K8sClient{Client: fake.NewSimpleClientset(), MetricsClient: fake2.NewSimpleClientset()}).AnyTimes()
	ps := mock.NewMockPubSub(m)
	ps.EXPECT().ToAll(gomock.Any()).Times(1)

	ws := mock.NewMockWsSender(m)
	ws.EXPECT().New("", "").Return(ps).AnyTimes()
	ws.EXPECT().Initialize(gomock.Any()).AnyTimes()
	app.EXPECT().RegisterAfterShutdownFunc(gomock.Any()).AnyTimes()
	app.EXPECT().GetPluginByName("test_ws").Return(ws).AnyTimes()

	for i := 0; i < 10; i++ {
		go NewWebsocketManager(1 * time.Second).TickClusterHealth()
	}
	time.Sleep(2 * time.Second)
}

func Test_Upgrader(t *testing.T) {
	assert.True(t, upgrader.CheckOrigin(nil))
}

func TestHandleWsProjectPodEvent(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ps := mock.NewMockPubSub(m)
	c := &WsConn{
		id:     "id",
		uid:    "uid",
		pubSub: ps,
	}
	marshal, _ := proto.Marshal(&websocket.ProjectPodEventJoinInput{
		Type:        websocket.Type_ProjectPodEvent,
		Join:        false,
		ProjectId:   1,
		NamespaceId: 1,
	})
	ps.EXPECT().Leave(int64(1), int64(1)).Times(1)
	HandleWsProjectPodEvent(c, websocket.Type_ProjectPodEvent, marshal)
	marshal2, _ := proto.Marshal(&websocket.ProjectPodEventJoinInput{
		Type:        websocket.Type_ProjectPodEvent,
		Join:        true,
		ProjectId:   2,
		NamespaceId: 2,
	})
	ps.EXPECT().Join(int64(2)).Times(1)
	HandleWsProjectPodEvent(c, websocket.Type_ProjectPodEvent, marshal2)

	ps.EXPECT().ToSelf(gomock.Any()).Times(1)
	HandleWsProjectPodEvent(c, websocket.Type_ProjectPodEvent, []byte("xxxxx"))
}

func TestWebsocketManager_Ws(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/test-ws", nil)
	NewWebsocketManager(1*time.Second).Ws(recorder, request)
	assert.Equal(t, 400, recorder.Code)
}

func Test_handleWsRead(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ps := mock.NewMockPubSub(m)

	ps.EXPECT().ToSelf(&websocket.WsMetadataResponse{
		Metadata: &websocket.Metadata{
			Type:    websocket.Type_HandleAuthorize,
			Result:  ResultSuccess,
			End:     false,
			Uid:     "uid",
			Id:      "",
			Message: "认证中，请稍等~",
		},
	}).Times(1)
	handleWsRead(&WsConn{
		uid:    "uid",
		pubSub: ps,
		user: contracts.UserInfo{
			ID: "",
		},
	}, &websocket.WsRequestMetadata{
		Type: websocket.Type_ApplyProject,
	}, nil, map[websocket.Type]HandleRequestFunc{
		websocket.Type_ApplyProject: func(c *WsConn, t websocket.Type, message []byte) {
		},
	})

	called := false
	handleWsRead(&WsConn{
		id:     "id",
		uid:    "uid",
		pubSub: ps,
		user: contracts.UserInfo{
			ID: "id",
		},
	}, &websocket.WsRequestMetadata{
		Type: websocket.Type_ApplyProject,
	}, nil, map[websocket.Type]HandleRequestFunc{
		websocket.Type_ApplyProject: func(c *WsConn, t websocket.Type, message []byte) {
			called = true
		},
	})
	assert.True(t, called)

	app := testutil.MockApp(m)
	app.EXPECT().IsDebug().Return(true).Times(1)
	assert.PanicsWithValue(t, "errxxx", func() {
		handleWsRead(&WsConn{
			id:     "id",
			uid:    "uid",
			pubSub: ps,
			user: contracts.UserInfo{
				ID: "id",
			},
		}, &websocket.WsRequestMetadata{
			Type: websocket.Type_ApplyProject,
		}, nil, map[websocket.Type]HandleRequestFunc{
			websocket.Type_ApplyProject: func(c *WsConn, t websocket.Type, message []byte) {
				panic("errxxx")
			},
		})
	})
}
