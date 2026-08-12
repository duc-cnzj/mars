package app

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	mcron "github.com/duc-cnzj/mars/v6/internal/cron"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/event"
	"github.com/duc-cnzj/mars/v6/internal/eventhandler"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/prometheus/client_golang/prometheus"
)

// hook 是 pre-run / shutdown 生命周期回调的相位键。
type hook string

// 注册在 app 上的生命周期钩子键。
const (
	beforeRunHook hook = "before_run"
	afterDownHook hook = "after_down"
)

var _ App = (*app)(nil)
var _ BootstrapDeps = (*app)(nil)

// app 实现 App 接口，装配全部运行时依赖与 server。
type app struct {
	done context.Context

	doneFunc func()
	servers  []Server
	// started 记录 Run 阶段成功启动的 server（Run 返回 nil 才追加），
	// Shutdown 只回收该集合，保证 Run 中途失败时未启动的 server 不会被错误 Shutdown。
	started       []Server
	bootstrappers []Bootstrapper
	excludeTags   []string

	hooksMu sync.RWMutex
	hooks   map[hook][]Callback

	timer              timer.Timer
	config             *config.Config
	logger             mlog.Logger
	authBiz            biz.AuthBiz
	dispatcher         event.Dispatcher
	cronManager        mcron.Manager
	cache              data.Cache
	data               data.Data
	projectRepo        biz.ProjectRepo
	k8sRepo            biz.K8sRepo
	pluginManager      PluginManager
	reg                *GrpcRegistry
	prometheusRegistry *prometheus.Registry
	httpHandler        HttpHandler
}

// Option 在构造期定制 app。
type Option func(*app)

// WithBootstrappers 设置自定义 bootstrapper。
func WithBootstrappers(bootstrappers ...Bootstrapper) Option {
	return func(app *app) {
		app.bootstrappers = bootstrappers
	}
}

// NewApp 返回 App。
func NewApp(
	config *config.Config,
	data data.Data,
	logger mlog.Logger,
	authBiz biz.AuthBiz,
	dispatcher event.Dispatcher,
	cronManager mcron.Manager,
	cache data.Cache,
	pm PluginManager,
	reg *GrpcRegistry,
	pr *prometheus.Registry,
	httpHandler HttpHandler,
	timer timer.Timer,
	projectRepo biz.ProjectRepo,
	k8sRepo biz.K8sRepo,
	// eventCoordinator 仅用于 wire 可达性：构造时在 dispatcher 上自注册 4 个
	// 跨域事件监听，随后由 dispatcher 持有监听闭包，app 无需保留引用。
	// 类型来自独立事件用例包（internal/eventhandler），组合根只关心装配不关心实现。
	eventCoordinator *eventhandler.EventCoordinator,
	opts ...Option,
) App {
	doneCtx, cancelFunc := context.WithCancel(context.TODO())
	appli := &app{
		timer:              timer,
		httpHandler:        httpHandler,
		done:               doneCtx,
		prometheusRegistry: pr,
		doneFunc:           cancelFunc,
		servers:            []Server{},
		excludeTags:        config.ExcludeServer.List(),
		hooksMu:            sync.RWMutex{},
		hooks:              map[hook][]Callback{},
		config:             config,
		logger:             logger.WithModule("app/app"),
		authBiz:            authBiz,
		dispatcher:         dispatcher,
		cronManager:        cronManager,
		cache:              cache,
		data:               data,
		projectRepo:        projectRepo,
		k8sRepo:            k8sRepo,
		pluginManager:      pm,
		reg:                reg,
	}

	for _, opt := range opts {
		opt(appli)
	}

	if len(appli.excludeTags) > 0 {
		// 排除失效显式化：拼错或文档外的 tag 不匹配任何 bootstrapper，告警而非静默。
		for _, tag := range unknownExcludeTags(appli.excludeTags, appli.bootstrappers) {
			appli.logger.Warningf("--exclude_server tag %q 不匹配任何 bootstrapper，排除为 no-op", tag)
		}
		appli.bootstrappers = excludeBootstrapperByTags(appli.excludeTags, appli.bootstrappers)
	}

	return appli
}

// PrometheusRegistry 实现 App 接口的 PrometheusRegistry。
func (app *app) PrometheusRegistry() *prometheus.Registry {
	return app.prometheusRegistry
}

// PluginManager 实现 App 接口的 PluginManager。
func (app *app) PluginManager() PluginManager {
	return app.pluginManager
}

// Data 实现 App（InfraProvider）与 BootstrapDeps 接口的 Data。
func (app *app) Data() data.Data {
	return app.data
}

// Logger 实现 App 接口的 Logger。
func (app *app) Logger() mlog.Logger {
	return app.logger
}

// Cache 实现 App（InfraProvider）接口的 Cache；picture_cartoon 插件经 Resolve 断言取用。
func (app *app) Cache() data.Cache {
	return app.cache
}

// ProjectRepo 实现 PluginApp 接口的 ProjectRepo。
func (app *app) ProjectRepo() biz.ProjectRepo {
	return app.projectRepo
}

// K8sRepo 返回 k8s 仓库，供 domainmanager syncsecret 插件经 Resolve 断言取用。
// 不再属于 PluginApp 公共接口（单插件独有能力走包内依赖视图），但组合根保留实现。
func (app *app) K8sRepo() biz.K8sRepo {
	return app.k8sRepo
}

// HttpHandler 实现 App 接口的 HttpHandler。
func (app *app) HttpHandler() HttpHandler {
	return app.httpHandler
}

// AuthBiz 实现 ServerDeps 接口的 AuthBiz。gRPC 拦截器经它做 token 校验与用户注入，
// 与 HTTP LoginHTTP 中间件共用 biz.Authenticate 同一鉴权核心。
func (app *app) AuthBiz() biz.AuthBiz {
	return app.authBiz
}

// GrpcRegistry 实现 App 接口的 GrpcRegistry。
func (app *app) GrpcRegistry() *GrpcRegistry {
	return app.reg
}

// CronManager 实现 App 接口的 CronManager。
func (app *app) CronManager() mcron.Manager {
	return app.cronManager
}

// Done 实现 App 接口的 Done。
func (app *app) Done() <-chan struct{} {
	return app.done.Done()
}

// Dispatcher 实现 App 接口的 Dispatcher。
func (app *app) Dispatcher() event.Dispatcher {
	return app.dispatcher
}

// bootTags 是一组 bootstrapper 用于排除过滤的 tag 列表。
type bootTags []string

// has 报告 tag 是否包含在 tag 列表中。
func (bt bootTags) has(tag string) bool {
	for _, t := range bt {
		if t == tag {
			return true
		}
	}
	return false
}

// excludeBootstrapperByTags 返回过滤掉 tags 中任一命中的 bootstrapper 后的剩余列表。
// 排除掉的 bootstrapper 无需回传：调用方只关心保留下来的集合。
func excludeBootstrapperByTags(tags []string, boots []Bootstrapper) []Bootstrapper {
	var newBoots []Bootstrapper
loop:
	for _, boot := range boots {
		bt := bootTags(boot.Tags())
		for _, tag := range tags {
			if bt.has(tag) {
				continue loop
			}
		}

		newBoots = append(newBoots, boot)
	}
	return newBoots
}

// unknownExcludeTags 返回排除项中匹配不到任何 bootstrapper tag 的 tag。
// 排除依赖 tag 精确匹配：拼错或文档外的 tag 会让排除静默失效（历史 P1 的根因类别），
// 把这类 tag 挑出来供 NewApp 告警，让"排除没生效"显式化。
func unknownExcludeTags(tags []string, boots []Bootstrapper) []string {
	matched := make(map[string]bool)
	for _, boot := range boots {
		for _, tag := range boot.Tags() {
			matched[tag] = true
		}
	}
	var unknown []string
	for _, tag := range tags {
		if !matched[tag] {
			unknown = append(unknown, tag)
		}
	}
	return unknown
}

// Bootstrap 实现 App 接口的 Bootstrap。
func (app *app) Bootstrap() error {
	for _, bootstrapper := range app.bootstrappers {
		err := func() error {
			defer func(t time.Time) {
				metrics.BootstrapperStartMetrics.With(prometheus.Labels{"bootstrapper": bootShortName(bootstrapper)}).Set(app.timer.Since(t).Seconds())
			}(app.timer.Now())
			return bootstrapper.Bootstrap(app)
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

// Config 实现 App 接口的 Config。
func (app *app) Config() *config.Config {
	return app.config
}

// AddServer 实现 App 接口的 AddServer。
func (app *app) AddServer(server Server) {
	app.servers = append(app.servers, server)
}

// serverStartupTimeout 是单个 server 的启动时限，兼作 Server.Run 异步契约的兜底闸。
// 契约要求 Run 异步启动、快速返回（见 types.go Server 注释）；若某实现写成同步阻塞，
// app.Run 会在该时限内拿到超时错误并带名字报错，而不是静默卡死后续 server 启动与整个 Shutdown。
// 用 var 而非 const：测试注入短超时，确定性触发该分支。
var serverStartupTimeout = 10 * time.Second

// Run 实现 App 接口的 Run。信号监听由调用方负责（signal.NotifyContext），
// 所有 server 统一跑在调用方传入的 ctx 上，消除 Run 内部自建 context 的双通道割裂。
// 每个 server 经 runOne 启动，带超时兜底。
func (app *app) Run(ctx context.Context) error {
	app.runServerHooks(beforeRunHook)

	for _, server := range app.servers {
		if err := app.runOne(ctx, server); err != nil {
			return err
		}
		// Run 成功返回后该 server 才算已启动，进入 Shutdown 回收集合。
		app.started = append(app.started, server)
	}

	return nil
}

// runOne 启动单个 server 并等待其 Run 返回，用 serverStartupTimeout 兜底异步契约。
// 不监听 ctx.Done：异步实现的 Run 无视 ctx 也立即返回，保持"Run 成功即全部启动"的既有语义。
// goroutine 内 recover 把 panic 转成错误，避免 panic 被误报成"启动超时"。
func (app *app) runOne(ctx context.Context, server Server) error {
	done := make(chan error, 1)
	go func(s Server) {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("server %s panicked on Run: %v", serverName(s), r)
			}
		}()
		done <- s.Run(ctx)
	}(server)

	t := time.NewTimer(serverStartupTimeout)
	defer t.Stop()
	select {
	case err := <-done:
		return err
	case <-t.C:
		return fmt.Errorf("server %s did not start within %s (Run likely blocks, violating the async contract)", serverName(server), serverStartupTimeout)
	}
}

// Shutdown 实现 App 接口的 Shutdown。
func (app *app) Shutdown() {
	app.doneFunc()

	wg := &sync.WaitGroup{}
	// 只回收 Run 阶段成功启动的 server：未启动的实例不在 started 内，跳过其 Shutdown。
	for _, server := range app.started {
		wg.Add(1)
		go func(server Server) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			name := serverName(server)
			defer app.logger.HandlePanic("[Remove]: " + name)

			if err := server.Shutdown(ctx); err != nil {
				app.logger.Warningf("[Remove]: %s %s", name, err)
			}
		}(server)
	}
	wg.Wait()

	app.runServerHooks(afterDownHook)

	app.logger.Info("server graceful shutdown.")
}

// RegisterAfterShutdownFunc 实现 App 接口的 RegisterAfterShutdownFunc。
func (app *app) RegisterAfterShutdownFunc(fn Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[afterDownHook] = append(app.hooks[afterDownHook], fn)
}

// runServerHooks 串行执行指定 phase 的全部钩子。
// 先在锁内快照 hook 列表再释放锁，随后串行执行全部回调：
//   - afterDown 清理类钩子按注册逆序执行（LIFO，镜像 bootstrap 的初始化顺序）：
//     后注册的资源先回收，保证插件先于 DB 关闭被销毁，避免清理竞态；
//   - 其余 phase（beforeRun）按注册顺序执行。
//
// 回调内再次注册 hook（RegisterAfterShutdownFunc 等取写锁）不会因外层持锁而死锁，
// 且本轮只执行快照内的 hook，注册发生在快照之后的钩子留到下一轮；
// 单个回调 panic 由 HandlePanic 隔离，不阻断后续钩子。
func (app *app) runServerHooks(hook hook) {
	app.hooksMu.RLock()
	hooks := append([]Callback(nil), app.hooks[hook]...)
	app.hooksMu.RUnlock()

	if hook == afterDownHook {
		// LIFO：逆序反转，清理顺序与初始化顺序镜像。
		for i, j := 0, len(hooks)-1; i < j; i, j = i+1, j-1 {
			hooks[i], hooks[j] = hooks[j], hooks[i]
		}
	}

	for _, cb := range hooks {
		func() {
			defer app.logger.HandlePanic("[RunServerHooks]: " + string(hook))
			cb()
		}()
	}
}

// BeforeServerRunHooks 实现 App 接口的 BeforeServerRunHooks。
func (app *app) BeforeServerRunHooks(cb Callback) {
	app.hooksMu.Lock()
	defer app.hooksMu.Unlock()
	app.hooks[beforeRunHook] = append(app.hooks[beforeRunHook], cb)
}

// bootShortName 获取 bootstrapper 的短名。
func bootShortName(boot Bootstrapper) string {
	if boot == nil {
		return ""
	}
	s := strings.Split(reflect.TypeOf(boot).String(), ".")
	return s[len(s)-1]
}

// serverName 返回 server 的运行时类型名，用于日志与错误信息。
func serverName(server Server) string {
	return reflect.TypeOf(server).String()
}
