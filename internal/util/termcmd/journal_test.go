package termcmd

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestJournal_FeedCommands 覆盖回车提交、退格、控制符与空白归一。
func TestJournal_FeedCommands(t *testing.T) {
	tests := []struct {
		name string
		feed []string
		want string
	}{
		{name: "普通命令", feed: []string{"kubectl get pods\r"}, want: "kubectl get pods"},
		{name: "换行提交", feed: []string{"ls\n"}, want: "ls"},
		{name: "退格", feed: []string{"lsX\x7f\r"}, want: "ls"},
		{name: "退格键 BS", feed: []string{"ab\b\r"}, want: "a"},
		{name: "空行退格忽略", feed: []string{"\x7fls\r"}, want: "ls"},
		{name: "Ctrl-C 放弃当前行", feed: []string{"rm -rf /\x03ls\r"}, want: "ls"},
		{name: "Ctrl-U 清空当前行", feed: []string{"abc\x15def\r"}, want: "def"},
		{name: "Ctrl-D 放弃当前行", feed: []string{"abc\x04def\r"}, want: "def"},
		{name: "Ctrl-W 删除最后一个词", feed: []string{"kubectl get\x17pods\r"}, want: "kubectl pods"},
		{name: "Ctrl-W 跳过尾随空格", feed: []string{"kubectl get \x17pods\r"}, want: "kubectl pods"},
		{name: "空白行忽略", feed: []string{"   \r", "ls\r"}, want: "ls"},
		{name: "连续空格归一", feed: []string{"kubectl   get    pods\r"}, want: "kubectl get pods"},
		{name: "Tab 忽略", feed: []string{"kubectl\tget\r"}, want: "kubectlget"},
		{name: "空输入", feed: []string{"\r"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j Journal
			for _, chunk := range tt.feed {
				j.Feed([]byte(chunk))
			}
			assert.Equal(t, tt.want, j.Snapshot(1024))
		})
	}
}

// TestJournal_FeedEscapes 覆盖方向键、SS3、OSC 标题与括号粘贴。
func TestJournal_FeedEscapes(t *testing.T) {
	tests := []struct {
		name string
		feed string
		want string
	}{
		{name: "方向键忽略", feed: "ech\x1b[Do hi\r", want: "echo hi"},
		{name: "SS3 功能键忽略", feed: "ls\x1bOP\r", want: "ls"},
		{name: "OSC 标题忽略", feed: "ls\x1b]0;title\x07\r", want: "ls"},
		{name: "OSC ST 结束", feed: "ls\x1b]0;title\x1b\\\r", want: "ls"},
		{name: "OSC 内 ESC 非 ST", feed: "ls\x1b]0;title\x1bX\x07\r", want: "ls"},
		{name: "括号粘贴", feed: "\x1b[200~kubectl get pods\x1b[201~\r", want: "kubectl get pods"},
		{name: "Alt 按键忽略一个字符", feed: "a\x1bxb\r", want: "ab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j Journal
			j.Feed([]byte(tt.feed))
			assert.Equal(t, tt.want, j.Snapshot(1024))
		})
	}
}

// TestJournal_EscapeInvalidState 覆盖内部转义状态异常时复位并丢弃该字节，
// 避免状态机卡死（正常路径不会产生非法状态）。
func TestJournal_EscapeInvalidState(t *testing.T) {
	var j Journal
	j.esc = escState(99)
	j.Feed([]byte("x\rls\r"))
	assert.Equal(t, "ls", j.Snapshot(1024))
	assert.Equal(t, escNone, j.esc)
}

// TestJournal_UTF8Split 覆盖 UTF-8 字符被拆到多个 stdin 包的情况。
func TestJournal_UTF8Split(t *testing.T) {
	var j Journal
	full := []byte("查看中文日志\r")
	j.Feed(full[:2])
	assert.Equal(t, "", j.Snapshot(1024), "不完整字符不应提前提交或产生乱码")
	j.Feed(full[2:])
	assert.Equal(t, "查看中文日志", j.Snapshot(1024))
}

// TestJournal_InvalidUTF8 覆盖非法字节不会堵住后续控制字节。
func TestJournal_InvalidUTF8(t *testing.T) {
	var j Journal
	j.Feed([]byte{0xff, 'l', 's', '\r'})
	assert.Equal(t, "ls", j.Snapshot(1024))
}

// TestJournal_LineAndHistoryLimits 覆盖单行截断与历史条数上限。
func TestJournal_LineAndHistoryLimits(t *testing.T) {
	var j Journal
	j.Feed([]byte(strings.Repeat("a", 300) + "\r"))
	assert.Equal(t, strings.Repeat("a", lineMaxRunes), j.Snapshot(1024))

	var h Journal
	for i := 0; i < historyMax+5; i++ {
		h.Feed([]byte("cmd" + string(rune('a'+i)) + "\r"))
	}
	got := h.Snapshot(1024)
	assert.NotContains(t, got, "cmda")
	assert.Contains(t, got, "cmd"+string(rune('a'+historyMax+4)))
	assert.Equal(t, historyMax, strings.Count(got, " ; ")+1)
}

// TestJournal_LineLimitBatchTrim 覆盖批量裁剪后对外仍只暴露最近 lineMaxRunes 个字符。
func TestJournal_LineLimitBatchTrim(t *testing.T) {
	var j Journal
	j.Feed([]byte(strings.Repeat("a", 500) + "END\r"))
	got := j.Snapshot(1024)
	assert.Equal(t, lineMaxRunes, len([]rune(got)))
	assert.Equal(t, strings.Repeat("a", lineMaxRunes-3)+"END", got)
}

// TestJournal_SnapshotBudget 覆盖按预算取最近命令与单命令中间截断。
func TestJournal_SnapshotBudget(t *testing.T) {
	var j Journal
	for _, cmd := range []string{"aa", "bb", "cc"} {
		j.Feed([]byte(cmd + "\r"))
	}
	tests := []struct {
		name string
		max  int
		want string
	}{
		{name: "全部放下", max: 12, want: "aa ; bb ; cc"},
		{name: "从最新往前取", max: 7, want: "bb ; cc"},
		{name: "只能放一条", max: 2, want: "cc"},
		{name: "一个字符", max: 1, want: "c"},
		{name: "零预算", max: 0, want: ""},
		{name: "负数预算", max: -1, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, j.Snapshot(tt.max))
		})
	}

	var long Journal
	long.Feed([]byte(strings.Repeat("x", 100) + "\r"))
	got := long.Snapshot(10)
	assert.Equal(t, 10, len([]rune(got)))
	assert.Contains(t, got, "…")
}

// TestTruncateMiddle 覆盖截断边界：非正预算、单字符预算、不超限与中间省略。
func TestTruncateMiddle(t *testing.T) {
	assert.Equal(t, "", TruncateMiddle("abc", 0))
	assert.Equal(t, "", TruncateMiddle("abc", -1))
	assert.Equal(t, "abc", TruncateMiddle("abc", 3))
	assert.Equal(t, "abc", TruncateMiddle("abc", 10))
	assert.Equal(t, "c", TruncateMiddle("abc", 1))
	assert.Equal(t, 4, len([]rune(TruncateMiddle("abcdefgh", 4))))
	assert.Equal(t, "ab…h", TruncateMiddle("abcdefgh", 4))
}

// TestJournal_ConcurrentFeed 覆盖并发写入的锁安全（配合 -race）。
func TestJournal_ConcurrentFeed(t *testing.T) {
	var j Journal
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			j.Feed([]byte("cmd" + string(rune('a'+id)) + "\r"))
			_ = j.Snapshot(64)
		}(i)
	}
	wg.Wait()
	assert.NotEmpty(t, j.Snapshot(1024))
}
