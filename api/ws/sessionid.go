package ws

// sessionid.go 提供终端会话 id（sessionID）的自动生成。
//
// 服务端（internal/services/websocket/terminal.go isValidSessionID）只校验
// sessionID 以 "<namespace>-<pod>-<container>:" 为前缀，randomID 部分自由。
// OpenTerminal 用本文件自动生成合法且唯一的 sessionID，调用方无需手拼。

import (
	"crypto/rand"
	"fmt"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
)

// newSessionID 生成符合服务端 "<namespace>-<pod>-<container>:<randomID>" 格式的
// 会话 id。randomID 用 crypto/rand 取 8 字节 hex，保证同名容器重复打开时
// 不会撞 sessionID（服务端按 sessionID 登记 pty 会话）。
func newSessionID(c *websocket_pb.Container) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%s-%s-%s:%x", c.Namespace, c.Pod, c.Container, b[:])
}
