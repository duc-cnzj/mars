// Package lock
//
// Laravel yyds!
package lock

import (
	"context"
	"math/rand"
	"time"

	"github.com/duc-cnzj/mars/internal/mlog"
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"gorm.io/gorm"
)

type timer interface {
	Unix() int64
}

type realTimers struct{}

func (r *realTimers) Unix() int64 {
	return time.Now().Unix()
}

type DatabaseLock struct {
	lottery [2]int
	timer   timer
	owner   string
	db      *gorm.DB
}

func NewDatabaseLock(lottery [2]int, db *gorm.DB) *DatabaseLock {
	return &DatabaseLock{lottery: lottery, db: db, owner: utils.RandomString(40), timer: &realTimers{}}
}

func (d *DatabaseLock) RenewalAcquire(key string, seconds int64, renewalSeconds int) (func(), bool) {
	if d.Acquire(key, seconds) {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		go func() {
			defer utils.HandlePanic("[lock]: key: " + key)

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

func (d *DatabaseLock) Acquire(key string, seconds int64) bool {
	var (
		acquired bool

		unix       = d.timer.Unix()
		expiration = unix + seconds
	)

	db := d.db.Create(&models.CacheLock{
		Key:        key,
		Owner:      d.owner,
		Expiration: expiration,
	})
	if db.Error == nil {
		acquired = true
	}
	if !acquired {
		updates := d.db.Model(&models.CacheLock{}).Where("`key` = ? AND `expiration` <= ?", key, unix).Updates(map[string]any{
			"owner":      d.owner,
			"expiration": expiration,
		})
		if updates.RowsAffected >= 1 {
			acquired = true
		}
	}

	if rand.Intn(d.lottery[1]) < d.lottery[0] {
		d.db.Where("`expiration` < ?", unix-60).Delete(&models.CacheLock{})
	}

	return acquired
}

func (d *DatabaseLock) renewalExistKey(key string, seconds int64) bool {
	var (
		acquired bool

		unix       = d.timer.Unix()
		expiration = unix + seconds
	)

	cl := &models.CacheLock{}
	first := d.db.Where("`key` = ?", key).First(cl)
	if first.Error != nil {
		return acquired
	}

	updates := d.db.Model(&models.CacheLock{}).Where("`key` = ? AND `owner` = ?", key, d.owner).Updates(map[string]any{
		"owner":      d.owner,
		"expiration": expiration,
	})
	if updates.RowsAffected >= 1 {
		acquired = true
	}

	return acquired
}

func (d *DatabaseLock) Release(key string) bool {
	if d.Owner(key) == d.owner {
		d.db.Where("`key` = ? AND `owner` = ?", key, d.owner).Delete(&models.CacheLock{})
		return true
	}

	return false
}

func (d *DatabaseLock) ForceRelease(key string) bool {
	d.db.Where("`key` = ?", key).Delete(&models.CacheLock{})
	return true
}

func (d *DatabaseLock) Owner(key string) string {
	cl := &models.CacheLock{}
	d.db.Where("`key` = ?", key).First(cl)

	return cl.Owner
}
