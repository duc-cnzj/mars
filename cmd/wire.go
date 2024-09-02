//go:build wireinject
// +build wireinject

package cmd

import (
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/duc-cnzj/mars/v5/internal/auth"
	"github.com/duc-cnzj/mars/v5/internal/cache"
	"github.com/duc-cnzj/mars/v5/internal/config"
	"github.com/duc-cnzj/mars/v5/internal/cron"
	"github.com/duc-cnzj/mars/v5/internal/data"
	"github.com/duc-cnzj/mars/v5/internal/event"
	"github.com/duc-cnzj/mars/v5/internal/locker"
	"github.com/duc-cnzj/mars/v5/internal/metrics"
	"github.com/duc-cnzj/mars/v5/internal/mlog"
	"github.com/duc-cnzj/mars/v5/internal/repo"
	"github.com/duc-cnzj/mars/v5/internal/services"
	"github.com/duc-cnzj/mars/v5/internal/socket"
	"github.com/duc-cnzj/mars/v5/internal/uploader"
	"github.com/google/wire"
)

func InitializeApp(*config.Config, mlog.Logger, []application.Bootstrapper) (application.App, error) {
	panic(
		wire.Build(
			NewSingleflight,
			locker.WireLocker,
			uploader.WireUploader,
			data.WireData,
			cache.WireCache,
			socket.WireSocket,
			metrics.WireMetrics,
			application.WireApp,
			event.WireEvent,
			repo.WireRepoSet,
			services.WireServiceSet,
			auth.WireAuth,
			cron.WireCron,
			newApp,
		),
	)
}
