package ws

// send_test.go：验证各出站方法序列化成的二进制帧能在服务端正确反解出字段。
// 每个用例的服务端 handler 读鉴权+发 SetUid 后，读一条帧并按期望类型反解断言。

import (
	"context"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// readFrame 从服务端连接读一条二进制帧并反解到 m。
func readFrame(t *testing.T, c *websocket.Conn, m proto.Message) {
	t.Helper()
	_, msg, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("读取帧失败: %v", err)
	}
	if err := proto.Unmarshal(msg, m); err != nil {
		t.Fatalf("反解帧失败: %v", err)
	}
}

// newReadyClient 起一个"读鉴权+发 SetUid"的直通服务端并返回已就绪客户端。
func newReadyClient(t *testing.T, onConn func(c *websocket.Conn)) *Client {
	t.Helper()
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		if onConn != nil {
			onConn(c)
		}
		// 保持连接直到客户端关闭。
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cli.waitReady(ctx); err != nil {
		t.Fatalf("客户端未就绪: %v", err)
	}
	return cli
}

func TestSend_ExecShell(t *testing.T) {
	received := make(chan struct{})
	cli := newReadyClient(t, func(c *websocket.Conn) {
		var in websocket_pb.WsHandleExecShellInput
		readFrame(t, c, &in)
		if in.Type != websocket_pb.Type_HandleExecShell {
			t.Fatalf("type 期望 HandleExecShell，实际 %v", in.Type)
		}
		if in.Container.GetNamespace() != "ns" || in.Container.GetPod() != "pod" || in.Container.GetContainer() != "c" {
			t.Fatalf("container 字段不匹配: %v", in.Container)
		}
		if in.SessionId != "ns-pod-c:1" {
			t.Fatalf("session_id 期望 ns-pod-c:1，实际 %q", in.SessionId)
		}
		close(received)
	})
	if err := cli.execShell(&websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"}, "ns-pod-c:1"); err != nil {
		t.Fatal(err)
	}
	<-received
}

func TestSend_TerminalMessages(t *testing.T) {
	// 依次验证 stdin / resize / close 三帧。
	expect := []struct {
		op   string
		data []byte
		h, w uint32
	}{
		{"stdin", []byte("ls -la\n"), 0, 0},
		{"resize", nil, 40, 120},
		{"close", nil, 0, 0},
	}
	done := make(chan struct{})
	cli := newReadyClient(t, func(c *websocket.Conn) {
		defer close(done)
		for i, e := range expect {
			var in websocket_pb.TerminalMessageInput
			readFrame(t, c, &in)
			if in.Message.GetOp() != e.op {
				t.Fatalf("[%d] op 期望 %q，实际 %q", i, e.op, in.Message.GetOp())
			}
			if string(in.Message.GetData()) != string(e.data) {
				t.Fatalf("[%d] data 不匹配: %q vs %q", i, in.Message.GetData(), e.data)
			}
			if in.Message.GetHeight() != e.h || in.Message.GetWidth() != e.w {
				t.Fatalf("[%d] 尺寸不匹配: %v", i, in.Message)
			}
		}
	})
	if err := cli.sendStdin("sid", []byte("ls -la\n")); err != nil {
		t.Fatal(err)
	}
	if err := cli.resize("sid", 40, 120); err != nil {
		t.Fatal(err)
	}
	if err := cli.closeShell("sid"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("服务端未收齐 3 帧")
	}
}
