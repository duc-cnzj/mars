// Package lock
//
// Laravel yyds!
package lock

import (
	"context"
	"math/rand"
	"time"

	"gorm.io/gorm"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"github.com/duc-cnzj/mars/internal/utils/recovery"
)

type timer interface {
	Unix() int64
}

type realTimers struct{}

func (r *realTimers) Unix() int64 {
	return time.Now().Unix()
}

type databaseLock struct {
	lottery [2]int
	timer   timer
	owner   string
	dbFunc  func() *gorm.DB
}

func NewDatabaseLock(lottery [2]int, dbFunc func() *gorm.DB) contracts.Locker {
	return &databaseLock{lottery: lottery, dbFunc: dbFunc, owner: utils.RandomString(40), timer: &realTimers{}}
}

func (d *databaseLock) RenewalAcquire(key string, seconds int64, renewalSeconds int64) (func(), bool) {
	if d.Acquire(key, seconds) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		go func() {
			defer recovery.HandlePanic("[lock]: key: " + key)

			ticker := time.NewTicker(time.Second * time.Duration(renewalSeconds))
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					mlog.Debug("[lock]: canceled: " + key)
					return
				case <-ticker.C:
					if !d.renewalExistKey(key, seconds) {
						mlog.Warning("[lock]: err renewal lock: " + key)
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

		unix       = d.timer.Unix()
		expiration = unix + seconds
	)

	db := d.dbFunc().Create(&models.CacheLock{
		Key:        key,
		Owner:      d.owner,
		Expiration: expiration,
	})
	if db.Error == nil {
		acquired = true
	}
	if !acquired {
		updates := d.dbFunc().Model(&models.CacheLock{}).Where("`key` = ? AND `expiration` <= ?", key, unix).Updates(map[string]any{
			"owner":      d.owner,
			"expiration": expiration,
		})
		if updates.RowsAffected >= 1 {
			acquired = true
		}
	}

	if rand.Intn(d.lottery[1]) < d.lottery[0] {
		d.dbFunc().Where("`expiration` < ?", unix-60).Delete(&models.CacheLock{})
	}

	return acquired
}

func (d *databaseLock) renewalExistKey(key string, seconds int64) bool {
	var (
		acquired bool

		unix       = d.timer.Unix()
		expiration = unix + seconds
	)

	cl := &models.CacheLock{}
	first := d.dbFunc().Where("`key` = ?", key).First(cl)
	if first.Error != nil {
		return acquired
	}

	updates := d.dbFunc().Model(&models.CacheLock{}).Where("`key` = ? AND `owner` = ?", key, d.owner).Updates(map[string]any{
		"owner":      d.owner,
		"expiration": expiration,
	})
	if updates.RowsAffected >= 1 {
		acquired = true
	}

	return acquired
}

func (d *databaseLock) Release(key string) bool {
	if d.Owner(key) == d.owner {
		d.dbFunc().Where("`key` = ? AND `owner` = ?", key, d.owner).Delete(&models.CacheLock{})
		return true
	}

	return false
}

func (d *databaseLock) ForceRelease(key string) bool {
	d.dbFunc().Where("`key` = ?", key).Delete(&models.CacheLock{})
	return true
}

func (d *databaseLock) Owner(key string) string {
	cl := &models.CacheLock{}
	d.dbFunc().Where("`key` = ?", key).First(cl)

	return cl.Owner
}
