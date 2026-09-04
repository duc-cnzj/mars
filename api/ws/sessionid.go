package ws

// sessionid.go 提供终端会话 id（sessionID）的自动生成。
//
// sessionID 是客户端生成、服务端（internal/services/websocket）只校验非空的
// 不透明关联键：客户端拿它标识一次终端会话，服务端按它登记 pty 并路由各帧。
// OpenTerminal 用本文件自动生成合法且唯一的 sessionID，调用方无需手拼，
// 也无需关心任何格式约定。

import (
	"crypto/rand"
	"fmt"
)

// newSessionID 生成一个随机的、对服务端不透明的会话 id（16 字节 hex）。
// 服务端按 sessionID 登记 pty 会话并据此路由 stdin/resize/close/stdout 帧，
// 因此随机性保证同名容器重复打开时不会撞 id。
func newSessionID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%x", b[:])
}
