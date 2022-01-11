package socket

import (
	"fmt"
	"sync"
	"time"
)

type Percentable interface {
	Current() int64
	Add()
	To(percent int64)
}

type processPercent struct {
	ProcessPercentMsger

	percentLock sync.RWMutex
	percent     int64
}

func newProcessPercent(sender ProcessPercentMsger) Percentable {
	return &processPercent{
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
		pp.SendProcessPercent(fmt.Sprintf("%d", pp.percent))
	}
}

func (pp *processPercent) To(percent int64) {
	pp.percentLock.Lock()
	defer pp.percentLock.Unlock()

	sleepTime := 100 * time.Millisecond
	for pp.percent < percent {
		time.Sleep(sleepTime)
		pp.percent++
		if sleepTime > 50*time.Millisecond {
			sleepTime = sleepTime / 2
		}
		pp.SendProcessPercent(fmt.Sprintf("%d", pp.percent))
	}
}
