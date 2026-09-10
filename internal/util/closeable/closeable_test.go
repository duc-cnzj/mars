package closeable

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloseable_IsClosed_Initially(t *testing.T) {
	c := &Closeable{}
	assert.False(t, c.IsClosed())
}

func TestCloseable_IsClosed_AfterClose(t *testing.T) {
	c := &Closeable{}
	c.Close()
	assert.True(t, c.IsClosed())
}

func TestCloseable_Close_Initially(t *testing.T) {
	c := &Closeable{}
	assert.True(t, c.Close())
}

func TestCloseable_Close_AfterClose(t *testing.T) {
	c := &Closeable{}
	c.Close()
	assert.False(t, c.Close())
}

func TestCloseable_ConcurrentClose(t *testing.T) {
	c := &Closeable{}
	var wg sync.WaitGroup
	var successCount int
	var failureCount int
	var mu sync.Mutex

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if c.Close() {
				mu.Lock()
				successCount++
				mu.Unlock()
			} else {
				mu.Lock()
				failureCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, 1, successCount)
	assert.Equal(t, 999, failureCount)
}
