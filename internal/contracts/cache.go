package contracts

//go:generate mockgen -destination ../mock/mock_cache.go -package mock github.com/duc-cnzj/mars/internal/contracts CacheInterface

type CacheInterface interface {
	Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error)
}
