package locker

import (
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
)

// defaultLottery 是默认的过期锁清理概率 [分子, 分母]，
// 即每次 Acquire 有 2/100 的概率触发一次僵尸锁清理。
var defaultLottery = [2]int{2, 100}

// Locker 是分布式锁的统一抽象，支持加锁、续期、释放与持有者查询。
type Locker interface {
	// ID 返回当前锁实例的持有者标识。
	ID() string
	// Type 返回锁驱动类型标识，与 CacheDriver 枚举值保持一致。
	Type() string
	// Acquire 尝试获取 key 锁并返回是否成功。
	Acquire(key string, seconds int64) bool
	// RenewalAcquire 获取 key 锁并启动后台续期协程，返回 release 函数与是否获取成功。
	RenewalAcquire(key string, seconds int64, renewalSeconds int64) (releaseFn func(), acquired bool)
	// Release 释放 key 锁。仅持有者能释放，成功返回 true。
	Release(key string) bool
	// ForceRelease 无条件释放 key 锁（不校验持有者）。
	ForceRelease(key string) bool
	// Owner 返回 key 锁当前的持有者，未加锁时返回空串。
	Owner(key string) string
}

// NewLocker 根据配置选择锁后端：CacheDriver 为 db 时使用数据库锁，否则使用内存锁。
//
// 特别地，sqlite 的 DBDriver 不支持数据库锁，会强制回退到内存锁。
func NewLocker(cfg *config.Config, data data.Data, logger mlog.Logger, timer timer.Timer) (Locker, error) {
	logger = logger.WithModule("locker/locker")
	if cfg.DBDriver == "sqlite" && cfg.CacheDriver == "db" {
		cfg.CacheDriver = "memory"
		logger.Warning(`使用 DBDriver 为 "sqlite" 时，CacheDriver,Locker 只能使用 "memory"!`)
	}
	driver := cfg.CacheDriver

	switch driver {
	case "db":
		return NewDatabaseLock(timer, defaultLottery, data, logger), nil
	case "memory":
		fallthrough
	default:
		return NewMemoryLock(timer, defaultLottery, NewMemStore(), logger), nil
	}
}
