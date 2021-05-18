package adapter

import (
	"strings"

	"github.com/DuC-cnZj/mars/pkg/mlog"
)

type GinWriter struct{}

func (e *GinWriter) Write(p []byte) (n int, err error) {
	if strings.Index(string(p), "[GIN-debug]") == 0 {
		return 0, nil
	}

	mlog.Debug(string(p))

	return len(p), nil
}
