package app

import (
	"github.com/duc-cnzj/mars/v4/internal/app/instance"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/contracts"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// App return contracts.ApplicationInterface
func App() contracts.ApplicationInterface {
	return instance.App()
}

// Auth return contracts.AuthInterface
func Auth() contracts.AuthInterface {
	return App().Auth()
}

// Oidc return contracts.OidcConfig
func Oidc() contracts.OidcConfig {
	return App().Oidc()
}

func Config() *config.Config {
	return App().Config()
}

func DB() *gorm.DB {
	return App().DBManager().DB()
}

// Uploader return contracts.Uploader
func Uploader() contracts.Uploader {
	return App().Uploader()
}

// LocalUploader return contracts.Uploader
func LocalUploader() contracts.Uploader {
	return App().LocalUploader()
}

// Event return contracts.DispatcherInterface
func Event() contracts.DispatcherInterface {
	return App().EventDispatcher()
}

// K8sClient return *contracts.K8sClient
func K8sClient() *contracts.K8sClient {
	return App().K8sClient()
}

// K8sClientSet return kubernetes.Interface
func K8sClientSet() kubernetes.Interface {
	return App().K8sClient().Client
}

// K8sMetrics return versioned.Interface
func K8sMetrics() versioned.Interface {
	return App().K8sClient().MetricsClient
}

// Singleflight return *singleflight.Group
func Singleflight() *singleflight.Group {
	return App().Singleflight()
}

// Cache return contracts.CacheInterface
func Cache() contracts.CacheInterface {
	return App().Cache()
}

// Tracer return trace.Tracer
func Tracer() trace.Tracer {
	return App().GetTracer()
}

// CacheLock return contracts.Locker
func CacheLock() contracts.Locker {
	return App().CacheLock()
}

// CronManager return contracts.CronManager
func CronManager() contracts.CronManager {
	return App().CronManager()
}

// Helmer return contracts.Helmer
func Helmer() contracts.Helmer {
	return App().Helmer()
}
