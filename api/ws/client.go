package ws

// client.go 定义 ws 客户端门面（Client）及其连接生命周期：
// HTTP/WS 握手升级 → HandleAuthorize 鉴权 → read 循环分发 → 断线自动重连。
//
// 帧协议与服务端 internal/services/websocket 对齐：一律为二进制 protobuf，
// 判别器是 proto 的 type 字段（field 1）。客户端出站帧反解成 WsRequestMetadata
// 取 field1 路由；服务端出站帧外层均可解成 WsMetadataResponse，具体 payload
// 按 metadata.type 二次解码。

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	apihttp "github.com/duc-cnzj/mars/api/v6/http"
	"github.com/duc-cnzj/mars/api/v6/proto/auth"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// 连接调优参数（对齐服务端 conn.go）。
const (
	// 允许对端发送的最大消息字节数。
	maxMessageSize int64 = 1024 * 1024 * 20 // 20MB
	// 允许读取对端下一个 pong 的时限；超时判定连接死亡触发重连。
	pongWait = 60 * time.Second
	// 每次握手/登录的等待时限。
	handshakeTimeout = 10 * time.Second
)

// Client 是 mars ws 客户端门面：持有连接、鉴权 token 来源与事件订阅注册表。
// 连接在后台 goroutine 常驻（run），断线按退避策略自动重连并重鉴权、重放
// pod 事件订阅；NewClient 返回后调用 waitReady 等待连接就绪。
type Client struct {
	url    string
	dialer *websocket.Dialer

	tokenProvider func(ctx context.Context) (string, error)
	httpAuth      *authCreds
	httpClient    *http.Client
	backoff       backoff.BackOff

	// 生命周期
	ctx       context.Context
	ctxCancel context.CancelFunc
	done      chan struct{}
	closed    atomic.Bool
	wg        sync.WaitGroup

	// 连接状态
	conn       *websocket.Conn
	writeMu    sync.Mutex
	ready      atomic.Bool
	uidVal     atomic.Value
	authFailed atomic.Bool
	readyOnce  sync.Once
	authOnce   sync.Once
	readyCh    chan struct{}
	authFailCh chan struct{}

	// 订阅注册表（mu 保护）
	mu              sync.RWMutex
	typeHandlers    map[websocket_pb.Type]map[string]func(*Event)
	sessionHandlers map[string]map[string]func(*Event)
}

// NewClient 建立到 url（形如 "ws://host:port/ws"）的 ws 客户端并立即在后台启动连接。
// 必须通过 WithBearerToken/WithAuth/WithTokenProvider 提供鉴权 token 来源，
// 否则返回错误。连接与鉴权是异步的，用 waitReady 等待就绪或鉴权失败。
func NewClient(url string, opts ...Option) (*Client, error) {
	c := &Client{
		url:             url,
		dialer:          websocket.DefaultDialer,
		backoff:         backoff.NewExponentialBackOff(),
		done:            make(chan struct{}),
		readyCh:         make(chan struct{}),
		authFailCh:      make(chan struct{}),
		typeHandlers:    make(map[websocket_pb.Type]map[string]func(*Event)),
		sessionHandlers: make(map[string]map[string]func(*Event)),
	}
	for _, opt := range opts {
		opt(c)
	}
	if err := c.initTokenProvider(); err != nil {
		return nil, err
	}
	c.ctx, c.ctxCancel = context.WithCancel(context.Background())
	c.wg.Add(1)
	go c.run()
	return c, nil
}

// initTokenProvider 收口鉴权来源：WithAuth 需要派生 HTTP base 并登录换取 token。
// WithBearerToken/WithTokenProvider 已在 Option 阶段设置 tokenProvider，这里不再覆盖。
func (c *Client) initTokenProvider() error {
	if c.httpAuth == nil {
		if c.tokenProvider == nil {
			return errors.New("ws: 需要 WithBearerToken/WithAuth/WithTokenProvider 之一提供鉴权 token")
		}
		return nil
	}
	base, err := wsURLToHTTPBase(c.url)
	if err != nil {
		return err
	}
	hc := c.httpClient
	if hc == nil {
		hc = http.DefaultClient
	}
	// 仅注入 http.Client（无用户名密码），NewClient 不会触发登录，不会返回错误。
	login, _ := apihttp.NewClient(base, apihttp.WithHTTPClient(hc))
	creds := c.httpAuth
	c.tokenProvider = func(ctx context.Context) (string, error) {
		res, err := login.Auth().Login(ctx, &auth.LoginRequest{
			Username: creds.username,
			Password: creds.password,
		})
		if err != nil {
			return "", err
		}
		return res.Token, nil
	}
	return nil
}

// run 是后台连接主循环：连接→读，读结束（断线/失败）后按退避策略重连，
// 客户端已关闭或鉴权失败时退出。
func (c *Client) run() {
	defer c.wg.Done()
	for {
		// 返回值只表示连接结束（正常/异常），重连与否由 closed/authFailed 决定。
		_ = c.connectOnce()
		if c.closed.Load() || c.authFailed.Load() {
			return
		}
		d := c.backoff.NextBackOff()
		if d == backoff.Stop {
			return
		}
		timer := time.NewTimer(d)
		select {
		case <-timer.C:
		case <-c.done:
			timer.Stop()
			return
		}
	}
}

// connectOnce 完成一次连接：握手升级、鉴权、重放 pod 事件、进入读循环直到断线。
func (c *Client) connectOnce() error {
	conn, _, err := c.dialer.DialContext(c.ctx, c.url, nil)
	if err != nil {
		return err
	}
	c.setConn(conn)
	defer c.closeConn()
	c.ready.Store(false)

	if err := c.authorize(); err != nil {
		return err
	}
	return c.readLoop(conn)
}

// authorize 从 tokenProvider 取 token 并发送 HandleAuthorize 帧。
// 鉴权失败由服务端异步回 InternalError 帧（authFailed），此处只负责发送。
func (c *Client) authorize() error {
	ctx, cancel := context.WithTimeout(c.ctx, handshakeTimeout)
	defer cancel()
	token, err := c.tokenProvider(ctx)
	if err != nil {
		return err
	}
	return c.writeMsg(&websocket_pb.AuthorizeTokenInput{
		Type:  websocket_pb.Type_HandleAuthorize,
		Token: token,
	})
}

// readLoop 循环读取二进制帧并分发；对端断开或超时返回错误触发重连。
func (c *Client) readLoop(conn *websocket.Conn) error {
	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		c.dispatch(message)
	}
}

// dispatch 把单条入站帧解码成 WsMetadataResponse 并分发：维护连接状态
// （SetUid/InternalError）并按 type/session 路由到订阅方。
func (c *Client) dispatch(message []byte) {
	var resp websocket_pb.WsMetadataResponse
	if err := proto.Unmarshal(message, &resp); err != nil {
		return
	}
	m := resp.Metadata
	if m == nil {
		return
	}
	switch m.Type {
	case websocket_pb.Type_SetUid:
		c.uidVal.Store(m.Message)
		c.ready.Store(true)
		c.readyOnce.Do(func() { close(c.readyCh) })
	case websocket_pb.Type_InternalError:
		c.authFailed.Store(true)
		c.authOnce.Do(func() { close(c.authFailCh) })
	}
	c.notify(m, &Event{Type: m.Type, Metadata: m, Raw: message})
}

// notify 按注册表分发一条已解码事件。
func (c *Client) notify(m *websocket_pb.Metadata, ev *Event) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 终端输出/关闭帧：按 terminal_message.session_id 路由。
	if m.Type == websocket_pb.Type_HandleExecShell ||
		m.Type == websocket_pb.Type_HandleExecShellMsg ||
		m.Type == websocket_pb.Type_HandleCloseShell {
		var shell websocket_pb.WsHandleShellResponse
		if proto.Unmarshal(ev.Raw, &shell) == nil && shell.TerminalMessage != nil {
			if hs := c.sessionHandlers[shell.TerminalMessage.SessionId]; len(hs) > 0 {
				for _, h := range hs {
					h(ev)
				}
			}
		}
	}
	if hs := c.typeHandlers[m.Type]; len(hs) > 0 {
		for _, h := range hs {
			h(ev)
		}
	}
}

// writeMsg 把一条 proto 消息序列化并以二进制帧写出（写锁串行化，对齐服务端写循环）。
func (c *Client) writeMsg(m proto.Message) error {
	// 入参均为本包构造的合法 proto 消息，Marshal 不会出错（丢弃死分支）。
	data, _ := proto.Marshal(m)
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.closed.Load() {
		return errors.New("ws: 客户端已关闭")
	}
	conn := c.conn
	if conn == nil {
		return errors.New("ws: 未连接")
	}
	return conn.WriteMessage(websocket.BinaryMessage, data)
}

// setConn 记录当前连接（写锁保护）。
func (c *Client) setConn(conn *websocket.Conn) {
	c.writeMu.Lock()
	c.conn = conn
	c.writeMu.Unlock()
}

// closeConn 清空并关闭当前连接。
func (c *Client) closeConn() {
	c.writeMu.Lock()
	conn := c.conn
	c.conn = nil
	c.writeMu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

// waitReady 阻塞直到连接就绪（收到首个 SetUid 帧）、鉴权失败或 ctx 取消。
// 连接就绪表示可以开始发送操作帧；鉴权失败返回错误。
func (c *Client) waitReady(ctx context.Context) error {
	if c.ready.Load() {
		return nil
	}
	select {
	case <-c.readyCh:
		return nil
	case <-c.authFailCh:
		return errors.New("ws: 鉴权失败")
	case <-c.done:
		return errors.New("ws: 客户端已关闭")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close 关闭客户端：停止重连、关闭当前连接并等待后台循环退出。幂等。
func (c *Client) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil
	}
	close(c.done)
	c.ctxCancel()
	c.closeConn()
	c.wg.Wait()
	return nil
}

// uid 返回连接的 uid（收到 SetUid 帧后有效）。
func (c *Client) uid() string {
	if v := c.uidVal.Load(); v != nil {
		return v.(string)
	}
	return ""
}

// wsURLToHTTPBase 把 ws/wss url 转成同主机的 http/https base（去掉路径），
// 供 WithAuth 登录复用 api/http 客户端。
func wsURLToHTTPBase(u string) (string, error) {
	p, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("ws: 解析 url %q 失败: %w", u, err)
	}
	switch p.Scheme {
	case "ws":
		p.Scheme = "http"
	case "wss":
		p.Scheme = "https"
	default:
		return "", fmt.Errorf("ws: 不支持的 scheme %q（需 ws/wss）", p.Scheme)
	}
	p.Path = ""
	p.RawPath = ""
	p.RawQuery = ""
	p.Fragment = ""
	return p.String(), nil
}
