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
	namespaceRepo := repo.NewNamespaceRepo(logger, dataData)
	projectRepo := repo.NewProjectRepo(logger, dataData)
	archiver := repo.NewDefaultArchiver()
	executorManager := repo.NewExecutorManager(dataData)
	k8sRepo := repo.NewK8sRepo(logger, dataData, uploaderUploader, archiver, executorManager)
	helmerRepo := repo.NewDefaultHelmer(k8sRepo, dataData, configConfig, logger)
	gitProjectRepo := repo.NewGitProjectRepo(logger, dataData)
	changelogRepo := repo.NewChangelogRepo(logger, dataData)
	eventRepo := repo.NewEventRepo(gitProjectRepo, changelogRepo, logger, dataData, dispatcher)
	toolRepo := repo.NewToolRepo()
	jobManager := socket.NewJobManager(logger, namespaceRepo, projectRepo, helmerRepo, lockerLocker, k8sRepo, eventRepo, toolRepo, pluginManger)
	wsRepo := repo.NewWsRepo(pluginManger)
	gitRepo := repo.NewGitRepo(logger, pluginManger, dataData)
	projectServer := services.NewProjectSvc(jobManager, projectRepo, wsRepo, gitRepo, k8sRepo, pluginManger, eventRepo, logger, helmerRepo, namespaceRepo)
	pictureRepo := repo.NewPictureRepo(logger, pluginManger)
	pictureServer := services.NewPictureSvc(pictureRepo)
	namespaceServer := services.NewNamespaceSvc(helmerRepo, namespaceRepo, k8sRepo, logger, eventRepo)
	metricsServer := services.NewMetricsSvc(k8sRepo, logger, projectRepo, namespaceRepo)
	gitConfigServer := services.NewGitConfigSvc(eventRepo, cacheCache, gitRepo, gitProjectRepo, logger)
	gitServer := services.NewGitSvc(eventRepo, logger, gitRepo, cacheCache, gitProjectRepo)
	cronRepo := repo.NewCronRepo(logger, eventRepo, dataData, uploaderUploader, helmerRepo, gitRepo, manager)
	fileRepo := repo.NewFileRepo(cronRepo, logger, dataData, uploaderUploader, timerTimer)
	fileServer := services.NewFileSvc(eventRepo, fileRepo, logger)
	eventServer := services.NewEventSvc(eventRepo)
	endpointRepo := repo.NewEndpointRepo(logger, dataData)
	endpointServer := services.NewEndpointSvc(logger, endpointRepo)
	containerServer := services.NewContainerSvc(eventRepo, k8sRepo, fileRepo, logger)
	clusterServer := services.NewClusterSvc(k8sRepo)
	changelogServer := services.NewChangelogSvc(changelogRepo)
	authServer := services.NewAuthSvc(eventRepo, logger, authAuth, dataData)
	accessTokenRepo := repo.NewAccessTokenRepo(timerTimer, logger, dataData)
	accessTokenServer := services.NewAccessTokenSvc(eventRepo, timerTimer, accessTokenRepo)
	repoRepo := repo.NewRepo(logger, dataData)
	repoServer := services.NewRepoSvc(logger, eventRepo, gitRepo, repoRepo)
	grpcRegistry := services.NewGrpcRegistry(versionServer, projectServer, pictureServer, namespaceServer, metricsServer, gitConfigServer, gitServer, fileServer, eventServer, endpointServer, containerServer, clusterServer, changelogServer, authServer, accessTokenServer, repoServer)
	wsServer := socket.NewWebsocketManager(logger, jobManager, dataData, pluginManger, authAuth, uploaderUploader, lockerLocker, k8sRepo, eventRepo, executorManager, fileRepo)
	app := newApp(configConfig, dataData, manager, arg, logger, uploaderUploader, authAuth, dispatcher, cacheCache, lockerLocker, group, pluginManger, grpcRegistry, wsServer)
	return app, nil
}
