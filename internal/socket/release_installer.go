package socket

import (
	"context"
	"fmt"
	"time"

	websocket_pb "github.com/duc-cnzj/mars/api/v4/websocket"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

type releaseInstaller struct {
	logger         mlog.Logger
	helmer         repo.HelmerRepo
	dryRun         bool
	chart          *chart.Chart
	timeoutSeconds int64
	releaseName    string
	namespace      string
	percenter      contracts.Percentable
	startTime      time.Time
	wait           *bool
	valueOpts      *values.Options
	logs           *timeOrderedSetString
	messageCh      contracts.SafeWriteMessageChInterface
}

func newReleaseInstaller(logger mlog.Logger, helmer repo.HelmerRepo, releaseName, namespace string, chart *chart.Chart, valueOpts *values.Options, wait *bool, timeoutSeconds int64, dryRun bool) *releaseInstaller {
	return &releaseInstaller{
		logger:         logger,
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
	defer r.logger.Debug("releaseInstaller exit")

	r.messageCh = messageCh
	r.percenter = percenter
	r.startTime = time.Now()

	var wait = false
	if r.wait != nil {
		wait = *r.wait
	}
	re, err := r.helmer.UpgradeOrInstall(stopCtx, r.releaseName, r.namespace, r.chart, r.valueOpts, r.loggerWrap(), wait, r.timeoutSeconds, r.dryRun, desc)
	if err == nil {
		return re, nil
	}
	r.logger.Debug(err)
	if !r.dryRun && !isNew {
		// 失败了，需要手动回滚
		r.logger.Debug("rollback project")
		if err := r.helmer.Rollback(r.releaseName, r.namespace, false, r.loggerWrap().UnWrap(), r.dryRun); err != nil {
			r.logger.Debug(err)
		}
	}
	if !r.dryRun && isNew {
		r.logger.Debug("uninstall project")
		if err := r.helmer.Uninstall(r.releaseName, r.namespace, r.loggerWrap().UnWrap()); err != nil {
			r.logger.Debug(err)
		}
	}
	return nil, err
}

func (r *releaseInstaller) Logs() []string {
	return r.logs.sortedItems()
}

func (r *releaseInstaller) loggerWrap() repo.WrapLogFn {
	return func(containers []*websocket_pb.Container, format string, v ...any) {
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
