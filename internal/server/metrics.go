package server

import (
	"context"
	"net/http"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metricsRunner struct {
	port   string
	s      HttpServer
	logger mlog.Logger
}

// NewMetricsRunner 构建 metrics 传输层启动器：在指定端口暴露 /metrics 端点
// （OpenMetrics 格式，promhttp 从 prometheus.Registry 拉取），server 在构造时装配
// （Handler 经 metricsHandler 构建），与 pprofRunner 的装配时机对齐。返回 app.Server。
func NewMetricsRunner(port string, logger mlog.Logger, reg *prometheus.Registry) app.Server {
	return &metricsRunner{
		port:   port,
		logger: logger.WithModule("server/metricsRunner"),
		s: &http.Server{
			Addr:              ":" + port,
			Handler:           metricsHandler(reg),
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

// Run 启动 metrics 服务：goroutine 内 ListenAndServe，非 ErrServerClosed 的启动错误
// 记录日志（与 apiGateway/pprofRunner 行为一致，避免端口冲突时指标服务静默下线）。
func (m *metricsRunner) Run(ctx context.Context) error {
	m.logger.Infof("[Server]: metrics running at :%s/metrics", m.port)
	go func() {
		if err := m.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.Error(err)
		}
	}()
	return nil
}

// metricsHandler 装配 /metrics 处理器：promhttp 从 registry 拉取指标并以 OpenMetrics
// 格式输出。独立成函数便于 httptest 直测端点行为，避免测试绑定真实端口。
func metricsHandler(reg *prometheus.Registry) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}))
	return mux
}

// Shutdown 优雅停止 metrics 服务。
func (m *metricsRunner) Shutdown(ctx context.Context) error {
	return m.s.Shutdown(ctx)
}
