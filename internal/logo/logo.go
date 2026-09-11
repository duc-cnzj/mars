package logo

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
)

// logo 是内嵌的启动 ASCII 横幅。
//
//go:embed logo.txt
var logo []byte

// banner 保存可选的追加文本，通过 Logo 的函数选项注入。
type banner struct {
	appends []byte
}

// Bytes 返回完整横幅内容：内嵌 logo 与追加文本。
// 每次显式分配新切片，避免追加时篡改内嵌 logo 的底层数组。
func (b *banner) Bytes() []byte {
	out := make([]byte, 0, len(logo)+len(b.appends))
	out = append(out, logo...)
	out = append(out, b.appends...)
	return out
}

// WithAppends 返回一个函数选项，在 logo 下方追加文本。
func WithAppends(appends []byte) func(*banner) {
	return func(b *banner) {
		b.appends = appends
	}
}

// rgb 表示一个 24 位真彩色，用于渐变着色。
type rgb struct {
	R, G, B uint8
}

// sprint 以 ANSI 真彩色转义序列包裹文本：ESC[38;2;R;G;Bm 文本 ESC[0m。
func (c rgb) sprint(text string) string {
	var sb strings.Builder
	sb.Grow(len(text) + 24)
	sb.WriteString("\x1b[38;2;")
	sb.WriteString(strconv.Itoa(int(c.R)))
	sb.WriteByte(';')
	sb.WriteString(strconv.Itoa(int(c.G)))
	sb.WriteByte(';')
	sb.WriteString(strconv.Itoa(int(c.B)))
	sb.WriteString("m")
	sb.WriteString(text)
	sb.WriteString("\x1b[0m")
	return sb.String()
}

// fade 把 current 在 [0, max] 上线性映射到 c（起点色）与 end（终点色）之间。
// current 到达 max（末行）时直接返回终点色，既保证终点色精确输出，也避免了 max 为 0 时的除零。
// 运算保持 float32 且先除后乘：改变顺序或精度会引入舍入差，使输出颜色与历史版本不同。
func (c rgb) fade(max, current float32, end rgb) rgb {
	if max == current {
		return end
	}
	// channel 对单个颜色通道做线性插值并取整。
	channel := func(from, to uint8) uint8 {
		return uint8(int(float32(from) + ((float32(to)-float32(from))/max)*current))
	}
	return rgb{
		R: channel(c.R, end.R),
		G: channel(c.G, end.G),
		B: channel(c.B, end.B),
	}
}

// Logo 渲染内嵌 ASCII 横幅，每行从上到下做青色到洋红的渐变着色。
func Logo(opts ...func(*banner)) string {
	b := new(banner)
	for _, opt := range opts {
		opt(b)
	}

	from := rgb{R: 0, G: 255, B: 255}
	to := rgb{R: 255, G: 0, B: 255}

	scanner := bufio.NewScanner(bytes.NewReader(b.Bytes()))
	scanner.Split(bufio.ScanLines)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var sb strings.Builder
	// fade 的 max 取 len(lines)-1，保证最后一行 current==max 命中终点洋红。
	for i, line := range lines {
		sb.WriteString(from.fade(float32(len(lines)-1), float32(i), to).sprint(line))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// WithAuthor 返回 logo 及右对齐的 "created by duc@2023." 署名行。
func WithAuthor() string {
	maxWidth := 0
	scanner := bufio.NewScanner(bytes.NewReader(logo))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		maxWidth = max(len(scanner.Bytes()), maxWidth)
	}

	return Logo(WithAppends([]byte("\n\n" + fmt.Sprintf("%*s", maxWidth, "created by duc@2023."))))
}
