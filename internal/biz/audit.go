package biz

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/duc-cnzj/mars/v6/internal/util/termcmd"
)

const (
	// auditMessageMaxRunes 是审计消息的安全长度上限：events.message 为 varchar(255)，
	// 留出余量避免不同数据库对长度边界处理不一致。
	auditMessageMaxRunes = 240
	// auditCommandMarker 是审计消息中命令段的固定前缀，事件列表按原样整串展示。
	auditCommandMarker = ", 命令: '"
	// auditCommandScanBytes 限制脱敏扫描的原始命令长度。最终展示只会保留头尾
	// 约 auditMessageMaxRunes 个字符，超出的部分先夹取再跑正则，避免为超长
	// 命令（ExecOnce 的命令来自请求体）付出线性扫描与分配开销。
	auditCommandScanBytes = 2048
)

// auditSecretPatterns 是命令脱敏规则。审计事件对管理员可见，命令里常见的
// 密码/令牌不能原样落库；只匹配“键=值 / 键: 值 / --flag 值 / URL 凭据”等形态，
// 避免误伤普通参数。
var auditSecretPatterns = []struct {
	re   *regexp.Regexp
	repl string
}{
	// 裸关键字：password=... / token: ...
	{regexp.MustCompile(`(?i)(password|passwd|token|secret|api[_-]?key)\s*[=:]\s*(?:'[^']*'|"[^"]*"|[^\s;|&'"]+)`), `${1}=***`},
	// 复合环境变量名：AWS_SECRET_ACCESS_KEY / MYSQL_PWD / GITHUB_TOKEN ...
	{regexp.MustCompile(`(?i)\b([a-z0-9_]*(?:password|passwd|secret|token|api[_-]?key|access[_-]?key|private[_-]?key|mysql_pwd)[a-z0-9_]*)\s*=\s*(?:'[^']*'|"[^"]*"|[^\s;|&'"]+)`), `${1}=***`},
	// 任意含敏感词的长选项：--password / --api-key / --client-secret ...
	{regexp.MustCompile(`(?i)(--[a-z0-9-]*(?:password|passwd|secret|token|api-?key)[a-z0-9-]*)(?:\s*=\s*|\s+)(?:'[^']*'|"[^"]*"|[^\s;|&'"]+)`), `${1}=***`},
	// URL 内嵌凭据：scheme://user:password@host
	{regexp.MustCompile(`(?i)(\b[a-z][a-z0-9+.-]*://[^/\s:@]+:)([^@\s/]+)(@)`), `${1}***${3}`},
	// mysql 系工具的行内短选项：mysql -uroot -psecret
	{regexp.MustCompile(`(?i)\b(mysql|mariadb|mysqldump|mysqladmin)\b([^;|&]*?)\s-p(\S+)`), `${1}${2} -p***`},
	// HTTP 认证头：Authorization: Bearer/Basic ...
	{regexp.MustCompile(`(?i)(authorization:\s*(?:bearer|basic)\s+)(?:'[^']*'|"[^"]*"|[^\s;|&'"]+)`), `${1}***`},
	// curl/wget 的行内凭据：curl -u user:password
	{regexp.MustCompile(`(?i)\b(curl|wget)\b([^;|&]*?)\s-u\s*(?:'[^']*'|"[^"]*"|[^\s;|&'"]+)`), `${1}${2} -u ***`},
	// redis-cli 的口令：redis-cli -a password
	{regexp.MustCompile(`(?i)\b(redis-cli)\b([^;|&]*?)\s-a\s*(?:'[^']*'|"[^"]*"|[^\s;|&'"]+)`), `${1}${2} -a ***`},
	// docker/podman login 的短口令选项：docker login -p password
	{regexp.MustCompile(`(?i)\b(docker|podman)\s+login\b([^;|&]*?)\s-p\s*(?:'[^']*'|"[^"]*"|[^\s;|&'"]+)`), `${1} login${2} -p ***`},
}

// AuditCommandBudget 返回 msg 之后还能容纳的命令字符数；<=0 表示无法追加。
// 调用方据此截取命令，避免先拼接再截断产生多余拷贝。
func AuditCommandBudget(msg string) int {
	return auditMessageMaxRunes -
		utf8.RuneCountInString(msg) -
		utf8.RuneCountInString(auditCommandMarker) -
		1 // 末尾闭合引号
}

// WithAuditCommands 在审计消息末尾追加用户执行的命令段。
//
// 命令先做空白归一与密钥脱敏，再按消息总长度上限截断；空命令或消息已超限时
// 不追加命令段，只把消息压到上限内。命令段来自终端 stdin 的尽力采集，可能不完整，
// 见 termcmd.Journal 的已知限制。
func WithAuditCommands(msg, commands string) string {
	commands = RedactAuditCommands(commands)
	if commands == "" {
		return clampAuditMessage(msg)
	}
	budget := AuditCommandBudget(msg)
	if budget <= 0 {
		return clampAuditMessage(msg)
	}
	return msg + auditCommandMarker + termcmd.TruncateMiddle(commands, budget) + "'"
}

// RedactAuditCommands 对命令做审计写入前的夹取、空白归一与密钥脱敏。
// 消息段与结构化审计字段（如 ExecOnce 的 new.command）共用同一口径，
// 避免只脱敏消息、却把原文写进变更详情。
func RedactAuditCommands(commands string) string {
	return normalizeAuditCommands(clampAuditCommands(commands))
}

// clampAuditMessage 把超过上限的审计消息按「头 + … + 尾」压到安全长度：
// events.message 是 varchar(255)，超长会让整条审计写入失败而丢失记录。
// 容器/命名空间/Pod 名理论上可长到 253 字节，正常路径不会触发，属兜底。
func clampAuditMessage(msg string) string {
	return termcmd.TruncateMiddle(msg, auditMessageMaxRunes)
}

// clampAuditCommands 把超长命令夹取为「头 + … + 尾」，两端按 UTF-8 边界对齐。
// 最终展示本来就只保留头尾，夹取不改变可见结果，只把脱敏扫描的输入限制在常量级。
func clampAuditCommands(commands string) string {
	if len(commands) <= auditCommandScanBytes {
		return commands
	}
	half := auditCommandScanBytes / 2
	head := commands[:half]
	tail := commands[len(commands)-half:]
	// 截断点可能落在多字节字符中间：头部去掉残缺尾字节，尾部跳过续字节。
	for len(head) > 0 {
		r, size := utf8.DecodeLastRuneInString(head)
		if r != utf8.RuneError || size > 1 {
			break
		}
		head = head[:len(head)-1]
	}
	for len(tail) > 0 && !utf8.RuneStart(tail[0]) {
		tail = tail[1:]
	}
	return head + "…" + tail
}

// ContainerExecAuditMessage 组装「用户进入容器执行命令」的审计消息。
// action 为空时不加前缀；Exec / ExecOnce / Shell 三个入口共用同一模板，
// 避免多处各自拼接导致格式漂移。
func ContainerExecAuditMessage(action, container, namespace, pod string) string {
	prefix := ""
	if action != "" {
		prefix = "[" + action + "]: "
	}
	return fmt.Sprintf("%s用户进入容器执行命令，container: '%s', namespace: '%s', pod： '%s'",
		prefix, container, namespace, pod)
}

// normalizeAuditCommands 把命令压成单行并脱敏常见密钥。
func normalizeAuditCommands(commands string) string {
	commands = strings.Join(strings.Fields(commands), " ")
	for _, pattern := range auditSecretPatterns {
		commands = pattern.re.ReplaceAllString(commands, pattern.repl)
	}
	return commands
}
