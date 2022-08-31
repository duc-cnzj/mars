package cache

import "github.com/duc-cnzj/mars/internal/contracts"

type NoCache struct{}

func (n *NoCache) Remember(key contracts.CacheKeyInterface, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	return fn()
}

func (n *NoCache) Clear(key contracts.CacheKeyInterface) error {
	return nil
}
