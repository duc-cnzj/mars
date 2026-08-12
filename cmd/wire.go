//go:build wireinject
// +build wireinject

package cmd

import (
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/cron"
	"github.com/duc-cnzj/mars/v6/internal/cronjob"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/event"
	"github.com/duc-cnzj/mars/v6/internal/eventhandler"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/services"
	"github.com/duc-cnzj/mars/v6/internal/services/websocket"
	"github.com/duc-cnzj/mars/v6/internal/uploader"
	"github.com/google/wire"
)

// InitializeApp 是 wire 注入入口：按 provider set 装配 App 依赖图，由 wire 生成器消费。
func InitializeApp(*config.Config, mlog.Logger, []app.Bootstrapper) (app.App, error) {
	panic(
		wire.Build(
			wire.Bind(new(biz.PictureProvider), new(app.PluginManager)),
			locker.WireLocker,
			uploader.WireUploader,
			data.WireDB,
			data.WireCache,
			deploy.WireDeploy,
			websocket.WireWebsocket,
			metrics.WireMetrics,
			app.WireApp,
			event.WireEvent,
			data.WireDataSet,
			provideEventDeps,
			provideCronDeps,
			providePodEventPublisher,
			provideGitServer,
			provideMinioClient,
			provideCacheDriver,
			provideDBGetter,
			biz.WireBizSet,
			services.WireServiceSet,
			cron.WireCron,
			cronjob.WireCronJob,
			eventhandler.WireEventHandler,
			newApp,
		),
	)
}
