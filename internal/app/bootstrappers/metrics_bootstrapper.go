package bootstrappers

import (
	"context"
	"net/http"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsBootstrapper struct{}

type metricsRunner struct{}

func (m *metricsRunner) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mlog.Debug("[Runner]: metrics running at :9091/metrics")
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":9091", mux)
	}()
	return nil
}

func (m *metricsRunner) Shutdown(ctx context.Context) error {
	return nil
}

func (m *MetricsBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	conns := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "mars",
		Subsystem: "pubsub",
		Name:      "current_connections",
		Help:      "当前 websocket 连接数",
	})
	prometheus.MustRegister(conns)

	app.SetMetrics(&metrics{ws: conns})
	app.AddServer(&metricsRunner{})

	return nil
}

type metrics struct {
	ws prometheus.Gauge
}

func (m *metrics) IncWebsocketConn() {
	m.WebsocketConn().Inc()
}

func (m *metrics) DecWebsocketConn() {
	m.WebsocketConn().Dec()
}

func (m *metrics) WebsocketConn() prometheus.Gauge {
	return m.ws
}
