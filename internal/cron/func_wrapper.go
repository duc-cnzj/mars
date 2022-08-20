package cron

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/duc-cnzj/mars/internal/utils/recovery"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/metrics"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultLockSeconds  int64 = 30
	defaultRenewSeconds int64 = 20
)

func Wrap(name string, fn func() error, lockerFn func() contracts.Locker) func() {
	label := prometheus.Labels{"cron_name": name}
	return func() {
		defer recovery.HandlePanicWithCallback("[CRON]: "+name, func(err error) {
			metrics.CronPanicCount.With(label).Inc()
		})

		time.Sleep(time.Duration(rand.Intn(20)) * time.Microsecond)
		releaseFn, acquired := lockerFn().RenewalAcquire(lockKey(name), defaultLockSeconds, defaultRenewSeconds)
		if acquired {
			now := time.Now()
			defer func(t time.Time) {
				mlog.Infof("[CRON-DONE: %s]: '%s' done, use %s.", lockerFn().ID(), name, time.Since(t))
				metrics.CronDuration.With(label).Observe(time.Since(t).Seconds())
				metrics.CronCommandCount.With(label).Inc()
			}(now)
			mlog.Infof("[CRON-START: %s]: '%s' start at %s.", lockerFn().ID(), name, now.Format("2006-01-02 15:04:05.000"))
			defer releaseFn()

			if err := fn(); err != nil {
				mlog.Errorf("[CRON]: '%s' err: '%v'", name, err)
				metrics.CronErrorCount.With(label).Inc()
			}
		}
	}
}

func lockKey(name string) string {
	return fmt.Sprintf("cron-%s", name)
}
