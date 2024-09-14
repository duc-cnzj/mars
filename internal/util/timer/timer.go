package timer

import "time"

type Timer interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

type realTimer struct{}

// NewReal return Timer
func NewReal() Timer {
	return &realTimer{}
}

// Now equal time.Now
func (r realTimer) Now() time.Time {
	return time.Now()
}

// Since equal time.Since
func (r realTimer) Since(t time.Time) time.Duration {
	return time.Since(t)
}
