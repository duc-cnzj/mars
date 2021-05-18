package bootstrappers

import (
	"github.com/DuC-cnZj/mars/pkg/contracts"
	"github.com/DuC-cnZj/mars/pkg/router"
	"github.com/gin-gonic/gin"
)

type RouterBootstrapper struct{}

func (r *RouterBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	router.Init(app.HttpHandler().(*gin.Engine))

	return nil
}
