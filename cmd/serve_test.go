package cmd

import (
	"context"
	"testing"

	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/app/bootstrappers"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/cron"
	"github.com/duc-cnzj/mars/v6/internal/cronjob"
	"github.com/duc-cnzj/mars/v6/internal/data"
	"github.com/duc-cnzj/mars/v6/internal/data/ent"
	"github.com/duc-cnzj/mars/v6/internal/eventhandler"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/timer"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
)

// TestExcludeServerDocumentedTagsCovered 锁住 --exclude_server 的文档承诺与真实 bootstrapper tag 的一致性。
// flag help（serve.go）与 README 声明可排除的服务名：api / metrics / cron / profile。
// 白名单来自 documentedExcludeTags（serve.go 的单一来源），flag help 也由它派生，
// 测试再引用同一变量，杜绝"文档写 A、代码用 B"的三处漂移。
func TestExcludeServerDocumentedTagsCovered(t *testing.T) {
	got := make(map[string]bool)
	for _, boot := range serverBootstrappers {
		for _, tag := range boot.Tags() {
			got[tag] = true
		}
	}

	for _, name := range documentedExcludeTags {
		if !got[name] {
			t.Errorf("--exclude_server 文档承诺的 %q 没有任何 bootstrapper tag 命中，排除会静默失效", name)
		}
	}
}

// TestServerBootstrappersPluginLast 锁住 plugin 必须是启动顺序最后一位的契约。
// 插件在 bootstrap 尾步完成 Initialize，wire 装配（serve.go 的惰性闭包）在触发时
// 才实时解析插件能力；把 plugin 插到中间会破坏初始化时序约定。
// 顺序是组合根（serve.go）的职责，契约锁在组合根一侧，不引入额外接口机制。
func TestServerBootstrappersPluginLast(t *testing.T) {
	if len(serverBootstrappers) == 0 {
		t.Fatal("serverBootstrappers 为空")
	}
	if _, ok := serverBootstrappers[len(serverBootstrappers)-1].(*bootstrappers.PluginBootstrapper); !ok {
		t.Fatalf("serverBootstrappers 最后一位必须是 PluginBootstrapper，实际是 %T", serverBootstrappers[len(serverBootstrappers)-1])
	}
}

// TestPodEventPublisher_LazyResolvesWs 验证 podPubAdapter 惰性解析：
// 构造期（wire 期，插件未加载）不触碰 Ws()；首次 Publish 才解析 ws 插件建
// 一次性 PubSub 并断言发布器；后续 Publish 复用同一发布器，不再重复解析。
func TestPodEventPublisher_LazyResolvesWs(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	pm := app.NewMockPluginManager(ctrl)
	pub := newPodEventPublisher(pm)
	// 构造期无 Ws() EXPECT：若被调用会 panic，反向证明惰性。

	ws := app.NewMockWsSender(ctrl)
	sub := app.NewMockPubSub(ctrl)
	pm.EXPECT().Ws().Return(ws)
	ws.EXPECT().New("", "").Return(sub)
	sub.EXPECT().Publish(int64(7), gomock.Any()).Return(nil)

	assert.NoError(t, pub.Publish(7, &corev1.Pod{}))

	// 第二次 Publish：once 已解析，复用同一发布器，Ws()/New 不再被调用。
	sub.EXPECT().Publish(int64(8), gomock.Any()).Return(nil)
	assert.NoError(t, pub.Publish(8, &corev1.Pod{}))
}

// TestProvideGitServer_LazyResolvesGit 验证 provideGitServer 惰性解析：
// 构造期（wire 期，插件未加载）不触碰 Git()；首次调用闭包才实时解析 git 插件，
// 替代原 GitServerHolder 快照 + 启动前 refreshGitServer 的机制。
func TestProvideGitServer_LazyResolvesGit(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	pm := app.NewMockPluginManager(ctrl)
	git := app.NewMockGitServer(ctrl)

	getGit := provideGitServer(pm)
	// 构造期无 Git() EXPECT：若被调用会 panic，反向证明惰性。

	pm.EXPECT().Git().Return(git)
	assert.Equal(t, data.GitServer(git), getGit())
}

// TestRegisterCronJobs 覆盖机械注册循环：每个 CronTask 的调度声明经 Schedule
// 落进真实 cron.Manager，命令名与 cron 表达式正确（表达式由各调度构造器产出）。
func TestRegisterCronJobs(t *testing.T) {
	cm := cron.NewManager(timer.NewReal(), nil, nil, mlog.NewForConfig(nil))
	RegisterCronJobs([]cronjob.CronTask{
		{Name: "daily", Schedule: func(cmd cron.Command) cron.Command { return cmd.DailyAt("2:00") }, Run: func() error { return nil }},
		{Name: "minutely", Schedule: cron.Command.EveryMinute, Run: func() error { return nil }},
		{Name: "two-min", Schedule: cron.Command.EveryTwoMinutes, Run: func() error { return nil }},
		{Name: "five-min", Schedule: cron.Command.EveryFiveMinutes, Run: func() error { return nil }},
		{Name: "ten-min", Schedule: cron.Command.EveryTenMinutes, Run: func() error { return nil }},
	}, cm)

	exprs := make(map[string]string)
	for _, c := range cm.List() {
		exprs[c.Name()] = c.Expression()
	}
	assert.Len(t, exprs, 5)
	assert.Equal(t, "0 00 2 * * *", exprs["daily"])
	assert.Equal(t, "0 * * * * *", exprs["minutely"])
	assert.Equal(t, "0 */2 * * * *", exprs["two-min"])
	assert.Equal(t, "0 */5 * * * *", exprs["five-min"])
	assert.Equal(t, "0 */10 * * * *", exprs["ten-min"])
}

// ---------------------------------------------------------------------------
// serve.go wire providers：惰性插件闭包
// ---------------------------------------------------------------------------

// TestNewGetCertsFunc_LazyResolvesDomain 验证 newGetCertsFunc 惰性解析：
// 构造期不触碰 Domain()；首次调用闭包才实时解析域名插件取证书。
func TestNewGetCertsFunc_LazyResolvesDomain(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	pm := app.NewMockPluginManager(ctrl)
	getCerts := newGetCertsFunc(pm)
	// 构造期无 Domain() EXPECT：若被调用会 panic，反向证明惰性。

	domain := app.NewMockDomainManager(ctrl)
	pm.EXPECT().Domain().Return(domain)
	domain.EXPECT().GetCerts().Return("name", "key", "crt")

	name, key, crt := getCerts()
	assert.Equal(t, "name", name)
	assert.Equal(t, "key", key)
	assert.Equal(t, "crt", crt)
}

// TestNewToAllFunc_LazyResolvesWs 验证 newToAllFunc 惰性解析：
// 构造期不触碰 Ws()；调用闭包时建一次性 PubSub、广播后关闭。
func TestNewToAllFunc_LazyResolvesWs(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	pm := app.NewMockPluginManager(ctrl)
	toAll := newToAllFunc(pm)
	// 构造期无 Ws() EXPECT：若被调用会 panic，反向证明惰性。

	ws := app.NewMockWsSender(ctrl)
	sub := app.NewMockPubSub(ctrl)
	pm.EXPECT().Ws().Return(ws)
	ws.EXPECT().New("", "").Return(sub)
	sub.EXPECT().Close().Return(nil)
	sub.EXPECT().ToAll(gomock.Any()).Return(nil)

	assert.NoError(t, toAll(&websocket_pb.WsProjectPodEventResponse{}))
}

// TestProvideEventDeps_WiresLazyClosures 验证 provideEventDeps 把 GetCerts/ToAll
// 两个惰性闭包装配进 PluginDeps，触发时实时解析插件。
func TestProvideEventDeps_WiresLazyClosures(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	pm := app.NewMockPluginManager(ctrl)
	deps := provideEventDeps(pm)
	require.NotNil(t, deps)
	require.NotNil(t, deps.GetCerts)
	require.NotNil(t, deps.ToAll)

	// GetCerts 惰性解析。
	domain := app.NewMockDomainManager(ctrl)
	pm.EXPECT().Domain().Return(domain)
	domain.EXPECT().GetCerts().Return("n", "k", "c")
	name, key, crt := deps.GetCerts()
	assert.Equal(t, "n", name)
	assert.Equal(t, "k", key)
	assert.Equal(t, "c", crt)

	// ToAll 惰性解析。
	ws := app.NewMockWsSender(ctrl)
	sub := app.NewMockPubSub(ctrl)
	pm.EXPECT().Ws().Return(ws)
	ws.EXPECT().New("", "").Return(sub)
	sub.EXPECT().Close().Return(nil)
	sub.EXPECT().ToAll(gomock.Any()).Return(nil)
	assert.NoError(t, deps.ToAll(&websocket_pb.WsProjectPodEventResponse{}))
}

// TestProvideCronDeps_WiresLazyClosures 验证 provideCronDeps 装配 GetCerts 惰性闭包。
func TestProvideCronDeps_WiresLazyClosures(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	pm := app.NewMockPluginManager(ctrl)
	deps := provideCronDeps(pm)
	require.NotNil(t, deps)
	require.NotNil(t, deps.GetCerts)

	domain := app.NewMockDomainManager(ctrl)
	pm.EXPECT().Domain().Return(domain)
	domain.EXPECT().GetCerts().Return("n", "k", "c")
	name, key, crt := deps.GetCerts()
	assert.Equal(t, "n", name)
	assert.Equal(t, "k", key)
	assert.Equal(t, "c", crt)
}

// TestProvidePodEventPublisher 验证 providePodEventPublisher 返回绑定 pm 的
// podPubAdapter（惰性发布器）。
func TestProvidePodEventPublisher(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	pm := app.NewMockPluginManager(ctrl)
	pub := providePodEventPublisher(pm)
	require.NotNil(t, pub)

	adapter, ok := pub.(*podPubAdapter)
	require.True(t, ok)
	assert.Same(t, pm, adapter.pm)
}

// fakeMinioGetter 实现 data.MinioGetter：记录是否被调用，返回 nil 客户端。
type fakeMinioGetter struct{ called bool }

// MinioCli 标记调用并返回 nil（仅验证委托，不构造真实客户端）。
func (f *fakeMinioGetter) MinioCli() *minio.Client { f.called = true; return nil }

// TestProvideMinioClient 验证 provideMinioClient 惰性取数委托给 MinioGetter 端口。
func TestProvideMinioClient(t *testing.T) {
	g := &fakeMinioGetter{}
	get := provideMinioClient(g)
	assert.Nil(t, get())
	assert.True(t, g.called)
}

// fakeDBGetter 实现 data.DBGetter：记录是否被调用，返回 nil 客户端。
type fakeDBGetter struct{ called bool }

// DB 标记调用并返回 nil（仅验证委托，不构造真实客户端）。
func (f *fakeDBGetter) DB() *ent.Client { f.called = true; return nil }

// TestProvideDBGetter 验证 provideDBGetter 惰性取数委托给 DBGetter 端口。
func TestProvideDBGetter(t *testing.T) {
	g := &fakeDBGetter{}
	get := provideDBGetter(g)
	assert.Nil(t, get())
	assert.True(t, g.called)
}

// TestProvideCacheDriver_SqliteDbFallback 验证 sqlite + db 组合强制回退内存锁。
func TestProvideCacheDriver_SqliteDbFallback(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	got := provideCacheDriver(&config.Config{DBDriver: "sqlite", CacheDriver: "db"}, logger)
	assert.Equal(t, locker.DriverMemory, got)
}

// TestProvideCacheDriver_Passthrough 验证其他组合原样返回 cache 驱动。
func TestProvideCacheDriver_Passthrough(t *testing.T) {
	logger := mlog.NewForConfig(nil)
	assert.Equal(t, locker.DriverDB, provideCacheDriver(&config.Config{DBDriver: "mysql", CacheDriver: "db"}, logger))
	assert.Equal(t, locker.DriverMemory, provideCacheDriver(&config.Config{DBDriver: "mysql", CacheDriver: "memory"}, logger))
}

// TestPodListenerServer_ShutdownNoCancel 验证未 Run 时 Shutdown 是幂等 no-op。
func TestPodListenerServer_ShutdownNoCancel(t *testing.T) {
	s := &podListenerServer{listener: nil}
	assert.NoError(t, s.Shutdown(context.Background()))
}

// TestPodListenerServer_RunAndShutdown 验证 Run 异步启动监听并持有 cancel，
// Shutdown 取消后监听退出（SubscribePodEvents 由 mock 提供，不触真实 informer）。
func TestPodListenerServer_RunAndShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	k8sRepo := data.NewMockK8sRepo(ctrl)
	ch := make(chan biz.PodEvent)
	k8sRepo.EXPECT().SubscribePodEvents("pod-watcher").Return(ch, func() {}).AnyTimes()

	listener := eventhandler.NewPodEventListener(mlog.NewForConfig(nil), k8sRepo, nil, nil)
	s, ok := newPodListenerServer(listener).(*podListenerServer)
	require.True(t, ok)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, s.Run(ctx))
	require.NotNil(t, s.cancel)
	assert.NoError(t, s.Shutdown(context.Background()))
}
