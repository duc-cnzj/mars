// Package locker
//
// Laravel yyds!
package locker

import (
	"context"
	"errors"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/ent"
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
	return &databaseLock{
		lottery: lottery,
		timer:   timer,
		owner:   rand.String(40),
		data:    data,
		logger:  logger,
	}
}

func (d *databaseLock) RenewalAcquire(key string, seconds int64, renewalSeconds int64) (func(), bool) {
	if d.Acquire(key, seconds) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		go d.renewalRoutine(ctx, key, seconds, renewalSeconds)
		return func() {
			cancelFunc()
			d.Release(key)
		}, true
	}
	return nil, false
}

func (d *databaseLock) renewalRoutine(ctx context.Context, key string, seconds, renewalSeconds int64) {
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
}

func (d *databaseLock) ID() string {
	return d.owner
}

func (d *databaseLock) Type() string {
	return "database"
}

func (d *databaseLock) Acquire(key string, seconds int64) bool {
	db := d.data.DB()
	now := d.timer.Now()
	expiredAt := now.Add(time.Duration(seconds) * time.Second)

	var acquired bool

	if d.createLock(db, key, expiredAt) {
		acquired = true
	}
	if !acquired && d.updateExpiredLock(db, key, expiredAt) {
		acquired = true
	}

	if rand.Intn(d.lottery[1]) < d.lottery[0] {
		d.cleanupExpiredLocks(db)
	}

	return acquired
}

func (d *databaseLock) createLock(db *ent.Client, key string, expiredAt time.Time) bool {
	_, err := db.CacheLock.Create().
		SetKey(key).
		SetOwner(d.owner).
		SetExpiredAt(expiredAt).
		Save(context.TODO())
	return err == nil
}

func (d *databaseLock) updateExpiredLock(db *ent.Client, key string, expiredAt time.Time) bool {
	rowsAffected, err := db.CacheLock.
		Update().
		Where(cachelock.Key(key), cachelock.ExpiredAtLTE(d.timer.Now())).
		SetOwner(d.owner).
		SetExpiredAt(expiredAt).
		Save(context.TODO())
	if err != nil {
		d.logger.Error(err)
		return false
	}
	return rowsAffected >= 1
}

func (d *databaseLock) cleanupExpiredLocks(db *ent.Client) {
	_, err := db.CacheLock.Delete().Where(cachelock.ExpiredAtLT(d.timer.Now().Add(-60 * time.Second))).Exec(context.TODO())
	if err != nil {
		d.logger.Error(err)
	}
}

func (d *databaseLock) renewalExistKey(key string, seconds int64) bool {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunc()
	if err := d.data.WithTx(ctx, func(db *ent.Tx) error {
		var (
			err  error
			item *ent.CacheLock
		)
		item, err = db.CacheLock.Query().Where(cachelock.Key(key)).ForUpdate().Only(ctx)
		if err != nil {
			return err
		}
		if item.Owner != d.owner {
			return errors.New("not owner")
		}

		_, err = item.Update().
			SetOwner(d.owner).
			SetExpiredAt(d.timer.Now().Add(time.Duration(seconds) * time.Second)).
			Save(ctx)
		return err
	}); err != nil {
		return false
	}

	return true
}

func (d *databaseLock) Release(key string) bool {
	if d.Owner(key) != d.owner {
		return false
	}
	d.data.DB().CacheLock.Delete().Where(cachelock.Key(key), cachelock.Owner(d.owner)).Exec(context.TODO())
	return true
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
