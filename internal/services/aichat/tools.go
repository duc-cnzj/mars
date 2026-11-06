package aichat

// tools.go 是服务端对话与 MCP 工具之间的适配层。对话循环只依赖 toolSession
// 这一个最小接口，生产实现包裹 go-sdk 的 ClientSession，单测可替换为内存 fake——
// 避免把 SDK 的协议细节（Content 联合类型、Params/Result 结构）泄漏进循环逻辑。

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// toolSpec 是对话循环看到的工具描述：名称、说明、JSON Schema 入参与只读标记。
// 只读标记决定是否需要人工确认——写工具（ReadOnlyHint=false）必须先暂停。
type toolSpec struct {
	// Name 是工具名，同时作为模型 function call 的 name。
	Name string
	// Description 是给模型看的用途说明，直接来自 MCP 工具定义。
	Description string
	// InputSchema 是 JSON Schema 原文，直接作为 OpenAI function.parameters。
	InputSchema json.RawMessage
	// ReadOnly 为 true 表示工具不修改环境，可直接执行。
	ReadOnly bool
}

// toolResult 是工具执行结果：Text 是工具返回原文，不做服务端截断；
// IsError 对应 MCP 的 isError 标记（业务失败，不是协议错误）。
type toolResult struct {
	// Text 是回吐给模型的文本内容。
	Text string
	// IsError 为 true 表示工具以业务错误结束（NotFound / PermissionDenied 等）。
	IsError bool
}

// toolSession 是对话循环依赖的工具会话抽象。生产实现是 mcp.Server 的进程内
// 会话（见 internal/services/mcp.OpenSession），单测用 fake 替换。
// 会话资源由 OpenSession 返回的 cleanup 释放，因此接口本身不暴露 Close。
type toolSession interface {
	// List 返回当前身份可见的工具集合（管理员与普通用户不同）。
	List(ctx context.Context) ([]toolSpec, error)
	// Call 执行一个工具。args 是模型给出的 JSON 参数原文。
	Call(ctx context.Context, name string, args json.RawMessage) (toolResult, error)
}

// mcpSession 是 toolSession 的 go-sdk 实现。
type mcpSession struct {
	session *sdk.ClientSession
}

// newMCPSession 把 go-sdk 的客户端会话包装成对话循环可用的 toolSession。
func newMCPSession(session *sdk.ClientSession) *mcpSession {
	return &mcpSession{session: session}
}

// List 把 SDK 的工具模型转成 toolSpec：只读标记取 annotations.readOnlyHint，
// 未声明 annotations 的工具按非只读处理（与 MCP 规范默认值一致）。
func (m *mcpSession) List(ctx context.Context) ([]toolSpec, error) {
	res, err := m.session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	specs := make([]toolSpec, 0, len(res.Tools))
	for _, tool := range res.Tools {
		// InputSchema 来自 JSON-RPC 解码，必然是合法 JSON，序列化不会失败。
		schema, _ := json.Marshal(tool.InputSchema)
		readOnly := tool.Annotations != nil && tool.Annotations.ReadOnlyHint
		specs = append(specs, toolSpec{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: schema,
			ReadOnly:    readOnly,
		})
	}
	return specs, nil
}

// Call 执行工具并把 MCP 的内容联盟收敛为纯文本。工具失败（isError）不返回
// Go error——那是模型需要看到的业务结果；只有协议层失败才返回 error。
func (m *mcpSession) Call(ctx context.Context, name string, args json.RawMessage) (toolResult, error) {
	var payload any
	if len(args) > 0 {
		if err := json.Unmarshal(args, &payload); err != nil {
			return toolResult{}, fmt.Errorf("工具 %s 的入参不是合法 JSON: %w", name, err)
		}
	}
	res, err := m.session.CallTool(ctx, &sdk.CallToolParams{Name: name, Arguments: payload})
	if err != nil {
		return toolResult{}, err
	}
	return toolResult{Text: collapseContent(res), IsError: res.IsError}, nil
}

// collapseContent 把 MCP 的内容联合类型收敛为纯文本：文本内容直接拼接，
// 图片/资源等非文本内容序列化后回吐（至少让模型知道返回了什么）；
// 全部为空时回退到 structuredContent。
func collapseContent(res *sdk.CallToolResult) string {
	var b strings.Builder
	for _, content := range res.Content {
		if text, ok := content.(*sdk.TextContent); ok {
			b.WriteString(text.Text)
			continue
		}
		raw, err := json.Marshal(content)
		if err == nil {
			b.WriteString(string(raw))
		}
	}
	if b.Len() == 0 && res.StructuredContent != nil {
		raw, err := json.Marshal(res.StructuredContent)
		if err == nil {
			b.WriteString(string(raw))
		}
	}
	return b.String()
}

// oneLine 把多行文本压成单行摘要，供确认条展示（不改变工具实际入参）。
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
