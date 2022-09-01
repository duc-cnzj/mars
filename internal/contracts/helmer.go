package contracts

import (
	"context"

	"github.com/duc-cnzj/mars-client/v4/types"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

//go:generate mockgen -destination ../mock/mock_helmer.go -package mock github.com/duc-cnzj/mars/internal/contracts Helmer

type LogFn func(container []*types.Container, format string, v ...any)

type Helmer interface {
	UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn LogFn, wait bool, timeoutSeconds int64, dryRun bool, lls []string) (*release.Release, error)
	Rollback(releaseName, namespace string, wait bool, log LogFn, dryRun bool) error
	Uninstall(releaseName, namespace string, log LogFn) error
	ReleaseStatus(releaseName, namespace string) types.Deploy
	PackageChart(path string, destDir string) (string, error)
}
