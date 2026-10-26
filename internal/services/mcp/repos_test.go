package mcp

import (
	"context"
	"errors"
	"net/http"
	"testing"

	repopb "github.com/duc-cnzj/mars/api/v6/proto/repo"
	apitypes "github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToolset_ListRepos 覆盖管理员列仓库工具：门禁先于查询、入参原样搬运
// （含显式 false 过滤）、出参直接回吐服务模型、错误透传。
func TestToolset_ListRepos(t *testing.T) {
	t.Run("管理员按条件分页查询并回吐服务结果", func(t *testing.T) {
		e := newTestEnv(t)
		want := &repopb.ListResponse{
			Page: 2, PageSize: 5, Count: 6,
			Items: []*apitypes.RepoModel{{Id: 9, Name: "app-repo", NeedGitRepo: true}},
		}
		var got *repopb.ListRequest
		e.repoSvc.listFn = func(_ context.Context, req *repopb.ListRequest) (*repopb.ListResponse, error) {
			got = req
			return want, nil
		}
		disabled := false
		aiAgent := true

		_, out, err := e.toolset().listRepos(ctxWithUser(adminUser()), nil, listReposIn{
			paginationIn: paginationIn{Page: 2, PageSize: 5},
			Name:         "app",
			Enabled:      &disabled,
			AiAgent:      &aiAgent,
		})
		require.NoError(t, err)

		require.NotNil(t, got)
		assert.Equal(t, int32(2), got.GetPage())
		assert.Equal(t, int32(5), got.GetPageSize())
		assert.Equal(t, "app", got.GetName())
		require.NotNil(t, got.Enabled, "显式 false 是有效筛选条件，必须原样下发")
		assert.False(t, got.GetEnabled())
		require.NotNil(t, got.AiAgent)
		assert.True(t, got.GetAiAgent())
		assert.Same(t, want, out, "出参直接回吐服务模型，不另造投影")
	})

	t.Run("缺省入参不附加筛选条件", func(t *testing.T) {
		e := newTestEnv(t)
		var got *repopb.ListRequest
		e.repoSvc.listFn = func(_ context.Context, req *repopb.ListRequest) (*repopb.ListResponse, error) {
			got = req
			return &repopb.ListResponse{}, nil
		}

		_, out, err := e.toolset().listRepos(ctxWithUser(adminUser()), nil, listReposIn{})
		require.NoError(t, err)

		require.NotNil(t, got)
		// page/page_size 恒下发零值指针，由 RepoSvc 的 InitByDefault 回退到 1/15；
		// 两个布尔过滤保持 nil 才是「不筛选」，不能被零值 false 静默变成「只看停用」。
		assert.Equal(t, int32(0), got.GetPage())
		assert.Equal(t, int32(0), got.GetPageSize())
		assert.Nil(t, got.Enabled)
		assert.Nil(t, got.AiAgent)
		assert.Empty(t, got.GetName())
		assert.Equal(t, &repopb.ListResponse{}, out)
	})

	t.Run("非管理员被拒且不查仓库", func(t *testing.T) {
		e := newTestEnv(t)
		called := false
		e.repoSvc.listFn = func(context.Context, *repopb.ListRequest) (*repopb.ListResponse, error) {
			called = true
			return nil, nil
		}

		_, _, err := e.toolset().listRepos(ctxWithUser(normalUser()), nil, listReposIn{})
		assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
		assert.False(t, called, "门禁必须先于仓库查询")
	})

	t.Run("List 错误透传", func(t *testing.T) {
		e := newTestEnv(t)
		boom := errors.New("db down")
		e.repoSvc.listFn = func(context.Context, *repopb.ListRequest) (*repopb.ListResponse, error) {
			return nil, boom
		}

		_, out, err := e.toolset().listRepos(ctxWithUser(adminUser()), nil, listReposIn{})
		assert.ErrorIs(t, err, boom)
		assert.Nil(t, out)
	})
}

// TestServer_CallTool_ListRepos 端到端证明管理员专属工具在传输层真实可达：
// HTTP → 鉴权 → 按管理员选中 admin server → list_repos → RepoSvc → 回吐 JSON。
// 只测 tools/list 可见性不够——若 per-request server 选择在 tools/call 路径上
// 走偏（例如选到 user server），客户端会看到工具却在调用时报 unknown tool。
func TestServer_CallTool_ListRepos(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(adminUser())
	e.repoSvc.listFn = func(_ context.Context, req *repopb.ListRequest) (*repopb.ListResponse, error) {
		assert.Equal(t, "app", req.GetName())
		// 显式 false 是有效筛选条件，必须穿过传输层原样下发。
		require.NotNil(t, req.Enabled, "enabled=false 必须下发，不能被 omitempty 吞掉")
		assert.False(t, req.GetEnabled())
		return &repopb.ListResponse{
			Page: 1, PageSize: 15, Count: 1,
			Items: []*apitypes.RepoModel{{Id: 9, Name: "app-repo", NeedGitRepo: true}},
		}, nil
	}

	sessionID := e.initSession(t, testToken)
	code, _, body := e.post(t, sessionID, testToken,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_repos","arguments":{"name":"app","enabled":false}}}`)
	require.Equal(t, http.StatusOK, code, body)

	res := sseJSON(t, body)["result"].(map[string]any)
	require.NotEqual(t, true, res["isError"], "list_repos 不应报错: %v", body)
	content := res["content"].([]any)[0].(map[string]any)["text"].(string)
	assert.Contains(t, content, `"id":9`)
	assert.Contains(t, content, `"name":"app-repo"`)
	assert.Contains(t, content, `"count":1`)
}
