package ws

// events.go 定义入站事件模型与订阅 API：
//   - Event：一条已解码的服务端帧（type/metadata/raw，raw 供订阅方二次解码）
//   - 订阅注册表：按 type / sessionID（终端）两类订阅（onType/onSession）
//
// 订阅方法均返回解绑函数；注册与解绑都在 c.mu 写锁下进行，与 notify 的读锁并发安全。
// 订阅是 OpenTerminal 的内部机制，SDK 对外只暴露 OpenTerminal 一个入口（见 README）。

import (
	"fmt"
	"sync/atomic"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
)

// Event 是服务端推送的一条已解码事件：Metadata 是外层统一元数据，
// Raw 保留原始二进制供订阅方按 metadata.type 二次解码具体 payload。
type Event struct {
	// Type 是事件的协议类型（Metadata.Type 的快捷副本）。
	Type websocket_pb.Type
	// Metadata 是外层 WsMetadataResponse 的元数据。
	Metadata *websocket_pb.Metadata
	// Raw 是原始帧字节，可用 proto.Unmarshal 解码为具体响应类型。
	Raw []byte
}

// nextHandlerID 生成全局单调递增的订阅句柄 id。
var nextHandlerID atomic.Uint64

// onType 注册按协议类型触发的处理器，返回解绑函数。
func (c *Client) onType(t websocket_pb.Type, h func(*Event)) func() {
	return register(c, &c.typeHandlers, t, h)
}

// onSession 注册按终端会话 sessionID 触发的处理器，用于接收终端 stdout/toast 帧。
func (c *Client) onSession(sessionID string, h func(*Event)) func() {
	return register(c, &c.sessionHandlers, sessionID, h)
}

// register 是各订阅方法的公共实现：在 c.mu 写锁下向给定注册表挂一个处理器，
// 返回解绑闭包（幂等删除、空集合时清理 key）。K 为 key 类型（Type/string/int32），
// H 恒为 func(*Event)。
func register[K comparable](c *Client, m *map[K]map[string]func(*Event), key K, h func(*Event)) func() {
	id := fmt.Sprintf("%d", nextHandlerID.Add(1))
	c.mu.Lock()
	if (*m)[key] == nil {
		(*m)[key] = make(map[string]func(*Event))
	}
	(*m)[key][id] = h
	c.mu.Unlock()
	return func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if set := (*m)[key]; set != nil {
			delete(set, id)
			if len(set) == 0 {
				delete(*m, key)
			}
		}
	}
}
