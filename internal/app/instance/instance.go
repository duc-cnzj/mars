package instance

import (
	"github.com/duc-cnzj/mars/internal/contracts"
)

var app contracts.ApplicationInterface

// SetInstance 不要再启动之后调用这个，not safe
func SetInstance(instance contracts.ApplicationInterface) {
	app = instance
}

func App() contracts.ApplicationInterface {
	return app
}
