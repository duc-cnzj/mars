package aichat

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/services/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOneLine 验证多行文本被压成单行。
func TestOneLine(t *testing.T) {
	assert.Equal(t, "a b c", oneLine("a\n  b\tc "))
	assert.Empty(t, oneLine("  \n "))
}

// TestCollapseContent 验证 MCP 内容联合类型的收敛规则。
func TestCollapseContent(t *testing.T) {
	t.Run("文本直接拼接", func(t *testing.T) {
		res := &sdk.CallToolResult{Content: []sdk.Content{
			&sdk.TextContent{Text: "a"},
			&sdk.TextContent{Text: "b"},
		}}
		assert.Equal(t, "ab", collapseContent(res))
	})

	t.Run("非文本序列化", func(t *testing.T) {
		res := &sdk.CallToolResult{Content: []sdk.Content{
			&sdk.ImageContent{Data: []byte("xx"), MIMEType: "image/png"},
		}}
		assert.Contains(t, collapseContent(res), "image/png")
	})

	t.Run("内容为空时回退 structuredContent", func(t *testing.T) {
		res := &sdk.CallToolResult{StructuredContent: map[string]any{"ok": true}}
		assert.JSONEq(t, `{"ok":true}`, collapseContent(res))
	})
}

// TestMCPSessionAdapter 用真实进程内会话验证工具清单与调用适配。
func TestMCPSessionAdapter(t *testing.T) {
	server := mcp.NewServer(mcp.Deps{Logger: testLogger()})
	ctx := biz.SetUser(context.Background(), testUser("adapter@example.com"))
	session, cleanup, err := server.OpenSession(ctx)
	require.NoError(t, err)
	defer cleanup()

	adapter := newMCPSession(session)
	specs, err := adapter.List(ctx)
	require.NoError(t, err)

	readOnly := map[string]bool{}
	for _, spec := range specs {
		readOnly[spec.Name] = spec.ReadOnly
	}
	assert.True(t, readOnly["get_current_user"])
	assert.False(t, readOnly["restart_pod"])

	result, err := adapter.Call(ctx, "get_current_user", json.RawMessage(`{}`))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Text, "adapter@example.com")

	_, err = adapter.Call(ctx, "get_current_user", json.RawMessage(`not-json`))
	assert.Error(t, err)
}

// TestMCPSessionClosedErrors 验证底层会话关闭后，列工具与调工具都返回协议错误
// （而不是静默返回空清单/空结果）。
func TestMCPSessionClosedErrors(t *testing.T) {
	server := mcp.NewServer(mcp.Deps{Logger: testLogger()})
	ctx := biz.SetUser(context.Background(), testUser("closed@example.com"))
	session, cleanup, err := server.OpenSession(ctx)
	require.NoError(t, err)
	defer cleanup()

	adapter := newMCPSession(session)
	require.NoError(t, session.Close())

	_, err = adapter.List(ctx)
	assert.Error(t, err)

	_, err = adapter.Call(ctx, "get_current_user", json.RawMessage(`{}`))
	assert.Error(t, err)
}
