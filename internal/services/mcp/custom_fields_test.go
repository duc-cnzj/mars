package mcp

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	repopb "github.com/duc-cnzj/mars/api/v6/proto/repo"
	apitypes "github.com/duc-cnzj/mars/api/v6/proto/types"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// customFieldRepo 造一个带完整仓库配置的模型：字段更新只应改动 elements，
// 其余配置（values.yaml/branches/分区/休眠）都必须原样保留，测试据此做回归。
func customFieldRepo() *apitypes.RepoModel {
	return &apitypes.RepoModel{
		Id:           5,
		Name:         "app-repo",
		NeedGitRepo:  true,
		GitProjectId: 42,
		Description:  "应用仓库",
		MarsConfig: &mars.Config{
			ValuesYaml: "image:\n  tag: v1",
			Branches:   []string{"main", "release"},
			Elements: []*mars.Element{
				{Path: "app->replicas", Type: mars.ElementType_ElementTypeInputNumber, Default: "2", Description: "副本数", Order: 1},
				{Path: "app->image", Type: mars.ElementType_ElementTypeInput, Default: "nginx", Order: 7},
			},
			GroupSettings: []*mars.GroupSetting{{Name: "基础配置", Order: 1, Collapsed: true}},
			Hibernate:     &mars.Hibernate{Enabled: true, ReplicaPaths: []string{"app->replicas"}, WindowNames: []string{"night"}},
		},
	}
}

// assertGroupSettings 逐条比较分区配置的语义值：proto 消息带运行期状态字段，
// 直接用 assert.Equal 比较会受反射缓存影响（-race 下尤其明显）而误判。
func assertGroupSettings(t *testing.T, want []*mars.GroupSetting, got []*mars.GroupSetting, msgAndArgs ...any) {
	t.Helper()
	require.Len(t, got, len(want), msgAndArgs...)
	for i := range want {
		assert.Equal(t, want[i].GetName(), got[i].GetName(), "第 %d 个分区 name", i)
		assert.Equal(t, want[i].GetOrder(), got[i].GetOrder(), "第 %d 个分区 order", i)
		assert.Equal(t, want[i].GetCollapsed(), got[i].GetCollapsed(), "第 %d 个分区 collapsed", i)
	}
}

// TestToolset_GetRepoCustomFields 覆盖管理员读工具的门禁、寻址、投影与错误透传。
func TestToolset_GetRepoCustomFields(t *testing.T) {
	t.Run("管理员读取成功并投影为类型名", func(t *testing.T) {
		e := newTestEnv(t)
		var shownID int32
		e.repoSvc.showFn = func(_ context.Context, req *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			shownID = req.GetId()
			return &repopb.ShowResponse{Item: &apitypes.RepoModel{
				Id:   5,
				Name: "app-repo",
				MarsConfig: &mars.Config{Elements: []*mars.Element{{
					Path:             "app->replicas",
					Type:             mars.ElementType_ElementTypeInputNumber,
					Default:          "2",
					Description:      "副本数",
					SelectValues:     []string{"1", "2"},
					Order:            3,
					TextareaLanguage: "yaml",
					Group:            "基础配置",
				}}},
			}}, nil
		}

		_, out, err := e.toolset().getRepoCustomFields(ctxWithUser(adminUser()), nil, getRepoCustomFieldsIn{RepoID: 5})
		require.NoError(t, err)
		assert.Equal(t, int32(5), shownID)
		assert.Equal(t, repoCustomFieldsOut{
			RepoID:        5,
			RepoName:      "app-repo",
			Branches:      []string{},
			GroupSettings: []repoGroupSetting{},
			Elements: []repoCustomField{{
				Path:             "app->replicas",
				Type:             "InputNumber",
				Default:          "2",
				Description:      "副本数",
				SelectValues:     []string{"1", "2"},
				Order:            3,
				TextareaLanguage: "yaml",
				Group:            "基础配置",
			}},
		}, out)
	})

	t.Run("回吐 values.yaml、分支、分区与休眠配置", func(t *testing.T) {
		e := newTestEnv(t)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{Item: customFieldRepo()}, nil
		}

		_, out, err := e.toolset().getRepoCustomFields(ctxWithUser(adminUser()), nil, getRepoCustomFieldsIn{RepoID: 5})
		require.NoError(t, err)
		got := out.(repoCustomFieldsOut)
		assert.Equal(t, "image:\n  tag: v1", got.ValuesYaml)
		assert.Equal(t, []string{"main", "release"}, got.Branches)
		assert.Equal(t, []repoGroupSetting{{Name: "基础配置", Order: 1, Collapsed: true}}, got.GroupSettings)
		require.NotNil(t, got.Hibernate)
		assert.Equal(t, repoHibernate{
			Enabled:      true,
			ReplicaPaths: []string{"app->replicas"},
			WindowNames:  []string{"night"},
		}, *got.Hibernate)
		assert.Len(t, got.Elements, 2)
	})

	t.Run("非管理员被拒且不读仓库", func(t *testing.T) {
		e := newTestEnv(t)
		called := false
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			called = true
			return nil, nil
		}

		_, _, err := e.toolset().getRepoCustomFields(ctxWithUser(normalUser()), nil, getRepoCustomFieldsIn{RepoID: 5})
		assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
		assert.False(t, called, "门禁必须先于仓库读取")
	})

	t.Run("repo_id 非法", func(t *testing.T) {
		for _, id := range []int32{0, -1} {
			e := newTestEnv(t)
			_, _, err := e.toolset().getRepoCustomFields(ctxWithUser(adminUser()), nil, getRepoCustomFieldsIn{RepoID: id})
			assert.Equal(t, codes.InvalidArgument, status.Code(err), "repo_id=%d 必须被拒", id)
		}
	})

	t.Run("MarsConfig 为 nil 时 elements 是空数组而非 null", func(t *testing.T) {
		e := newTestEnv(t)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{Item: &apitypes.RepoModel{Id: 5, Name: "empty-repo"}}, nil
		}
		_, out, err := e.toolset().getRepoCustomFields(ctxWithUser(adminUser()), nil, getRepoCustomFieldsIn{RepoID: 5})
		require.NoError(t, err)
		got := out.(repoCustomFieldsOut)
		assert.NotNil(t, got.Elements, "空列表必须序列化成 []，不能是 null")
		assert.Empty(t, got.Elements)
		assert.Empty(t, got.ValuesYaml)
		assert.NotNil(t, got.Branches, "空分支列表必须序列化成 []，不能是 null")
		assert.Empty(t, got.Branches)
		assert.NotNil(t, got.GroupSettings, "空分区列表必须序列化成 []，不能是 null")
		assert.Empty(t, got.GroupSettings)
		assert.Nil(t, got.Hibernate, "未配置休眠时保持 null，与 enabled=false 区分")
	})

	t.Run("Show 错误透传", func(t *testing.T) {
		e := newTestEnv(t)
		boom := errors.New("db down")
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return nil, boom
		}
		_, _, err := e.toolset().getRepoCustomFields(ctxWithUser(adminUser()), nil, getRepoCustomFieldsIn{RepoID: 5})
		assert.ErrorIs(t, err, boom)
	})

	t.Run("Show 返回空项目视为不存在", func(t *testing.T) {
		e := newTestEnv(t)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{}, nil
		}
		_, _, err := e.toolset().getRepoCustomFields(ctxWithUser(adminUser()), nil, getRepoCustomFieldsIn{RepoID: 5})
		assert.True(t, errs.IsNotFound(err))
	})
}

// TestServer_CallTool_GetRepoCustomFieldsReturnsConfig 端到端证明管理员读工具经传输层
// 回吐完整仓库配置：values.yaml、分支、分区、休眠与字段定义必须真实出现在 JSON 里——
// 字段名或嵌套形状写错时，模型拿到的就是残缺配置，只断言结构体发现不了这类问题。
func TestServer_CallTool_GetRepoCustomFieldsReturnsConfig(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(adminUser())
	e.repoSvc.showFn = func(_ context.Context, req *repopb.ShowRequest) (*repopb.ShowResponse, error) {
		assert.Equal(t, int32(5), req.GetId())
		return &repopb.ShowResponse{Item: customFieldRepo()}, nil
	}

	sessionID := e.initSession(t, testToken)
	code, _, body := e.post(t, sessionID, testToken,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_repo_custom_fields","arguments":{"repo_id":5}}}`)
	require.Equal(t, http.StatusOK, code, body)

	res := sseJSON(t, body)["result"].(map[string]any)
	require.NotEqual(t, true, res["isError"], "get_repo_custom_fields 不应报错: %v", body)
	content := res["content"].([]any)[0].(map[string]any)["text"].(string)
	assert.Contains(t, content, `"repo_name":"app-repo"`)
	assert.Contains(t, content, `"values_yaml":"image:\n  tag: v1"`)
	assert.Contains(t, content, `"branches":["main","release"]`)
	assert.Contains(t, content, `"group_settings":[{"name":"基础配置","order":1,"collapsed":true}]`)
	// encoding/json 默认对 > 做 HTML 转义（\u003e），解码后仍是 app->replicas；
	// 这里按线上真实字节断言，避免测试与序列化行为脱节。
	assert.Contains(t, content, `"hibernate":{"enabled":true,"replica_paths":["app-\u003ereplicas"],"window_names":["night"]}`)
	assert.Contains(t, content, `"path":"app-\u003ereplicas"`)
}

// TestServer_CallTool_UpdateRepoCustomFieldsGroupSettings 端到端锁定分区配置的入参
// 字段名（upsert_group_settings / name / collapsed）与落库结果：JSON 标签一旦写错，
// 模型按 schema 传参会被静默忽略，只有走真实 tools/call 链路才能发现。
func TestServer_CallTool_UpdateRepoCustomFieldsGroupSettings(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(adminUser())
	var captured *repopb.UpdateRequest
	e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
		return &repopb.ShowResponse{Item: customFieldRepo()}, nil
	}
	e.repoSvc.updateFn = func(_ context.Context, req *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
		captured = req
		return &repopb.UpdateResponse{Item: &apitypes.RepoModel{
			Id: req.GetId(), Name: req.GetName(), MarsConfig: req.GetMarsConfig(),
		}}, nil
	}

	sessionID := e.initSession(t, testToken)
	code, _, body := e.post(t, sessionID, testToken,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"update_repo_custom_fields","arguments":{`+
			`"repo_id":5,`+
			`"upsert_elements":[{"path":"app->log_level","type":"Select","default":"info","description":"日志级别","select_values":["info","debug"],"group":"日志配置"}],`+
			`"upsert_group_settings":[{"name":"日志配置","collapsed":true}]}}}`)
	require.Equal(t, http.StatusOK, code, body)
	res := sseJSON(t, body)["result"].(map[string]any)
	require.NotEqual(t, true, res["isError"], "update_repo_custom_fields 不应报错: %v", body)

	require.NotNil(t, captured, "必须调用 RepoSvc.Update")
	cfg := captured.GetMarsConfig()
	require.Len(t, cfg.GetElements(), 3)
	assert.Equal(t, "日志配置", cfg.GetElements()[2].GetGroup())
	require.Len(t, cfg.GetGroupSettings(), 2)
	assertGroupSettings(t, []*mars.GroupSetting{{Name: "日志配置", Order: 2, Collapsed: true}}, cfg.GetGroupSettings()[1:])
}

// TestServer_CallTool_UpdateRepoCustomFieldsExplicitZero 锁定分区配置指针入参的零值语义
// 穿透 MCP 传输层：`collapsed: false` / `order: 0` 与「未传」在 Go 里零值相同，若解码
// 层不能区分（或 omitempty 误生效），管理员就无法把折叠态改回 false、把分区排到最前。
// 用「原值 order=1、collapsed=true」做底，显式传 0/false 必须覆盖成功才有区分度。
func TestServer_CallTool_UpdateRepoCustomFieldsExplicitZero(t *testing.T) {
	e := newTestEnv(t)
	e.expectAuth(adminUser())
	var captured *repopb.UpdateRequest
	e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
		return &repopb.ShowResponse{Item: customFieldRepo()}, nil
	}
	e.repoSvc.updateFn = func(_ context.Context, req *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
		captured = req
		return &repopb.UpdateResponse{Item: &apitypes.RepoModel{
			Id: req.GetId(), Name: req.GetName(), MarsConfig: req.GetMarsConfig(),
		}}, nil
	}

	sessionID := e.initSession(t, testToken)
	code, _, body := e.post(t, sessionID, testToken,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"update_repo_custom_fields","arguments":{`+
			`"repo_id":5,"upsert_group_settings":[{"name":"基础配置","order":0,"collapsed":false}]}}}`)
	require.Equal(t, http.StatusOK, code, body)
	res := sseJSON(t, body)["result"].(map[string]any)
	require.NotEqual(t, true, res["isError"], "显式零值不应报错: %v", body)

	require.NotNil(t, captured, "必须调用 RepoSvc.Update")
	gs := captured.GetMarsConfig().GetGroupSettings()
	require.Len(t, gs, 1)
	assert.Equal(t, uint32(0), gs[0].GetOrder(), "显式 order=0 必须覆盖原值 1")
	assert.False(t, gs[0].GetCollapsed(), "显式 collapsed=false 必须覆盖原值 true")
}

// TestToolset_UpdateRepoCustomFields 覆盖写工具的增量语义：整条替换、末尾追加、
// 删除，以及「只改 elements、其余配置原样回填」这条整段替换链路的关键约束。
func TestToolset_UpdateRepoCustomFields(t *testing.T) {
	// captureUpdate 装配 Show + Update 替身，返回取值函数：取 UpdateRequest 前先断言
	// Update 确实被调用过——直接解引用 nil 只会得到空配置，把「压根没调用」伪装成断言失败。
	captureUpdate := func(t *testing.T, e *testEnv, current *apitypes.RepoModel) func() *repopb.UpdateRequest {
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{Item: current}, nil
		}
		var captured *repopb.UpdateRequest
		e.repoSvc.updateFn = func(_ context.Context, req *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
			captured = req
			return &repopb.UpdateResponse{Item: &apitypes.RepoModel{
				Id: req.GetId(), Name: req.GetName(), MarsConfig: req.GetMarsConfig(),
			}}, nil
		}
		return func() *repopb.UpdateRequest {
			t.Helper()
			require.NotNil(t, captured, "必须调用 RepoSvc.Update")
			return captured
		}
	}

	t.Run("替换已有字段并原值回填其余配置", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		captured := captureUpdate(t, e, current)

		_, out, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{
				Path: "app->replicas", Type: "InputNumber", Default: "3", Description: "副本数（新）",
			}},
		})
		require.NoError(t, err)

		req := captured()
		assert.Equal(t, int32(5), req.GetId())
		assert.Equal(t, "app-repo", req.GetName())
		assert.True(t, req.GetNeedGitRepo())
		require.NotNil(t, req.GitProjectId)
		assert.Equal(t, int32(42), req.GetGitProjectId())
		assert.Equal(t, "应用仓库", req.GetDescription())

		cfg := req.GetMarsConfig()
		assert.NotSame(t, current.GetMarsConfig(), cfg, "必须在克隆上修改，不能改 Show 返回的共享对象")
		// 除 elements 外的仓库配置必须原样保留。
		assert.Equal(t, "image:\n  tag: v1", cfg.GetValuesYaml())
		assert.Equal(t, []string{"main", "release"}, cfg.GetBranches())
		assertGroupSettings(t, []*mars.GroupSetting{{Name: "基础配置", Order: 1, Collapsed: true}}, cfg.GetGroupSettings())
		require.NotNil(t, cfg.GetHibernate())
		assert.True(t, cfg.GetHibernate().GetEnabled())
		assert.Equal(t, []string{"app->replicas"}, cfg.GetHibernate().GetReplicaPaths())
		// 目标字段整条替换，未传 order 时保留原位序；同列表的其它字段不受影响。
		require.Len(t, cfg.GetElements(), 2)
		assert.Equal(t, &mars.Element{
			Path: "app->replicas", Type: mars.ElementType_ElementTypeInputNumber,
			Default: "3", Description: "副本数（新）", Order: 1,
		}, cfg.GetElements()[0])
		assert.Equal(t, "app->image", cfg.GetElements()[1].GetPath())

		// 出参是更新后的投影：字段类型投影成类型名，其余仓库配置一并回吐。
		got := out.(repoCustomFieldsOut)
		assert.Equal(t, int32(5), got.RepoID)
		assert.Equal(t, "InputNumber", got.Elements[0].Type)
		assert.Equal(t, "image:\n  tag: v1", got.ValuesYaml)
		assert.Equal(t, []string{"main", "release"}, got.Branches)
		require.NotNil(t, got.Hibernate)
		assert.Equal(t, []string{"night"}, got.Hibernate.WindowNames)
	})

	t.Run("新增字段未传 order 追加到末尾", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{Path: "app->log_level", Type: "Select", Default: "info", Description: "日志级别", SelectValues: []string{"info", "debug"}}},
		})
		require.NoError(t, err)

		els := captured().GetMarsConfig().GetElements()
		require.Len(t, els, 3)
		assert.Equal(t, "app->log_level", els[2].GetPath())
		assert.Equal(t, uint32(8), els[2].GetOrder(), "新增字段 order = 当前最大值 7 + 1")
	})

	t.Run("新增字段显式 order 以入参为准", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())
		order := uint32(100)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{
				Path: "app->log_level", Type: "ElementTypeSelect", Default: "info", Description: "日志级别",
				SelectValues: []string{"info", "debug"}, Order: &order,
			}},
		})
		require.NoError(t, err)

		els := captured().GetMarsConfig().GetElements()
		require.Len(t, els, 3)
		assert.Equal(t, uint32(100), els[2].GetOrder())
	})

	t.Run("显式 order 之后追加的新字段仍排最后", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())
		order := uint32(100)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{
				{Path: "app->a", Type: "Input", Description: "字段 A", Order: &order},
				{Path: "app->b", Type: "Input", Description: "字段 B"},
			},
		})
		require.NoError(t, err)

		els := captured().GetMarsConfig().GetElements()
		require.Len(t, els, 4)
		assert.Equal(t, uint32(100), els[2].GetOrder())
		assert.Equal(t, uint32(101), els[3].GetOrder(), "不传 order 的新字段要排在显式 order 之后")
	})

	t.Run("替换已有字段抬高 order 后新增字段仍排最后", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())
		order := uint32(100)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{
				{Path: "app->image", Type: "Input", Description: "镜像", Order: &order},
				{Path: "app->b", Type: "Input", Description: "字段 B"},
			},
		})
		require.NoError(t, err)

		els := captured().GetMarsConfig().GetElements()
		require.Len(t, els, 3)
		assert.Equal(t, uint32(100), els[1].GetOrder())
		assert.Equal(t, uint32(101), els[2].GetOrder())
	})

	t.Run("group 恰好 64 字节可写入", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())
		group := strings.Repeat("a", maxGroupNameBytes)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{Path: "app->log_level", Type: "Input", Description: "日志级别", Group: group}},
		})
		require.NoError(t, err)

		els := captured().GetMarsConfig().GetElements()
		require.Len(t, els, 3)
		assert.Equal(t, group, els[2].GetGroup(), "边界值必须原样保留，不能被截断")
	})

	t.Run("新增字段与分区配置同时写入", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())
		collapsed := true

		_, out, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{
				Path: "app->log_level", Type: "Select", Default: "info", Description: "日志级别",
				SelectValues: []string{"info", "debug"}, Group: "日志配置",
			}},
			UpsertGroups: []repoGroupSettingIn{{Name: "日志配置", Collapsed: &collapsed}},
		})
		require.NoError(t, err)

		cfg := captured().GetMarsConfig()
		els := cfg.GetElements()
		require.Len(t, els, 3)
		assert.Equal(t, "日志配置", els[2].GetGroup())
		assertGroupSettings(t, []*mars.GroupSetting{
			{Name: "基础配置", Order: 1, Collapsed: true},
			{Name: "日志配置", Order: 2, Collapsed: true}, // 新增分区未传 order 时追加到当前最大值 1 + 1
		}, cfg.GetGroupSettings())

		got := out.(repoCustomFieldsOut)
		assert.Equal(t, []repoGroupSetting{
			{Name: "基础配置", Order: 1, Collapsed: true},
			{Name: "日志配置", Order: 2, Collapsed: true},
		}, got.GroupSettings, "出参必须回吐更新后的分区配置")
	})

	t.Run("更新已有分区保留未传字段并规范化名称", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())
		order := uint32(5)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID:       5,
			UpsertGroups: []repoGroupSettingIn{{Name: " 基础配置 ", Order: &order}},
		})
		require.NoError(t, err)

		// 名称裁剪后落库，未传的 collapsed 保留原值 true
		assertGroupSettings(t, []*mars.GroupSetting{{Name: "基础配置", Order: 5, Collapsed: true}},
			captured().GetMarsConfig().GetGroupSettings())
	})

	t.Run("显式 collapsed=false 覆盖原值", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())
		collapsed := false

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID:       5,
			UpsertGroups: []repoGroupSettingIn{{Name: "基础配置", Collapsed: &collapsed}},
		})
		require.NoError(t, err)

		// 未传的 order 保留原值 1
		assertGroupSettings(t, []*mars.GroupSetting{{Name: "基础配置", Order: 1, Collapsed: false}},
			captured().GetMarsConfig().GetGroupSettings())
	})

	t.Run("新增分区显式 order 以入参为准", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())
		order := uint32(0)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID:       5,
			Upsert:       []repoElementIn{{Path: "app->a", Type: "Input", Description: "字段 A", Group: "日志配置"}},
			UpsertGroups: []repoGroupSettingIn{{Name: "日志配置", Order: &order}},
		})
		require.NoError(t, err)

		groups := captured().GetMarsConfig().GetGroupSettings()
		require.Len(t, groups, 2)
		assertGroupSettings(t, []*mars.GroupSetting{{Name: "日志配置", Order: 0}}, groups[1:])
	})

	t.Run("删除未被引用的存量分区", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5, RemoveGroupNames: []string{" 基础配置 "},
		})
		require.NoError(t, err)

		assert.Empty(t, captured().GetMarsConfig().GetGroupSettings(), "分区配置应被删除，其余配置不受影响")
		require.Len(t, captured().GetMarsConfig().GetElements(), 2)
	})

	t.Run("分区 order 达到上限时拒绝自动追加", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		current.MarsConfig.GroupSettings[0].Order = math.MaxUint32
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{Item: current}, nil
		}
		e.repoSvc.updateFn = func(context.Context, *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
			t.Fatal("order 溢出时必须拒绝，不得调用 Update")
			return nil, nil
		}

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID:       5,
			Upsert:       []repoElementIn{{Path: "app->a", Type: "Input", Description: "字段 A", Group: "日志配置"}},
			UpsertGroups: []repoGroupSettingIn{{Name: "日志配置"}},
		})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "order 已达上限")
	})

	t.Run("分区名恰好 64 字节可写入", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())
		name := strings.Repeat("a", maxGroupNameBytes)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID:       5,
			Upsert:       []repoElementIn{{Path: "app->a", Type: "Input", Description: "字段 A", Group: name}},
			UpsertGroups: []repoGroupSettingIn{{Name: name}},
		})
		require.NoError(t, err)

		assertGroupSettings(t, []*mars.GroupSetting{
			{Name: "基础配置", Order: 1, Collapsed: true},
			{Name: name, Order: 2},
		}, captured().GetMarsConfig().GetGroupSettings())
	})

	t.Run("新增分区可被存量字段引用", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		current.MarsConfig.Elements[1].Group = "日志配置"
		captured := captureUpdate(t, e, current)
		collapsed := true

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID:       5,
			UpsertGroups: []repoGroupSettingIn{{Name: "日志配置", Collapsed: &collapsed}},
		})
		require.NoError(t, err)

		assertGroupSettings(t, []*mars.GroupSetting{
			{Name: "基础配置", Order: 1, Collapsed: true},
			{Name: "日志配置", Order: 2, Collapsed: true},
		}, captured().GetMarsConfig().GetGroupSettings())
		require.Len(t, captured().GetMarsConfig().GetElements(), 2, "元素未被触碰")
	})

	t.Run("分区改名：搬字段、建新分区、删旧分区一次完成", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		current.MarsConfig.Elements[0].Group = "旧分区"
		current.MarsConfig.GroupSettings = []*mars.GroupSetting{{Name: "旧分区", Order: 1, Collapsed: true}}
		captured := captureUpdate(t, e, current)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{
				Path: "app->replicas", Type: "InputNumber", Default: "2", Description: "副本数", Group: "新分区",
			}},
			UpsertGroups:     []repoGroupSettingIn{{Name: "新分区"}},
			RemoveGroupNames: []string{"旧分区"},
		})
		require.NoError(t, err)

		cfg := captured().GetMarsConfig()
		assert.Equal(t, "新分区", cfg.GetElements()[0].GetGroup())
		assertGroupSettings(t, []*mars.GroupSetting{{Name: "新分区", Order: 1}},
			cfg.GetGroupSettings(), "旧分区删除后重新从 1 起算追加")
	})

	t.Run("MarsConfig 为 nil 时仅写分区配置", func(t *testing.T) {
		e := newTestEnv(t)
		current := &apitypes.RepoModel{Id: 5, Name: "empty-repo"}
		captured := captureUpdate(t, e, current)
		collapsed := false

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID:       5,
			Upsert:       []repoElementIn{{Path: "app->a", Type: "Input", Description: "字段 A", Group: "日志配置"}},
			UpsertGroups: []repoGroupSettingIn{{Name: "日志配置", Collapsed: &collapsed}},
		})
		require.NoError(t, err)

		req := captured()
		assert.Nil(t, req.GitProjectId, "不需要 git 仓库时不回填绑定")
		require.NotNil(t, req.GetMarsConfig())
		assertGroupSettings(t, []*mars.GroupSetting{{Name: "日志配置", Order: 1, Collapsed: false}},
			req.GetMarsConfig().GetGroupSettings())
	})

	t.Run("删除仅路径带首尾空白的存量字段", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		current.MarsConfig.Elements[1].Path = " app->image "
		captured := captureUpdate(t, e, current)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5, RemovePaths: []string{"app->image"},
		})
		require.NoError(t, err)

		els := captured().GetMarsConfig().GetElements()
		require.Len(t, els, 1)
		assert.Equal(t, "app->replicas", els[0].GetPath())
	})

	t.Run("upsert 命中带空白的存量字段并规范化 path", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		current.MarsConfig.Elements[1].Path = " app->image "
		captured := captureUpdate(t, e, current)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{Path: "app->image", Type: "Input", Default: "busybox", Description: "镜像（新）"}},
		})
		require.NoError(t, err)

		els := captured().GetMarsConfig().GetElements()
		require.Len(t, els, 2, "命中存量字段必须原位替换，不能追加出近重复")
		assert.Equal(t, "app->image", els[1].GetPath(), "落库 path 必须规范化为裁剪后的形态")
		assert.Equal(t, "镜像（新）", els[1].GetDescription())
	})

	t.Run("存量 path 撞车时拒绝覆盖", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		current.MarsConfig.Elements = append(current.MarsConfig.Elements,
			&mars.Element{Path: " app->image ", Type: mars.ElementType_ElementTypeInput, Description: "旧字段", Order: 8})
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{Item: current}, nil
		}
		e.repoSvc.updateFn = func(context.Context, *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
			t.Fatal("撞车时必须拒绝，不得调用 Update")
			return nil, nil
		}

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{Path: "app->image", Type: "Input", Description: "镜像"}},
		})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "多个仅空白不同")
	})

	t.Run("存量 path 撞车时拒绝删除", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		current.MarsConfig.Elements = append(current.MarsConfig.Elements,
			&mars.Element{Path: " app->image ", Type: mars.ElementType_ElementTypeInput, Description: "旧字段", Order: 8})
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{Item: current}, nil
		}
		e.repoSvc.updateFn = func(context.Context, *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
			t.Fatal("撞车时必须拒绝，不得调用 Update")
			return nil, nil
		}

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5, RemovePaths: []string{"app->image"},
		})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "无法安全删除")
	})

	t.Run("局部撞车不阻断不相关字段的修改", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		current.MarsConfig.Elements = append(current.MarsConfig.Elements,
			&mars.Element{Path: " app->image ", Type: mars.ElementType_ElementTypeInput, Description: "旧字段", Order: 8})
		captured := captureUpdate(t, e, current)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{Path: "app->new", Type: "Input", Description: "新字段"}},
		})
		require.NoError(t, err)

		els := captured().GetMarsConfig().GetElements()
		require.Len(t, els, 4, "未命中的存量字段（包括脏 path）必须原样保留")
		assert.Equal(t, "app->new", els[3].GetPath())
		assert.Equal(t, uint32(9), els[3].GetOrder(), "追加 order = 现存最大值 8 + 1")
	})

	t.Run("order 达到上限时拒绝自动追加", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		current.MarsConfig.Elements[1].Order = math.MaxUint32
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{Item: current}, nil
		}
		e.repoSvc.updateFn = func(context.Context, *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
			t.Fatal("order 溢出时必须拒绝，不得调用 Update")
			return nil, nil
		}

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{Path: "app->new", Type: "Input", Description: "新字段"}},
		})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "order 已达上限")
	})

	t.Run("移除最大 order 字段后仍可自动追加", func(t *testing.T) {
		e := newTestEnv(t)
		current := customFieldRepo()
		current.MarsConfig.Elements[1].Order = math.MaxUint32
		captured := captureUpdate(t, e, current)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID:      5,
			RemovePaths: []string{"app->image"},
			Upsert:      []repoElementIn{{Path: "app->new", Type: "Input", Description: "新字段"}},
		})
		require.NoError(t, err)

		els := captured().GetMarsConfig().GetElements()
		require.Len(t, els, 2)
		assert.Equal(t, uint32(1), els[0].GetOrder())
		assert.Equal(t, uint32(2), els[1].GetOrder(), "被删除字段的 order 不应继续占用最大值")
	})

	t.Run("删除字段成功", func(t *testing.T) {
		e := newTestEnv(t)
		captured := captureUpdate(t, e, customFieldRepo())

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5, RemovePaths: []string{"  app->image  "},
		})
		require.NoError(t, err)

		els := captured().GetMarsConfig().GetElements()
		require.Len(t, els, 1)
		assert.Equal(t, "app->replicas", els[0].GetPath())
	})

	t.Run("MarsConfig 为 nil 时新建配置", func(t *testing.T) {
		e := newTestEnv(t)
		// 不绑定 git 仓库时 GitProjectId 传 nil：data 层见到 nil 会清空绑定，
		// 与「本来就没有绑定」语义一致。
		current := &apitypes.RepoModel{Id: 5, Name: "empty-repo"}
		captured := captureUpdate(t, e, current)

		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{Path: "app->replicas", Type: "InputNumber", Default: "1", Description: "副本数"}},
		})
		require.NoError(t, err)

		req := captured()
		assert.Nil(t, req.GitProjectId, "不需要 git 仓库时不回填绑定")
		require.NotNil(t, req.GetMarsConfig())
		require.Len(t, req.GetMarsConfig().GetElements(), 1)
		assert.Equal(t, uint32(1), req.GetMarsConfig().GetElements()[0].GetOrder())
	})

	t.Run("Update 错误透传且不回吐投影", func(t *testing.T) {
		e := newTestEnv(t)
		boom := errors.New("update failed")
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{Item: customFieldRepo()}, nil
		}
		e.repoSvc.updateFn = func(context.Context, *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
			return nil, boom
		}

		_, out, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{Path: "app->replicas", Type: "InputNumber", Default: "1", Description: "副本数"}},
		})
		assert.ErrorIs(t, err, boom)
		assert.Nil(t, out)
	})

	t.Run("Show 错误透传", func(t *testing.T) {
		e := newTestEnv(t)
		boom := errors.New("db down")
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return nil, boom
		}
		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{Path: "app->replicas", Type: "InputNumber", Default: "1", Description: "副本数"}},
		})
		assert.ErrorIs(t, err, boom)
	})

	t.Run("Show 返回空项目视为不存在", func(t *testing.T) {
		e := newTestEnv(t)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{}, nil
		}
		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5,
			Upsert: []repoElementIn{{Path: "app->replicas", Type: "InputNumber", Default: "1", Description: "副本数"}},
		})
		assert.True(t, errs.IsNotFound(err))
	})
}

// TestToolset_UpdateRepoCustomFields_Rejects 覆盖写工具的全部拒绝场景：非法入参
// 必须在触达仓库读/写之前返回，且不产生任何副作用。
func TestToolset_UpdateRepoCustomFields_Rejects(t *testing.T) {
	t.Run("非管理员", func(t *testing.T) {
		e := newTestEnv(t)
		called := false
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			called = true
			return nil, nil
		}
		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(normalUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5, RemovePaths: []string{"app->replicas"},
		})
		assert.ErrorIs(t, err, errs.ErrorPermissionDenied)
		assert.False(t, called, "门禁必须先于仓库读取")
	})

	// 纯入参校验用例：Show 替身刻意留空——若校验缺失导致继续读仓库会直接 panic，
	// 以此证明「非法字段不触达仓库」。
	noRepoCases := []struct {
		name string
		in   updateRepoCustomFieldsIn
		want string
	}{
		{"repo_id 为 0", updateRepoCustomFieldsIn{RepoID: 0}, "repo_id"},
		{"repo_id 为负数", updateRepoCustomFieldsIn{RepoID: -1}, "repo_id"},
		{"upsert 与 remove 都为空", updateRepoCustomFieldsIn{RepoID: 5}, "至少提供"},
		{"upsert path 为空", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{{Type: "Input"}}}, "path 不能为空"},
		{"upsert path 仅空白", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{{Path: "   ", Type: "Input"}}}, "path 不能为空"},
		{"path 以分隔符开头", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{{Path: "->a", Type: "Input"}}}, "空路径段"},
		{"path 以分隔符结尾", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{{Path: "a->", Type: "Input"}}}, "空路径段"},
		{"path 含连续分隔符", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{{Path: "a->->b", Type: "Input"}}}, "空路径段"},
		{"path 段前导空白", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{{Path: "a-> b", Type: "Input"}}}, "首尾不能有空白"},
		{"path 段尾随空白", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{{Path: "a ->b", Type: "Input"}}}, "首尾不能有空白"},
		{"未知类型", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{{Path: "a", Type: "Bogus"}}}, "未知的自定义字段类型"},
		{"group 超过 64 字节", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "Input", Group: strings.Repeat("a", 65)},
		}}, "分区名过长"},
		{"group 按字节而非字符计算", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "Input", Group: strings.Repeat("分", 22)},
		}}, "分区名过长"},
		{"upsert 内 path 重复", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "Input", Description: "字段 A"}, {Path: "a", Type: "Input", Description: "字段 A"},
		}}, "path 重复"},
		{"path 去空格后重复", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: " a ", Type: "Input", Description: "字段 A"}, {Path: "a", Type: "Input", Description: "字段 A"},
		}}, "path 重复"},
		{"description 为空", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "Input"},
		}}, "描述不能为空"},
		{"description 仅空白", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "Input", Description: "   "},
		}}, "描述不能为空"},
		{"InputNumber 默认值为空", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "InputNumber", Description: "数量"},
		}}, "默认值不能为空"},
		{"Switch 默认值为空", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "Switch", Description: "开关"},
		}}, "默认值不能为空"},
		{"Switch 默认值非 bool", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "Switch", Default: "maybe", Description: "开关"},
		}}, "应该为 bool"},
		{"InputNumber 默认值非整数", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "InputNumber", Default: "1.5", Description: "数量"},
		}}, "应该为整数"},
		{"Select 默认值不在白名单", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "Select", Default: "x", Description: "选项", SelectValues: []string{"y"}},
		}}, "必须在"},
		{"NumberSelect 默认值非整数", updateRepoCustomFieldsIn{RepoID: 5, Upsert: []repoElementIn{
			{Path: "a", Type: "NumberSelect", Default: "x", Description: "数值选项", SelectValues: []string{"x"}},
		}}, "非 number 类型"},
		{"分区 name 为空", updateRepoCustomFieldsIn{RepoID: 5, UpsertGroups: []repoGroupSettingIn{{}}}, "分区 name 不能为空"},
		{"分区 name 仅空白", updateRepoCustomFieldsIn{RepoID: 5, UpsertGroups: []repoGroupSettingIn{{Name: "   "}}}, "分区 name 不能为空"},
		{"分区名超过 64 字节", updateRepoCustomFieldsIn{RepoID: 5, UpsertGroups: []repoGroupSettingIn{
			{Name: strings.Repeat("a", 65)},
		}}, "名称过长"},
		{"分区名按字节而非字符计算", updateRepoCustomFieldsIn{RepoID: 5, UpsertGroups: []repoGroupSettingIn{
			{Name: strings.Repeat("分", 22)},
		}}, "名称过长"},
		{"upsert 分区 name 重复", updateRepoCustomFieldsIn{RepoID: 5, UpsertGroups: []repoGroupSettingIn{
			{Name: "日志配置"}, {Name: " 日志配置 "},
		}}, "name 重复"},
		{"remove_group_names 含空 name", updateRepoCustomFieldsIn{RepoID: 5, RemoveGroupNames: []string{"   "}}, "空 name"},
		{"remove_group_names 重复", updateRepoCustomFieldsIn{RepoID: 5, RemoveGroupNames: []string{"基础配置", " 基础配置 "}}, "重复"},
		{"分区同时 upsert 与 remove", updateRepoCustomFieldsIn{RepoID: 5,
			UpsertGroups: []repoGroupSettingIn{{Name: "基础配置"}}, RemoveGroupNames: []string{" 基础配置 "},
		}, "同时出现"},
	}
	for _, tc := range noRepoCases {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestEnv(t)
			_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, tc.in)
			require.Error(t, err)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Contains(t, err.Error(), tc.want)
		})
	}

	// 删除类校验依赖当前字段列表，需要 Show 替身。
	removeCases := []struct {
		name string
		in   updateRepoCustomFieldsIn
		want string
	}{
		{"删除不存在的 path", updateRepoCustomFieldsIn{RepoID: 5, RemovePaths: []string{"app->missing"}}, "不存在"},
		{"remove 包含空 path", updateRepoCustomFieldsIn{RepoID: 5, RemovePaths: []string{""}}, "空 path"},
		{"remove 仅空白 path", updateRepoCustomFieldsIn{RepoID: 5, RemovePaths: []string{"   "}}, "空 path"},
		{"remove path 重复", updateRepoCustomFieldsIn{RepoID: 5, RemovePaths: []string{"app->image", "app->image"}}, "重复"},
		{"同一 path 同时 upsert 与 remove", updateRepoCustomFieldsIn{
			RepoID:      5,
			Upsert:      []repoElementIn{{Path: "app->image", Type: "Input", Description: "镜像"}},
			RemovePaths: []string{"app->image"},
		}, "同时出现"},
		{"去空格后同一 path 冲突", updateRepoCustomFieldsIn{
			RepoID:      5,
			Upsert:      []repoElementIn{{Path: " app->image ", Type: "Input", Description: "镜像"}},
			RemovePaths: []string{"app->image  "},
		}, "同时出现"},
	}
	for _, tc := range removeCases {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestEnv(t)
			e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
				return &repopb.ShowResponse{Item: customFieldRepo()}, nil
			}
			e.repoSvc.updateFn = func(context.Context, *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
				t.Fatal("校验失败时不得调用 Update")
				return nil, nil
			}
			_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, tc.in)
			require.Error(t, err)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Contains(t, err.Error(), tc.want)
		})
	}

	// 分区类校验依赖当前配置与字段归属，同样需要 Show 替身；校验失败不得调用 Update。
	groupCases := []struct {
		name   string
		mutate func(*apitypes.RepoModel)
		in     updateRepoCustomFieldsIn
		want   string
	}{
		{
			"删除不存在的分区",
			func(*apitypes.RepoModel) {},
			updateRepoCustomFieldsIn{RepoID: 5, RemoveGroupNames: []string{"不存在"}},
			"不存在",
		},
		{
			"删除仍被字段引用的分区",
			func(r *apitypes.RepoModel) { r.MarsConfig.Elements[0].Group = "基础配置" },
			updateRepoCustomFieldsIn{RepoID: 5, RemoveGroupNames: []string{"基础配置"}},
			"仍被自定义字段",
		},
		{
			// 元素 group 带首尾空白（脏数据）时也要按裁剪后的名称判定引用，
			// 否则会误删配置、把字段推进无配置分区。
			"元素分区名带空白仍阻止删除分区",
			func(r *apitypes.RepoModel) { r.MarsConfig.Elements[0].Group = " 基础配置 " },
			updateRepoCustomFieldsIn{RepoID: 5, RemoveGroupNames: []string{"基础配置"}},
			"仍被自定义字段",
		},
		{
			"新增分区未被任何字段引用",
			func(*apitypes.RepoModel) {},
			updateRepoCustomFieldsIn{RepoID: 5, UpsertGroups: []repoGroupSettingIn{{Name: "日志配置"}}},
			"未被任何自定义字段引用",
		},
		{
			"存量分区撞车时拒绝更新",
			func(r *apitypes.RepoModel) {
				r.MarsConfig.GroupSettings = append(r.MarsConfig.GroupSettings,
					&mars.GroupSetting{Name: " 基础配置 ", Order: 2})
			},
			updateRepoCustomFieldsIn{RepoID: 5, UpsertGroups: []repoGroupSettingIn{{Name: "基础配置"}}},
			"无法安全更新",
		},
		{
			"存量分区撞车时拒绝删除",
			func(r *apitypes.RepoModel) {
				r.MarsConfig.GroupSettings = append(r.MarsConfig.GroupSettings,
					&mars.GroupSetting{Name: " 基础配置 ", Order: 2})
			},
			updateRepoCustomFieldsIn{RepoID: 5, RemoveGroupNames: []string{"基础配置"}},
			"无法安全删除",
		},
	}
	for _, tc := range groupCases {
		t.Run(tc.name, func(t *testing.T) {
			e := newTestEnv(t)
			current := customFieldRepo()
			tc.mutate(current)
			e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
				return &repopb.ShowResponse{Item: current}, nil
			}
			e.repoSvc.updateFn = func(context.Context, *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
				t.Fatal("校验失败时不得调用 Update")
				return nil, nil
			}
			_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, tc.in)
			require.Error(t, err)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Contains(t, err.Error(), tc.want)
		})
	}

	// need_git_repo=true 但未绑定 git_project_id 的仓库无法通过 Update：data 层
	// 会用 0 去解析 GitLab 项目并失败（nil 与 0 等价）。必须在写库前给出可操作的错误。
	t.Run("需要 git 仓库但未绑定 git_project_id", func(t *testing.T) {
		e := newTestEnv(t)
		e.repoSvc.showFn = func(context.Context, *repopb.ShowRequest) (*repopb.ShowResponse, error) {
			return &repopb.ShowResponse{Item: &apitypes.RepoModel{Id: 5, Name: "no-binding", NeedGitRepo: true}}, nil
		}
		e.repoSvc.updateFn = func(context.Context, *repopb.UpdateRequest) (*repopb.UpdateResponse, error) {
			t.Fatal("前置校验失败时不得调用 Update")
			return nil, nil
		}
		_, _, err := e.toolset().updateRepoCustomFields(ctxWithUser(adminUser()), nil, updateRepoCustomFieldsIn{
			RepoID: 5, Upsert: []repoElementIn{{Path: "app->replicas", Type: "InputNumber", Default: "1", Description: "副本数"}},
		})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "git_project_id")
	})
}

// TestElementTypeMapping 锁定枚举名与类型名的双向映射：读工具靠它给出可读类型名，
// 写工具靠它反解入参；枚举新增而映射未补时必须显式暴露为 Unknown。
func TestElementTypeMapping(t *testing.T) {
	assert.Equal(t, "Switch", elementTypeName(mars.ElementType_ElementTypeSwitch))
	assert.Equal(t, "Unknown", elementTypeName(mars.ElementType_ElementTypeUnknown))

	// 防漂移：proto 枚举每新增一个取值，映射表与反解都必须同步，否则模型会看到
	// "Unknown" 或传进来的类型名反解失败。这里直接遍历生成代码的枚举表。
	for value, name := range mars.ElementType_name {
		typ := mars.ElementType(value)
		if typ == mars.ElementType_ElementTypeUnknown {
			continue
		}
		assert.NotEqual(t, "Unknown", elementTypeName(typ), "枚举 %s 未登记对外类型名", name)
		parsed, ok := parseElementType(name)
		assert.True(t, ok, "枚举全名 %s 必须能反解", name)
		assert.Equal(t, typ, parsed, "枚举 %s 反解结果不一致", name)
	}

	typ, ok := parseElementType("Input")
	assert.True(t, ok)
	assert.Equal(t, mars.ElementType_ElementTypeInput, typ)

	// 全名与前后空格都要能反解。
	typ, ok = parseElementType(" ElementTypeNumberRadio ")
	assert.True(t, ok)
	assert.Equal(t, mars.ElementType_ElementTypeNumberRadio, typ)

	_, ok = parseElementType("Bogus")
	assert.False(t, ok)
}
