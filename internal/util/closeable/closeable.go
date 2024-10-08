package closeable

import "sync/atomic"

const closed int64 = 1

type Closeable struct {
	closed int64
}

func (c *Closeable) IsClosed() bool {
	return atomic.LoadInt64(&c.closed) == closed
}

func (c *Closeable) Close() bool {
	return atomic.CompareAndSwapInt64(&c.closed, 0, closed)
}
