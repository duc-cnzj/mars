package data

// aiagent_test.go 覆盖 biz.AgentActivity 的缓存实现：Touch 的禁用/成功/失败分支，
// IsActive 的命中/未命中/后端故障分支，以及 TTL 秒数向上取整。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestAgentActivity_Touch_Disabled 空项目名或空闲关闭禁用时不触碰缓存。
func TestAgentActivity_Touch_Disabled(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cache := NewMockCache(m)
	a := NewAgentActivityRepo(cache, &config.Config{AI: config.AIConfig{IdleTimeout: -time.Minute}})

	assert.NoError(t, a.Touch(context.TODO(), "agent-x"))
	assert.NoError(t, a.Touch(context.TODO(), ""))
}

// TestAgentActivity_Touch_WritesTTL 按 ai.idle_timeout 向上取整写入缓存标记。
func TestAgentActivity_Touch_WritesTTL(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cache := NewMockCache(m)
	a := &agentActivity{cache: cache, ttl: 90*time.Second + 500*time.Millisecond}

	cache.EXPECT().SetWithTTL(NewKey("agent_activity:%s", "agent-x"), []byte("1"), 91).Return(nil)

	assert.NoError(t, a.Touch(context.TODO(), "agent-x"))
}

// TestAgentActivity_Touch_DefaultTTL 未配置 idle_timeout 时按默认 30 分钟写入。
func TestAgentActivity_Touch_DefaultTTL(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cache := NewMockCache(m)
	a := NewAgentActivityRepo(cache, &config.Config{})

	cache.EXPECT().SetWithTTL(NewKey("agent_activity:%s", "agent-x"), []byte("1"), int(config.DefaultAgentIdleTimeout/time.Second)).Return(nil)

	assert.NoError(t, a.Touch(context.TODO(), "agent-x"))
}

// TestAgentActivity_Touch_CacheError 缓存写失败返回带上下文的包装错误。
func TestAgentActivity_Touch_CacheError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cache := NewMockCache(m)
	a := NewAgentActivityRepo(cache, &config.Config{})

	cache.EXPECT().SetWithTTL(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("boom"))

	err := a.Touch(context.TODO(), "agent-x")
	assert.ErrorContains(t, err, "touch agent activity")
}

// TestAgentActivity_IsActive 覆盖命中（活跃）、未命中（空闲）、后端故障三种语义。
func TestAgentActivity_IsActive(t *testing.T) {
	cases := map[string]struct {
		setup func(*MockCache)
		want  bool
		err   string
	}{
		"命中": {
			setup: func(c *MockCache) {
				c.EXPECT().Get(NewKey("agent_activity:%s", "agent-x")).Return([]byte("1"), nil)
			},
			want: true,
		},
		"未命中": {
			setup: func(c *MockCache) {
				c.EXPECT().Get(NewKey("agent_activity:%s", "agent-x")).Return(nil, ErrCacheMiss)
			},
			want: false,
		},
		"后端故障": {
			setup: func(c *MockCache) {
				c.EXPECT().Get(NewKey("agent_activity:%s", "agent-x")).Return(nil, errors.New("db down"))
			},
			want: false,
			err:  "get agent activity",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			m := gomock.NewController(t)
			defer m.Finish()
			cache := NewMockCache(m)
			a := NewAgentActivityRepo(cache, &config.Config{})
			tc.setup(cache)

			got, err := a.IsActive(context.TODO(), "agent-x")
			assert.Equal(t, tc.want, got)
			if tc.err == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tc.err)
		})
	}
}

// TestAgentActivity_IsActive_Disabled 空项目名或空闲关闭禁用时恒为非活跃且不读缓存。
func TestAgentActivity_IsActive_Disabled(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cache := NewMockCache(m)
	a := NewAgentActivityRepo(cache, &config.Config{AI: config.AIConfig{IdleTimeout: -time.Minute}})

	got, err := a.IsActive(context.TODO(), "agent-x")
	assert.NoError(t, err)
	assert.False(t, got)

	got, err = a.IsActive(context.TODO(), "")
	assert.NoError(t, err)
	assert.False(t, got)
}
