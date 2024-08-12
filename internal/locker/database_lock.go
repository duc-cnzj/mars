// Package locker
//
// Laravel yyds!
package locker

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent/cachelock"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/util/rand"
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
)

type databaseLock struct {
	lottery [2]int
	timer   timer.Timer
	owner   string
	data    data.Data
	logger  mlog.Logger
}

func NewDatabaseLock(timer timer.Timer, lottery [2]int, data data.Data, logger mlog.Logger) Locker {
	return &databaseLock{lottery: lottery, data: data, owner: rand.String(40), timer: timer, logger: logger}
}

func (d *databaseLock) RenewalAcquire(key string, seconds int64, renewalSeconds int64) (func(), bool) {
	if d.Acquire(key, seconds) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		go func() {
			defer d.logger.HandlePanic("[lock]: key: " + key)

			ticker := time.NewTicker(time.Second * time.Duration(renewalSeconds))
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					d.logger.Debug("[lock]: canceled: " + key)
					return
				case <-ticker.C:
					if !d.renewalExistKey(key, seconds) {
						d.logger.Error("[lock]: err renewal lock: " + key)
						return
					}
				}
			}
		}()
		return func() {
			cancelFunc()
			d.Release(key)
		}, true
	}
	return nil, false
}

func (d *databaseLock) ID() string {
	return d.owner
}

func (d *databaseLock) Type() string {
	return "database"
}

func (d *databaseLock) Acquire(key string, seconds int64) bool {
	var (
		acquired bool

		db        = d.data.DB()
		now       = d.timer.Now()
		expiredAt = now.Add(time.Duration(seconds) * time.Second)
	)

	_, err := db.CacheLock.Create().
		SetKey(key).
		SetOwner(d.owner).
		SetExpiredAt(expiredAt).
		Save(context.TODO())
	if err == nil {
		acquired = true
	}
	if !acquired {
		rowsAffected, _ := db.CacheLock.
			Update().
			Where(cachelock.Key(key), cachelock.ExpiredAtLTE(now)).
			SetOwner(d.owner).
			SetExpiredAt(d.timer.Now().Add(time.Duration(seconds) * time.Second)).
			Save(context.TODO())
		if rowsAffected >= 1 {
			acquired = true
		}
	}

	if rand.Intn(d.lottery[1]) < d.lottery[0] {
		db.CacheLock.Delete().Where(cachelock.ExpiredAtLT(d.timer.Now().Add(-60 * time.Second))).Exec(context.TODO())
	}

	return acquired
}

func (d *databaseLock) renewalExistKey(key string, seconds int64) bool {
	var (
		acquired bool

		db = d.data.DB()
	)

	_, err := db.CacheLock.Query().Where(cachelock.Key(key)).Only(context.TODO())
	if err != nil {
		d.logger.Error(err)
		return acquired
	}

	rowsAffected, err := db.CacheLock.Update().Where(cachelock.Key(key)).SetOwner(d.owner).SetExpiredAt(d.timer.Now().Add(time.Duration(seconds) * time.Second)).Save(context.TODO())
	if err != nil {
		d.logger.Error(err)
	}
	if rowsAffected >= 1 {
		acquired = true
	}

	return acquired
}

func (d *databaseLock) Release(key string) bool {
	if d.Owner(key) == d.owner {
		d.data.DB().CacheLock.Delete().Where(cachelock.Key(key), cachelock.Owner(d.owner)).Exec(context.TODO())
		return true
	}

	return false
}

func (d *databaseLock) ForceRelease(key string) bool {
	d.data.DB().CacheLock.Delete().Where(cachelock.Key(key)).Exec(context.TODO())

	return true
}

func (d *databaseLock) Owner(key string) string {
	cl, err := d.data.DB().CacheLock.Query().Where(cachelock.Key(key)).First(context.TODO())
	if err != nil {
		return ""
	}

	return cl.Owner
}
