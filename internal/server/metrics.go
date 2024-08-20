package server

import (
	"context"
	"net/http"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HttpServer interface {
	Shutdown(ctx context.Context) error
	ListenAndServe() error
}

type metricsRunner struct {
	port   string
	s      HttpServer
	logger mlog.Logger
	reg    *prometheus.Registry
}

func NewMetricsRunner(port string, logger mlog.Logger, reg *prometheus.Registry) application.Server {
	return &metricsRunner{
		port:   port,
		logger: logger.WithModule("server/metricsRunner"),
		reg:    reg,
	}
}

func (m *metricsRunner) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	m.logger.Infof("[Server]: metrics running at :%s/metrics", m.port)

	mux.Handle(
		"/metrics", promhttp.HandlerFor(
			m.reg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			}),
	)

	m.s = &http.Server{Addr: ":" + m.port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		m.s.ListenAndServe()
	}()
	return nil
}

func (m *metricsRunner) Shutdown(ctx context.Context) error {
	return m.s.Shutdown(ctx)
}
