package bootstrappers

import (
	"context"
	"net/http"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
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
}

func (m *metricsRunner) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mlog.Infof("[Server]: metrics running at :%s/metrics", m.port)
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":"+m.port, mux)
	}()
	return nil
}

func (m *metricsRunner) Shutdown(ctx context.Context) error {
	return nil
}
