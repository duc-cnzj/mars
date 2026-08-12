package logo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLogo 断言 Logo 的渐变输出结构：
// 首行为起点青色，末行为终点洋红，原始 ASCII 逐行保留，默认无署名。
func TestLogo(t *testing.T) {
	out := Logo()

	assert.NotEmpty(t, out)
	// 渐变两端：起点青色 / 终点洋红。
	// 终点洋红是渐变差一的回归断言——若 max 取 len(lines) 而非 len(lines)-1，终点色永不输出。
	assert.True(t, strings.HasPrefix(out, "\x1b[38;2;0;255;255m"), "首行应为起点青色")
	assert.True(t, strings.Contains(out, "\x1b[38;2;255;0;255m"), "末行应为终点洋红")
	// 原始 ASCII 行在 strip 后完整保留（每行末带换行）。
	assert.True(t, strings.HasSuffix(out, "\n"))
	assert.True(t, strings.Contains(stripANSI(out), "::::    ::::      :::     :::::::::   ::::::::"))
	assert.True(t, strings.Contains(stripANSI(out), "###       ### ###     ### ###    ###  ########"))
	// 默认不追加署名。
	assert.False(t, strings.Contains(out, "created by duc@2023."))
}

// TestLogoWithAppends 直测 WithAppends：追加内容出现在输出中，且继续参与渐变。
func TestLogoWithAppends(t *testing.T) {
	out := Logo(WithAppends([]byte("\nTAIL")))

	assert.True(t, strings.Contains(stripANSI(out), "TAIL"))
	// 追加行作为末行，同样命中终点洋红。
	assert.True(t, strings.Contains(out, "\x1b[38;2;255;0;255mTAIL"))
}

// TestWithAuthor 断言署名行存在且按 logo 最宽行右对齐。
func TestWithAuthor(t *testing.T) {
	out := WithAuthor()

	assert.True(t, strings.Contains(out, "created by duc@2023."))
	// 右对齐：署名前补空格到 logo 最宽行(47)宽度。
	const maxWidth = 47
	const signature = "created by duc@2023."
	expected := "\x1b[38;2;255;0;255m" + strings.Repeat(" ", maxWidth-len(signature)) + signature
	assert.True(t, strings.Contains(out, expected))
}

// stripANSI 去掉 pterm 的 ANSI 转义序列，便于断言纯文本内容。
func stripANSI(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '\x1b' {
			j := i
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				i = j + 1
				continue
			}
		}
		sb.WriteByte(s[i])
		i++
	}
	return sb.String()
}
