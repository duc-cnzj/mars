package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// renderInspectTable 用给定表头与数据行渲染表格，返回渲染结果。
func renderInspectTable(t *testing.T, header []string, rows ...[]string) string {
	t.Helper()
	var buf bytes.Buffer
	tbl := newInspectTable(header)
	tbl.out = &buf
	for _, row := range rows {
		tbl.Append(row)
	}
	tbl.Render()
	return buf.String()
}

// TestInspectTable_Render 逐字节锁定表格渲染格式：表头居中大写、数据左对齐、
// 行间带分隔线，防止后续改动造成输出漂移。
func TestInspectTable_Render(t *testing.T) {
	got := renderInspectTable(t,
		[]string{"ID", "Name", "Tags"},
		[]string{"1", "EventBootstrapper", ""},
		[]string{"10", "GrpcBootstrapper", "api,grpc"},
	)

	want := "┌────┬───────────────────┬──────────┐\n" +
		"│ ID │       NAME        │   TAGS   │\n" +
		"├────┼───────────────────┼──────────┤\n" +
		"│ 1  │ EventBootstrapper │          │\n" +
		"├────┼───────────────────┼──────────┤\n" +
		"│ 10 │ GrpcBootstrapper  │ api,grpc │\n" +
		"└────┴───────────────────┴──────────┘\n"
	assert.Equal(t, want, got)
}

// TestInspectTable_Empty 无数据行时只渲染表头，不出现行分隔线。
func TestInspectTable_Empty(t *testing.T) {
	got := renderInspectTable(t, []string{"ID", "Name", "Expression"})

	want := "┌────┬──────┬────────────┐\n" +
		"│ ID │ NAME │ EXPRESSION │\n" +
		"└────┴──────┴────────────┘\n"
	assert.Equal(t, want, got)
}

// TestInspectTable_RaggedRow 数据行列数少于表头时，缺列按空单元格渲染；
// 超出表头的列被忽略。列宽仍由表头与已有数据决定。
func TestInspectTable_RaggedRow(t *testing.T) {
	got := renderInspectTable(t,
		[]string{"ID", "Event Name", "Listener Names", "Listener Count"},
		[]string{"1", "EventProjectCreated", "A B", "2"},
		[]string{"2", "X", "Y"},
		[]string{"3", "Z", "W", "1", "ignored"},
	)

	want := "┌────┬─────────────────────┬────────────────┬────────────────┐\n" +
		"│ ID │     EVENT NAME      │ LISTENER NAMES │ LISTENER COUNT │\n" +
		"├────┼─────────────────────┼────────────────┼────────────────┤\n" +
		"│ 1  │ EventProjectCreated │ A B            │ 2              │\n" +
		"├────┼─────────────────────┼────────────────┼────────────────┤\n" +
		"│ 2  │ X                   │ Y              │                │\n" +
		"├────┼─────────────────────┼────────────────┼────────────────┤\n" +
		"│ 3  │ Z                   │ W              │ 1              │\n" +
		"└────┴─────────────────────┴────────────────┴────────────────┘\n"
	assert.Equal(t, want, got)
}

// TestInspectTable_StarMarker 插件表用 "⭐︎" 标记当前启用的插件；
// 该标记两个 rune 的宽度与终端显示宽度一致，列宽由更长的表头决定。
func TestInspectTable_StarMarker(t *testing.T) {
	got := renderInspectTable(t,
		[]string{"ID", "Plugin", "Current"},
		[]string{"2", "gitlab", "⭐︎"},
		[]string{"1", "ws_sender_redis", ""},
	)

	want := "┌────┬─────────────────┬─────────┐\n" +
		"│ ID │     PLUGIN      │ CURRENT │\n" +
		"├────┼─────────────────┼─────────┤\n" +
		"│ 2  │ gitlab          │ ⭐︎      │\n" +
		"├────┼─────────────────┼─────────┤\n" +
		"│ 1  │ ws_sender_redis │         │\n" +
		"└────┴─────────────────┴─────────┘\n"
	assert.Equal(t, want, got)
}

// TestDisplayWidth 按 rune 计数返回宽度；inspect 表格只承载 ASCII 数据与 "⭐︎"，
// 二者 rune 数与终端显示宽度一致。CJK 宽字符会被少算（如 "中文" 记 2），属已知取舍。
func TestDisplayWidth(t *testing.T) {
	assert.Equal(t, 0, displayWidth(""))
	assert.Equal(t, 7, displayWidth("CURRENT"))
	assert.Equal(t, 2, displayWidth("⭐︎"))
	assert.Equal(t, 2, displayWidth("中文"))
}
