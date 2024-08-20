package server

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
)

type pprofRunner struct {
	server HttpServer
	logger mlog.Logger
}

func NewPprofRunner(logger mlog.Logger) application.Server {
	return &pprofRunner{
		logger: logger.WithModule("server/pprofRunner"),
		server: &http.Server{
			Addr:              "localhost:6060",
			ReadHeaderTimeout: 5 * time.Second,
			Handler:           pprofMux(),
		}}
}

func (p *pprofRunner) Run(ctx context.Context) error {
	p.logger.Info("[Server]: start pprofRunner runner.")
	go func() {
		p.logger.Info("Starting pprof server on localhost:6060.")
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.Error(err)
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
	p.logger.Info("[Server]: shutdown pprofRunner runner.")
	return p.server.Shutdown(ctx)
}
