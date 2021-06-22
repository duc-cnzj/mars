package bootstrappers

import (
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/pkg/adapter"
	"github.com/duc-cnzj/mars/pkg/contracts"
	"github.com/duc-cnzj/mars/pkg/mlog"

	"github.com/gin-gonic/gin"
)

var (
	DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		mlog.Debugf("%-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	ginWriter = &adapter.GinWriter{}
)

type WebBootstrapper struct{}

func (a *WebBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	var middlewares []gin.HandlerFunc

	gin.DebugPrintRouteFunc = DebugPrintRouteFunc
	gin.DefaultWriter = ginWriter

	if app.IsDebug() {
		mlog.Debug("debug mode.")
		gin.SetMode(gin.DebugMode)
		middlewares = append(middlewares, Logger())
	} else {
		mlog.Info("release mode.")
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery()).Use(middlewares...)
	app.SetHttpHandler(engine)
	mlog.Debug("WebBootstrapper booted!")

	return nil
}

func Logger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			if param.Latency > time.Minute {
				param.Latency = param.Latency - param.Latency%time.Second
			}
			return fmt.Sprintf("| %3d | %13v | %15s | %-7s %#v\n%s",
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				param.Path,
				param.ErrorMessage,
			)
		},
	})
}
