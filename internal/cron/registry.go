package cron

import (
	"sync"

	"github.com/duc-cnzj/mars/internal/contracts"
)

var (
	registry []callback
	mu       sync.RWMutex
)

type callback func(manager contracts.CronManager, app contracts.ApplicationInterface)

func Register(cb callback) {
	mu.Lock()
	defer mu.Unlock()
	registry = append(registry, cb)
}

func RegisteredCronJobs() []callback {
	mu.RLock()
	defer mu.RUnlock()
	return registry
}
