// Package termcmd 从终端 stdin 字节流中抽取用户实际提交的命令行。
//
// 它只关心“用户按回车执行了什么”，不采集容器返回的输出，也不依赖任何具体
// 传输层：websocket 终端与 gRPC Exec 流都复用同一套解析规则。
//
// 已知限制：命令行来自终端按键流。方向键调历史、Tab 补全、Ctrl-R 搜索等编辑器
// 行为不会体现在 stdin 里，记录可能不完整，甚至与容器内实际执行的命令不一致
// （例如先输入 A、再按 ↑ 调出 B 并回车）；无回显提示（sudo/数据库口令）下输入的
// 明文也会被当成一行命令。因此命令段只用于辅助排查，不构成强审计证据。
package termcmd

import (
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	// lineMaxRunes 限制单行命令长度：超出时保留尾部，路径/参数通常在后半段。
	lineMaxRunes = 240
	// historyMax 限制保留的命令条数，避免长会话无限占用内存。
	historyMax = 20
)

// escState 是转义序列解析状态：方向键、功能键、OSC 标题等都不应进入命令行。
type escState uint8

const (
	escNone escState = iota
	escStart
	escCSI
	escOSC
	escOSCEsc
)

// Journal 是并发安全的终端命令日志。
//
// 零值可用；Feed 与 Snapshot 可被多个 goroutine 并发调用。终端输入是原始
// 击键流（含退格、方向键、括号粘贴），Journal 只在用户回车时提交一条命令。
type Journal struct {
	mu       sync.Mutex
	pending  []byte
	esc      escState
	line     []rune
	commands []string
}

// Feed 消费一段原始 stdin 字节。
//
// 处理规则：可打印字符追加到当前行；退格删除最后一个字符；Ctrl-C/Ctrl-D/Ctrl-U
// 放弃当前行；Ctrl-W 删除最后一个词；方向键/OSC 等转义序列直接忽略；回车把当前行
// 作为一条命令提交。UTF-8 字符被拆包时留在 pending 里等下一段，不会产生乱码。
func (j *Journal) Feed(data []byte) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.pending = append(j.pending, data...)
	for len(j.pending) > 0 {
		b := j.pending[0]
		if j.consumeEscape(b) {
			j.pending = j.pending[1:]
			continue
		}
		if b == 0x1b {
			j.esc = escStart
			j.pending = j.pending[1:]
			continue
		}
		if b < utf8.RuneSelf {
			j.consumeASCII(b)
			j.pending = j.pending[1:]
			continue
		}
		r, size := utf8.DecodeRune(j.pending)
		if r == utf8.RuneError && size == 1 {
			// 不足一个完整 UTF-8 字符：等下一段；非法字节则丢弃，避免污染命令。
			// FullRune 对非法编码返回 true，因此不会把后续控制字节一直堵在缓冲里。
			if !utf8.FullRune(j.pending) {
				return
			}
			j.pending = j.pending[1:]
			continue
		}
		j.appendRune(r)
		j.pending = j.pending[size:]
	}
}

// Snapshot 返回最近提交的命令，按时间顺序拼接，总长度不超过 maxRunes。
//
// 从最新一条向前取，装不下就停：推荐场景更关心用户刚做过什么。单条命令本身
// 超限时保留头尾、中间省略，避免只剩路径尾部或只剩命令名。
func (j *Journal) Snapshot(maxRunes int) string {
	j.mu.Lock()
	defer j.mu.Unlock()
	if maxRunes <= 0 || len(j.commands) == 0 {
		return ""
	}
	kept := make([]string, 0, len(j.commands))
	total := 0
	for i := len(j.commands) - 1; i >= 0; i-- {
		cmd := j.commands[i]
		cost := utf8.RuneCountInString(cmd)
		if len(kept) > 0 {
			cost += len(" ; ")
		}
		if total+cost > maxRunes {
			if len(kept) == 0 {
				return TruncateMiddle(cmd, maxRunes)
			}
			break
		}
		kept = append(kept, cmd)
		total += cost
	}
	for i, k := 0, len(kept)-1; i < k; i, k = i+1, k-1 {
		kept[i], kept[k] = kept[k], kept[i]
	}
	return strings.Join(kept, " ; ")
}

// consumeEscape 消费转义序列中的当前字节，返回是否已消费。
func (j *Journal) consumeEscape(b byte) bool {
	switch j.esc {
	case escNone:
		return false
	case escStart:
		switch b {
		case '[':
			j.esc = escCSI
		case ']':
			j.esc = escOSC
		case 'O':
			// SS3（如 F1-F4）：下一字节是最终字节，按 CSI 同样消费。
			j.esc = escCSI
		default:
			j.esc = escNone
		}
		return true
	case escCSI:
		if b >= 0x40 && b <= 0x7e {
			j.esc = escNone
		}
		return true
	case escOSC:
		if b == 0x07 {
			j.esc = escNone
			return true
		}
		if b == 0x1b {
			j.esc = escOSCEsc
		}
		return true
	case escOSCEsc:
		if b == '\\' {
			j.esc = escNone
			return true
		}
		j.esc = escOSC
		return true
	default:
		j.esc = escNone
		return true
	}
}

// consumeASCII 处理单个 ASCII 控制字节或可打印字符。
func (j *Journal) consumeASCII(b byte) {
	switch b {
	case '\r', '\n':
		j.commit()
	case 0x7f, 0x08:
		j.backspace()
	case 0x03, 0x04, 0x15:
		// Ctrl-C / Ctrl-D / Ctrl-U：放弃当前未提交的行。
		j.line = j.line[:0]
	case 0x17:
		j.deleteWord()
	default:
		if b >= 0x20 {
			j.appendRune(rune(b))
		}
	}
}

// appendRune 把可打印字符追加到当前行；超长时丢掉最旧的字符，保留尾部。
//
// 裁剪放到 2 倍上限再批量执行：逐字符裁剪会在每个新字符上搬移 lineMaxRunes 个
// rune，大段粘贴（单条 WS 消息最大 20MB）会退化成 O(240·n) 的 CPU 放大。
func (j *Journal) appendRune(r rune) {
	j.line = append(j.line, r)
	if len(j.line) > 2*lineMaxRunes {
		j.line = append(j.line[:0], j.line[len(j.line)-lineMaxRunes:]...)
	}
}

// backspace 删除当前行最后一个字符。
func (j *Journal) backspace() {
	if len(j.line) == 0 {
		return
	}
	j.line = j.line[:len(j.line)-1]
}

// deleteWord 删除当前行最后一个词（先跳过尾部空白，再删到上一个空白）。
func (j *Journal) deleteWord() {
	end := len(j.line)
	for end > 0 && j.line[end-1] == ' ' {
		end--
	}
	start := end
	for start > 0 && j.line[start-1] != ' ' {
		start--
	}
	j.line = j.line[:start]
}

// commit 提交当前行：空白行忽略，连续空白归一为单个空格，保留最近 historyMax 条。
// 提交前统一裁到 lineMaxRunes，保证批量裁剪策略下对外仍只暴露最近 240 个字符。
func (j *Journal) commit() {
	if len(j.line) > lineMaxRunes {
		j.line = j.line[len(j.line)-lineMaxRunes:]
	}
	cmd := strings.Join(strings.Fields(string(j.line)), " ")
	j.line = j.line[:0]
	if cmd == "" {
		return
	}
	if len(j.commands) >= historyMax {
		j.commands = j.commands[1:]
	}
	j.commands = append(j.commands, cmd)
}

// TruncateMiddle 保留头尾、中间用省略号连接，避免命令只剩命令名或只剩路径尾部。
// maxRunes 是返回的字符数上限；maxRunes <= 0 返回空串，命令本身不超限时原样返回。
// 审计命令段与终端命令日志共用这一份截断语义。
func TruncateMiddle(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	if maxRunes == 1 {
		return string(runes[len(runes)-1:])
	}
	head := maxRunes * 2 / 3
	tail := maxRunes - head - 1
	return string(runes[:head]) + "…" + string(runes[len(runes)-tail:])
}
