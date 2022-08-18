package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/lock"
)

type DistributedLocksBootstrapper struct{}

func (d *DistributedLocksBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	app.SetDistributedLocks(lock.NewDatabaseLock([2]int{2, 100}, app.DBManager().DB()))
	// app.SetDistributedLocks(lock.NewMemoryLock([2]int{2, 100}, nil))

	return nil
}
