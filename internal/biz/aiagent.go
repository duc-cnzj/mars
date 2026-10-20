package biz

// aiagent.go 定义 AI Agent 沙箱的领域常量、命名派生与活跃度端口。这些契约被
// services（LLM 代理/MCP/配置/启用入口）与 cronjob（空闲关闭对账）共同消费，
// 收在 biz 避免 services 子包与 cronjob 之间互相引用。

import (
	"context"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/hasher"
)

const (
	// AgentNamespace 是 AI Agent 项目的逻辑命名空间名（data 层会追加 ns_prefix）。
	AgentNamespace = "mars-agent"
	// AgentProjectPrefix 是 agent 项目名前缀，后接可读的邮箱本地部分与哈希后缀。
	AgentProjectPrefix = "agent-"
	// AgentChartName 是沙箱 chart 名：Deployment 名由 helm fullname 规则派生时使用。
	AgentChartName = "mars-agent"
	// AgentSessionTokenSecretKey 是沙箱 Secret 中存放会话令牌的键（与 chart 约定一致）。
	AgentSessionTokenSecretKey = "mars-session-token"

	// maxK8sNameLength 是 k8s 资源名的长度上限。
	maxK8sNameLength = 63
	// agentProjectHashLen 是项目名中邮箱哈希后缀的十六进制字符数。
	agentProjectHashLen = 12
	// agentProjectLocalPartMaxLen 是项目名中可读邮箱部分的长度上限。
	// 总名 = "agent-" + 可读部分 + "-" + 哈希，控制在 52 字符内（helm release 名上限 53），
	// 保证 AgentDeploymentName 追加 "-mars-agent" 后仍不超过 k8s 63 字符上限。
	agentProjectLocalPartMaxLen = 52 - len(AgentProjectPrefix) - 1 - agentProjectHashLen
)

// AgentProjectName 由用户邮箱派生可读且全局唯一的 agent 项目名：
// agent-<邮箱本地部分净化>-<sha256(email) 前 12 位十六进制>。
// 可读部分让管理员一眼看出项目归属用户；哈希后缀保证不同邮箱（即使本地部分相同）不冲突。
func AgentProjectName(email string) string {
	return AgentProjectPrefix + agentProjectLocalPart(email) + "-" + hasher.Hash(email)[:agentProjectHashLen]
}

// agentProjectLocalPart 取邮箱 @ 前的本地部分并净化为 DNS-1123 标签风格：
// 转小写、非 [a-z0-9] 字符折叠为单个 "-"、去掉首尾 "-"、超长截断并再次去除尾部 "-"。
// 净化后为空（如本地部分全是符号/非 ASCII）时回落到 "user"，保证项目名仍可读。
func agentProjectLocalPart(email string) string {
	local, _, _ := strings.Cut(email, "@")
	var b strings.Builder
	lastWasDash := false
	for _, r := range strings.ToLower(local) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastWasDash = false
			continue
		}
		if b.Len() > 0 && !lastWasDash {
			b.WriteByte('-')
			lastWasDash = true
		}
	}
	part := strings.Trim(b.String(), "-")
	if part == "" {
		return "user"
	}
	if len(part) > agentProjectLocalPartMaxLen {
		part = strings.Trim(part[:agentProjectLocalPartMaxLen], "-")
	}
	return part
}

// AgentDeploymentName 返回 agent 项目对应的 Deployment 名，复刻 chart 的
// fullname 规则：Release 名已包含 chart 名时不再拼接，否则追加 chart 名；
// 最终截断到 63 字符并去掉截断产生的尾部连字符（与 helm 的 trunc/trimSuffix 一致）。
func AgentDeploymentName(projectName string) string {
	name := projectName
	if !strings.Contains(name, AgentChartName) {
		name = projectName + "-" + AgentChartName
	}
	if len(name) > maxK8sNameLength {
		name = strings.TrimSuffix(name[:maxK8sNameLength], "-")
	}
	return name
}

// IsAgentProjectOf 判断项目是否是由 email 派生的 agent 沙箱项目：项目名必须
// 精确等于 AgentProjectName(email)。用「名 = 邮箱派生结果」自校验，无需依赖
// 命名空间名（实际 k8s 名带 ns_prefix 前缀）；creator 为空时恒为 false。
//
// 部署系统变量用它把沙箱回连 mars 的令牌身份锚定到项目创建者：管理员进入
// 别人的 agent 项目重新部署时，令牌不能被改写成管理员身份。
func IsAgentProjectOf(projectName, email string) bool {
	if email == "" {
		return false
	}
	return projectName == AgentProjectName(email)
}

// AgentActivity 是 agent 沙箱的活跃度端口：入口侧记录最近操作时间，定时任务
// 据此判定沙箱是否空闲。实现方以 TTL 缓存表达「最近 Touch 后 N 分钟内有效」。
type AgentActivity interface {
	// Touch 刷新项目的最近活跃时间，记录失败返回错误（调用方可忽略）。
	Touch(ctx context.Context, projectName string) error
	// IsActive 判断项目是否仍在活跃窗口内；键不存在/已过期返回 false。
	IsActive(ctx context.Context, projectName string) (bool, error)
}

// TouchAgentActivity 按邮箱派生的 agent 项目刷新活跃度，供 LLM 代理、MCP、配置
// 下发与启用入口共用。email 为空时静默跳过；Touch 失败只记 Warning，不返回错误——
// 活跃度是空闲缩容的优化依据，不能变成请求的故障点。
func TouchAgentActivity(ctx context.Context, activity AgentActivity, logger mlog.Logger, email string) {
	if email == "" {
		return
	}
	if err := activity.Touch(ctx, AgentProjectName(email)); err != nil {
		logger.WarningCtxf(ctx, "刷新 AI Agent 活跃度失败: %v", err)
	}
}
