// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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
	"github.com/duc-cnzj/mars/v5/internal/util/counter"
	"github.com/duc-cnzj/mars/v5/internal/util/timer"
)

// Injectors from wire.go:

func InitializeApp(configConfig *config.Config, logger mlog.Logger, arg []application.Bootstrapper) (application.App, error) {
	dataData := data.NewData(configConfig, logger)
	runner := cron.NewRobfigCronV3Runner(logger)
	timerTimer := timer.NewRealTimer()
	lockerLocker, err := locker.NewLocker(configConfig, dataData, logger, timerTimer)
	if err != nil {
		return nil, err
	}
	manager := cron.NewManager(runner, lockerLocker, logger)
	uploaderUploader, err := uploader.NewUploader(configConfig, logger, dataData)
	if err != nil {
		return nil, err
	}
	authAuth, err := auth.NewAuthn(dataData)
	if err != nil {
		return nil, err
	}
	dispatcher := event.NewDispatcher(logger)
	group := NewSingleflight()
	cacheCache := cache.NewCacheImpl(configConfig, dataData, logger, group)
	pluginManger, err := application.NewPluginManager(configConfig, logger)
	if err != nil {
		return nil, err
	}
	versionServer := services.NewVersionSvc()
	gitRepo := repo.NewGitRepo(logger, cacheCache, pluginManger, dataData)
	repoRepo := repo.NewRepo(logger, dataData, gitRepo)
	fileRepo := repo.NewFileRepo(logger, dataData, cacheCache, uploaderUploader, timerTimer)
	archiver := repo.NewDefaultArchiver()
	executorManager := repo.NewExecutorManager(dataData)
	k8sRepo := repo.NewK8sRepo(logger, timerTimer, dataData, fileRepo, uploaderUploader, archiver, executorManager)
	helmerRepo := repo.NewDefaultHelmer(k8sRepo, dataData, configConfig, logger)
	releaseInstaller := socket.NewReleaseInstaller(logger, helmerRepo, dataData, timerTimer)
	namespaceRepo := repo.NewNamespaceRepo(logger, dataData)
	projectRepo := repo.NewProjectRepo(logger, dataData)
	changelogRepo := repo.NewChangelogRepo(logger, dataData)
	eventRepo := repo.NewEventRepo(projectRepo, k8sRepo, pluginManger, changelogRepo, logger, dataData, dispatcher)
	jobManager := socket.NewJobManager(dataData, timerTimer, logger, releaseInstaller, repoRepo, namespaceRepo, projectRepo, helmerRepo, uploaderUploader, lockerLocker, k8sRepo, eventRepo, pluginManger)
	projectServer := services.NewProjectSvc(repoRepo, pluginManger, jobManager, projectRepo, gitRepo, k8sRepo, eventRepo, logger, helmerRepo, namespaceRepo)
	pictureRepo := repo.NewPictureRepo(logger, pluginManger)
	pictureServer := services.NewPictureSvc(pictureRepo)
	namespaceServer := services.NewNamespaceSvc(helmerRepo, namespaceRepo, k8sRepo, logger, eventRepo)
	metricsServer := services.NewMetricsSvc(timerTimer, k8sRepo, logger, projectRepo, namespaceRepo)
	gitServer := services.NewGitSvc(repoRepo, eventRepo, logger, gitRepo, cacheCache)
	fileServer := services.NewFileSvc(eventRepo, fileRepo, logger)
	eventServer := services.NewEventSvc(logger, eventRepo)
	endpointRepo := repo.NewEndpointRepo(logger, dataData, projectRepo)
	endpointServer := services.NewEndpointSvc(logger, endpointRepo)
	containerServer := services.NewContainerSvc(eventRepo, k8sRepo, fileRepo, logger)
	clusterServer := services.NewClusterSvc(k8sRepo, logger)
	changelogServer := services.NewChangelogSvc(changelogRepo)
	authRepo := repo.NewAuthRepo(authAuth, logger, dataData)
	authServer := services.NewAuthSvc(eventRepo, logger, authRepo)
	accessTokenRepo := repo.NewAccessTokenRepo(timerTimer, logger, dataData)
	accessTokenServer := services.NewAccessTokenSvc(logger, eventRepo, timerTimer, accessTokenRepo)
	repoServer := services.NewRepoSvc(logger, eventRepo, gitRepo, repoRepo)
	grpcRegistry := services.NewGrpcRegistry(versionServer, projectServer, pictureServer, namespaceServer, metricsServer, gitServer, fileServer, eventServer, endpointServer, containerServer, clusterServer, changelogServer, authServer, accessTokenServer, repoServer)
	counterCounter := counter.NewCounter()
	wsHttpServer := socket.NewWebsocketManager(logger, counterCounter, projectRepo, repoRepo, namespaceRepo, jobManager, dataData, pluginManger, authAuth, uploaderUploader, lockerLocker, k8sRepo, eventRepo, executorManager, fileRepo)
	registry := metrics.NewRegistry()
	cronRepo := repo.NewCronRepo(logger, fileRepo, cacheCache, repoRepo, namespaceRepo, k8sRepo, pluginManger, eventRepo, dataData, uploaderUploader, helmerRepo, gitRepo, manager)
	app := newApp(configConfig, dataData, manager, arg, logger, uploaderUploader, authAuth, dispatcher, cacheCache, lockerLocker, group, pluginManger, grpcRegistry, wsHttpServer, registry, cronRepo)
	return app, nil
}
