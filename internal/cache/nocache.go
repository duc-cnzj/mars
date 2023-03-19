package cache

import "github.com/duc-cnzj/mars/v4/internal/contracts"

type NoCache struct{}

func (n *NoCache) Remember(key contracts.CacheKeyInterface, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	return fn()
}

func (n *NoCache) Clear(key contracts.CacheKeyInterface) error {
	return nil
}

func (n *NoCache) SetWithTTL(key contracts.CacheKeyInterface, value []byte, seconds int) error {
	return nil
}
