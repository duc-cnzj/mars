package bootstrappers

import (
	"github.com/duc-cnzj/mars/pkg/contracts"
	"github.com/duc-cnzj/mars/pkg/router"
	"github.com/gin-gonic/gin"
)

type RouterBootstrapper struct{}

func (r *RouterBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	router.Init(app.HttpHandler().(*gin.Engine))

	return nil
}
