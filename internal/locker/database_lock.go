// Package locker 提供进程内内存与数据库两种后端的分布式锁实现，
// 通过 Locker 接口统一暴露加锁、续期与释放能力。
package locker

import (
	"context"
	"errors"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/cachelock"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
)

// databaseLock 是基于数据库缓存表 cache_locks 的 Locker 实现，适用于多实例场景。
type databaseLock struct {
	lottery [2]int
	timer   timer.Timer
	owner   string
	data    data.Data
	logger  mlog.Logger
}

// NewDatabaseLock 创建一个基于数据库的锁。
//
// lottery 是 [分子, 分母] 组合：每次 Acquire 有 lottery[0]/lottery[1] 的概率触发一次过期锁清理。
func NewDatabaseLock(timer timer.Timer, lottery [2]int, data data.Data, logger mlog.Logger) Locker {
	return &databaseLock{
		lottery: lottery,
		timer:   timer,
		owner:   rand.String(40),
		data:    data,
		logger:  logger,
	}
}

// RenewalAcquire 获取 key 锁并启动后台续期协程，返回 release 函数与是否获取成功。
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

// renewalRoutine 在后台周期续期锁，直到 ctx 取消或续期失败。
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
			if err := d.renewalExistKey(key, seconds); err != nil {
				d.logger.Errorf("[lock]: err renewal lock: %v", err)
				return
			}
		}
	}
}

// ID 返回当前锁实例的持有者标识。
func (d *databaseLock) ID() string {
	return d.owner
}

// Type 返回锁驱动类型标识，与 CacheDriver 枚举值保持一致。
func (d *databaseLock) Type() string {
	return "db"
}

// Acquire 尝试获取 key 锁并返回是否成功：
// 先尝试插入新锁，失败则接管已过期的锁，并以 lottery 概率触发一次过期锁清理。
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

// createLock 尝试插入一条新锁记录，key 已存在（唯一约束）时失败。
func (d *databaseLock) createLock(db *ent.Client, key string, expiredAt time.Time) bool {
	_, err := db.CacheLock.Create().
		SetKey(key).
		SetOwner(d.owner).
		SetExpiredAt(expiredAt).
		Save(context.TODO())
	return err == nil
}

// updateExpiredLock 若 key 对应的锁已过期则接管（CAS 语义），返回是否接管成功。
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

// cleanupExpiredLocks 删除所有过期超过 60 秒的锁记录，回收僵尸锁。
func (d *databaseLock) cleanupExpiredLocks(db *ent.Client) {
	_, err := db.CacheLock.Delete().Where(cachelock.ExpiredAtLT(d.timer.Now().Add(-60 * time.Second))).Exec(context.TODO())
	if err != nil {
		d.logger.Error(err)
	}
}

// renewalExistKey 在事务中校验并续期 key 锁：锁存在且持有者为当前 owner 时顺延过期时间。
func (d *databaseLock) renewalExistKey(key string, seconds int64) error {
	ctx, cancelFunc := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancelFunc()
	return d.data.WithTx(ctx, func(db *ent.Tx) error {
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
	})
}

// Release 释放 key 锁。仅持有者能释放，成功返回 true。
func (d *databaseLock) Release(key string) bool {
	if d.Owner(key) != d.owner {
		return false
	}
	d.data.DB().CacheLock.Delete().Where(cachelock.Key(key), cachelock.Owner(d.owner)).Exec(context.TODO())
	return true
}

// ForceRelease 无条件释放 key 锁（不校验持有者）。
func (d *databaseLock) ForceRelease(key string) bool {
	d.data.DB().CacheLock.Delete().Where(cachelock.Key(key)).Exec(context.TODO())
	return true
}

// Owner 返回 key 锁当前的持有者，未加锁时返回空串。
func (d *databaseLock) Owner(key string) string {
	cl, err := d.data.DB().CacheLock.Query().Where(cachelock.Key(key)).First(context.TODO())
	if err != nil {
		return ""
	}
	return cl.Owner
}
