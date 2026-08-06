package closeable

import "sync/atomic"

// closed 表示 Closeable 已关闭的原子标记值。
const closed int64 = 1

// Closeable 提供并发安全的单向关闭标记。
type Closeable struct {
	closed int64
}

// IsClosed 返回该对象是否已关闭。
func (c *Closeable) IsClosed() bool {
	return atomic.LoadInt64(&c.closed) == closed
}

// Close 尝试关闭并返回是否为首次关闭（之前未关闭）。
func (c *Closeable) Close() bool {
	return atomic.CompareAndSwapInt64(&c.closed, 0, closed)
}
