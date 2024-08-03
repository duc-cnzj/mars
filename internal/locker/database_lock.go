// Package locker
//
// Laravel yyds!
package locker

import (
	"context"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/ent"
	"github.com/duc-cnzj/mars/v4/internal/ent/cachelock"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/utils/rand"
	"github.com/duc-cnzj/mars/v4/internal/utils/timer"
)

type databaseLock struct {
	lottery [2]int
	timer   timer.Timer
	owner   string
	db      *ent.Client
	logger  mlog.Logger
}

func NewDatabaseLock(timer timer.Timer, lottery [2]int, db *ent.Client, logger mlog.Logger) Locker {
	return &databaseLock{lottery: lottery, db: db, owner: rand.String(40), timer: timer, logger: logger}
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
						d.logger.Warning("[lock]: err renewal lock: " + key)
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

		db = d.db
	)

	_, err := db.CacheLock.Create().SetKey(key).SetOwner(d.owner).SetExpiredAt(d.timer.Now().Add(time.Duration(seconds) * time.Second)).Save(context.TODO())
	if err == nil {
		acquired = true
	}
	if !acquired {
		rowsAffected, _ := db.CacheLock.Update().Where(cachelock.Key(key)).SetOwner(d.owner).SetExpiredAt(d.timer.Now().Add(time.Duration(seconds) * time.Second)).Save(context.TODO())
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

		db = d.db
	)

	_, err := db.CacheLock.Query().Where(cachelock.Key(key)).Only(context.TODO())
	if err != nil {
		return acquired
	}

	rowsAffected, _ := db.CacheLock.Update().Where(cachelock.Key(key)).SetOwner(d.owner).SetExpiredAt(d.timer.Now().Add(time.Duration(seconds) * time.Second)).Save(context.TODO())
	if rowsAffected >= 1 {
		acquired = true
	}

	return acquired
}

func (d *databaseLock) Release(key string) bool {
	if d.Owner(key) == d.owner {
		d.db.CacheLock.Delete().Where(cachelock.Key(key), cachelock.Owner(d.owner)).Exec(context.TODO())
		return true
	}

	return false
}

func (d *databaseLock) ForceRelease(key string) bool {
	d.db.CacheLock.Delete().Where(cachelock.Key(key)).Exec(context.TODO())

	return true
}

func (d *databaseLock) Owner(key string) string {
	cl, _ := d.db.CacheLock.Query().Where(cachelock.Key(key)).First(context.TODO())

	return cl.Owner
}
