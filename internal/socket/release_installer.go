package socket

import (
	"context"
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/internal/utils"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/release"
)

type ReleaseInstaller interface {
	Chart() *chart.Chart
	Run(stopCtx context.Context, messageCh chan MessageItem, percenter Percentable) (*release.Release, error)

	Logs() []string
}

type releaseInstaller struct {
	chart       *chart.Chart
	releaseName string
	namespace   string
	percenter   Percentable
	startTime   time.Time
	atomic      bool
	valueOpts   *values.Options
	logs        *timeOrderedSetString
	messageCh   chan MessageItem
}

func newReleaseInstaller(releaseName, namespace string, chart *chart.Chart, valueOpts *values.Options, atomic bool) ReleaseInstaller {
	return &releaseInstaller{
		chart:       chart,
		valueOpts:   valueOpts,
		releaseName: releaseName,
		atomic:      atomic,
		namespace:   namespace,
		logs:        NewTimeOrderedSetString(),
	}
}

func (r *releaseInstaller) Chart() *chart.Chart {
	return r.chart
}

func (r *releaseInstaller) Run(stopCtx context.Context, messageCh chan MessageItem, percenter Percentable) (*release.Release, error) {
	defer utils.HandlePanic("releaseInstaller: Run")

	r.messageCh = messageCh
	r.percenter = percenter
	r.startTime = time.Now()
	return utils.UpgradeOrInstall(stopCtx, r.releaseName, r.namespace, r.chart, r.valueOpts, r.logger(), r.atomic)
}

func (r *releaseInstaller) Logs() []string {
	return r.logs.sortedItems()
}

func (r *releaseInstaller) logger() func(format string, v ...interface{}) {
	return func(format string, v ...interface{}) {
		if r.percenter.Current() < 99 {
			r.percenter.Add()
		}

		msg := fmt.Sprintf(format, v...)

		if time.Now().Sub(r.startTime.Add(time.Minute*3)).Seconds() > 0 {
			msg = fmt.Sprintf("[如果长时间未成功，请试试 debug 模式]: %s", msg)
		}

		if !r.logs.has(msg) {
			r.logs.add(msg)
			select {
			case r.messageCh <- MessageItem{
				Msg:  msg,
				Type: MessageText,
			}:
			default:
			}
		}
	}
}
