package commands

import (
	"testing"

	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	assert.Len(t, RegisteredCronJobs(), 0)
	Register(func(manager contracts.CronManager, app contracts.ApplicationInterface) {})
	assert.Len(t, RegisteredCronJobs(), 1)
}
