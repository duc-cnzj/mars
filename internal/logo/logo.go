package logo

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"

	"github.com/pterm/pterm"
)

//go:embed logo.txt
var logo []byte

type logoS struct {
	b       []byte
	appends []byte
}

func (l *logoS) Bytes() []byte {
	return append(l.b, l.appends...)
}

func WithAppends(appends []byte) func(s *logoS) {
	return func(s *logoS) {
		s.appends = appends
	}
}

func Logo(opts ...func(s *logoS)) string {
	l := &logoS{
		b: logo,
	}
	for _, opt := range opts {
		opt(l)
	}
	from := pterm.NewRGB(0, 255, 255)
	to := pterm.NewRGB(255, 0, 255)

	maxLine := len(bytes.Split(l.Bytes(), []byte("\n")))
	scanner := bufio.NewScanner(bytes.NewReader(l.Bytes()))
	scanner.Split(bufio.ScanLines)
	i := 0
	var logoOutput string
	for scanner.Scan() {
		logoOutput += from.Fade(0, float32(maxLine), float32(i), to).Sprintf(scanner.Text() + "\n")
		i++
	}

	return logoOutput
}

func WithAuthor() string {
	maxWidth := 0
	scanner := bufio.NewScanner(bytes.NewReader(logo))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		maxWidth = max(len(scanner.Bytes()), maxWidth)
	}

	return Logo(WithAppends([]byte("\n\n" + fmt.Sprintf(fmt.Sprintf("%%%ds", maxWidth), "created by duc@2023."))))
}
