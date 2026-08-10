package server

import (
	"context"
	"net/http"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/application"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metricsRunner struct {
	port   string
	s      HttpServer
	logger mlog.Logger
	reg    *prometheus.Registry
}

// NewMetricsRunner 构建 metrics 传输层启动器：在指定端口暴露 /metrics 端点
// （OpenMetrics 格式，promhttp 从 prometheus.Registry 拉取）。返回 application.Server。
func NewMetricsRunner(port string, logger mlog.Logger, reg *prometheus.Registry) application.Server {
	return &metricsRunner{
		port:   port,
		logger: logger.WithModule("server/metricsRunner"),
		reg:    reg,
	}
}

// Run 启动 metrics 服务：注册 /metrics 处理器并 goroutine 内 ListenAndServe。
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

// Shutdown 优雅停止 metrics 服务。
func (m *metricsRunner) Shutdown(ctx context.Context) error {
	return m.s.Shutdown(ctx)
}
