package utils

import "sync/atomic"

const closed int64 = 1

type Closeable struct {
	closed int64
}

func (c *Closeable) IsClosed() bool {
	return atomic.LoadInt64(&c.closed) == closed
}

func (c *Closeable) Close() bool {
	if c.IsClosed() {
		return false
	}
	atomic.StoreInt64(&c.closed, closed)
	return true
}
