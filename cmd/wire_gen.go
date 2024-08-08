// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package cmd

import (
	"github.com/duc-cnzj/mars/v4/internal/application"
	"github.com/duc-cnzj/mars/v4/internal/auth"
	"github.com/duc-cnzj/mars/v4/internal/cache"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/cron"
	"github.com/duc-cnzj/mars/v4/internal/data"
	"github.com/duc-cnzj/mars/v4/internal/event"
	"github.com/duc-cnzj/mars/v4/internal/locker"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"github.com/duc-cnzj/mars/v4/internal/repo"
	"github.com/duc-cnzj/mars/v4/internal/services"
	"github.com/duc-cnzj/mars/v4/internal/socket"
	"github.com/duc-cnzj/mars/v4/internal/uploader"
	"github.com/duc-cnzj/mars/v4/internal/util/counter"
	"github.com/duc-cnzj/mars/v4/internal/util/timer"
)

// Injectors from wire.go:

func InitializeApp(configConfig *config.Config, logger mlog.Logger, arg []application.Bootstrapper) (application.App, error) {
	dataData, err := data.NewData(configConfig, logger)
	if err != nil {
		return nil, err
	}
	runner := cron.NewRobfigCronV3Runner(logger)
	timerTimer := timer.NewRealTimer()
	lockerLocker, err := locker.NewLocker(configConfig, dataData, logger, timerTimer)
	if err != nil {
		return nil, err
	}
	manager := cron.NewManager(runner, lockerLocker, logger)
	group := NewSingleflight()
	cacheCache := cache.NewCacheImpl(configConfig, dataData, logger, group)
	uploaderUploader, err := uploader.NewUploader(configConfig, logger, dataData, cacheCache)
	if err != nil {
		return nil, err
	}
	authAuth, err := auth.NewAuthn(configConfig, dataData)
	if err != nil {
		return nil, err
	}
	dispatcher := event.NewDispatcher(logger)
	pluginManger, err := application.NewPluginManager(configConfig, logger)
	if err != nil {
		return nil, err
	}
	versionServer := services.NewVersionSvc()
	gitRepo := repo.NewGitRepo(logger, pluginManger, dataData)
	repoImp := repo.NewRepo(logger, dataData, gitRepo)
	archiver := repo.NewDefaultArchiver()
	executorManager := repo.NewExecutorManager(dataData)
	k8sRepo := repo.NewK8sRepo(logger, dataData, uploaderUploader, archiver, executorManager)
	helmerRepo := repo.NewDefaultHelmer(k8sRepo, dataData, configConfig, logger)
	releaseInstaller := socket.NewReleaseInstaller(logger, helmerRepo, dataData, timerTimer)
	namespaceRepo := repo.NewNamespaceRepo(logger, dataData)
	projectRepo := repo.NewProjectRepo(logger, dataData)
	changelogRepo := repo.NewChangelogRepo(logger, dataData)
	eventRepo := repo.NewEventRepo(projectRepo, pluginManger, changelogRepo, logger, dataData, dispatcher)
	toolRepo := repo.NewToolRepo()
	jobManager := socket.NewJobManager(dataData, timerTimer, logger, releaseInstaller, repoImp, namespaceRepo, projectRepo, helmerRepo, uploaderUploader, lockerLocker, k8sRepo, eventRepo, toolRepo, pluginManger)
	wsRepo := repo.NewWsRepo(pluginManger)
	projectServer := services.NewProjectSvc(repoImp, jobManager, projectRepo, wsRepo, gitRepo, k8sRepo, pluginManger, eventRepo, logger, helmerRepo, namespaceRepo)
	pictureRepo := repo.NewPictureRepo(logger, pluginManger)
	pictureServer := services.NewPictureSvc(pictureRepo)
	namespaceServer := services.NewNamespaceSvc(helmerRepo, namespaceRepo, k8sRepo, logger, eventRepo)
	metricsServer := services.NewMetricsSvc(k8sRepo, logger, projectRepo, namespaceRepo)
	gitServer := services.NewGitSvc(repoImp, eventRepo, logger, gitRepo, cacheCache)
	cronRepo := repo.NewCronRepo(logger, eventRepo, dataData, uploaderUploader, helmerRepo, gitRepo, manager)
	fileRepo := repo.NewFileRepo(cronRepo, logger, dataData, uploaderUploader, timerTimer)
	fileServer := services.NewFileSvc(eventRepo, fileRepo, logger)
	eventServer := services.NewEventSvc(eventRepo)
	endpointRepo := repo.NewEndpointRepo(logger, dataData, projectRepo)
	endpointServer := services.NewEndpointSvc(logger, endpointRepo)
	containerServer := services.NewContainerSvc(eventRepo, k8sRepo, fileRepo, logger)
	clusterServer := services.NewClusterSvc(k8sRepo)
	changelogServer := services.NewChangelogSvc(changelogRepo)
	authServer := services.NewAuthSvc(eventRepo, logger, authAuth, dataData)
	accessTokenRepo := repo.NewAccessTokenRepo(timerTimer, logger, dataData)
	accessTokenServer := services.NewAccessTokenSvc(eventRepo, timerTimer, accessTokenRepo)
	repoServer := services.NewRepoSvc(logger, eventRepo, gitRepo, repoImp)
	grpcRegistry := services.NewGrpcRegistry(versionServer, projectServer, pictureServer, namespaceServer, metricsServer, gitServer, fileServer, eventServer, endpointServer, containerServer, clusterServer, changelogServer, authServer, accessTokenServer, repoServer)
	counterCounter := counter.NewCounter()
	wsServer := socket.NewWebsocketManager(logger, counterCounter, repoImp, namespaceRepo, jobManager, dataData, pluginManger, authAuth, uploaderUploader, lockerLocker, k8sRepo, eventRepo, executorManager, fileRepo)
	app := newApp(configConfig, dataData, manager, arg, logger, uploaderUploader, authAuth, dispatcher, cacheCache, lockerLocker, group, pluginManger, grpcRegistry, wsServer)
	return app, nil
}
