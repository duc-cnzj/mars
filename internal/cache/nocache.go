package cache

type NoCache struct{}

// Remember TODO.
func (n *NoCache) Remember(key CacheKey, seconds int, fn func() ([]byte, error), force bool) ([]byte, error) {
	return fn()
}

// Clear TODO.
func (n *NoCache) Clear(key CacheKey) error {
	return nil
}

// SetWithTTL TODO.
func (n *NoCache) SetWithTTL(key CacheKey, value []byte, seconds int) error {
	return nil
}

// Store TODO.
func (n *NoCache) Store() Store {
	return nil
}
