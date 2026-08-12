// Package locker 提供进程内内存与数据库两种后端的分布式锁实现，
// 通过 Locker 接口统一暴露加锁、续期与释放能力。
package locker

import (
	"context"
	"errors"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/data/ent/cachelock"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/rand"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
)

// databaseLock 是基于数据库缓存表 cache_locks 的 Locker 实现，适用于多实例场景。
// getDB 以闭包形式注入：每次操作实时获取当前 *ent.Client，因此锁可以在
// DB 初始化（DBBootstrapper → InitDB）之前构造，只要首次调用发生在就绪之后。
type databaseLock struct {
	lottery [2]int
	timer   timer.Timer
	owner   string
	getDB   func() *ent.Client
	logger  mlog.Logger
}

// NewDatabaseLock 创建一个基于数据库的锁。
//
// getDB 是获取当前 *ent.Client 的闭包（例如 data.Data 的 DB 方法值），
// 允许锁实例在 DB 尚未初始化时构造；但首次 Acquire 前必须完成初始化，
// 否则 DB 相关操作返回失败（false/空串/error），而不是 panic。
// lottery 是 [分子, 分母] 组合：每次 Acquire 有 lottery[0]/lottery[1] 的概率触发一次过期锁清理。
func NewDatabaseLock(timer timer.Timer, lottery [2]int, getDB func() *ent.Client, logger mlog.Logger) Locker {
	return &databaseLock{
		lottery: lottery,
		timer:   timer,
		owner:   rand.String(40),
		getDB:   getDB,
		logger:  logger,
	}
}

// db 返回当前 *ent.Client；DB 未初始化（getDB 返回 nil）时记录警告并返回 nil。
func (d *databaseLock) db() *ent.Client {
	db := d.getDB()
	if db == nil {
		d.logger.Warning("[lock]: database not initialized")
	}
	return db
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
	db := d.db()
	if db == nil {
		return false
	}
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
	db := d.db()
	if db == nil {
		return errors.New("db not initialized")
	}
	tx, err := db.Tx(ctx)
	if err != nil {
		return err
	}
	// 已提交后回滚返回 ErrTxDone，忽略即可；未提交则回滚释放事务。
	defer func() { _ = tx.Rollback() }()

	item, err := tx.CacheLock.Query().Where(cachelock.Key(key)).ForUpdate().Only(ctx)
	if err != nil {
		return err
	}
	if item.Owner != d.owner {
		return errors.New("not owner")
	}
	if _, err = item.Update().
		SetOwner(d.owner).
		SetExpiredAt(d.timer.Now().Add(time.Duration(seconds) * time.Second)).
		Save(ctx); err != nil {
		return err
	}
	return tx.Commit()
}

// Release 释放 key 锁。仅持有者能释放，成功返回 true。
// DELETE 执行失败时返回 false，避免在锁实际仍存在的情况下谎报已释放。
func (d *databaseLock) Release(key string) bool {
	db := d.db()
	if db == nil {
		return false
	}
	cl, err := db.CacheLock.Query().Where(cachelock.Key(key)).First(context.TODO())
	if err != nil || cl.Owner != d.owner {
		return false
	}
	if _, err := db.CacheLock.Delete().Where(cachelock.Key(key), cachelock.Owner(d.owner)).Exec(context.TODO()); err != nil {
		d.logger.Error(err)
		return false
	}
	return true
}

// ForceRelease 无条件释放 key 锁（不校验持有者）。
// DELETE 执行失败时返回 false。
func (d *databaseLock) ForceRelease(key string) bool {
	db := d.db()
	if db == nil {
		return false
	}
	if _, err := db.CacheLock.Delete().Where(cachelock.Key(key)).Exec(context.TODO()); err != nil {
		d.logger.Error(err)
		return false
	}
	return true
}

// Owner 返回 key 锁当前的持有者，未加锁时返回空串。
func (d *databaseLock) Owner(key string) string {
	db := d.db()
	if db == nil {
		return ""
	}
	cl, err := db.CacheLock.Query().Where(cachelock.Key(key)).First(context.TODO())
	if err != nil {
		return ""
	}
	return cl.Owner
}
