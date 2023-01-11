package socket

import (
	"context"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

type DefaultHelmer struct{}

func (d *DefaultHelmer) UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn contracts.WrapLogFn, wait bool, timeoutSeconds int64, dryRun bool, desc string) (*release.Release, error) {
	var podSelectors []string
	if wait && !dryRun {
		re, err := utils.UpgradeOrInstall(context.TODO(), releaseName, namespace, ch, valueOpts, func(container []*types.Container, format string, v ...any) {}, false, timeoutSeconds, true, nil, desc)
		if err != nil {
			return nil, err
		}

		podSelectors = getPodSelectorsByManifest(utils.SplitManifests(re.Manifest))
	}

	return utils.UpgradeOrInstall(ctx, releaseName, namespace, ch, valueOpts, fn, wait, timeoutSeconds, dryRun, podSelectors, desc)
}

func (d *DefaultHelmer) Rollback(releaseName, namespace string, wait bool, log contracts.LogFn, dryRun bool) error {
	return utils.Rollback(releaseName, namespace, wait, log, dryRun)
}

func (d *DefaultHelmer) PackageChart(path string, destDir string) (string, error) {
	return utils.PackageChart(path, destDir)
}

func (d *DefaultHelmer) Uninstall(releaseName, namespace string, log contracts.LogFn) error {
	return utils.UninstallRelease(releaseName, namespace, log)
}

func (d *DefaultHelmer) ReleaseStatus(releaseName, namespace string) types.Deploy {
	return utils.ReleaseStatus(releaseName, namespace)
}

type releaseInstaller struct {
	helmer         contracts.Helmer
	dryRun         bool
	chart          *chart.Chart
	timeoutSeconds int64
	releaseName    string
	namespace      string
	percenter      contracts.Percentable
	startTime      time.Time
	wait           bool
	valueOpts      *values.Options
	logs           *timeOrderedSetString
	messageCh      contracts.SafeWriteMessageChInterface
}

func newReleaseInstaller(releaseName, namespace string, chart *chart.Chart, valueOpts *values.Options, wait bool, timeoutSeconds int64, dryRun bool) *releaseInstaller {
	return &releaseInstaller{
		helmer:         &DefaultHelmer{},
		dryRun:         dryRun,
		chart:          chart,
		valueOpts:      valueOpts,
		releaseName:    releaseName,
		wait:           wait,
		namespace:      namespace,
		logs:           NewTimeOrderedSetString(time.Now),
		timeoutSeconds: timeoutSeconds,
	}
}

func (r *releaseInstaller) Chart() *chart.Chart {
	return r.chart
}

func (r *releaseInstaller) Run(stopCtx context.Context, messageCh contracts.SafeWriteMessageChInterface, percenter contracts.Percentable, isNew bool, desc string) (*release.Release, error) {
	defer mlog.Debug("releaseInstaller exit")

	r.messageCh = messageCh
	r.percenter = percenter
	r.startTime = time.Now()

	re, err := r.helmer.UpgradeOrInstall(stopCtx, r.releaseName, r.namespace, r.chart, r.valueOpts, r.logger(), r.wait, r.timeoutSeconds, r.dryRun, desc)
	if err == nil {
		return re, nil
	}
	mlog.Debug(err)
	if !r.dryRun && !isNew {
		// 失败了，需要手动回滚
		mlog.Debug("rollback project")
		if err := r.helmer.Rollback(r.releaseName, r.namespace, false, r.logger().UnWrap(), r.dryRun); err != nil {
			mlog.Debug(err)
		}
	}
	if !r.dryRun && isNew {
		mlog.Debug("uninstall project")
		if err := r.helmer.Uninstall(r.releaseName, r.namespace, r.logger().UnWrap()); err != nil {
			mlog.Debug(err)
		}
	}
	return nil, err
}

func (r *releaseInstaller) Logs() []string {
	return r.logs.sortedItems()
}

func (r *releaseInstaller) logger() contracts.WrapLogFn {
	return func(containers []*types.Container, format string, v ...any) {
		if r.percenter.Current() < 99 {
			r.percenter.Add()
		}

		msg := fmt.Sprintf(format, v...)

		if !r.logs.has(msg) {
			r.logs.add(msg)
			r.messageCh.Send(contracts.MessageItem{
				Msg:        msg,
				Containers: containers,
				Type:       contracts.MessageText,
			})
		}
	}
}
