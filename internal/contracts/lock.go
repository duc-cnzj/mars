package contracts

type Locker interface {
	Acquire(key string, seconds int64) bool
	Release(key string) bool
	ForceRelease(key string) bool
	Owner(key string) string
}
