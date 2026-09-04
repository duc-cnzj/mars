package ws

// send.go 定义客户端出站帧封装。所有方法把参数组装成对应的输入 proto
// 并复用 writeMsg 以二进制帧写出；字段与 api/proto/websocket 定义对齐。

import (
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
)

// execShell 请求在指定容器内拉起一个交互终端 shell。sessionID 是不透明关联键，
// 由 newSessionID 生成，服务端只校验非空。
func (c *Client) execShell(container *websocket_pb.Container, sessionID string) error {
	return c.writeMsg(&websocket_pb.WsHandleExecShellInput{
		Type:      websocket_pb.Type_HandleExecShell,
		Container: container,
		SessionId: sessionID,
	})
}

// sendStdin 向终端会话写入 stdin 数据（击键/粘贴）。
func (c *Client) sendStdin(sessionID string, data []byte) error {
	return c.writeMsg(&websocket_pb.TerminalMessageInput{
		Type: websocket_pb.Type_HandleExecShellMsg,
		Message: &websocket_pb.TerminalMessage{
			Op:        "stdin",
			SessionId: sessionID,
			Data:      data,
		},
	})
}

// resize 调整终端会话的窗口尺寸。
func (c *Client) resize(sessionID string, height, width uint32) error {
	return c.writeMsg(&websocket_pb.TerminalMessageInput{
		Type: websocket_pb.Type_HandleExecShellMsg,
		Message: &websocket_pb.TerminalMessage{
			Op:        "resize",
			SessionId: sessionID,
			Height:    height,
			Width:     width,
		},
	})
}

// closeShell 请求服务端主动关闭指定终端会话。
func (c *Client) closeShell(sessionID string) error {
	return c.writeMsg(&websocket_pb.TerminalMessageInput{
		Type: websocket_pb.Type_HandleCloseShell,
		Message: &websocket_pb.TerminalMessage{
			Op:        "close",
			SessionId: sessionID,
		},
	})
}
