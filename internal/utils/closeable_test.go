package utils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Closeable_IsClosed(t *testing.T) {
	c := Closeable{}
	assert.False(t, c.IsClosed())
	assert.True(t, c.Close())
	assert.True(t, c.IsClosed())
	assert.False(t, c.Close())
}

func Test_Closeable_race_test(t *testing.T) {
	c := &Closeable{}
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Close()
		}()
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.IsClosed()
		}()
	}
	wg.Wait()
}

//BenchmarkCloseable_IsClosed-8           1000000000               0.5395 ns/op
//BenchmarkCloseable_Close-8              1000000000               0.5496 ns/op

//BenchmarkSyncCloseable_IsClosed-8       83941765                14.17 ns/op
//BenchmarkSyncCloseable_Close-8          63248329                19.09 ns/op

func BenchmarkCloseable_IsClosed(b *testing.B) {
	c := &Closeable{}
	for i := 0; i < b.N; i++ {
		c.IsClosed()
	}
}

func BenchmarkSyncCloseable_IsClosed(b *testing.B) {
	c := &SyncCloseable{}
	for i := 0; i < b.N; i++ {
		c.IsClosed()
	}
}

func BenchmarkCloseable_Close(b *testing.B) {
	c := &Closeable{}
	for i := 0; i < b.N; i++ {
		c.Close()
	}
}

func BenchmarkSyncCloseable_Close(b *testing.B) {
	c := &SyncCloseable{}
	for i := 0; i < b.N; i++ {
		c.Close()
	}
}

type SyncCloseable struct {
	sync.RWMutex
	closed bool
}

func (c *SyncCloseable) IsClosed() bool {
	c.RLock()
	defer c.RUnlock()
	return c.closed
}

func (c *SyncCloseable) Close() bool {
	c.Lock()
	defer c.Unlock()
	if c.closed {
		return false
	}
	c.closed = true
	return true
}
