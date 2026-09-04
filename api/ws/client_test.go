package ws

// client_test.go：用 httptest + gorilla Upgrader 起真实内存 ws 服务端，
// 对客户端做集成式验证（连接/鉴权/重连/就绪/关闭）。其余测试文件复用本文件
// 的 newWsServer / wsURL 等 harness。

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// upgrader 允许任意来源（对齐服务端 CheckOrigin 恒真）。
var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

// newWsServer 起一个 httptest 内存服务端，每个 ws 连接都交给 onConn 处理。
func newWsServer(t *testing.T, onConn func(c *websocket.Conn)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		onConn(c)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// wsURL 把 httptest 的 http URL 转成 ws URL。
func wsURL(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	return "ws" + strings.TrimPrefix(srv.URL, "http")
}

// readAuthorize 读客户端首帧（HandleAuthorize）并返回 token。
func readAuthorize(t *testing.T, c *websocket.Conn) string {
	t.Helper()
	_, msg, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("读取鉴权帧失败: %v", err)
	}
	var in websocket_pb.AuthorizeTokenInput
	if err := proto.Unmarshal(msg, &in); err != nil {
		t.Fatalf("解析鉴权帧失败: %v", err)
	}
	if in.Type != websocket_pb.Type_HandleAuthorize {
		t.Fatalf("首帧 type 应为 HandleAuthorize，实际 %v", in.Type)
	}
	return in.Token
}

// sendSetUid 发送 SetUid 就绪帧。
func sendSetUid(t *testing.T, c *websocket.Conn, uid string) {
	t.Helper()
	data, _ := proto.Marshal(&websocket_pb.WsMetadataResponse{
		Metadata: &websocket_pb.Metadata{Type: websocket_pb.Type_SetUid, Message: uid},
	})
	if err := c.WriteMessage(websocket.BinaryMessage, data); err != nil {
		t.Fatalf("发送 SetUid 失败: %v", err)
	}
}

// sendInternalError 发送鉴权失败帧。
func sendInternalError(t *testing.T, c *websocket.Conn) {
	t.Helper()
	data, _ := proto.Marshal(&websocket_pb.WsMetadataResponse{
		Metadata: &websocket_pb.Metadata{Type: websocket_pb.Type_InternalError},
	})
	if err := c.WriteMessage(websocket.BinaryMessage, data); err != nil {
		t.Fatalf("发送 InternalError 失败: %v", err)
	}
}

func TestNewClient_RequiresTokenProvider(t *testing.T) {
	if _, err := NewClient("ws://localhost:1/ws"); err == nil {
		t.Fatal("缺少 token 来源时应返回错误")
	}
}

func TestNewClient_InvalidWSURLForAuth(t *testing.T) {
	// WithAuth 要求 ws/wss scheme，非法 scheme 应在构造期报错。
	if _, err := NewClient("http://localhost:1/ws", WithAuth("u", "p")); err == nil {
		t.Fatal("WithAuth + 非 ws scheme 应返回错误")
	}
}

func TestWaitReady_AuthorizeSuccess(t *testing.T) {
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "uid-1")
	})
	cli, err := NewClient(wsURL(t, srv)+"/ws", WithBearerToken("tok"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cli.waitReady(ctx); err != nil {
		t.Fatalf("WaitReady 应成功: %v", err)
	}
	if got := cli.uid(); got != "uid-1" {
		t.Fatalf("UID 应为 uid-1，实际 %q", got)
	}
}

func TestWaitReady_AuthFailed(t *testing.T) {
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendInternalError(t, c)
	})
	cli, err := NewClient(wsURL(t, srv)+"/ws", WithBearerToken("bad"))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cli.waitReady(ctx); err == nil {
		t.Fatal("鉴权失败时 WaitReady 应返回错误")
	}
}

func TestClose_Idempotent(t *testing.T) {
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		<-time.After(time.Hour) // 保持连接，直到客户端关闭
	})
	cli, err := NewClient(wsURL(t, srv)+"/ws", WithBearerToken("t"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cli.waitReady(ctx); err != nil {
		t.Fatal(err)
	}
	if err := cli.Close(); err != nil {
		t.Fatalf("首次 Close 失败: %v", err)
	}
	if err := cli.Close(); err != nil {
		t.Fatalf("二次 Close 应幂等返回 nil: %v", err)
	}
}

func TestReconnect_Reauthorize(t *testing.T) {
	var authorizeCalls atomic.Int32
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		authorizeCalls.Add(1)
		sendSetUid(t, c, "u")
		if authorizeCalls.Load() == 1 {
			// 第一次连接立刻断开，模拟服务端掉线，触发客户端重连。
			return
		}
		// 之后保持连接，读到断线为止。
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				return
			}
		}
	})

	cli, err := NewClient(wsURL(t, srv)+"/ws",
		WithBearerToken("t"),
		WithReconnectBackoff(backoff.NewConstantBackOff(5*time.Millisecond)))
	if err != nil {
		t.Fatal(err)
	}
	defer cli.Close()

	// 轮询等待第二次连接完成重鉴权。
	deadline := time.Now().Add(3 * time.Second)
	for authorizeCalls.Load() < 2 {
		if time.Now().After(deadline) {
			t.Fatalf("客户端未在期限内重连重鉴权，calls=%d", authorizeCalls.Load())
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestWSURLToHTTPBase(t *testing.T) {
	cases := []struct{ in, want string }{
		{"ws://localhost:4000/ws", "http://localhost:4000"},
		{"wss://mars.example.com/ws", "https://mars.example.com"},
	}
	for _, tc := range cases {
		got, err := wsURLToHTTPBase(tc.in)
		if err != nil {
			t.Fatalf("%s: 意外错误 %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("%s: 期望 %s，实际 %s", tc.in, tc.want, got)
		}
	}
	if _, err := wsURLToHTTPBase("http://localhost:4000/ws"); err == nil {
		t.Fatal("非 ws/wss scheme 应报错")
	}
}
