package biz

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// TestWithAuditCommands_Appends 覆盖命令段的拼接格式。
func TestWithAuditCommands_Appends(t *testing.T) {
	msg := "用户进入容器执行命令，container: 'web', namespace: 'uat', pod： 'web-1'"
	got := WithAuditCommands(msg, "kubectl get pods\r\n")
	assert.Equal(t, msg+", 命令: 'kubectl get pods'", got)
}

// TestWithAuditCommands_Empty 覆盖空命令（含纯空白）时原样返回。
func TestWithAuditCommands_Empty(t *testing.T) {
	msg := "用户进入容器执行命令"
	assert.Equal(t, msg, WithAuditCommands(msg, "  \r\n  "))
	assert.Equal(t, msg, WithAuditCommands(msg, ""))
}

// TestWithAuditCommands_RedactsSecrets 覆盖常见密钥形态的脱敏，且不误伤普通参数。
func TestWithAuditCommands_RedactsSecrets(t *testing.T) {
	tests := []struct {
		name     string
		commands string
		want     string
	}{
		{name: "环境变量", commands: "export TOKEN=abc123", want: "export TOKEN=***"},
		{name: "复合环境变量", commands: "export AWS_SECRET_ACCESS_KEY=abc", want: "export AWS_SECRET_ACCESS_KEY=***"},
		{name: "密码参数", commands: "mysql --password hunter2 -u root", want: "mysql --password=*** -u root"},
		{name: "api key", commands: "curl --api-key=xyz https://example.com", want: "curl --api-key=*** https://example.com"},
		{name: "client secret", commands: "app --client-secret s3cr3t", want: "app --client-secret=***"},
		{name: "mysql 简写", commands: "mysql -uroot -phunter2 -e 'select 1'", want: "mysql -uroot -p*** -e 'select 1'"},
		{name: "URL 凭据", commands: "curl https://user:pass@example.com", want: "curl https://user:***@example.com"},
		{name: "引号包裹", commands: "login password='p@ss word'", want: "login password=***"},
		{name: "Bearer", commands: "curl -H 'Authorization: Bearer abc.def' x", want: "curl -H 'Authorization: Bearer ***' x"},
		{name: "Basic", commands: "curl -H 'Authorization: Basic dXNlcjpwYXNz' x", want: "curl -H 'Authorization: Basic ***' x"},
		{name: "curl 行内凭据", commands: "curl -u admin:s3cret https://x", want: "curl -u *** https://x"},
		{name: "redis-cli 口令", commands: "redis-cli -h db -a s3cret ping", want: "redis-cli -h db -a *** ping"},
		{name: "docker login 短口令", commands: "docker login -u me -p s3cret registry", want: "docker login -u me -p *** registry"},
		{name: "普通参数不误伤", commands: "grep password /app/config", want: "grep password /app/config"},
		{name: "普通命令不误伤", commands: "kubectl get pods -n uat", want: "kubectl get pods -n uat"},
		{name: "空白归一", commands: "kubectl   get\npods", want: "kubectl get pods"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithAuditCommands("命令", tt.commands)
			assert.Equal(t, "命令, 命令: '"+tt.want+"'", got)
		})
	}
}

// TestWithAuditCommands_Budget 覆盖按消息上限截断并保留头尾。
func TestWithAuditCommands_Budget(t *testing.T) {
	msg := strings.Repeat("前", 200)
	got := WithAuditCommands(msg, strings.Repeat("命", 100))
	assert.LessOrEqual(t, utf8.RuneCountInString(got), auditMessageMaxRunes)
	assert.Contains(t, got, "…")

	full := strings.Repeat("满", auditMessageMaxRunes)
	assert.Equal(t, full, WithAuditCommands(full, "ls"))
	assert.True(t, AuditCommandBudget(full) <= 0)
	assert.True(t, AuditCommandBudget("短消息") > 0)
}

// TestWithAuditCommands_ClampsLargeInput 覆盖超长命令先夹取再脱敏：头尾保留、
// 中间省略，尾部密钥仍会脱敏，多字节字符跨截断点也不会产生非法 UTF-8。
func TestWithAuditCommands_ClampsLargeInput(t *testing.T) {
	got := WithAuditCommands("命令", strings.Repeat("a", 4096)+" password=hunter2 END")
	assert.LessOrEqual(t, utf8.RuneCountInString(got), auditMessageMaxRunes)
	assert.Contains(t, got, "END")
	assert.NotContains(t, got, "hunter2")
	assert.True(t, utf8.ValidString(got))

	// 4096 个三字节字符：截断点必然落在字符中间，夹取后仍需是合法 UTF-8。
	got = WithAuditCommands("命令", strings.Repeat("中", 4096))
	assert.True(t, utf8.ValidString(got))
	assert.NotContains(t, got, "\uFFFD")
}

// TestContainerExecAuditMessage 覆盖三个入口共用的审计消息模板。
func TestContainerExecAuditMessage(t *testing.T) {
	assert.Equal(t,
		"[Exec]: 用户进入容器执行命令，container: 'c', namespace: 'n', pod： 'p'",
		ContainerExecAuditMessage("Exec", "c", "n", "p"))
	assert.Equal(t,
		"用户进入容器执行命令，container: 'c', namespace: 'n', pod： 'p'",
		ContainerExecAuditMessage("", "c", "n", "p"))
}

// TestRedactAuditCommands 覆盖结构化命令字段的空白归一、脱敏与超长夹取。
func TestRedactAuditCommands(t *testing.T) {
	assert.Equal(t, "mysql -uroot -p***", RedactAuditCommands("  mysql   -uroot -psecret "))
	assert.Equal(t, "", RedactAuditCommands(" \r\n "))
	got := RedactAuditCommands(strings.Repeat("a", 4096))
	assert.Less(t, utf8.RuneCountInString(got), 4096)
	assert.Contains(t, got, "…")
}

// TestWithAuditCommands_ClampsOversizedMessage 覆盖消息本身超上限时压到 varchar 安全
// 长度：容器/命名空间/Pod 名理论上可长到 253 字节，超长会让整条审计写入失败。
func TestWithAuditCommands_ClampsOversizedMessage(t *testing.T) {
	msg := strings.Repeat("前", 300)
	got := WithAuditCommands(msg, "ls")
	assert.LessOrEqual(t, utf8.RuneCountInString(got), auditMessageMaxRunes)
	assert.Contains(t, got, "…")

	withoutCmd := WithAuditCommands(msg, "")
	assert.LessOrEqual(t, utf8.RuneCountInString(withoutCmd), auditMessageMaxRunes)
	assert.Contains(t, withoutCmd, "…")
}

// TestWithAuditCommands_BudgetOne 预算只剩 1 个字符时保留命令尾部（参数/路径比命令名
// 更有排查价值），与 termcmd.TruncateMiddle 的边界语义保持一致。
func TestWithAuditCommands_BudgetOne(t *testing.T) {
	msg := strings.Repeat("前", auditMessageMaxRunes-
		utf8.RuneCountInString(auditCommandMarker)-2)
	assert.Equal(t, 1, AuditCommandBudget(msg))

	got := WithAuditCommands(msg, "abc")
	assert.Equal(t, msg+auditCommandMarker+"c'", got)
	assert.Equal(t, auditMessageMaxRunes, utf8.RuneCountInString(got))
}
