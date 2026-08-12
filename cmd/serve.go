package cmd

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/cron"
	"github.com/duc-cnzj/mars/v6/internal/cronjob"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/event"
	"github.com/duc-cnzj/mars/v6/internal/eventhandler"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/minio/minio-go/v7"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
)

// documentedExcludeTags 是 --exclude_server 文档承诺的可排除服务 tag 白名单。
// flag help 文案与回归测试（TestExcludeServerDocumentedTagsCovered）都从这里取，
// 杜绝"文档写 A、代码用 B"的历史 P1（pprof tag 曾与文档不一致导致排除静默失效）。
// 注意：grpc/apigateway 额外挂着 gateway/grpc/trace 等 tag，技术上也可排除，
// 但白名单只承诺最常用的四个服务名。
var documentedExcludeTags = []string{"api", "metrics", "cron", "profile"}

// serverBootstrappers 是启动 bootstrapper 的执行顺序，顺序即依赖拓扑，不可随意调整：
//   - Event 最先：事件分发器作为 server 先 Run，且是后续 bootstrapper 派发/消费事件的依赖；
//   - 基础设施先行（DB/K8s）：数据库连接与 k8s 客户端需在依赖它们的服务注册（网关/gRPC）之前就绪；
//   - 服务注册与附属能力随后（ApiGateway/Pprof/Grpc/Metrics/Tracing/Cron/SSO/S3）；
//   - Plugin 殿后：插件 Load 依赖 DB/K8s/事件等全部基础设施就绪，其回收钩子经 afterDown
//     LIFO 逆序最先执行，保证插件先于 DB 关闭被销毁。
//
// 新增 bootstrapper 按"被依赖者在前"插入，并同步更新本注释。
var serverBootstrappers = []app.Bootstrapper{
	&bootstrappers.EventBootstrapper{},
	&bootstrappers.DBBootstrapper{},
	&bootstrappers.K8sBootstrapper{},
	&bootstrappers.ApiGatewayBootstrapper{},
	&bootstrappers.PprofBootstrapper{},
	&bootstrappers.GrpcBootstrapper{},
	&bootstrappers.MetricsBootstrapper{},
	&bootstrappers.TracingBootstrapper{},
	&bootstrappers.CronBootstrapper{},
	&bootstrappers.SSOBootstrapper{},
	&bootstrappers.S3Bootstrapper{},
	&bootstrappers.PluginBootstrapper{},
}

var apiGatewayCmd = &cobra.Command{
	Use:   "serve",
	Short: "start mars server use grpc.",
	PreRun: func(cmd *cobra.Command, args []string) {
		fmt.Println(logo)
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(log.Lshortfile)
		cfg := config.Init(viper.GetString("config"))
		logger := mlog.NewForConfig(cfg)
		// 对齐 kratos/log.With：在 main 声明随请求变化的日志字段，Valuer 按调用时
		// ctx 求值。TraceID/SpanID 由 mlog 提供，用户字段来源（biz.GetUser）在此注入，
		// 避免 mlog 反向依赖 auth。
		logger = mlog.With(logger,
			"trace.id", mlog.TraceID(),
			"span.id", mlog.SpanID(),
			"user_name", func(ctx context.Context) any {
				if u, err := biz.GetUser(ctx); err == nil {
					return u.Name
				}
				return ""
			},
			"email", func(ctx context.Context) any {
				if u, err := biz.GetUser(ctx); err == nil {
					return u.Email
				}
				return ""
			},
		)
		app, err := InitializeApp(cfg, logger, serverBootstrappers)
		if err != nil {
			logger.Fatal(err)
		}
		if err := app.Bootstrap(); err != nil {
			logger.Fatal(err)
		}
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
		defer stop()

		if err := app.Run(ctx); err != nil {
			logger.Fatal(err)
		}
		<-ctx.Done()
		stop()
		app.Shutdown()
		// 进程优雅退出前冲刷日志缓冲：zap 后端 Sync() 落盘、logrus 为 no-op。
		// Fatal 路径走 os.Exit 跳过此处，但各后端写 stderr/stdout 均同步落盘，无丢失。
		_ = logger.Flush()
	},
}

func newApp(
	cfg *config.Config,
	data data.Data,
	cron cron.Manager,
	bootstrappers []app.Bootstrapper,
	logger mlog.Logger,
	authBiz biz.AuthBiz,
	dispatcher event.Dispatcher,
	cache data.Cache,
	pm app.PluginManager,
	reg *app.GrpcRegistry,
	pr *prometheus.Registry,
	// cron 任务由 cronjob.Registry 声明（条件任务就地过滤），cmd 只做机械注册。
	cronTasks *cronjob.Tasks,
	httpHandler app.HttpHandler,
	timer timer.Timer,
	projectRepo biz.ProjectRepo,
	k8sRepo biz.K8sRepo,
	// 事件用例由 eventhandler 包提供：EventCoordinator 仅作 wire 可达性，
	// PodEventListener 包装成 Server 随 app 生命周期常驻启停。
	eventCoordinator *eventhandler.EventCoordinator,
	podListener *eventhandler.PodEventListener,
) app.App {
	RegisterCronJobs(cronjob.Registry(cronTasks, cfg), cron)
	app := app.NewApp(
		cfg,
		data,
		logger,
		authBiz,
		dispatcher,
		cron,
		cache,
		pm,
		reg,
		pr,
		httpHandler,
		timer,
		projectRepo,
		k8sRepo,
		eventCoordinator,
		app.WithBootstrappers(bootstrappers...),
	)
	// git 插件经 provideGitServer 惰性闭包在首次调用时实时解析，无需启动前刷新。
	app.AddServer(newPodListenerServer(podListener))
	return app
}

// podListenerServer 把常驻 Pod 事件监听包装成 app.Server：
// Run 异步启动消费 goroutine（满足 Server 异步契约），Shutdown 取消其 ctx，
// 触发 PodEventListener 注销 informer 订阅。
type podListenerServer struct {
	listener *eventhandler.PodEventListener
	cancel   context.CancelFunc
}

// newPodListenerServer 构造 Pod 监听 server。
func newPodListenerServer(l *eventhandler.PodEventListener) app.Server {
	return &podListenerServer{listener: l}
}

// Run 实现 app.Server 的 Run：异步启动常驻监听，立即返回 nil。
func (s *podListenerServer) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	go func() {
		_ = s.listener.Run(runCtx)
	}()
	return nil
}

// Shutdown 实现 app.Server 的 Shutdown：取消监听 ctx。
func (s *podListenerServer) Shutdown(_ context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

// RegisterCronJobs 把注册表 Registry 产出的 []CronTask 机械注册进 cron.Manager。
// 任务名与调度声明已收敛到任务层（CronTask.Schedule），cmd 作为组合根只做解释，
// 不关心任何调度细节，也无导入环。常驻的 Pod 事件监听已归位 eventhandler
// （随 app 启停），不再经 cron 触发。
func RegisterCronJobs(tasks []cronjob.CronTask, mgr cron.Manager) {
	for _, task := range tasks {
		task.Schedule(mgr.NewCommand(task.Name, task.Run))
	}
}

// provideEventDeps 把 PluginManager 的插件能力包成事件用例的惰性闭包结构体。
// 插件实例在 PluginBootstrapper（bootstrap 阶段最后一步）才完成 Initialize，
// wire 期 pm.Domain()/pm.Ws() 恒为 nil，因此闭包在触发时才实时解析 pm——
// 不需要服务器启动前二次刷新（原 PluginDeps 快照 + refreshPluginDeps 机制的替代）。
func provideEventDeps(pm app.PluginManager) *eventhandler.PluginDeps {
	return &eventhandler.PluginDeps{
		GetCerts: newGetCertsFunc(pm),
		ToAll:    newToAllFunc(pm),
	}
}

// provideCronDeps 把 PluginManager 的插件能力包成定时任务用例的惰性闭包结构体。
func provideCronDeps(pm app.PluginManager) *cronjob.PluginDeps {
	return &cronjob.PluginDeps{
		GetCerts: newGetCertsFunc(pm),
	}
}

// providePodEventPublisher 把 PluginManager 的 ws 插件惰性适配为 Pod 事件发布器。
func providePodEventPublisher(pm app.PluginManager) eventhandler.PodEventPublisher {
	return newPodEventPublisher(pm)
}

// newGetCertsFunc 返回惰性取证书闭包：触发时解析已加载的域名插件 GetCerts。
func newGetCertsFunc(pm app.PluginManager) func() (string, string, string) {
	return func() (string, string, string) {
		return pm.Domain().GetCerts()
	}
}

// newToAllFunc 返回惰性 ws 广播闭包：触发时经已加载的 ws 插件建一次性 PubSub
// 广播消息后关闭。负载由 eventhandler 侧以 WebsocketMessage 断言。
func newToAllFunc(pm app.PluginManager) func(proto.Message) error {
	return func(msg proto.Message) error {
		sub := pm.Ws().New("", "")
		defer sub.Close()
		return sub.ToAll(msg.(app.WebsocketMessage))
	}
}

// podPubAdapter 惰性适配 PluginManager.Ws() 为 eventhandler.PodEventPublisher：
// 首次 Publish 时经 once 解析已加载的 ws 插件建一次性 PubSub 并断言发布器，
// 后续复用同一发布器。构造发生在 wire 期（pm.Ws() 为 nil），惰性规避空指针。
type podPubAdapter struct {
	pm   app.PluginManager
	once sync.Once
	pub  app.ProjectPodEventPublisher
}

// newPodEventPublisher 构造惰性 Pod 事件发布器。
func newPodEventPublisher(pm app.PluginManager) eventhandler.PodEventPublisher {
	return &podPubAdapter{pm: pm}
}

// Publish 发布某个项目的 pod 变更事件。
func (a *podPubAdapter) Publish(nsID int64, pod *corev1.Pod) error {
	a.once.Do(func() {
		a.pub = a.pm.Ws().New("", "").(app.ProjectPodEventPublisher)
	})
	return a.pub.Publish(nsID, pod)
}

// provideGitServer 返回惰性取 git 插件闭包：插件在 PluginBootstrapper 阶段才
// 完成 Initialize，wire 期 pm.Git() 恒为 nil，gitRepo 首次调用方法时才实时解析。
// 替代原 GitServerHolder 快照 + refreshGitServer 的二次刷新机制，与 event/cron/
// minio/db 的惰性取数模式对齐。
func provideGitServer(pm app.PluginManager) func() data.GitServer {
	return func() data.GitServer {
		return pm.Git()
	}
}

// provideMinioClient 返回惰性取数函数而非客户端本身：minio 客户端由
// S3Bootstrapper 在 bootstrap 阶段才初始化，此刻拿到的还是 nil，必须让
// uploader 在首次 S3 操作时实时解析。取数入口是 MinioGetter 窄端口
// （MinioCli 不在 Data/dataStore 门面上，仅 MinioGetter 提供），仅限 cmd 装配期使用。
func provideMinioClient(d data.MinioGetter) func() *minio.Client {
	return d.MinioCli
}

// provideCacheDriver 解析配置中的锁驱动，处理 sqlite 不支持数据库锁的回退规则。
func provideCacheDriver(cfg *config.Config, logger mlog.Logger) locker.Driver {
	return locker.ResolveDriver(cfg.DBDriver, cfg.CacheDriver, logger)
}

// provideDBGetter 返回惰性获取 *ent.Client 的闭包：锁在 DB 初始化前即可构造，
// 首次操作时实时取到初始化后的客户端。取数入口是 DBGetter 窄端口
// （DB 在 dataStore 而非 Data 门面上），同样仅限 cmd 装配期使用。
func provideDBGetter(d data.DBGetter) func() *ent.Client {
	return d.DB
}

func init() {
	apiGatewayCmd.Flags().BoolP("debug", "", true, "debug mode.")
	apiGatewayCmd.Flags().StringP("metrics_port", "", "9091", "metrics port")
	apiGatewayCmd.Flags().StringP("app_port", "", "6000", "app port.")
	apiGatewayCmd.Flags().StringP("kubeconfig", "", "", "kubeconfig.")
	apiGatewayCmd.Flags().StringP("grpc_port", "", "", "grpc port.")
	apiGatewayCmd.Flags().StringP("exclude_server", "", "", "do not start these services("+strings.Join(documentedExcludeTags, "/")+"), join with ','.")

	viper.BindPFlag("debug", apiGatewayCmd.Flags().Lookup("debug"))
	viper.BindPFlag("metrics_port", apiGatewayCmd.Flags().Lookup("metrics_port"))
	viper.BindPFlag("app_port", apiGatewayCmd.Flags().Lookup("app_port"))
	viper.BindPFlag("kubeconfig", apiGatewayCmd.Flags().Lookup("kubeconfig"))
	viper.BindPFlag("grpc_port", apiGatewayCmd.Flags().Lookup("grpc_port"))
	viper.BindPFlag("exclude_server", apiGatewayCmd.Flags().Lookup("exclude_server"))
}
