package contracts

type Locker interface {
	ID() string
	Type() string
	Acquire(key string, seconds int64) bool
	RenewalAcquire(key string, seconds int64, renewalSeconds int) (releaseFn func(), acquired bool)
	Release(key string) bool
	ForceRelease(key string) bool
	Owner(key string) string
}
