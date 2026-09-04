package ws

// terminal_test.go：验证终端高层抽象 OpenTerminal 的完整生命周期——
// 发起 shell、接收 stdout/toast、服务端关闭触发 Done、Write 转发 stdin。

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// sendShell 发送一条 WsHandleShellResponse（终端输出/关闭帧）。
func sendShell(t *testing.T, c *websocket.Conn, wsType websocket_pb.Type, op, sid string, data []byte) {
	t.Helper()
	msg, _ := proto.Marshal(&websocket_pb.WsHandleShellResponse{
		Metadata: &websocket_pb.Metadata{Type: wsType},
		TerminalMessage: &websocket_pb.TerminalMessage{
			Op:        op,
			SessionId: sid,
			Data:      data,
		},
	})
	if err := c.WriteMessage(websocket.BinaryMessage, msg); err != nil {
		t.Fatalf("发送 shell 帧失败: %v", err)
	}
}

// assertTerminalID 校验 OpenTerminal 自动生成的 sessionID 符合
// "<namespace>-<pod>-<container>:" 前缀（服务端校验规则）。
func assertTerminalID(t *testing.T, id string) {
	t.Helper()
	if !strings.HasPrefix(id, "ns-pod-c:") {
		t.Fatalf("sessionID 应带 'ns-pod-c:' 前缀，实际 %q", id)
	}
}

func TestTerminal_FullLifecycle(t *testing.T) {
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		var in websocket_pb.WsHandleExecShellInput
		readFrame(t, c, &in) // OpenTerminal 内部发出的 ExecShell
		sid := in.SessionId
		sendShell(t, c, websocket_pb.Type_HandleExecShell, "", sid, nil) // shell-open → 确认 opened
		sendShell(t, c, websocket_pb.Type_HandleExecShellMsg, "stdout", sid, []byte("out-1"))
		sendShell(t, c, websocket_pb.Type_HandleExecShellMsg, "toast", sid, []byte("note"))
		sendShell(t, c, websocket_pb.Type_HandleCloseShell, "stdout", sid, []byte("bye"))
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
	defer cli.Close()
	ready(t, cli)

	term, err := cli.OpenTerminal(context.Background(), &websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	assertTerminalID(t, term.ID())

	select {
	case d := <-term.Stdout():
		if string(d) != "out-1" {
			t.Fatalf("stdout 期望 out-1，实际 %q", d)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stdout 通道未收到数据")
	}
	select {
	case d := <-term.Toast():
		if string(d) != "note" {
			t.Fatalf("toast 期望 note，实际 %q", d)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("toast 通道未收到数据")
	}
	select {
	case <-term.Done():
		// 服务端关闭帧触发 Done
	case <-time.After(2 * time.Second):
		t.Fatal("服务端关闭帧后 Done 未关闭")
	}
}

func TestTerminal_Write(t *testing.T) {
	received := make(chan string, 1)
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		var exec websocket_pb.WsHandleExecShellInput
		readFrame(t, c, &exec)
		var stdin websocket_pb.TerminalMessageInput
		readFrame(t, c, &stdin)
		if stdin.Message.GetOp() != "stdin" {
			t.Fatalf("op 期望 stdin，实际 %q", stdin.Message.GetOp())
		}
		if stdin.Message.GetSessionId() != exec.SessionId {
			t.Fatalf("stdin 应带自动生成的 sessionID %q，实际 %q", exec.SessionId, stdin.Message.GetSessionId())
		}
		received <- string(stdin.Message.GetData())
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
	defer cli.Close()
	ready(t, cli)

	term, err := cli.OpenTerminal(context.Background(), &websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()
	if n, err := term.Write([]byte("echo hi\n")); err != nil || n != 8 {
		t.Fatalf("Write 返回 %d,%v（期望 8,nil）", n, err)
	}
	select {
	case d := <-received:
		if d != "echo hi\n" {
			t.Fatalf("stdin 期望 'echo hi\\n'，实际 %q", d)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("服务端未收到 stdin 帧")
	}
}

// TestTerminal_Pump 验证数据面编排快乐路径：in 字节源 → 远端 stdin、
// 远端 stdout/toast → 消费闭包，且返回的 stop 幂等可安全多次调用。
func TestTerminal_Pump(t *testing.T) {
	outCh := make(chan []byte, 4)
	toastCh := make(chan []byte, 4)
	stdinCh := make(chan string, 1)

	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		var exec websocket_pb.WsHandleExecShellInput
		readFrame(t, c, &exec)
		sid := exec.SessionId
		sendShell(t, c, websocket_pb.Type_HandleExecShell, "", sid, nil) // shell-open

		// 收一条 pump 转发的 stdin 帧。
		var stdin websocket_pb.TerminalMessageInput
		readFrame(t, c, &stdin)
		stdinCh <- string(stdin.Message.GetData())

		// 回 stdout / toast 帧。
		sendShell(t, c, websocket_pb.Type_HandleExecShellMsg, "stdout", sid, []byte("out-1"))
		sendShell(t, c, websocket_pb.Type_HandleExecShellMsg, "toast", sid, []byte("note"))

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
	defer cli.Close()
	ready(t, cli)

	term, err := cli.OpenTerminal(context.Background(), &websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	stop := term.Pump(strings.NewReader("ab\n"),
		func(d []byte) { outCh <- d },
		func(d []byte) { toastCh <- d },
	)

	select {
	case d := <-stdinCh:
		if d != "ab\n" {
			t.Fatalf("stdin 期望 'ab\\n'，实际 %q", d)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pump 未把 in 转发成 stdin 帧")
	}
	select {
	case d := <-outCh:
		if string(d) != "out-1" {
			t.Fatalf("stdout 闭包期望 out-1，实际 %q", d)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stdout 闭包未收到数据")
	}
	select {
	case d := <-toastCh:
		if string(d) != "note" {
			t.Fatalf("toast 闭包期望 note，实际 %q", d)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("toast 闭包未收到数据")
	}
	stop() // 幂等：二次调用不 panic
	stop()
}

// TestTerminal_Pump_NilHandlers 验证 stdout/toast 闭包传 nil 时不启转发、
// in 读到 EOF 静默退出，stop 幂等。
func TestTerminal_Pump_NilHandlers(t *testing.T) {
	stdinCh := make(chan string, 1)
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		var exec websocket_pb.WsHandleExecShellInput
		readFrame(t, c, &exec)
		sendShell(t, c, websocket_pb.Type_HandleExecShell, "", exec.SessionId, nil)
		var stdin websocket_pb.TerminalMessageInput
		readFrame(t, c, &stdin)
		stdinCh <- string(stdin.Message.GetData())
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
	defer cli.Close()
	ready(t, cli)

	term, err := cli.OpenTerminal(context.Background(), &websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	stop := term.Pump(strings.NewReader("x"), nil, nil)
	select {
	case d := <-stdinCh:
		if d != "x" {
			t.Fatalf("stdin 期望 x，实际 %q", d)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pump 未把 in 转发成 stdin 帧")
	}
	stop()
	stop()
}

// TestTerminal_Pump_WriteError 验证远端已关（Write 失败）时 in 转发静默停止。
func TestTerminal_Pump_WriteError(t *testing.T) {
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		var exec websocket_pb.WsHandleExecShellInput
		readFrame(t, c, &exec)
		sendShell(t, c, websocket_pb.Type_HandleExecShell, "", exec.SessionId, nil)
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
	ready(t, cli)

	term, err := cli.OpenTerminal(context.Background(), &websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"})
	if err != nil {
		t.Fatal(err)
	}

	// 关闭客户端连接后，pump 的 in 转发 Write 会失败并静默退出，不 panic。
	stop := term.Pump(strings.NewReader("y"), func([]byte) {}, func([]byte) {})
	_ = cli.Close()
	stop()
}

// TestTerminal_Pump_NoRaw 验证 WithRawMode(false) 时跳过 raw 切换分支，stop 幂等。
func TestTerminal_Pump_NoRaw(t *testing.T) {
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		var exec websocket_pb.WsHandleExecShellInput
		readFrame(t, c, &exec)
		sendShell(t, c, websocket_pb.Type_HandleExecShell, "", exec.SessionId, nil)
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
	defer cli.Close()
	ready(t, cli)

	term, err := cli.OpenTerminal(context.Background(), &websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	stop := term.Pump(strings.NewReader(""), func([]byte) {}, func([]byte) {}, WithRawMode(false))
	stop()
	stop()
}

// TestTerminal_Pump_NonTTYFile 验证 in 是普通文件（非 tty）时不切 raw（IsTerminal 为 false），
// in 内容仍被转发、stop 幂等。
func TestTerminal_Pump_NonTTYFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "pump")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString("hi"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	stdinCh := make(chan string, 1)
	srv := newWsServer(t, func(c *websocket.Conn) {
		defer c.Close()
		readAuthorize(t, c)
		sendSetUid(t, c, "u")
		var exec websocket_pb.WsHandleExecShellInput
		readFrame(t, c, &exec)
		sendShell(t, c, websocket_pb.Type_HandleExecShell, "", exec.SessionId, nil)
		var stdin websocket_pb.TerminalMessageInput
		readFrame(t, c, &stdin)
		stdinCh <- string(stdin.Message.GetData())
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
	defer cli.Close()
	ready(t, cli)

	term, err := cli.OpenTerminal(context.Background(), &websocket_pb.Container{Namespace: "ns", Pod: "pod", Container: "c"})
	if err != nil {
		t.Fatal(err)
	}
	defer term.Close()

	stop := term.Pump(f, func([]byte) {}, func([]byte) {})
	select {
	case d := <-stdinCh:
		if d != "hi" {
			t.Fatalf("stdin 期望 hi，实际 %q", d)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pump 未把文件内容转发成 stdin 帧")
	}
	stop()
	stop()
}
