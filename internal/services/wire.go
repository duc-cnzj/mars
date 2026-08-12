package services

//go:generate go tool mockgen -destination ./mock_metrics_stream_test.go -package services github.com/duc-cnzj/mars/api/v6/proto/metrics Metrics_StreamTopPodServer
import (
	"github.com/duc-cnzj/mars/api/v6/proto/auth"
	"github.com/duc-cnzj/mars/api/v6/proto/changelog"
	"github.com/duc-cnzj/mars/api/v6/proto/cluster"
	"github.com/duc-cnzj/mars/api/v6/proto/container"
	"github.com/duc-cnzj/mars/api/v6/proto/endpoint"
	"github.com/duc-cnzj/mars/api/v6/proto/event"
	"github.com/duc-cnzj/mars/api/v6/proto/file"
	"github.com/duc-cnzj/mars/api/v6/proto/git"
	"github.com/duc-cnzj/mars/api/v6/proto/metrics"
	"github.com/duc-cnzj/mars/api/v6/proto/namespace"
	"github.com/duc-cnzj/mars/api/v6/proto/picture"
	"github.com/duc-cnzj/mars/api/v6/proto/project"
	"github.com/duc-cnzj/mars/api/v6/proto/repo"
	"github.com/duc-cnzj/mars/api/v6/proto/token"
	"github.com/duc-cnzj/mars/api/v6/proto/version"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

// WireServiceSet 是传输层服务装配集：供 wire 注入，组合 14 个 gRPC service
// （XxxSvcDeps wire.Struct 全字段注入 + 对应 NewXxxSvc 构造器）与 HTTP 处理器、
// gRPC 注册表，构成 services 包对外可注入的全部传输层能力。
var WireServiceSet = wire.NewSet(
	wire.Struct(new(ProjectSvcDeps), "*"),
	wire.Struct(new(HttpHandlerDeps), "*"),
	wire.Struct(new(MetricsSvcDeps), "*"),
	wire.Struct(new(ChangelogSvcDeps), "*"),
	wire.Struct(new(EndpointSvcDeps), "*"),
	wire.Struct(new(NamespaceSvcDeps), "*"),
	wire.Struct(new(ContainerSvcDeps), "*"),
	wire.Struct(new(AccessTokenSvcDeps), "*"),
	wire.Struct(new(AuthSvcDeps), "*"),
	wire.Struct(new(FileSvcDeps), "*"),
	wire.Struct(new(GitSvcDeps), "*"),
	wire.Struct(new(RepoSvcDeps), "*"),
	wire.Struct(new(ClusterSvcDeps), "*"),
	wire.Struct(new(EventSvcDeps), "*"),
	wire.Struct(new(PictureSvcDeps), "*"),
	wire.Struct(new(NewGrpcRegistryDeps), "*"),
	NewAccessTokenSvc,
	NewAuthSvc,
	NewChangelogSvc,
	NewClusterSvc,
	NewRepoSvc,
	NewContainerSvc,
	NewEndpointSvc,
	NewEventSvc,
	NewFileSvc,
	NewHttpHandler,
	NewGitSvc,
	NewMetricsSvc,
	NewNamespaceSvc,
	NewPictureSvc,
	NewProjectSvc,
	NewVersionSvc,
	NewGrpcRegistry,
)

// NewGrpcRegistryDeps 收口 NewGrpcRegistry 的 15 个 gRPC 服务实现，由 wire 按字段
// 注入。与全包 wire.Struct("*") 模式对齐：裸位置参数存在"错一个位置编译期全过、
// 运行期串服务"的隐患，Deps struct 根除这一风险。
type NewGrpcRegistryDeps struct {
	Version     version.VersionServer
	Project     project.ProjectServer
	Picture     picture.PictureServer
	Namespace   namespace.NamespaceServer
	Metrics     metrics.MetricsServer
	Git         git.GitServer
	File        file.FileServer
	Event       event.EventServer
	Endpoint    endpoint.EndpointServer
	Container   container.ContainerServer
	Cluster     cluster.ClusterServer
	Changelog   changelog.ChangelogServer
	Auth        auth.AuthServer
	AccessToken token.AccessTokenServer
	Repo        repo.RepoServer
}

// NewGrpcRegistry 组装 gRPC 网关路由表与服务器注册函数：把各服务实现注册到
// grpc.Server（RegistryFunc）并挂载 HTTP-gateway 端点（EndpointFuncs）。
// 由 wire 自动调用，消费方为 server 装配层。
func NewGrpcRegistry(deps NewGrpcRegistryDeps) *app.GrpcRegistry {
	return &app.GrpcRegistry{
		EndpointFuncs: []app.EndpointFunc{
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
			repo.RegisterRepoServer(s, deps.Repo)
			container.RegisterContainerServer(s, deps.Container)
			cluster.RegisterClusterServer(s, deps.Cluster)
			endpoint.RegisterEndpointServer(s, deps.Endpoint)
			event.RegisterEventServer(s, deps.Event)
			file.RegisterFileServer(s, deps.File)
			git.RegisterGitServer(s, deps.Git)
			metrics.RegisterMetricsServer(s, deps.Metrics)
			namespace.RegisterNamespaceServer(s, deps.Namespace)
			picture.RegisterPictureServer(s, deps.Picture)
			project.RegisterProjectServer(s, deps.Project)
			version.RegisterVersionServer(s, deps.Version)
			changelog.RegisterChangelogServer(s, deps.Changelog)
			auth.RegisterAuthServer(s, deps.Auth)
			token.RegisterAccessTokenServer(s, deps.AccessToken)
		},
	}
}
