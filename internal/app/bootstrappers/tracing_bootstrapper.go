package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

const serviceName = "mars"

type TracingBootstrapper struct{}

func (t *TracingBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	if app.Config().JaegerAgentHostPort == "" {
		return nil
	}
	var (
		samplerType  string  = jaeger.SamplerTypeProbabilistic
		samplerParam float64 = 0.3
	)
	if app.IsDebug() {
		samplerType = jaeger.SamplerTypeConst
		samplerParam = 1
	}
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  samplerType,
			Param: samplerParam,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: app.Config().JaegerAgentHostPort,
			User:               app.Config().JaegerUser,
			Password:           app.Config().JaegerPassword,
		},
	}
	// Initialize tracer with a logger and a metrics factory
	closer, err := cfg.InitGlobalTracer(serviceName)
	if err != nil {
		return err
	}
	app.RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
		err := closer.Close()
		mlog.Infof("[Tracer]: shutdown. %v", err)
	})

	return nil
}
