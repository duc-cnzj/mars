package contracts

//go:generate mockgen -destination ../mock/mock_cache_locker.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts Locker

type Locker interface {
	ID() string
	Type() string
	Acquire(key string, seconds int64) bool
	RenewalAcquire(key string, seconds int64, renewalSeconds int64) (releaseFn func(), acquired bool)
	Release(key string) bool
	ForceRelease(key string) bool
	Owner(key string) string
}
