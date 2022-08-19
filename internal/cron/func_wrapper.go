package cron

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/metrics"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/utils/recovery"

	"github.com/prometheus/client_golang/prometheus"
)

func Wrap(name string, fn func(), locker contracts.Locker) func() {
	label := prometheus.Labels{"cron_name": name}
	return func() {
		defer recovery.HandlePanicWithCallback("[CRON]: "+name, func(err error) {
			metrics.CronPanicCount.With(label).Inc()
		})

		time.Sleep(time.Duration(rand.Intn(20)) * time.Microsecond)
		key := fmt.Sprintf("cron-%s", name)
		releaseFn, acquired := locker.RenewalAcquire(key, 30, 20)
		if acquired {
			defer func(t time.Time) {
				metrics.CronDuration.With(label).Observe(time.Since(t).Seconds())
				metrics.CronCommandCount.With(label).Inc()
			}(time.Now())
			defer releaseFn()

			mlog.Debugf("[CRON]: run %s at %s", name, time.Now())
			fn()
		}
	}
}
