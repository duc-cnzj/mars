package cron

import (
	"testing"

	"github.com/duc-cnzj/mars/v4/internal/contracts"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	registry = nil
	assert.Len(t, RegisteredCronJobs(), 0)
	Register(func(manager contracts.CronManager, app contracts.ApplicationInterface) {})
	assert.Len(t, RegisteredCronJobs(), 1)
}
