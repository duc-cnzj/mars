package biz

// aiagent_test.go 覆盖 agent 项目名/Deployment 名派生：可读邮箱部分与哈希后缀的
// 稳定性、净化规则与长度约束，以及 helm fullname 规则的复刻（含 chart 名判断、
// 63 字符截断与尾部连字符清理）。

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/hasher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentProjectName 项目名 = agent-<可读本地部分>-<12 位哈希>：可读且唯一稳定。
func TestAgentProjectName(t *testing.T) {
	name := AgentProjectName("alice@mars.local")
	assert.Equal(t, AgentProjectPrefix+"alice-"+hasher.Hash("alice@mars.local")[:12], name)
	assert.Equal(t, name, AgentProjectName("alice@mars.local"))
	assert.NotEqual(t, name, AgentProjectName("bob@mars.local"))
	// 本地部分相同、域名不同：哈希后缀必须区分开。
	assert.NotEqual(t, name, AgentProjectName("alice@other.local"))
}

// TestAgentProjectName_SanitizedLocalPart 非 DNS 字符折叠为单个连字符并去首尾，
// 本地部分为空时回落 user。
func TestAgentProjectName_SanitizedLocalPart(t *testing.T) {
	assert.Equal(t, AgentProjectPrefix+"john-doe-test-"+hasher.Hash("john.doe+test@mars.local")[:12],
		AgentProjectName("john.doe+test@mars.local"))
	assert.Equal(t, AgentProjectPrefix+"a-b-"+hasher.Hash("A__B@mars.local")[:12],
		AgentProjectName("A__B@mars.local"))
	assert.Equal(t, AgentProjectPrefix+"user-"+hasher.Hash("@mars.local")[:12],
		AgentProjectName("@mars.local"))
	assert.Equal(t, AgentProjectPrefix+"user-"+hasher.Hash("张三@mars.local")[:12],
		AgentProjectName("张三@mars.local"))
}

// TestAgentProjectName_Length “可读部分 + 哈希”总长受控：项目名不超过 52 字符，
// 追加 "-mars-agent" 后仍不超过 k8s 63 字符上限，且截断不留尾部连字符。
func TestAgentProjectName_Length(t *testing.T) {
	longEmail := strings.Repeat("a", 100) + "@mars.local"
	name := AgentProjectName(longEmail)

	require.Len(t, name, 6+agentProjectLocalPartMaxLen+1+12)
	assert.LessOrEqual(t, len(name), 52)
	assert.True(t, strings.HasPrefix(name, AgentProjectPrefix+"a"))
	assert.True(t, strings.HasSuffix(name, hasher.Hash(longEmail)[:12]))
	assert.LessOrEqual(t, len(AgentDeploymentName(name)), maxK8sNameLength)
	assert.False(t, strings.HasSuffix(strings.TrimSuffix(name, "-"+hasher.Hash(longEmail)[:12]), "-"))
}

// TestAgentDeploymentName 复刻 helm fullname：不含 chart 名时追加 mars-agent，
// 含 chart 名时保持原样，超长时截断到 63 字符并去掉截断尾部的连字符。
func TestAgentDeploymentName(t *testing.T) {
	assert.Equal(t, "agent-abcdef123456-mars-agent", AgentDeploymentName("agent-abcdef123456"))
	assert.Equal(t, "custom-mars-agent-x", AgentDeploymentName("custom-mars-agent-x"))

	long := strings.Repeat("a", 62) + "-" + AgentChartName
	got := AgentDeploymentName(long)
	require.Len(t, got, 62)
	assert.False(t, strings.HasSuffix(got, "-"))

	over := strings.Repeat("a", 70)
	assert.Len(t, AgentDeploymentName(over), maxK8sNameLength)
}

// TestIsAgentProjectOf 判定必须同时满足「项目名 = 该邮箱派生的 agent 名」且邮箱非空：
// 空 creator（历史项目）或普通项目名都不能被误判成 agent 项目。
func TestIsAgentProjectOf(t *testing.T) {
	email := "alice@mars.local"
	assert.True(t, IsAgentProjectOf(AgentProjectName(email), email))
	assert.False(t, IsAgentProjectOf("my-app", email))
	assert.False(t, IsAgentProjectOf(AgentProjectName(email), "bob@mars.local"))
	assert.False(t, IsAgentProjectOf(AgentProjectName(""), ""))
}

// fakeAgentActivity 记录 Touch 调用的项目名，并可注入错误。
type fakeAgentActivity struct {
	touched string
	err     error
}

// Touch 记录项目名并返回预设错误。
func (f *fakeAgentActivity) Touch(_ context.Context, projectName string) error {
	f.touched = projectName
	return f.err
}

// IsActive 未被本测试使用，恒返回未活跃。
func (f *fakeAgentActivity) IsActive(context.Context, string) (bool, error) { return false, nil }

// TestTouchAgentActivity 活跃度刷新入口：按邮箱派生项目名；空邮箱跳过；失败只记日志。
func TestTouchAgentActivity(t *testing.T) {
	logger := mlog.NewForConfig(nil)

	t.Run("按邮箱派生的项目名刷新", func(t *testing.T) {
		activity := &fakeAgentActivity{}
		TouchAgentActivity(context.Background(), activity, logger, "alice@mars.local")
		assert.Equal(t, AgentProjectName("alice@mars.local"), activity.touched)
	})

	t.Run("空邮箱不触碰依赖", func(t *testing.T) {
		activity := &fakeAgentActivity{}
		TouchAgentActivity(context.Background(), activity, logger, "")
		assert.Empty(t, activity.touched)
	})

	t.Run("失败只记日志", func(t *testing.T) {
		activity := &fakeAgentActivity{err: errors.New("cache down")}
		assert.NotPanics(t, func() {
			TouchAgentActivity(context.Background(), activity, logger, "alice@mars.local")
		})
	})
}
