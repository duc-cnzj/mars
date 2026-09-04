package ws

// edge_test.go：补齐覆盖 100% 的边界/错误分支——
// option 副作用、writeMsg 失败分支、WaitReady 各出口、
// dispatch 坏帧、authorize/connectOnce/run 错误路径、Terminal 错误分支。

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// ---- option.go 副作用 ----

func TestOption_SideEffects(t *testing.T) {
	c := &Client{}

	tokFn := func(context.Context) (string, error) { return "tk", nil }
	WithTokenProvider(tokFn)(c)
	if c.tokenProvider == nil {
		t.Fatal("WithTokenProvider 应设置 tokenProvider")
	}

	hc := &http.Client{}
	WithHTTPClient(hc)(c)
	if c.httpClient != hc {
		t.Fatal("WithHTTPClient 应设置 httpClient")
	}

	dl := &websocket.Dialer{HandshakeTimeout: time.Second}
	WithDialer(dl)(c)
	if c.dialer != dl {
		t.Fatal("WithDialer 应设置 dialer")
	}
	WithDialer(nil)(c) // nil 不覆盖
	if c.dialer != dl {
		t.Fatal("WithDialer(nil) 不应覆盖已有 dialer")
	}

	bb := backoff.NewConstantBackOff(time.Second)
	WithReconnectBackoff(bb)(c)
	if c.backoff != bb {
		t.Fatal("WithReconnectBackoff 应设置 backoff")
	}
	WithReconnectBackoff(nil)(c)
	if c.backoff != bb {
		t.Fatal("WithReconnectBackoff(nil) 不应覆盖已有 backoff")
	}
}

// ---- writeMsg / send 失败分支 ----

func TestWriteMsg_ClosedAndNotConnected(t *testing.T) {
	c := &Client{done: make(chan struct{})}
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())

	c.closed.Store(true)
	if err := c.writeMsg(&websocket_pb.AuthorizeTokenInput{}); err == nil {
		t.Fatal("closed 时 writeMsg 应报错")
	}
	c.closed.Store(false)

	if err := c.writeMsg(&websocket_pb.AuthorizeTokenInput{}); err == nil {
		t.Fatal("未连接时 writeMsg 应报错")
	}
}

// ---- authorize / connectOnce / run 错误路径 ----

func TestAuthorize_TokenProviderError(t *testing.T) {
	c := &Client{ctx: context.Background()}
	c.tokenProvider = func(context.Context) (string, error) { return "", errors.New("token 获取失败") }
	if err := c.authorize(); err == nil {
		t.Fatal("tokenProvider 出错时 authorize 应返回错误")
	}
}

func TestConnectOnce_DialError(t *testing.T) {
	c := &Client{
		url:    "ws://127.0.0.1:1/ws",
		dialer: websocket.DefaultDialer,
		ctx:    context.Background(),
		done:   make(chan struct{}),
	}
	if err := c.connectOnce(); err == nil {
		t.Fatal("dial 到关闭端口应报错")
	}
}

func TestRun_StopBackoff(t *testing.T) {
	// 退避策略返回 Stop 时 run 应退出（91.7% 缺口）。
	c := &Client{
		url:             "ws://127.0.0.1:1/ws",
		dialer:          websocket.DefaultDialer,
		backoff:         &backoff.StopBackOff{},
		done:            make(chan struct{}),
		readyCh:         make(chan struct{}),
		authFailCh:      make(chan struct{}),
		typeHandlers:    map[websocket_pb.Type]map[string]func(*Event){},
		sessionHandlers: map[string]map[string]func(*Event){},
	}
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())
	c.wg.Add(1)
	go c.run()
	c.wg.Wait() // Stop 退避后应立即退出，不会挂死
}

func TestRun_CloseDuringBackoff(t *testing.T) {
	// 退避等待期间 Close（done 关闭）应退出（case <-c.done 分支）。
	c := &Client{
		url:     "ws://127.0.0.1:1/ws",
		dialer:  websocket.DefaultDialer,
		backoff: backoff.NewConstantBackOff(time.Hour),
		done:    make(chan struct{}),
	}
	c.tokenProvider = func(context.Context) (string, error) { return "", errors.New("no") }
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())
	c.wg.Add(1)
	go c.run()
	time.Sleep(10 * time.Millisecond) // 让 run 进入退避等待
	_ = c.Close()
	c.wg.Wait()
}

// ---- dispatch 坏帧 ----

func TestDispatch_BadFrames(t *testing.T) {
	c := &Client{
		done:            make(chan struct{}),
		readyCh:         make(chan struct{}),
		authFailCh:      make(chan struct{}),
		typeHandlers:    map[websocket_pb.Type]map[string]func(*Event){},
		sessionHandlers: map[string]map[string]func(*Event){},
	}
	// 不可反解为 WsMetadataResponse 的垃圾帧 → 直接丢弃。
	c.dispatch([]byte{0xff, 0xfe, 0xfd})
	// 可反解但 metadata 为 nil → 直接丢弃。
	garbage, _ := proto.Marshal(&websocket_pb.WsMetadataResponse{})
	c.dispatch(garbage)
}

// ---- WaitReady 各出口 ----

func TestWaitReady_AlreadyReady(t *testing.T) {
	c := &Client{done: make(chan struct{})}
	c.ready.Store(true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.waitReady(ctx); err != nil {
		t.Fatalf("已就绪时应立即返回 nil: %v", err)
	}
}

func TestWaitReady_Closed(t *testing.T) {
	c := &Client{done: make(chan struct{}), readyCh: make(chan struct{}), authFailCh: make(chan struct{})}
	close(c.done)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := c.waitReady(ctx); err == nil {
		t.Fatal("closed 时 WaitReady 应返回错误")
	}
}

func TestWaitReady_CtxCancel(t *testing.T) {
	c := &Client{done: make(chan struct{}), readyCh: make(chan struct{}), authFailCh: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 已取消的 ctx
	if err := c.waitReady(ctx); err == nil {
		t.Fatal("ctx 取消时 WaitReady 应返回错误")
	}
}

func TestUID_Empty(t *testing.T) {
	c := &Client{}
	if got := c.uid(); got != "" {
		t.Fatalf("未收到 SetUid 时 UID 应为空，实际 %q", got)
	}
}

// ---- wsURLToHTTPBase url.Parse 错误 ----

func TestWSURLToHTTPBase_ParseError(t *testing.T) {
	if _, err := wsURLToHTTPBase("://bad"); err == nil {
		t.Fatal("非法 url 应报错")
	}
}

// ---- initTokenProvider WithAuth 全链路 ----

// loginServer 起一个同时处理登录 POST 与 ws 升级的 httptest 服务端。
// loginFail 为 true 时登录返回 500；否则返回合法 token。
func loginServer(t *testing.T, loginFail bool, token string, onWS func(c *websocket.Conn)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/auth/login" {
			if loginFail {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"token":"` + token + `"}`))
			return
		}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		onWS(c)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestWithAuth_LoginFlow(t *testing.T) {
	// 用登录换取 token → 鉴权帧携带该 token → 就绪。
	srv := loginServer(t, false, "abc", func(c *websocket.Conn) {
		defer c.Close()
		tok := readAuthorize(t, c)
		if tok != "abc" {
			t.Fatalf("鉴权 token 期望 abc，实际 %q", tok)
		}
		sendSetUid(t, c, "u")
	})
	cli, err := NewClient(wsURL(t, srv)+"/ws", WithAuth("u", "p"),
		WithReconnectBackoff(backoff.NewConstantBackOff(time.Millisecond)))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	ready(t, cli)
}

func TestWithAuth_LoginFailure(t *testing.T) {
	// 登录失败 → tokenProvider 报错 → authorize 返回错误。
	srv := loginServer(t, true, "", func(c *websocket.Conn) {})
	cli, err := NewClient(wsURL(t, srv)+"/ws", WithAuth("u", "p"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	if err := cli.authorize(); err == nil {
		t.Fatal("登录失败时 authorize 应报错")
	}
}

// ---- connectOnce authorize 错误分支 ----

func TestConnectOnce_AuthorizeError(t *testing.T) {
	// tokenProvider 失败 → 客户端不会发鉴权帧，服务端只升级即返回。
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
	})
	cli, err := NewClient(wsURL(t, srv)+"/ws",
		WithTokenProvider(func(context.Context) (string, error) { return "", errors.New("取 token 失败") }))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	if err := cli.connectOnce(); err == nil {
		t.Fatal("authorize 失败时 connectOnce 应返回错误")
	}
}

// ---- readLoop pong handler ----

func TestReadLoop_PongHandler(t *testing.T) {
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		// 发一个 pong 控制帧触发客户端的 PongHandler。
		_ = c.WriteControl(websocket.PongMessage, []byte("pong"), time.Now().Add(time.Second))
	})
	cli, err := NewClient(wsURL(t, srv)+"/ws", WithBearerToken("t"))
	if err != nil {
		t.Fatal(err)
	}
	ready(t, cli)
	_ = cli.Close() // 关闭连接触发 readLoop 的读错误返回
}

// ---- Terminal 错误分支 ----

func TestTerminal_OpenErrors(t *testing.T) {
	// OpenTerminal 的守卫：nil container 直接报错；ctx 取消/连接未就绪时 waitReady 报错。
	c := &Client{done: make(chan struct{})}
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())

	if _, err := c.OpenTerminal(context.Background(), nil); err == nil {
		t.Fatal("nil container 时 OpenTerminal 应报错")
	}

	// ctx 已取消 → OpenTerminal 内部 waitReady 返回 ctx.Err()。
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.OpenTerminal(ctx, &websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"}); err == nil {
		t.Fatal("ctx 取消时 OpenTerminal 应报错")
	}

	// 直接构造 Terminal 验证 Write 的 sendStdin 错误分支（closed 客户端）。
	c.closed.Store(true)
	tm := &Terminal{client: c, sessionID: "s"}
	if n, err := tm.Write([]byte("x")); err == nil || n != 0 {
		t.Fatalf("closed 时 Write 应返回 (0,err)，实际 (%d,%v)", n, err)
	}
}

func TestTerminal_ResizeAndHandleBadFrame(t *testing.T) {
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		var in websocket_pb.WsHandleExecShellInput
		readFrame(t, c, &in)
		var resize websocket_pb.TerminalMessageInput
		readFrame(t, c, &resize)
		if resize.Message.GetOp() != "resize" {
			t.Fatalf("op 期望 resize，实际 %q", resize.Message.GetOp())
		}
		if resize.Message.GetSessionId() != in.SessionId {
			t.Fatalf("resize 应带自动生成的 sessionID %q，实际 %q", in.SessionId, resize.Message.GetSessionId())
		}
	})
	cli, err := NewClient(wsURL(t, srv)+"/ws", WithBearerToken("t"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()
	ready(t, cli)

	term, err := cli.OpenTerminal(context.Background(), &websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	if err := term.Resize(40, 120); err != nil {
		t.Fatalf("Resize 失败: %v", err)
	}

	// handle 的坏帧分支：不可反解的 Raw 应被忽略，不 panic。
	term.handle(&Event{Type: websocket_pb.Type_HandleExecShellMsg, Raw: []byte{0x01, 0x02}})
	term.handle(&Event{Type: websocket_pb.Type_HandleCloseShell}) // Raw 为空 → TerminalMessage 为 nil，同样忽略
}

// ---- OpenTerminal execShell 错误分支（waitReady 通过但连接缺失）----

func TestOpenTerminal_ExecShellError(t *testing.T) {
	c := &Client{done: make(chan struct{})}
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())
	c.ready.Store(true) // waitReady 直接通过
	// conn 为 nil → execShell → writeMsg 返回"未连接"。
	if _, err := c.OpenTerminal(context.Background(),
		&websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"}); err == nil {
		t.Fatal("未连接时 OpenTerminal 应报错")
	}
}

// ---- retryGate（鉴权竞态兜底）各出口 ----

// newRetryClient 构造一个已关闭的 Client（让兜底 goroutine 的 execShell 静默失败），
// 并返回绑定的 Terminal（opened 未关闭、retries 可配）。
func newRetryClient() *Client {
	c := &Client{done: make(chan struct{})}
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())
	c.closed.Store(true)
	return c
}

func TestRetryGate_NonGateFrames(t *testing.T) {
	t2 := &Terminal{client: newRetryClient(), opened: make(chan struct{})}
	t2.retryGate(&Event{Metadata: nil}) // metadata 为 nil → 直接返回
	t2.retryGate(&Event{Metadata: &websocket_pb.Metadata{Message: "hello"}})
	if t2.retries != 0 {
		t.Fatalf("非 gate 帧不应触发重试，实际 %d", t2.retries)
	}
}

func TestRetryGate_Opened(t *testing.T) {
	t2 := &Terminal{client: newRetryClient(), opened: make(chan struct{})}
	close(t2.opened) // shell 已开启 → 直接返回
	t2.retryGate(&Event{Metadata: &websocket_pb.Metadata{Message: "认证中，请稍等~"}})
	if t2.retries != 0 {
		t.Fatalf("opened 后不应重试，实际 %d", t2.retries)
	}
}

func TestRetryGate_MaxRetries(t *testing.T) {
	t2 := &Terminal{client: newRetryClient(), opened: make(chan struct{}), retries: maxOpenRetries}
	t2.retryGate(&Event{Metadata: &websocket_pb.Metadata{Message: "认证中，请稍等~"}})
	if t2.retries != maxOpenRetries {
		t.Fatalf("超过最大重试不应再重发，实际 %d", t2.retries)
	}
}

func TestRetryGate_Retries(t *testing.T) {
	// 正常路径：retries 递增 + 独立 goroutine 里 sleep 后重发 execShell。
	t2 := &Terminal{
		client:    newRetryClient(),
		opened:    make(chan struct{}),
		sessionID: "ns-pod-c:sid",
		container: &websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"},
	}
	t2.retryGate(&Event{Metadata: &websocket_pb.Metadata{Message: "认证中，请稍等~"}})
	if t2.retries != 1 {
		t.Fatalf("应重试一次，实际 %d", t2.retries)
	}
	time.Sleep(200 * time.Millisecond) // 等 goroutine 的 sleep+execShell 跑完
}

// ---- handle 缓冲满丢弃分支 ----

func TestTerminal_handle_BufferFull(t *testing.T) {
	c := &Client{done: make(chan struct{})}
	t2 := &Terminal{client: c, opened: make(chan struct{}), stdout: make(chan []byte, 1), toast: make(chan []byte, 1)}
	t2.stdout <- []byte("full") // 缓冲满 → stdout 帧走 default 丢弃
	msg, _ := proto.Marshal(&websocket_pb.WsHandleShellResponse{
		Metadata:        &websocket_pb.Metadata{Type: websocket_pb.Type_HandleExecShellMsg},
		TerminalMessage: &websocket_pb.TerminalMessage{Op: "stdout", Data: []byte("x")},
	})
	t2.handle(&Event{Type: websocket_pb.Type_HandleExecShellMsg, Raw: msg})
	t2.toast <- []byte("full") // 缓冲满 → toast 帧走 default 丢弃
	msg2, _ := proto.Marshal(&websocket_pb.WsHandleShellResponse{
		Metadata:        &websocket_pb.Metadata{Type: websocket_pb.Type_HandleExecShellMsg},
		TerminalMessage: &websocket_pb.TerminalMessage{Op: "toast", Data: []byte("y")},
	})
	t2.handle(&Event{Type: websocket_pb.Type_HandleExecShellMsg, Raw: msg2})
}
