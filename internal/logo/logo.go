package logo

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"github.com/pterm/pterm"
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

// Logo 渲染内嵌 ASCII 横幅，每行从上到下做青色到洋红的渐变着色。
func Logo(opts ...func(*banner)) string {
	b := new(banner)
	for _, opt := range opts {
		opt(b)
	}

	from := pterm.NewRGB(0, 255, 255)
	to := pterm.NewRGB(255, 0, 255)

	scanner := bufio.NewScanner(bytes.NewReader(b.Bytes()))
	scanner.Split(bufio.ScanLines)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var sb strings.Builder
	// Fade 的 max 取 len(lines)-1，保证最后一行 current==max 命中终点洋红。
	for i, line := range lines {
		sb.WriteString(from.Fade(0, float32(len(lines)-1), float32(i), to).Sprint(line))
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
