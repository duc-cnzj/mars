package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenSession 验证进程内会话的完整链路：Connect 建连、tools/list 按身份裁剪，
// 且 ctx 中的用户确实传到工具 handler（get_current_user 回吐注入的用户）。
func TestOpenSession(t *testing.T) {
	server := NewServer(Deps{Logger: mlog.NewForConfig(nil)})

	t.Run("普通用户看不到管理员工具", func(t *testing.T) {
		ctx := biz.SetUser(context.Background(), &biz.UserInfo{Email: "user@example.com"})
		session, cleanup, err := server.OpenSession(ctx)
		require.NoError(t, err)
		defer cleanup()

		res, err := session.ListTools(ctx, nil)
		require.NoError(t, err)
		names := toolNames(res)
		assert.Contains(t, names, "get_current_user")
		assert.Contains(t, names, "list_namespaces")
		assert.NotContains(t, names, "list_repos")
		assert.NotContains(t, names, "update_repo_custom_fields")
	})

	t.Run("管理员拿到全量工具", func(t *testing.T) {
		ctx := biz.SetUser(context.Background(), &biz.UserInfo{Email: "admin@example.com", Roles: []string{"mars_admin"}})
		session, cleanup, err := server.OpenSession(ctx)
		require.NoError(t, err)
		defer cleanup()

		res, err := session.ListTools(ctx, nil)
		require.NoError(t, err)
		names := toolNames(res)
		assert.Contains(t, names, "list_repos")
		assert.Contains(t, names, "update_repo_custom_fields")
	})

	t.Run("工具 handler 能拿到 ctx 中的用户", func(t *testing.T) {
		ctx := biz.SetUser(context.Background(), &biz.UserInfo{Email: "who@example.com"})
		session, cleanup, err := server.OpenSession(ctx)
		require.NoError(t, err)
		defer cleanup()

		result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "get_current_user"})
		require.NoError(t, err)
		require.False(t, result.IsError)
		assert.Contains(t, firstText(t, result), "who@example.com")
	})
}

// TestInstructionsFor 两条链路共用同一份口径：管理员在通用排查说明之后
// 追加仓库管理段，普通用户绝不出现管理员专属工具的措辞。
func TestInstructionsFor(t *testing.T) {
	user := InstructionsFor(false)
	admin := InstructionsFor(true)
	assert.NotContains(t, user, "管理员专属工具")
	assert.Contains(t, admin, "管理员专属工具")
	assert.True(t, strings.HasPrefix(admin, user), "管理员说明必须以通用说明为前缀")
}

// TestOpenSession_ConnectError ctx 已取消时建连失败：错误带中文上下文返回，
// 且不返回半成品 session/cleanup，调用方不会误用未建立的会话。
func TestOpenSession_ConnectError(t *testing.T) {
	server := NewServer(Deps{Logger: mlog.NewForConfig(nil)})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	session, cleanup, err := server.OpenSession(ctx)
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Nil(t, cleanup)
}

// toolNames 提取 tools/list 返回的工具名。
func toolNames(res *sdk.ListToolsResult) []string {
	names := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	return names
}

// firstText 返回工具结果中的第一段文本内容。
func firstText(t *testing.T, res *sdk.CallToolResult) string {
	t.Helper()
	for _, content := range res.Content {
		if text, ok := content.(*sdk.TextContent); ok {
			return text.Text
		}
	}
	t.Fatal("工具结果不含文本内容")
	return ""
}
