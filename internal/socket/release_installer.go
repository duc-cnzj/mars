package socket

import (
	"context"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/api/v4/types"
	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

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

func newReleaseInstaller(helmer contracts.Helmer, releaseName, namespace string, chart *chart.Chart, valueOpts *values.Options, wait bool, timeoutSeconds int64, dryRun bool) *releaseInstaller {
	return &releaseInstaller{
		helmer:         helmer,
		dryRun:         dryRun,
		chart:          chart,
		valueOpts:      valueOpts,
		releaseName:    releaseName,
		wait:           wait,
		namespace:      namespace,
		logs:           newTimeOrderedSetString(time.Now),
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
