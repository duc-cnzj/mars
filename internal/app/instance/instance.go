package instance

import (
	"sync"

	"github.com/duc-cnzj/mars/internal/contracts"
)

var (
	app  contracts.ApplicationInterface
	once sync.Once
)

func SetInstance(instance contracts.ApplicationInterface) {
	once.Do(func() {
		app = instance
	})
}

func App() contracts.ApplicationInterface {
	return app
}
