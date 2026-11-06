package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseSkill 验证 frontmatter 与正文解析、CRLF 归一化与各错误分支。
func TestParseSkill(t *testing.T) {
	t.Run("正常解析", func(t *testing.T) {
		sk, err := parseSkill([]byte("---\nname: demo\ndescription: 演示\n---\n# 正文\n内容\n"))
		require.NoError(t, err)
		assert.Equal(t, "demo", sk.Name)
		assert.Equal(t, "演示", sk.Description)
		assert.Equal(t, "# 正文\n内容", sk.Body)
	})

	t.Run("CRLF 归一化", func(t *testing.T) {
		sk, err := parseSkill([]byte("---\r\nname: demo\r\n---\r\n正文\r\n"))
		require.NoError(t, err)
		assert.Equal(t, "正文", sk.Body)
	})

	t.Run("缺少 frontmatter", func(t *testing.T) {
		_, err := parseSkill([]byte("正文"))
		assert.Error(t, err)
	})

	t.Run("frontmatter 未闭合", func(t *testing.T) {
		_, err := parseSkill([]byte("---\nname: demo\n"))
		assert.Error(t, err)
	})

	t.Run("frontmatter 非法 YAML", func(t *testing.T) {
		_, err := parseSkill([]byte("---\nname: [\n---\n正文"))
		assert.Error(t, err)
	})

	t.Run("正文为空", func(t *testing.T) {
		_, err := parseSkill([]byte("---\nname: demo\n---\n   "))
		assert.Error(t, err)
	})
}

// TestNewSkillRegistry 验证目录加载、排序与各类结构错误。
func TestNewSkillRegistry(t *testing.T) {
	t.Run("加载并按名称排序", func(t *testing.T) {
		fsys := fstest.MapFS{
			"skills/b/SKILL.md": {Data: []byte("---\nname: b\ndescription: B\n---\nB 正文")},
			"skills/a/SKILL.md": {Data: []byte("---\nname: a\ndescription: A\n---\nA 正文")},
			"skills/a/other.md": {Data: []byte("忽略")},
		}
		registry, err := newSkillRegistry(fsys, "skills")
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, registry.order)
		assert.Contains(t, registry.catalog(), "- a：A")
	})

	t.Run("缺少 name", func(t *testing.T) {
		fsys := fstest.MapFS{
			"skills/a/SKILL.md": {Data: []byte("---\ndescription: A\n---\n正文")},
		}
		_, err := newSkillRegistry(fsys, "skills")
		assert.Error(t, err)
	})

	t.Run("名称重复", func(t *testing.T) {
		fsys := fstest.MapFS{
			"skills/a/SKILL.md": {Data: []byte("---\nname: dup\n---\nA")},
			"skills/b/SKILL.md": {Data: []byte("---\nname: dup\n---\nB")},
		}
		_, err := newSkillRegistry(fsys, "skills")
		assert.Error(t, err)
	})

	t.Run("根目录不存在", func(t *testing.T) {
		_, err := newSkillRegistry(os.DirFS("/definitely-not-exists-mars"), ".")
		assert.Error(t, err)
	})

	t.Run("读取 SKILL.md 失败", func(t *testing.T) {
		fsys := readFileErrorFS{fstest.MapFS{
			"skills/a/SKILL.md": {Data: []byte("---\nname: a\n---\n正文")},
		}}
		_, err := newSkillRegistry(fsys, "skills")
		assert.ErrorIs(t, err, errReadBoom)
	})

	t.Run("正文解析失败", func(t *testing.T) {
		fsys := fstest.MapFS{
			"skills/a/SKILL.md": {Data: []byte("没有 frontmatter")},
		}
		_, err := newSkillRegistry(fsys, "skills")
		assert.Error(t, err)
	})
}

// errReadBoom 是读取文件失败的注入错误。
var errReadBoom = errors.New("read boom")

// readFileErrorFS 让 WalkDir 正常遍历、但读取文件恒失败。
type readFileErrorFS struct {
	fstest.MapFS
}

// ReadFile 恒返回注入的错误。
func (readFileErrorFS) ReadFile(string) ([]byte, error) {
	return nil, errReadBoom
}

// TestMustRegistryPanic 验证内嵌内容损坏时启动即 panic，而不是静默少一个 skill。
func TestMustRegistryPanic(t *testing.T) {
	fsys := fstest.MapFS{
		"skills/a/SKILL.md": {Data: []byte("没有 frontmatter")},
	}
	assert.Panics(t, func() { mustRegistry(fsys, "skills") })
}

// TestEmbeddedSkills 验证随二进制内嵌的两份手册可加载，且未知名称返回错误。
func TestEmbeddedSkills(t *testing.T) {
	registry := mustLoadEmbeddedSkills()
	assert.Equal(t, []string{"crashloop-triage", "mars-resource-model"}, registry.order)

	body, err := registry.load("crashloop-triage")
	require.NoError(t, err)
	assert.Contains(t, body, "CrashLoopBackOff")

	_, err = registry.load("unknown")
	assert.Error(t, err)

	// 空目录时 catalog 返回空串，调用方据此省略整段提示。
	empty := &skillRegistry{}
	assert.Empty(t, empty.catalog())
}

// TestSkillToolSpec 验证 load_skill 的 schema 枚举与只读标记。
func TestSkillToolSpec(t *testing.T) {
	registry := mustLoadEmbeddedSkills()
	spec := registry.toolSpec()
	require.True(t, spec.ReadOnly)
	assert.Equal(t, skillToolName, spec.Name)

	var schema struct {
		Properties struct {
			Name struct {
				Enum []string `json:"enum"`
			} `json:"name"`
		} `json:"properties"`
	}
	require.NoError(t, json.Unmarshal(spec.InputSchema, &schema))
	assert.Equal(t, registry.order, schema.Properties.Name.Enum)
}

// TestToolboxSkill 验证本地工具分发与 MCP 会话分发。
func TestToolboxSkill(t *testing.T) {
	session := &fakeToolSession{
		specs: []toolSpec{{Name: "list_projects", ReadOnly: true}},
		callFn: func(context.Context, string, json.RawMessage) (toolResult, error) {
			return toolResult{Text: "mcp-result"}, nil
		},
	}
	tools := &toolbox{session: session, skills: mustLoadEmbeddedSkills()}

	specs, err := tools.list(context.Background())
	require.NoError(t, err)
	assert.Len(t, specs, 2)
	assert.Equal(t, skillToolName, specs[1].Name)

	result, err := tools.call(context.Background(), "list_projects", json.RawMessage(`{}`))
	require.NoError(t, err)
	assert.Equal(t, "mcp-result", result.Text)
	assert.Equal(t, []string{"list_projects"}, session.calls)

	result, err = tools.call(context.Background(), skillToolName, json.RawMessage(`{"name":"crashloop-triage"}`))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Text, "CrashLoopBackOff")

	result, _ = tools.call(context.Background(), skillToolName, json.RawMessage(`{"name":"unknown"}`))
	assert.True(t, result.IsError)

	result, _ = tools.call(context.Background(), skillToolName, json.RawMessage(`not-json`))
	assert.True(t, result.IsError)

	result, _ = tools.call(context.Background(), skillToolName, nil)
	assert.True(t, result.IsError)
}

// TestToolboxListError 验证会话列工具失败时错误向上返回。
func TestToolboxListError(t *testing.T) {
	tools := &toolbox{
		session: &fakeToolSession{listErr: assert.AnError},
		skills:  mustLoadEmbeddedSkills(),
	}
	_, err := tools.list(context.Background())
	assert.ErrorIs(t, err, assert.AnError)
}
