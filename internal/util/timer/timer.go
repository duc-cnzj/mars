package timer

import "time"

type Timer interface {
	Now() time.Time
}

type realTimer struct{}

// NewRealTimer return Timer
func NewRealTimer() Timer {
	return &realTimer{}
}

// Now equal time.Now
func (r realTimer) Now() time.Time {
	return time.Now()
}
