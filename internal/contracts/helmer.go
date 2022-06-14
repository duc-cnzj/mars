package contracts

import (
	"context"

	"github.com/duc-cnzj/mars-client/v4/types"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

//go:generate mockgen -destination ../mock/mock_helmer.go -package mock github.com/duc-cnzj/mars/internal/contracts  Helmer

type Helmer interface {
	UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn func(format string, v ...any), wait bool, timeoutSeconds int64, dryRun bool) (*release.Release, error)
	Rollback(releaseName, namespace string, wait bool, log action.DebugLog, dryRun bool) error
	Uninstall(releaseName, namespace string, log action.DebugLog) error
	ReleaseStatus(namespace, releaseName string) types.Deploy
	PackageChart(path string, destDir string) (string, error)
}
