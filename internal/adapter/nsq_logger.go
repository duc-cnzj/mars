package adapter

import (
	"strings"

	"github.com/duc-cnzj/mars/internal/mlog"
)

type NsqLoggerAdapter struct{}

func (*NsqLoggerAdapter) Output(calldepth int, s string) error {
	if strings.Contains(s, "TOPIC_NOT_FOUND") {
		mlog.Debug(s)
	} else {
		mlog.Error(s)
	}
	return nil
}
