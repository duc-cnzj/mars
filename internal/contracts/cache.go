package contracts

//go:generate mockgen -destination ../mock/mock_cache.go -package mock github.com/duc-cnzj/mars/internal/contracts CacheInterface

type CacheKeyInterface interface {
	String() string
	Slug() string
}

type CacheInterface interface {
	Remember(key CacheKeyInterface, seconds int, fn func() ([]byte, error)) ([]byte, error)

	Clear(key CacheKeyInterface) error
}
