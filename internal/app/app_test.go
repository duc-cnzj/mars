package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	config2 "github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/cron"
	data2 "github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/event"
	"github.com/duc-cnzj/mars/v6/internal/eventhandler"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type testBoot struct {
	tags []string
	err  error
}

func (t *testBoot) Bootstrap(BootstrapDeps) error {
	return t.err
}

func (t *testBoot) Tags() []string {
	return t.tags
}

func TestNewAppWithValidConfig(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	config := &config2.Config{
		Debug:         true,
		ExcludeServer: config2.ExcludeServerTags("cron"),
	}
	data := data2.NewMockData(m)
	logger := mlog.NewForConfig(nil)
	authBiz := biz.NewMockAuthBiz(m)
	dispatcher := event.NewMockDispatcher(m)
	cronManager := cron.NewMockManager(m)
	cache := data2.NewMockCache(m)
	pm := NewMockPluginManager(m)
	reg := &GrpcRegistry{}
	httpHandler := NewMockHttpHandler(m)
	pr := &prometheus.Registry{}

	b1 := &testBoot{
		tags: []string{"cron"},
	}
	appli := NewApp(
		config,
		data,
		logger,
		authBiz,
		dispatcher,
		cronManager,
		cache,
		pm,
		reg,
		pr,
		httpHandler,
		timer.NewReal(),
		data2.NewMockProjectRepo(m),
		data2.NewMockK8sRepo(m),
		&eventhandler.EventCoordinator{},
		WithBootstrappers(b1, &testBoot{}),
	)

	assert.NotNil(t, appli)
	assert.NotNil(t, appli.Data())
	assert.NotNil(t, appli.Logger())
	assert.NotNil(t, appli.(*app).AuthBiz())
	assert.NotNil(t, appli.Dispatcher())
	assert.NotNil(t, appli.CronManager())
	assert.NotNil(t, appli.Cache())
	assert.NotNil(t, appli.PluginManager())
	assert.NotNil(t, appli.GrpcRegistry())
	assert.NotNil(t, appli.HttpHandler())
	assert.NotNil(t, appli.PrometheusRegistry())
	assert.NotNil(t, appli.(*app).ProjectRepo())
	assert.NotNil(t, appli.(*app).K8sRepo())

	assert.NotNil(t, appli.(*app).timer)
	assert.Len(t, appli.(*app).bootstrappers, 1)
	assert.NotContains(t, appli.(*app).bootstrappers, b1)
}

func TestNewAppWithoutExcludeTags(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	config := &config2.Config{}
	b1 := &testBoot{
		tags: []string{"cron"},
	}
	b2 := &testBoot{}
	appli := NewApp(
		config,
		data2.NewMockData(m),
		mlog.NewForConfig(nil),
		biz.NewMockAuthBiz(m),
		event.NewMockDispatcher(m),
		cron.NewMockManager(m),
		data2.NewMockCache(m),
		NewMockPluginManager(m),
		&GrpcRegistry{},
		&prometheus.Registry{},
		NewMockHttpHandler(m),
		timer.NewReal(),
		data2.NewMockProjectRepo(m),
		data2.NewMockK8sRepo(m),
		&eventhandler.EventCoordinator{},
		WithBootstrappers(b1, b2),
	)

	assert.NotNil(t, appli)
	assert.Len(t, appli.(*app).bootstrappers, 2)
	assert.Contains(t, appli.(*app).bootstrappers, b1)
	assert.Contains(t, appli.(*app).bootstrappers, b2)
}

func TestWithBootstrappers(t *testing.T) {
	a := &app{}
	WithBootstrappers(&testBoot{})(a)
	assert.Len(t, a.bootstrappers, 1)
}

type testServer struct {
	Server
}

func Test_app_AddServer(t *testing.T) {
	a := &app{}
	a.AddServer(&testServer{})
	assert.Len(t, a.servers, 1)
}

func Test_app_AuthBiz(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	authBiz := biz.NewMockAuthBiz(m)
	a := &app{
		authBiz: authBiz,
	}
	assert.NotNil(t, a.AuthBiz())
}

func Test_app_BeforeServerRunHooks(t *testing.T) {
	a := &app{hooks: map[hook][]Callback{}}
	a.BeforeServerRunHooks(func() {})
	assert.Len(t, a.hooks, 1)
}

func Test_app_Bootstrap(t *testing.T) {
	a := &app{bootstrappers: []Bootstrapper{&testBoot{}}, timer: timer.NewReal()}
	assert.Nil(t, a.Bootstrap())
}

func Test_app_Cache(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cache := data2.NewMockCache(m)
	a := &app{
		cache: cache,
	}
	assert.NotNil(t, a.Cache())
}

func Test_app_Config(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	config := &config2.Config{}
	a := &app{
		config: config,
	}
	assert.NotNil(t, a.Config())
}

func Test_app_CronManager(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	cronManager := cron.NewMockManager(m)
	a := &app{
		cronManager: cronManager,
	}
	assert.NotNil(t, a.CronManager())
}

func Test_app_Data(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	data := data2.NewMockData(m)
	a := &app{data: data}
	assert.NotNil(t, a.Data())
}

func Test_app_Dispatcher(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	dispatcher := event.NewMockDispatcher(m)
	a := &app{
		dispatcher: dispatcher,
	}
	assert.NotNil(t, a.Dispatcher())
}

func Test_app_Done(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	a := &app{done: ctx}
	assert.NotNil(t, a.Done())
}

func Test_app_GrpcRegistry(t *testing.T) {
	a := &app{reg: &GrpcRegistry{}}
	assert.NotNil(t, a.GrpcRegistry())
}

func Test_app_Logger(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	logger := mlog.NewForConfig(nil)
	a := &app{
		logger: logger,
	}
	assert.NotNil(t, a.Logger())
}

func Test_app_PluginManager(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	pm := NewMockPluginManager(m)
	a := &app{pluginManager: pm}
	assert.NotNil(t, a.PluginManager())
}

func Test_app_PrometheusRegistry(t *testing.T) {
	a := &app{prometheusRegistry: &prometheus.Registry{}}
	assert.NotNil(t, a.PrometheusRegistry())
}

func Test_app_RegisterAfterShutdownFunc(t *testing.T) {
	a := &app{hooks: map[hook][]Callback{}}
	a.RegisterAfterShutdownFunc(func() {})
	assert.Len(t, a.hooks[afterDownHook], 1)
}

func Test_app_RunServerHooks(t *testing.T) {
	called := false
	a := &app{
		logger: mlog.NewForConfig(nil),
		hooks: map[hook][]Callback{
			afterDownHook: {func() {
				called = true
			}},
		},
	}
	a.runServerHooks(afterDownHook)
	assert.True(t, called)
}

// 回调内再次注册 hook 不死锁（旧实现持 RLock 执行回调到注册处写锁自锁）；
// 新注册发生在快照之后，本轮不执行，仅已有 hook 执行。
func Test_app_RunServerHooks_ReentrantRegisterNoDeadlock(t *testing.T) {
	var order []string
	a := &app{
		logger: mlog.NewForConfig(nil),
		hooks:  map[hook][]Callback{},
	}
	a.RegisterAfterShutdownFunc(func() {
		order = append(order, "first")
		a.RegisterAfterShutdownFunc(func() {
			order = append(order, "nested")
		})
	})
	a.runServerHooks(afterDownHook)
	assert.Equal(t, []string{"first"}, order)
	assert.Len(t, a.hooks[afterDownHook], 2)
}

// afterDown 清理类钩子串行逆序执行（LIFO）：全部执行到，断言元素集合而非顺序。
func Test_app_RunServerHooks_RunsAllHooks(t *testing.T) {
	var ran []string
	a := &app{
		logger: mlog.NewForConfig(nil),
		hooks: map[hook][]Callback{
			afterDownHook: {
				func() { ran = append(ran, "a") },
				func() { ran = append(ran, "b") },
				func() { ran = append(ran, "c") },
			},
		},
	}
	a.runServerHooks(afterDownHook)
	assert.ElementsMatch(t, []string{"a", "b", "c"}, ran)
}

// 单个 hook panic 被 HandlePanic 隔离，不阻断其余 hook 串行执行。
func Test_app_RunServerHooks_PanicIsolated(t *testing.T) {
	var ran []string
	a := &app{
		logger: mlog.NewForConfig(nil),
		hooks: map[hook][]Callback{
			afterDownHook: {
				func() { ran = append(ran, "a") },
				func() { panic("boom") },
				func() { ran = append(ran, "c") },
			},
		},
	}
	a.runServerHooks(afterDownHook)
	assert.ElementsMatch(t, []string{"a", "c"}, ran)
}

// afterDown 按注册逆序执行（LIFO）：后注册的钩子先回收，镜像 bootstrap 初始化顺序。
func Test_app_RunServerHooks_AfterDownLIFO(t *testing.T) {
	var ran []string
	a := &app{
		logger: mlog.NewForConfig(nil),
		hooks: map[hook][]Callback{
			afterDownHook: {
				func() { ran = append(ran, "first-registered") },
				func() { ran = append(ran, "last-registered") },
			},
		},
	}
	a.runServerHooks(afterDownHook)
	assert.Equal(t, []string{"last-registered", "first-registered"}, ran)
}

// beforeRun 按注册顺序串行执行。
func Test_app_RunServerHooks_BeforeRunRegistrationOrder(t *testing.T) {
	var ran []string
	a := &app{
		logger: mlog.NewForConfig(nil),
		hooks: map[hook][]Callback{
			beforeRunHook: {
				func() { ran = append(ran, "a") },
				func() { ran = append(ran, "b") },
			},
		},
	}
	a.runServerHooks(beforeRunHook)
	assert.Equal(t, []string{"a", "b"}, ran)
}

type mockServer struct {
	Server
	called bool
	err    error
}

func (m *mockServer) Shutdown(context.Context) error {
	m.called = true
	return m.err
}

func Test_app_Shutdown(t *testing.T) {
	called := false
	started := []Server{&mockServer{}, &mockServer{err: errors.New("x")}}
	a := &app{
		hooks:  map[hook][]Callback{},
		logger: mlog.NewForConfig(nil), doneFunc: func() {
			called = true
		},
		started: started,
	}
	a.Shutdown()
	assert.True(t, called)
	for _, server := range started {
		assert.True(t, server.(*mockServer).called)
	}
}

// Run 中途失败后从未启动的 server 不在 started 集合内，Shutdown 不得回收它们。
func Test_app_Shutdown_SkipsUnstartedServers(t *testing.T) {
	started := &mockServer{}
	unstarted := &mockServer{}
	a := &app{
		hooks:    map[hook][]Callback{},
		logger:   mlog.NewForConfig(nil),
		doneFunc: func() {},
		servers:  []Server{started, unstarted},
		started:  []Server{started},
	}
	a.Shutdown()
	assert.True(t, started.called)
	assert.False(t, unstarted.called)
}

func Test_app_HttpHandler(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	httpHandler := NewMockHttpHandler(m)
	a := &app{
		httpHandler: httpHandler,
	}
	assert.NotNil(t, a.HttpHandler())
}

func Test_bootShortName(t *testing.T) {
	assert.Empty(t, bootShortName(nil))
	assert.Equal(t, "testBoot", bootShortName(&testBoot{}))
}

func Test_bootTags_has(t *testing.T) {
	assert.True(t, bootTags{"test"}.has("test"))
	assert.False(t, bootTags{"test"}.has("test1"))
}

func Test_excludeBootstrapperByTags(t *testing.T) {
	boots := []Bootstrapper{&testBoot{tags: []string{"test"}}, &testBoot{tags: []string{"test1"}}}
	res := excludeBootstrapperByTags([]string{"test"}, boots)
	assert.Len(t, res, 1)
	assert.Equal(t, "test1", res[0].Tags()[0])

	// 命中的 bootstrapper 被排除，其余原样保留。
	all := excludeBootstrapperByTags([]string{"no-such-tag"}, boots)
	assert.Len(t, all, 2)
	// 全部命中时返回空列表而非 nil。
	empty := excludeBootstrapperByTags([]string{"test", "test1"}, boots)
	assert.Empty(t, empty)
}

func Test_unknownExcludeTags(t *testing.T) {
	boots := []Bootstrapper{
		&testBoot{tags: []string{"api"}},
		&testBoot{tags: []string{"cron", "profile"}},
	}
	// 全部命中：无未知项。
	assert.Empty(t, unknownExcludeTags([]string{"api", "profile"}, boots))
	// 未知项原样返回。
	assert.Equal(t, []string{"nope"}, unknownExcludeTags([]string{"nope"}, boots))
	// 命中与未知混排：只返回未知项，且顺序稳定。
	assert.Equal(t, []string{"nope", "x"}, unknownExcludeTags([]string{"nope", "x", "api"}, boots))
	// 空排除列表。
	assert.Empty(t, unknownExcludeTags(nil, boots))
}

func TestNewApp_WarnsUnknownExcludeTag(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	appli := NewApp(
		&config2.Config{ExcludeServer: config2.ExcludeServerTags("not-a-tag")},
		data2.NewMockData(m),
		mlog.NewForConfig(nil),
		biz.NewMockAuthBiz(m),
		event.NewMockDispatcher(m),
		cron.NewMockManager(m),
		data2.NewMockCache(m),
		NewMockPluginManager(m),
		&GrpcRegistry{},
		&prometheus.Registry{},
		NewMockHttpHandler(m),
		timer.NewReal(),
		data2.NewMockProjectRepo(m),
		data2.NewMockK8sRepo(m),
		&eventhandler.EventCoordinator{},
		WithBootstrappers(&testBoot{tags: []string{"cron"}}),
	)
	// 未知 tag 不匹配任何 bootstrapper：排除为 no-op，bootstrapper 原样保留，
	// 同时触发 NewApp 的告警分支。
	assert.Len(t, appli.(*app).bootstrappers, 1)
}

func Test_app_Bootstrap_Error(t *testing.T) {
	a := &app{
		timer: timer.NewReal(),
		bootstrappers: []Bootstrapper{&testBoot{
			err: errors.New("x"),
		}},
	}
	assert.Error(t, a.Bootstrap())
}

type runRecorderServer struct {
	Server
	called bool
	err    error
}

func (r *runRecorderServer) Run(context.Context) error {
	r.called = true
	return r.err
}

func Test_app_Run_RunsBeforeHooksAndAllServers(t *testing.T) {
	beforeRun := false
	a := &app{
		logger: mlog.NewForConfig(nil),
		hooks: map[hook][]Callback{
			beforeRunHook: {func() { beforeRun = true }},
		},
		servers: []Server{&runRecorderServer{}, &runRecorderServer{}},
	}
	assert.Nil(t, a.Run(context.TODO()))
	assert.True(t, beforeRun)
	for _, s := range a.servers {
		assert.True(t, s.(*runRecorderServer).called)
	}
	// 全部 server 启动成功，进入 started 回收集合。
	assert.Len(t, a.started, 2)
}

func Test_app_Run_StopsOnServerError(t *testing.T) {
	a := &app{
		logger: mlog.NewForConfig(nil),
		hooks:  map[hook][]Callback{},
		servers: []Server{
			&runRecorderServer{},
			&runRecorderServer{err: errors.New("x")},
		},
	}
	assert.EqualError(t, a.Run(context.TODO()), "x")
	// 串行启动：失败前已启动的 server 照常启动，Run 立即上抛首个错误。
	assert.True(t, a.servers[0].(*runRecorderServer).called)
	assert.True(t, a.servers[1].(*runRecorderServer).called)
	// 只有 Run 成功的 server 进入 started；失败的那个不被回收。
	assert.Len(t, a.started, 1)
	assert.Equal(t, a.servers[0], a.started[0])
}

// blockingServer 的 Run 阻塞在 block 关闭前，用于验证启动超时闸门。
type blockingServer struct {
	Server
	entered chan struct{} // 进入 Run 即关闭：测试靠它同步 called 的写读，避免竞态
	block   chan struct{} // 关闭前 Run 一直阻塞
	called  bool
}

// Run 记录调用后阻塞，模拟违反"异步启动"契约的同步阻塞实现。
func (b *blockingServer) Run(context.Context) error {
	b.called = true
	close(b.entered)
	<-b.block
	return nil
}

// panicServer 的 Run 直接 panic，用于验证 panic 被转成启动错误而非误报超时。
type panicServer struct {
	Server
	called bool
}

// Run 记录调用后 panic，模拟启动期崩溃的 server 实现。
func (p *panicServer) Run(context.Context) error {
	p.called = true
	panic("boom")
}

func Test_app_Run_ServerStartupTimeout(t *testing.T) {
	orig := serverStartupTimeout
	serverStartupTimeout = 10 * time.Millisecond
	t.Cleanup(func() { serverStartupTimeout = orig })

	doneCtx, doneFunc := context.WithCancel(context.TODO())
	blocked := &blockingServer{entered: make(chan struct{}), block: make(chan struct{})}
	a := &app{
		done:     doneCtx,
		doneFunc: doneFunc,
		logger:   mlog.NewForConfig(nil),
		hooks:    map[hook][]Callback{},
		servers:  []Server{blocked},
	}
	err := a.Run(context.TODO())
	t.Cleanup(func() { close(blocked.block) }) // 放行被阻塞的 Run，避免 goroutine 泄漏
	// 等待 Run 真正进入阻塞：证明超时闸门拦截的是一个"已开始启动"的 server，
	// 且 channel 关闭与 called 的读构成 happens-before，消除竞态。
	<-blocked.entered
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "did not start within")
	}
	assert.True(t, blocked.called)
	// 超时的 server 不算启动成功，不进 started 回收集合。
	assert.Empty(t, a.started)
	// Shutdown 不回收该 server，且不因它阻塞。
	a.Shutdown()
}

func Test_app_Run_ServerPanicBecomesError(t *testing.T) {
	a := &app{
		logger:  mlog.NewForConfig(nil),
		hooks:   map[hook][]Callback{},
		servers: []Server{&panicServer{}},
	}
	err := a.Run(context.TODO())
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "panicked on Run")
		assert.Contains(t, err.Error(), "boom")
	}
	// panic 的 server 未启动成功，不进 started。
	assert.Empty(t, a.started)
}
