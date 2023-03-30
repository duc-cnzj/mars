package cron

import (
	"sync"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

var (
	registry []callback
	mu       sync.RWMutex
)

type callback func(manager contracts.CronManager, app contracts.ApplicationInterface)

// Register cron.
func Register(cb callback) {
	mu.Lock()
	defer mu.Unlock()
	registry = append(registry, cb)
}

// RegisteredCronJobs list cron.
func RegisteredCronJobs() []callback {
	mu.RLock()
	defer mu.RUnlock()
	return registry
}
