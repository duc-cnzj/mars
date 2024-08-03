package bootstrappers

import (
	"fmt"
	"time"

	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/server"
	"github.com/duc-cnzj/mars/v4/internal/socket"
)

type ApiGatewayBootstrapper struct{}

func (a *ApiGatewayBootstrapper) Tags() []string {
	return []string{"api", "gateway"}
}

func (a *ApiGatewayBootstrapper) Bootstrap(appli application.App) error {
	appli.AddServer(server.NewApiGateway(fmt.Sprintf("localhost:%s", appli.Config().GrpcPort), appli))
	appli.RegisterAfterShutdownFunc(func(app application.App) {
		t := time.NewTimer(5 * time.Second)
		defer t.Stop()
		ch := make(chan struct{})
		go func() {
			socket.Wait.Wait()
			close(ch)
		}()
		select {
		case <-ch:
			app.Logger().Info("[Websocket]: all socket connection closed")
		case <-t.C:
			app.Logger().Warningf("[Websocket]: 等待超时, 未等待所有 socket 连接退出，当前剩余连接 %v 个。", socket.Wait.Count())
		}
	})

	return nil
}
