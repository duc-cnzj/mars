package data

// aiagent.go 是 biz.AgentActivity 的缓存实现：用带 TTL 的缓存键表达
// 「agent 项目最近一次 Touch 后的活跃窗口」，不做任何持久化。

import (
	"context"
	"errors"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/errs"
)

var _ biz.AgentActivity = (*agentActivity)(nil)

// agentActivity 基于通用缓存实现活跃度端口：Touch 写入 TTL=空闲阈值的键，
// IsActive 读取该键是否存在。缓存后端（memory/db）由 Cache 门面决定。
type agentActivity struct {
	cache Cache
	ttl   time.Duration
}

// NewAgentActivityRepo 构造 agent 活跃度实现；ttl 为 0 表示空闲自动关闭已禁用，
// 此时 Touch 不写缓存、IsActive 恒为 false（对账任务本身也不会注册）。
func NewAgentActivityRepo(cache Cache, cfg *config.Config) biz.AgentActivity {
	return &agentActivity{cache: cache, ttl: cfg.AgentIdleTimeout()}
}

// agentActivityKey 返回项目活跃度缓存键；slug 为固定格式串，避免按项目名分裂指标分组。
func agentActivityKey(projectName string) CacheKey {
	return NewKey("agent_activity:%s", projectName)
}

// Touch 刷新项目活跃度：写入 TTL 为空闲阈值的标记。空项目名或功能禁用时直接返回 nil。
func (a *agentActivity) Touch(ctx context.Context, projectName string) error {
	if projectName == "" || a.ttl <= 0 {
		return nil
	}
	seconds := int((a.ttl + time.Second - 1) / time.Second)
	return errs.Wrap(a.cache.SetWithTTL(agentActivityKey(projectName), []byte("1"), seconds), "touch agent activity")
}

// IsActive 判断项目是否在活跃窗口内。键不存在/已过期返回 false；缓存后端故障
// 上抛错误，由调用方按「跳过关闭」降级（宁可漏关，不可误关正在使用的沙箱）。
func (a *agentActivity) IsActive(ctx context.Context, projectName string) (bool, error) {
	if projectName == "" || a.ttl <= 0 {
		return false, nil
	}
	_, err := a.cache.Get(agentActivityKey(projectName))
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, ErrCacheMiss):
		return false, nil
	default:
		return false, errs.Wrap(err, "get agent activity")
	}
}
