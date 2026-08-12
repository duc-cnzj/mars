package timer

import "time"

// Timer 抽象当前时间获取能力，便于测试注入假时钟。
type Timer interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

// realTimer 是 Timer 的真实实现，直接返回系统时钟。
type realTimer struct{}

// NewReal 构造使用系统时钟的真实 Timer。
func NewReal() Timer {
	return &realTimer{}
}

// Now 返回当前系统时间，等价于 time.Now。
func (r realTimer) Now() time.Time {
	return time.Now()
}

// Since 返回自 t 以来的时长，等价于 time.Since。
func (r realTimer) Since(t time.Time) time.Duration {
	return time.Since(t)
}
