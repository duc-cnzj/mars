package contracts

type CacheInterface interface {
	Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error)
}
