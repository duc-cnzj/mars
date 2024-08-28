package services

//go:generate mockgen -destination ./mock_svc.go -package services github.com/duc-cnzj/mars/api/v5/metrics Metrics_StreamTopPodServer
import (
	"github.com/duc-cnzj/mars/api/v5/auth"
	"github.com/duc-cnzj/mars/api/v5/changelog"
	"github.com/duc-cnzj/mars/api/v5/cluster"
	"github.com/duc-cnzj/mars/api/v5/container"
	"github.com/duc-cnzj/mars/api/v5/endpoint"
	"github.com/duc-cnzj/mars/api/v5/event"
	"github.com/duc-cnzj/mars/api/v5/file"
	"github.com/duc-cnzj/mars/api/v5/git"
	"github.com/duc-cnzj/mars/api/v5/metrics"
	"github.com/duc-cnzj/mars/api/v5/namespace"
	"github.com/duc-cnzj/mars/api/v5/picture"
	"github.com/duc-cnzj/mars/api/v5/project"
	"github.com/duc-cnzj/mars/api/v5/repo"
	"github.com/duc-cnzj/mars/api/v5/token"
	"github.com/duc-cnzj/mars/api/v5/version"
	"github.com/duc-cnzj/mars/v5/internal/application"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

var WireServiceSet = wire.NewSet(
	NewAccessTokenSvc,
	NewAuthSvc,
	NewChangelogSvc,
	NewDefaultAuthProvider,
	NewClusterSvc,
	NewRepoSvc,
	NewContainerSvc,
	NewEndpointSvc,
	NewEventSvc,
	NewFileSvc,
	NewGitSvc,
	NewMetricsSvc,
	NewNamespaceSvc,
	NewPictureSvc,
	NewProjectSvc,
	NewVersionSvc,
	NewGrpcRegistry,
)

func NewGrpcRegistry(
	v version.VersionServer,
	server project.ProjectServer,
	pictureServer picture.PictureServer,
	namespaceServer namespace.NamespaceServer,
	ms metrics.MetricsServer,
	gitServer git.GitServer,
	fileServer file.FileServer,
	eventServer event.EventServer,
	endpointServer endpoint.EndpointServer,
	containerServer container.ContainerServer,
	clusterServer cluster.ClusterServer,
	changelogServer changelog.ChangelogServer,
	authServer auth.AuthServer,
	tokenServer token.AccessTokenServer,
	repoServer repo.RepoServer,
) *application.GrpcRegistry {
	return &application.GrpcRegistry{
		EndpointFuncs: []application.EndpointFunc{
			repo.RegisterRepoHandlerFromEndpoint,
			container.RegisterContainerHandlerFromEndpoint,
			cluster.RegisterClusterHandlerFromEndpoint,
			endpoint.RegisterEndpointHandlerFromEndpoint,
			event.RegisterEventHandlerFromEndpoint,
			file.RegisterFileHandlerFromEndpoint,
			git.RegisterGitHandlerFromEndpoint,
			metrics.RegisterMetricsHandlerFromEndpoint,
			namespace.RegisterNamespaceHandlerFromEndpoint,
			picture.RegisterPictureHandlerFromEndpoint,
			project.RegisterProjectHandlerFromEndpoint,
			version.RegisterVersionHandlerFromEndpoint,
			changelog.RegisterChangelogHandlerFromEndpoint,
			auth.RegisterAuthHandlerFromEndpoint,
			token.RegisterAccessTokenHandlerFromEndpoint,
		},
		RegistryFunc: func(s grpc.ServiceRegistrar) {
			repo.RegisterRepoServer(s, repoServer)
			container.RegisterContainerServer(s, containerServer)
			cluster.RegisterClusterServer(s, clusterServer)
			endpoint.RegisterEndpointServer(s, endpointServer)
			event.RegisterEventServer(s, eventServer)
			file.RegisterFileServer(s, fileServer)
			git.RegisterGitServer(s, gitServer)
			metrics.RegisterMetricsServer(s, ms)
			namespace.RegisterNamespaceServer(s, namespaceServer)
			picture.RegisterPictureServer(s, pictureServer)
			project.RegisterProjectServer(s, server)
			version.RegisterVersionServer(s, v)
			changelog.RegisterChangelogServer(s, changelogServer)
			auth.RegisterAuthServer(s, authServer)
			token.RegisterAccessTokenServer(s, tokenServer)
		},
	}
}
