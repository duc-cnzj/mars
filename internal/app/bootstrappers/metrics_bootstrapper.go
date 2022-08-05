package bootstrappers

import (
	"context"
	"net/http"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsBootstrapper struct{}

func (m *MetricsBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.AddServer(&metricsRunner{})

	return nil
}

type metricsRunner struct{}

func (m *metricsRunner) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mlog.Info("[Server]: metrics running at :9091/metrics")
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":9091", mux)
	}()
	return nil
}

func (m *metricsRunner) Shutdown(ctx context.Context) error {
	return nil
}
