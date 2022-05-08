package bootstrappers

import (
	"errors"

	"github.com/duc-cnzj/mars/internal/adapter"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type LogBootstrapper struct{}

func (a *LogBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	switch app.Config().LogChannel {
	case "", "logrus":
		mlog.SetLogger(adapter.NewLogrusLogger(app))
	case "zap":
		mlog.SetLogger(adapter.NewZapLogger(app))
	default:
		return errors.New("log channel not exists: " + app.Config().LogChannel)
	}
	mlog.Debug("LogBootstrapper booted!")

	return nil
}
