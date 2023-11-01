package bootstrappers

import (
	"context"
	"net/http"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsBootstrapper struct{}

func (m *MetricsBootstrapper) Tags() []string {
	return []string{"metrics"}
}

func (m *MetricsBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(&metricsRunner{port: app.Config().MetricsPort})

	return nil
}

type metricsRunner struct {
	port string
	s    httpServer
}

func (m *metricsRunner) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mlog.Infof("[Server]: metrics running at :%s/metrics", m.port)
	mux.Handle("/metrics", promhttp.Handler())
	m.s = &http.Server{Addr: ":" + m.port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		m.s.ListenAndServe()
	}()
	return nil
}

func (m *metricsRunner) Shutdown(ctx context.Context) error {
	return m.s.Shutdown(ctx)
}
