package cache

type NoCache struct{}

func (n *NoCache) Remember(key string, seconds int, fn func() ([]byte, error)) ([]byte, error) {
	return fn()
}
