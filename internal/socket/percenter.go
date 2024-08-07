package socket

import (
	"sync"
	"time"
)

type Percentable interface {
	Current() int64
	Add()
	To(percent int64)
}

var _ Percentable = (*processPercent)(nil)

type processPercent struct {
	msger DeployMsger

	s           Sleeper
	percentLock sync.RWMutex
	percent     int64
}

func NewProcessPercent(sender DeployMsger, s Sleeper) Percentable {
	return &processPercent{
		s:     s,
		msger: sender,
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
		pp.msger.SendProcessPercent(pp.percent)
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
		pp.msger.SendProcessPercent(pp.percent)
	}
	if pp.percent != percent {
		pp.msger.SendProcessPercent(pp.percent)
		pp.percent = percent
	}
}
