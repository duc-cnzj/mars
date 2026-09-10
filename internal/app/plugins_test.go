package app

import (
	"errors"
	"testing"

	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type testPlugin struct {
	GitServer
	WsSender
	Picture
	DomainManager
}

func (tp *testPlugin) Initialize(pluginApp PluginApp, args map[string]any) error {
	return nil
}

func (tp *testPlugin) Destroy() error {
	return nil
}

func (tp *testPlugin) Name() string {
	return "test"
}

// registerTestPlugin 注册测试插件，测试结束时清空整个注册表。
// RegisterPlugin 对同名注册会 panic，而 -count=N 重跑与测试间会重复注册同名插件，
// 故测试统一走本 helper，用 t.Cleanup 清空 pluginSet 保证隔离。
func registerTestPlugin(t *testing.T, name string, p Plugin) {
	t.Helper()
	mu.Lock()
	if _, dup := pluginSet[name]; dup {
		mu.Unlock()
		t.Fatalf("test plugin %q already registered", name)
	}
	pluginSet[name] = p
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		clear(pluginSet)
		mu.Unlock()
	})
}

func TestRegisterPlugin_DuplicatePanics(t *testing.T) {
	registerTestPlugin(t, "dup_test", &testPlugin{})
	// 同名二次注册（如两个插件包 init 撞名）→ panic 显式失败，而非静默覆盖。
	assert.Panics(t, func() {
		RegisterPlugin("dup_test", &testPlugin{})
	})
}

func TestRegisterPlugin_Success(t *testing.T) {
	// RegisterPlugin 正常路径：全新名字注册进注册表，可被 GetPlugins 拿到。
	// 生产侧真实入口是各插件包 init()，这里直接调注册函数补成功分支覆盖。
	mu.Lock()
	if _, existed := pluginSet["success_test"]; existed {
		mu.Unlock()
		t.Fatalf("test plugin %q already registered", "success_test")
	}
	mu.Unlock()
	RegisterPlugin("success_test", &testPlugin{})
	t.Cleanup(func() {
		mu.Lock()
		delete(pluginSet, "success_test")
		mu.Unlock()
	})
	assert.Contains(t, GetPlugins(), "success_test")
}

func TestPluginManagerLoad(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	registerTestPlugin(t, "test", &testPlugin{})

	app := NewMockApp(m)
	cfg := &config.Config{
		DomainManagerPlugin: config.Plugin{Name: "test"},
		WsSenderPlugin:      config.Plugin{Name: "test"},
		GitServerPlugin:     config.Plugin{Name: "test"},
		PicturePlugin:       config.Plugin{Name: "test"},
	}

	logger := mlog.NewForConfig(nil)

	manager, err := NewPluginManager(cfg, logger)
	assert.NoError(t, err)

	err = manager.Load(app)
	assert.NoError(t, err)

	assert.NotNil(t, manager.Domain())
	assert.NotNil(t, manager.Ws())
	assert.NotNil(t, manager.Git())
	assert.NotNil(t, manager.Picture())
}

func TestManagerGetPlugins(t *testing.T) {
	registerTestPlugin(t, "mgr_get_plugins", &testPlugin{})
	got := (&manager{}).GetPlugins()
	assert.Contains(t, got, "mgr_get_plugins")
}

func TestGetPluginsReturnsCopy(t *testing.T) {
	registerTestPlugin(t, "copy_test", &testPlugin{})
	before := GetPlugins()
	before["hacked"] = &testPlugin{}
	after := GetPlugins()
	_, ok := after["hacked"]
	assert.False(t, ok)
}

func TestGetPluginUnregistered(t *testing.T) {
	_, err := GetPlugin[WsSender](config.Plugin{Name: "not_registered_plugin"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not_registered_plugin")
	assert.Contains(t, err.Error(), "available plugins")
}

func TestManagerDestroyNilSafe(t *testing.T) {
	ma := &manager{logger: mlog.NewForConfig(nil)}
	assert.NotPanics(t, ma.Destroy)
}

func TestManagerDestroyOrderAndContinuesOnError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()

	pic := NewMockPicture(m)
	domain := NewMockDomainManager(m)
	ws := NewMockWsSender(m)
	git := NewMockGitServer(m)

	// 逆序销毁 picture→domain→ws→git；第一个失败不中断其余插件回收。
	gomock.InOrder(
		pic.EXPECT().Destroy().Return(errors.New("boom")),
		domain.EXPECT().Destroy().Return(nil),
		ws.EXPECT().Destroy().Return(nil),
		git.EXPECT().Destroy().Return(nil),
	)

	ma := &manager{logger: mlog.NewForConfig(nil), pic: pic, domain: domain, ws: ws, git: git}
	assert.NotPanics(t, ma.Destroy)
}

func TestPluginManagerLoad_MidwayError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	ma := &manager{
		logger:  mlog.NewForConfig(nil),
		gitFunc: func(PluginApp) (GitServer, error) { return nil, errors.New("git load failed") },
	}
	err := ma.Load(NewMockApp(m))
	assert.EqualError(t, err, "git load failed")
	assert.Nil(t, ma.git)
}

func TestPluginManagerLoad_WsError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	git := NewMockGitServer(m)
	git.EXPECT().Destroy().Return(nil) // 回滚：ws 失败时已加载的 git 必须被销毁
	ma := &manager{
		logger:  mlog.NewForConfig(nil),
		gitFunc: func(PluginApp) (GitServer, error) { return git, nil },
		wsFunc:  func(PluginApp) (WsSender, error) { return nil, errors.New("ws load failed") },
	}
	err := ma.Load(NewMockApp(m))
	assert.EqualError(t, err, "ws load failed")
	assert.NotNil(t, ma.git)
	assert.Nil(t, ma.ws)
}

func TestPluginManagerLoad_DomainError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	git := NewMockGitServer(m)
	ws := NewMockWsSender(m)
	gomock.InOrder(
		ws.EXPECT().Destroy().Return(nil), // 逆序回滚：ws 先于 git
		git.EXPECT().Destroy().Return(nil),
	)
	ma := &manager{
		logger:     mlog.NewForConfig(nil),
		gitFunc:    func(PluginApp) (GitServer, error) { return git, nil },
		wsFunc:     func(PluginApp) (WsSender, error) { return ws, nil },
		domainFunc: func(PluginApp) (DomainManager, error) { return nil, errors.New("domain load failed") },
	}
	err := ma.Load(NewMockApp(m))
	assert.EqualError(t, err, "domain load failed")
	assert.NotNil(t, ma.git)
	assert.NotNil(t, ma.ws)
	assert.Nil(t, ma.domain)
	assert.Nil(t, ma.pic)
}

func TestPluginManagerLoad_PictureError(t *testing.T) {
	m := gomock.NewController(t)
	defer m.Finish()
	git := NewMockGitServer(m)
	ws := NewMockWsSender(m)
	domain := NewMockDomainManager(m)
	gomock.InOrder(
		domain.EXPECT().Destroy().Return(nil), // 逆序回滚：domain→ws→git
		ws.EXPECT().Destroy().Return(nil),
		git.EXPECT().Destroy().Return(nil),
	)
	ma := &manager{
		logger:     mlog.NewForConfig(nil),
		gitFunc:    func(PluginApp) (GitServer, error) { return git, nil },
		wsFunc:     func(PluginApp) (WsSender, error) { return ws, nil },
		domainFunc: func(PluginApp) (DomainManager, error) { return domain, nil },
		picFunc:    func(PluginApp) (Picture, error) { return nil, errors.New("pic load failed") },
	}
	err := ma.Load(NewMockApp(m))
	assert.EqualError(t, err, "pic load failed")
	// 中途成功的前三个已赋值，失败的 pic 保持 nil。
	assert.NotNil(t, ma.git)
	assert.NotNil(t, ma.ws)
	assert.NotNil(t, ma.domain)
	assert.Nil(t, ma.pic)
}

type errInitializePlugin struct {
	WsSender
}

func (p *errInitializePlugin) Initialize(pluginApp PluginApp, args map[string]any) error {
	return errors.New("init failed")
}

func (p *errInitializePlugin) Destroy() error { return nil }

func (p *errInitializePlugin) Name() string { return "err_init" }

func TestGetPlugin_InitializeError(t *testing.T) {
	registerTestPlugin(t, "err_init", &errInitializePlugin{})
	closure, err := GetPlugin[WsSender](config.Plugin{Name: "err_init"})
	assert.NoError(t, err)
	_, err = closure(nil)
	assert.EqualError(t, err, "init failed")
}

// minimalPlugin 只实现 Plugin，用于验证 GetPlugin 对"名字已注册但接口不符"的配置错误返回 error 而非 panic。
type minimalPlugin struct{}

func (minimalPlugin) Name() string                               { return "minimal" }
func (minimalPlugin) Initialize(PluginApp, map[string]any) error { return nil }
func (minimalPlugin) Destroy() error                             { return nil }

func TestGetPlugin_TypeMismatch(t *testing.T) {
	registerTestPlugin(t, "only_plugin", &minimalPlugin{})
	closure, err := GetPlugin[WsSender](config.Plugin{Name: "only_plugin"})
	assert.NoError(t, err)
	_, err = closure(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only_plugin")
}

func TestNewPluginManager_Unregistered(t *testing.T) {
	_, err := NewPluginManager(&config.Config{
		DomainManagerPlugin: config.Plugin{Name: "no_such_plugin"},
	}, mlog.NewForConfig(nil))
	assert.Error(t, err)
}

func TestNewPluginManager_GitUnregistered(t *testing.T) {
	registerTestPlugin(t, "pm_git_test", &testPlugin{})
	// domain/ws 已注册、git 未注册 → err 来自 git 分支（后段的 GetPlugin）。
	_, err := NewPluginManager(&config.Config{
		DomainManagerPlugin: config.Plugin{Name: "pm_git_test"},
		WsSenderPlugin:      config.Plugin{Name: "pm_git_test"},
		GitServerPlugin:     config.Plugin{Name: "no_such_git"},
	}, mlog.NewForConfig(nil))
	assert.Error(t, err)
}

func TestNewPluginManager_WsUnregistered(t *testing.T) {
	registerTestPlugin(t, "pm_ws_test", &testPlugin{})
	// domain 已注册、ws 未注册 → err 来自 ws 分支（中段的 GetPlugin）。
	_, err := NewPluginManager(&config.Config{
		DomainManagerPlugin: config.Plugin{Name: "pm_ws_test"},
		WsSenderPlugin:      config.Plugin{Name: "no_such_ws"},
	}, mlog.NewForConfig(nil))
	assert.Error(t, err)
}

func TestNewPluginManager_PictureUnregistered(t *testing.T) {
	registerTestPlugin(t, "pm_pic_test", &testPlugin{})
	// domain/ws/git 已注册、pic 未注册 → err 来自 pic 分支（末段的 GetPlugin）。
	_, err := NewPluginManager(&config.Config{
		DomainManagerPlugin: config.Plugin{Name: "pm_pic_test"},
		WsSenderPlugin:      config.Plugin{Name: "pm_pic_test"},
		GitServerPlugin:     config.Plugin{Name: "pm_pic_test"},
		PicturePlugin:       config.Plugin{Name: "no_such_pic"},
	}, mlog.NewForConfig(nil))
	assert.Error(t, err)
}

// resolverApp 仅实现 PluginApp 收窄后的两方法（Logger+ProjectRepo），不实现 K8sRepo，
// 用于覆盖 Resolve 的断言成功与断言失败（panic）两条分支。
type resolverApp struct{}

func (resolverApp) Logger() mlog.Logger          { return mlog.NewForConfig(nil) }
func (resolverApp) ProjectRepo() biz.ProjectRepo { return nil }

func TestResolve(t *testing.T) {
	// 成功：动态类型实现窄视图，返回视图实例。
	d := Resolve[interface{ Logger() mlog.Logger }](resolverApp{})
	assert.NotNil(t, d.Logger())

	// 失败：动态类型未实现窄视图要求的 K8sRepo → panic 显式暴露装配配置错误。
	assert.Panics(t, func() {
		Resolve[interface{ K8sRepo() biz.K8sRepo }](resolverApp{})
	})
}
