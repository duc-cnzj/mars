package locker

import (
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
)

// defaultLottery 是默认的过期锁清理概率 [分子, 分母]，
// 即每次 Acquire 有 2/100 的概率触发一次僵尸锁清理。
var defaultLottery = [2]int{2, 100}

// Driver 是锁后端驱动类型，与 CacheDriver 枚举值保持一致。
// 定义为命名类型而非裸 string，是为了让 wire 注入时能与 provideAdminPassword
// 之类的 string provider 区分开，避免出现多个 string provider 的注入歧义。
type Driver string

const (
	// DriverDB 表示数据库锁后端（cache_locks 表），适用于多实例部署。
	DriverDB Driver = "db"
	// DriverMemory 表示进程内内存锁后端，适用于单实例或 sqlite 场景。
	DriverMemory Driver = "memory"
)

// ResolveDriver 根据 DB 驱动与 Cache 驱动解析出最终的锁驱动。
//
// 特别地，sqlite 的 DBDriver 不支持数据库锁：当 DBDriver 为 sqlite 且
// CacheDriver 为 db 时，强制回退到内存锁，并记录告警日志。
// 其他情况下原样返回 CacheDriver 对应的驱动。
func ResolveDriver(dbDriver, cacheDriver string, logger mlog.Logger) Driver {
	if dbDriver == "sqlite" && cacheDriver == string(DriverDB) {
		logger.Warning(`使用 DBDriver 为 "sqlite" 时，CacheDriver,Locker 只能使用 "memory"!`)
		return DriverMemory
	}
	return Driver(cacheDriver)
}

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

// NewLocker 根据已解析的锁驱动选择后端：DriverDB 时使用数据库锁，否则使用内存锁。
//
// 两个后端都是纯内存构造（数据库锁只是持有闭包，不触发 I/O），故不返回 error。
// getDB 以闭包形式注入：锁在 DB 初始化前即可构造，首次操作时实时取到初始化后的
// *ent.Client；DB 未就绪时相关操作返回失败（false/空串/error）而非 panic。
// driver 由 ResolveDriver 解析（处理 sqlite 回退），本函数不再关心 DB 驱动。
func NewLocker(driver Driver, getDB func() *ent.Client, logger mlog.Logger, timer timer.Timer) Locker {
	logger = logger.WithModule("locker/locker")
	switch driver {
	case DriverDB:
		return NewDatabaseLock(timer, defaultLottery, getDB, logger)
	case DriverMemory:
		fallthrough
	default:
		return NewMemoryLock(timer, defaultLottery, NewMemStore(), logger)
	}
}
