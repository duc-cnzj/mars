package bootstrappers

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
)

type PprofBootstrapper struct{}

func (p *PprofBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	if app.Config().ProfileEnabled {
		app.AddServer(&pprofRunner{})
	}

	return nil
}

type pprofRunner struct{}

func (p *pprofRunner) Run(ctx context.Context) error {
	mlog.Debug("[Runner]: start pprofRunner runner.")

	go func() {
		mlog.Info("Starting pprof server on localhost:6060.")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil && err != http.ErrServerClosed {
			mlog.Error(err)
		}
	}()

	return nil
}

func (p *pprofRunner) Shutdown(ctx context.Context) error {
	mlog.Info("[Runner]: shutdown pprofRunner runner.")

	return nil
}
