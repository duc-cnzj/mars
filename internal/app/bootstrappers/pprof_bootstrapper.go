package bootstrappers

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type PprofBootstrapper struct{}

func (p *PprofBootstrapper) Tags() []string {
	return []string{"profile"}
}

func (p *PprofBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(&pprofRunner{
		server: &http.Server{
			Addr:              "localhost:6060",
			ReadHeaderTimeout: 5 * time.Second,
			Handler:           pprofMux(),
		},
	})

	return nil
}

type pprofRunner struct {
	server httpServer
}

func (p *pprofRunner) Run(ctx context.Context) error {
	mlog.Info("[Server]: start pprofRunner runner.")
	go func() {
		mlog.Info("Starting pprof server on localhost:6060.")
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mlog.Error(err)
		}
	}()

	return nil
}

func pprofMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}

func (p *pprofRunner) Shutdown(ctx context.Context) error {
	mlog.Info("[Server]: shutdown pprofRunner runner.")
	return p.server.Shutdown(ctx)
}
