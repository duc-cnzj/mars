// Package lock
//
// Laravel yyds!
package lock

import (
	"github.com/duc-cnzj/mars/internal/models"
	"github.com/duc-cnzj/mars/internal/utils"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type timer interface {
	Unix() int64
	Now() time.Time
}

type realTimers struct{}

func (r *realTimers) Unix() int64 {
	return time.Now().Unix()
}

func (r *realTimers) Now() time.Time {
	return time.Now()
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
		updates := d.db.Model(&models.CacheLock{}).Where("`key` = ? AND (`owner` = ? or `expiration` <= ?)", key, d.owner, unix).Updates(map[string]any{
			"owner":      d.owner,
			"expiration": expiration,
		})
		if updates.RowsAffected >= 1 {
			acquired = true
		}
	}

	if rand.Intn(d.lottery[1]) < d.lottery[0] {
		d.db.Where("`expiration` < ?", unix).Delete(&models.CacheLock{})
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
