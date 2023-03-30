package cache

import "github.com/duc-cnzj/mars/v4/internal/contracts"

type NoCache struct{}

// Remember TODO.
func (n *NoCache) Remember(key contracts.CacheKeyInterface, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	return fn()
}

// Clear TODO.
func (n *NoCache) Clear(key contracts.CacheKeyInterface) error {
	return nil
}

// SetWithTTL TODO.
func (n *NoCache) SetWithTTL(key contracts.CacheKeyInterface, value []byte, seconds int) error {
	return nil
}

// Store TODO.
func (n *NoCache) Store() contracts.Store {
	return nil
}
