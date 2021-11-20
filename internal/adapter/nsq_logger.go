package adapter

import "github.com/duc-cnzj/mars/internal/mlog"

type NsqLoggerAdapter struct {
}

func (*NsqLoggerAdapter) Output(calldepth int, s string) error {
	mlog.Error(s)
	return nil
}
