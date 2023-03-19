package contracts

//go:generate mockgen -destination ../mock/mock_cache.go -package mock github.com/duc-cnzj/mars/v4/internal/contracts CacheInterface

type CacheKeyInterface interface {
	String() string
	Slug() string
}

type CacheInterface interface {
	SetWithTTL(key CacheKeyInterface, value []byte, seconds int) error
	Remember(key CacheKeyInterface, seconds int, fn func() ([]byte, error)) ([]byte, error)
	Clear(key CacheKeyInterface) error
}
