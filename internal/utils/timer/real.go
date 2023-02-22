package timer

import "time"

type Timer interface {
	Now() time.Time
}

type realTimer struct{}

func NewRealTimer() *realTimer {
	return &realTimer{}
}

func (r realTimer) Now() time.Time {
	return time.Now()
}
