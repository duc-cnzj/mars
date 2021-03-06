package socket

import (
	"context"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars-client/v4/types"

	"helm.sh/helm/v3/pkg/action"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

type DefaultHelmer struct{}

func (d *DefaultHelmer) UpgradeOrInstall(ctx context.Context, releaseName, namespace string, ch *chart.Chart, valueOpts *values.Options, fn func(format string, v ...any), wait bool, timeoutSeconds int64, dryRun bool) (*release.Release, error) {
	return utils.UpgradeOrInstall(ctx, releaseName, namespace, ch, valueOpts, fn, wait, timeoutSeconds, dryRun)
}

func (d *DefaultHelmer) Rollback(releaseName, namespace string, wait bool, log action.DebugLog, dryRun bool) error {
	return utils.Rollback(releaseName, namespace, wait, log, dryRun)
}

func (d *DefaultHelmer) PackageChart(path string, destDir string) (string, error) {
	return utils.PackageChart(path, destDir)
}

func (d *DefaultHelmer) Uninstall(releaseName, namespace string, log action.DebugLog) error {
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

func (r *releaseInstaller) Run(stopCtx context.Context, messageCh contracts.SafeWriteMessageChInterface, percenter contracts.Percentable, isNew bool) (*release.Release, error) {
	defer utils.HandlePanic("releaseInstaller: Run")
	defer mlog.Debug("releaseInstaller exit")

	r.messageCh = messageCh
	r.percenter = percenter
	r.startTime = time.Now()
	re, err := r.helmer.UpgradeOrInstall(stopCtx, r.releaseName, r.namespace, r.chart, r.valueOpts, r.logger(), r.wait, r.timeoutSeconds, r.dryRun)
	if err == nil {
		return re, nil
	}
	mlog.Debug(err)
	if !r.dryRun && !isNew {
		// ??????????????????????????????
		mlog.Debug("rollback project")
		if err := r.helmer.Rollback(r.releaseName, r.namespace, false, r.logger(), r.dryRun); err != nil {
			mlog.Debug(err)
		}
	}
	if !r.dryRun && isNew {
		mlog.Debug("uninstall project")
		if err := r.helmer.Uninstall(r.releaseName, r.namespace, r.logger()); err != nil {
			mlog.Debug(err)
		}
	}
	return nil, err
}

func (r *releaseInstaller) Logs() []string {
	return r.logs.sortedItems()
}

func (r *releaseInstaller) logger() func(format string, v ...any) {
	return func(format string, v ...any) {
		if r.percenter.Current() < 99 {
			r.percenter.Add()
		}

		msg := fmt.Sprintf(format, v...)

		if time.Since(r.startTime.Add(time.Minute*3)).Seconds() > 0 {
			msg = fmt.Sprintf("[???????????????????????????????????? debug ??????]: %s", msg)
		}

		if !r.logs.has(msg) {
			r.logs.add(msg)
			r.messageCh.Send(contracts.MessageItem{
				Msg:  msg,
				Type: contracts.MessageText,
			})
		}
	}
}
