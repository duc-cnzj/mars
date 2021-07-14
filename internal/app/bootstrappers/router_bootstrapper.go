package bootstrappers

import (
	"github.com/duc-cnzj/mars/internal/contracts"
	"github.com/duc-cnzj/mars/internal/router"
	"github.com/gin-gonic/gin"
)

type RouterBootstrapper struct{}

func (r *RouterBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	router.Init(app.HttpHandler().(*gin.Engine))

	return nil
}
