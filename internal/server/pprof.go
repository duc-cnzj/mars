package server

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
)

type pprofRunner struct {
	server HttpServer
	logger mlog.Logger
}

// NewPprofRunner 构建 pprof 传输层启动器：在 localhost:6060 暴露 Go 性能剖析端点，
// 仅本机可访问，用于生产定位 CPU/内存/阻塞问题。返回 app.Server。
func NewPprofRunner(logger mlog.Logger) app.Server {
	return &pprofRunner{
		logger: logger.WithModule("server/pprofRunner"),
		server: &http.Server{
			Addr:              "localhost:6060",
			ReadHeaderTimeout: 5 * time.Second,
			Handler:           pprofMux(),
		}}
}

// Run 启动 pprof 服务：goroutine 内 ListenAndServe，非正常关闭的错误才记录。
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

// pprofMux 装配 /debug/pprof 系列剖析端点（Index/Cmdline/Profile/Symbol/Trace）。
func pprofMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}

// Shutdown 优雅停止 pprof 服务。
func (p *pprofRunner) Shutdown(ctx context.Context) error {
	p.logger.Info("[Server]: shutdown pprofRunner runner.")
	return p.server.Shutdown(ctx)
}
