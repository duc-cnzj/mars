package app

import (
	"context"
	"net/http"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/cron"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/event"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

// EndpointFunc 向 HTTP mux 注册单个 gRPC 服务端点。
type EndpointFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

// RegistryFunc 向 gRPC registrar 注册服务。
type RegistryFunc = func(s grpc.ServiceRegistrar)

// GrpcRegistry 保存网关的端点与服务注册函数。
type GrpcRegistry struct {
	EndpointFuncs []EndpointFunc
	RegistryFunc  RegistryFunc
}

// Callback 是生命周期钩子，无参：钩子在注册时捕获自身依赖，不需要组合根门面，
// 因此不把整个 App 递给回调，避免 service locator 泄漏到业务闭包。
type Callback func()

// Server 定义了可启动/可停机的服务单元（grpc、http、metrics 等）。
// Run/Shutdown 配对约定：
//   - Run 必须异步启动：内部起 goroutine 完成 serve，并立即返回 nil。
//     这是强契约而非"允许"——app.Run 对每个 server 施加 serverStartupTimeout 启动闸门，
//     同步阻塞的 Run（超时内不返回）会判启动失败并返回错误，阻塞后续启动。
//   - 调用方保证：只有 Run 成功返回 nil 的 server 才会被 Shutdown；
//   - 实现方仍应让 Shutdown 对「未 Run 或 Run 已失败」的实例安全（判空/no-op），
//     因为调用方只做"Run 成功后回收"的保证，不保证所有实现都严格遵循调用序列。
type Server interface {
	// Run 启动 server。必须异步启动并快速返回；同步阻塞会被 app.Run 的启动超时闸门判失败。
	Run(context.Context) error
	// Shutdown 优雅停机。
	Shutdown(context.Context) error
}

// BootstrapDeps 是 bootstrapper 在启动期可用的最小依赖面（12 个 bootstrapper 实际调用的能力并集）。
// 嵌入 ServerDeps（构造 grpc/gateway server 需整体传入）与 PluginApp（plugin.Load 需整体传入），
// 外加数据/调度/插件/注册表/生命周期注册。
// 相比 App 门面，此接口砍掉了 bootstrapper 永不触碰的能力：Run、Shutdown、Bootstrap、BeforeServerRunHooks。
type BootstrapDeps interface {
	ServerDeps
	PluginApp

	// Data 返回 app 数据（初始化能力面）。原属 PluginApp，但插件不使用，
	// 唯一消费者是 db/s3/sso/k8s 四个 bootstrapper（InitDB/Migrate/InitOidcProvider/InitS3/InitK8s），
	// 故上收至 bootstrapper 依赖面。
	Data() data.Data

	// Dispatcher 事件分发（event bootstrapper）。
	Dispatcher() event.Dispatcher
	// CronManager 定时任务（cron bootstrapper）。
	CronManager() cron.Manager
	// Done 进程结束信号（k8s 客户端后台任务）。
	Done() <-chan struct{}
	// PrometheusRegistry 指标注册表（metrics/tracing bootstrapper）。
	PrometheusRegistry() *prometheus.Registry
	// PluginManager 插件管理器（plugin bootstrapper）。
	PluginManager() PluginManager
	// AddServer 注册一个 server（全部 bootstrapper）。
	AddServer(Server)
	// RegisterAfterShutdownFunc 注册停机钩子（db/plugin/tracing bootstrapper）。
	RegisterAfterShutdownFunc(Callback)
}

// Bootstrapper 启动。
type Bootstrapper interface {
	// Bootstrap 在 app 启动时执行。
	Bootstrap(BootstrapDeps) error
	// Tags 返回启动标签。
	Tags() []string
}

// App 应用组合根门面：聚合角色接口，供 bootstrapper 与顶层入口使用。
// 不再直接暴露 DB()/ent，需要数据库访问的组件应通过 Data() 端口获取。
//
// 除角色接口外，App 还显式声明 AuthBiz/ProjectRepo 两个访问器：二者分别是
// ServerDeps/PluginApp 的成员，组合根实现（*app）天然持有。把她们显式写进
// App 接口，使 mockgen 生成的 MockApp 一并具备这几个方法，测试才能把
// MockApp 直接当作 ServerDeps/PluginApp/BootstrapDeps 的替身使用——
// 否则每次重生成 mock 都会丢失手补方法、打断这三个包的单测编译。
type App interface {
	ConfigProvider
	RegistryHost
	InfraProvider
	Lifecycle

	// PluginManager 返回插件管理器。
	PluginManager() PluginManager
	// AuthBiz 返回鉴权业务逻辑（ServerDeps 成员：gRPC/HTTP 中间件共用）。
	AuthBiz() biz.AuthBiz
	// ProjectRepo 返回项目仓库（PluginApp 成员：wssender 插件按项目路由命名空间）。
	ProjectRepo() biz.ProjectRepo
}

// ConfigProvider 配置与日志。
type ConfigProvider interface {
	// Config 返回 app 配置。
	Config() *config.Config

	// Logger 返回日志器。
	Logger() mlog.Logger
}

// RegistryHost 注册表宿主：grpc、http handler、prometheus。
type RegistryHost interface {
	// GrpcRegistry 返回注册表。
	GrpcRegistry() *GrpcRegistry

	// HttpHandler 返回 http handler。
	HttpHandler() HttpHandler

	// PrometheusRegistry 返回 prometheus 注册表。
	PrometheusRegistry() *prometheus.Registry
}

// InfraProvider 基础设施端口。
type InfraProvider interface {
	// Data 返回 app 数据。
	Data() data.Data

	// Cache 返回缓存。
	Cache() data.Cache

	// Dispatcher 返回事件分发器。
	Dispatcher() event.Dispatcher

	// CronManager 返回定时任务管理器。
	CronManager() cron.Manager
}

// Lifecycle 应用生命周期。
type Lifecycle interface {
	// Bootstrap 启动全部。
	Bootstrap() error

	// AddServer 添加启动 server。
	AddServer(Server)

	// Run servers, ctx 由调用方提供（信号监听等）。
	Run(context.Context) error

	// Shutdown 停机全部 server。
	Shutdown()

	// Done 返回 done 通道。
	Done() <-chan struct{}

	// BeforeServerRunHooks 注册钩子。
	BeforeServerRunHooks(Callback)

	// RegisterAfterShutdownFunc 注册钩子。
	RegisterAfterShutdownFunc(Callback)
}

// PluginApp 插件初始化的统一宽入口，按「复用度分流」只保留多插件共享的能力：
//   - Logger：全部插件共用；
//   - ProjectRepo：三个 wssender 后端共用。
//
// 单插件独有能力（如 Cache、K8sRepo）不进本接口——对应插件在自己包内声明窄 deps
// 视图接口，经 Resolve 断言取用（见 plugins.go 的 Resolve）。这样新增单插件依赖
// 只改该插件自己的包，不逼全仓测试 stub 补实现。
type PluginApp interface {
	// Logger 返回日志器。
	Logger() mlog.Logger
	// ProjectRepo 返回项目仓库（wssender 插件按项目路由命名空间用）。
	ProjectRepo() biz.ProjectRepo
}

// ServerDeps server 构造所需的最小依赖面。
type ServerDeps interface {
	// Config 返回 app 配置。
	Config() *config.Config
	// Logger 返回日志器。
	Logger() mlog.Logger
	// AuthBiz 返回鉴权业务逻辑。gRPC 拦截器经它（biz.Authenticate）做
	// token 校验、用户注入与生效角色解析，与 HTTP LoginHTTP 中间件共用同一鉴权核心。
	AuthBiz() biz.AuthBiz
	// GrpcRegistry 返回注册表。
	GrpcRegistry() *GrpcRegistry
	// HttpHandler 返回 http handler。
	HttpHandler() HttpHandler
}

// WsHttpServer websocket 与集群健康检查的 HTTP server 面。
type WsHttpServer interface {
	// TickClusterHealth 周期性检查集群健康，直至 done 关闭。
	TickClusterHealth(done <-chan struct{})
	// Info 向响应写入服务信息。
	Info(writer http.ResponseWriter, request *http.Request)
	// Serve 处理 websocket http 请求。
	Serve(w http.ResponseWriter, r *http.Request)
	// Shutdown 优雅停机。
	Shutdown(ctx context.Context) error
}

// HttpHandler HTTP handler 面：ws、swagger UI 与文件路由。
type HttpHandler interface {
	WsHttpServer

	// RegisterWsRoute 在 mux 上注册 websocket 路由。
	RegisterWsRoute(mux *mux.Router)
	// RegisterSwaggerUIRoute 在 mux 上注册 swagger UI 路由。
	RegisterSwaggerUIRoute(mux *mux.Router)
	// RegisterFileRoute 在 mux 上注册文件服务路由。
	RegisterFileRoute(mux *runtime.ServeMux)
}
