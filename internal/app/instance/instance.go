package instance

import (
	"github.com/duc-cnzj/mars/v4/internal/contracts"
)

var app contracts.ApplicationInterface

// SetInstance 不要再启动之后调用这个，not safe
func SetInstance(instance contracts.ApplicationInterface) {
	app = instance
}

// App return contracts.ApplicationInterface, not safe.
func App() contracts.ApplicationInterface {
	return app
}
