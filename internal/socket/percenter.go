package socket

import (
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

type processPercent struct {
	contracts.ProcessPercentMsger

	s           Sleeper
	percentLock sync.RWMutex
	percent     int64
}

type Sleeper interface {
	Sleep(time.Duration)
}

type realSleeper struct{}

func (r *realSleeper) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

func newProcessPercent(sender contracts.ProcessPercentMsger, s Sleeper) contracts.Percentable {
	return &processPercent{
		s:                   s,
		percent:             0,
		ProcessPercentMsger: sender,
	}
}

func (pp *processPercent) Current() int64 {
	pp.percentLock.RLock()
	defer pp.percentLock.RUnlock()

	return pp.percent
}

func (pp *processPercent) Add() {
	pp.percentLock.Lock()
	defer pp.percentLock.Unlock()

	if pp.percent < 100 {
		pp.percent++
		pp.SendProcessPercent(pp.percent)
	}
}

func (pp *processPercent) To(percent int64) {
	pp.percentLock.Lock()
	defer pp.percentLock.Unlock()

	sleepTime := 100 * time.Millisecond
	var step int64 = 2
	for pp.percent+step <= percent {
		pp.s.Sleep(sleepTime)
		pp.percent += step
		if sleepTime > 50*time.Millisecond {
			sleepTime = sleepTime / 2
		}
		pp.SendProcessPercent(pp.percent)
	}
	if pp.percent != percent {
		pp.SendProcessPercent(pp.percent)
		pp.percent = percent
	}
}
