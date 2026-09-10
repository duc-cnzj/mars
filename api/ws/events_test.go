package ws

// events_test.go：验证入站帧按 type/sessionID 分发与订阅解绑。用 goSend 通道
// 协调，避免连接时序竞争。

import (
	"context"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// sendMeta 以 WsMetadataResponse 包装一个 metadata 发送给客户端。
func sendMeta(t *testing.T, c *websocket.Conn, m *websocket_pb.Metadata) {
	t.Helper()
	data, _ := proto.Marshal(&websocket_pb.WsMetadataResponse{Metadata: m})
	if err := c.WriteMessage(websocket.BinaryMessage, data); err != nil {
		t.Fatalf("发送帧失败: %v", err)
	}
}

// newCoordClient 起一个服务端：读鉴权→发 SetUid→等 goSend→执行 sendFn。
func newCoordClient(t *testing.T, sendFn func(c *websocket.Conn)) (*Client, chan struct{}) {
	t.Helper()
	goSend := make(chan struct{})
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		<-goSend
		sendFn(c)
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				return
			}
		}
	})
	cli, err := NewClient(wsURL(t, srv)+"/ws", WithBearerToken("t"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cli.Close() })
	return cli, goSend
}

// ready 等待客户端就绪。
func ready(t *testing.T, cli *Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cli.waitReady(ctx); err != nil {
		t.Fatalf("客户端未就绪: %v", err)
	}
}

func TestOnType_Dispatch(t *testing.T) {
	got := make(chan *Event, 1)
	cli, goSend := newCoordClient(t, func(c *websocket.Conn) {
		sendMeta(t, c, &websocket_pb.Metadata{Type: websocket_pb.Type_ProcessPercent, Percent: 50})
	})
	defer cli.onType(websocket_pb.Type_ProcessPercent, func(ev *Event) { got <- ev })()
	ready(t, cli)
	close(goSend)
	select {
	case ev := <-got:
		if ev.Metadata.GetPercent() != 50 {
			t.Fatalf("percent 期望 50，实际 %d", ev.Metadata.GetPercent())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("OnType 未收到帧")
	}
}

func TestOnSession_Dispatch(t *testing.T) {
	got := make(chan *Event, 1)
	cli, goSend := newCoordClient(t, func(c *websocket.Conn) {
		data, _ := proto.Marshal(&websocket_pb.WsHandleShellResponse{
			Metadata: &websocket_pb.Metadata{Type: websocket_pb.Type_HandleExecShellMsg},
			TerminalMessage: &websocket_pb.TerminalMessage{
				Op:        "stdout",
				SessionId: "sid-1",
				Data:      []byte("hello"),
			},
		})
		_ = c.WriteMessage(websocket.BinaryMessage, data)
	})
	defer cli.onSession("sid-1", func(ev *Event) { got <- ev })()
	ready(t, cli)
	close(goSend)
	select {
	case <-got:
		// 命中即通过
	case <-time.After(2 * time.Second):
		t.Fatal("OnSession 未收到帧")
	}
}

func TestUnsubscribe_StopsDelivery(t *testing.T) {
	got := make(chan *Event, 1)
	cli, goSend := newCoordClient(t, func(c *websocket.Conn) {
		sendMeta(t, c, &websocket_pb.Metadata{Type: websocket_pb.Type_ProcessPercent})
	})
	cancel := cli.onType(websocket_pb.Type_ProcessPercent, func(ev *Event) { got <- ev })
	ready(t, cli)
	cancel() // 先解绑再放行，应不再派发
	close(goSend)
	select {
	case <-got:
		t.Fatal("解绑后不应再收到帧")
	case <-time.After(200 * time.Millisecond):
		// 通过
	}
}
