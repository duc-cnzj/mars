package app

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/config"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
)

// Plugin 可加载插件：随 app 初始化，停机时销毁。
type Plugin interface {
	// Name 插件名。
	Name() string
	// Initialize 初始化插件。
	Initialize(pluginApp PluginApp, args map[string]any) error
	// Destroy 销毁插件。
	Destroy() error
}

// 全局插件注册表，由 mu 保护。
var (
	mu        sync.RWMutex
	pluginSet = make(map[string]Plugin)
)

// newFunc 在加载时基于 PluginApp 构建插件实例。
type newFunc[T Plugin] func(PluginApp) (T, error)

// GetPlugins 获取全部已注册插件。
func GetPlugins() map[string]Plugin {
	mu.RLock()
	defer mu.RUnlock()
	// 返回拷贝，避免调用方拿到内部 map 的引用后与注册写并发。
	result := make(map[string]Plugin, len(pluginSet))
	for name, plugin := range pluginSet {
		result[name] = plugin
	}
	return result
}

// GetPluginNames 返回全部已注册插件的排序名字。
func GetPluginNames() []string {
	ps := GetPlugins()
	names := make([]string, 0, len(ps))
	for name := range ps {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterPlugin 注册插件；同名注册直接 panic——显式失败优于静默覆盖，
// 撞名是装配配置错误（两个包注册同一插件名），init 期 panic 是合法信号。
func RegisterPlugin(name string, pluginInterface Plugin) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := pluginSet[name]; dup {
		panic(fmt.Sprintf("plugin %q already registered", name))
	}
	pluginSet[name] = pluginInterface
}

// WsSender websocket 发送插件。
type WsSender interface {
	Plugin

	// New 新建 PubSub。
	New(uid, id string) PubSub
}

// WebsocketMessage 带元数据的 websocket 消息。
type WebsocketMessage interface {
	proto.Message
	// GetMetadata 返回 websocket 消息元数据。
	GetMetadata() *websocket.Metadata
}

// PubSub ws 发送者的 pubsub 通道。
type PubSub interface {
	ProjectPodEventSubscriber
	ProjectPodEventPublisher

	// Info 返回通道信息。
	Info() any
	// Uid 返回通道属主 uid。
	Uid() string
	// ID 返回通道 id。
	ID() string
	// ToSelf 仅向通道属主发送消息。
	ToSelf(WebsocketMessage) error
	// ToAll 向全部通道成员发送消息。
	ToAll(WebsocketMessage) error
	// Subscribe 返回消息接收通道。
	Subscribe() <-chan []byte
	// Close 关闭通道。
	Close() error
}

// ProjectPodEventSubscriber 订阅项目 pod 事件。
type ProjectPodEventSubscriber interface {
	// Join 订阅某个项目的 pod 事件。
	Join(projectID int64) error
	// Leave 退订某个项目的 pod 事件。
	Leave(nsID int64, projectID int64) error
	// Run 持续消费 pod 事件直至 ctx 结束。
	Run(ctx context.Context) error
}

// ProjectPodEventPublisher 发布项目 pod 事件。
type ProjectPodEventPublisher interface {
	// Publish 发布某个项目的 pod 变更事件。
	Publish(nsID int64, pod *corev1.Pod) error
}

// PictureItem biz 图片项的别名。
type PictureItem = biz.PictureItem

// Picture 随机图片插件。
type Picture interface {
	Plugin
	// Get 返回随机或指定图片。
	Get(ctx context.Context, random bool) (*biz.PictureItem, error)
}

// DomainManager 域名插件。
type DomainManager interface {
	Plugin

	// GetDomainByIndex 参数：domainSuffix: test.com，project: mars，namespace: default，index: 0,1,2...，preOccupiedLen: 预占用的长度
	// index 传 -1 生成不带 index 的子域名。
	GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string

	// GetCertSecretName 获取 HTTPS 证书对应的 secret
	GetCertSecretName(projectName string, index int) string

	// GetClusterIssuer CertManager 要用
	GetClusterIssuer() string

	// GetCerts 在 namespace 创建的时候注入证书信息
	GetCerts() (name, key, crt string)
}

// GitServer git 插件。
type GitServer interface {
	Plugin

	// GetProject 返回单个项目。
	GetProject(pid string) (*biz.GitProject, error)
	// AllProjects 返回全部项目。
	AllProjects() ([]*biz.GitProject, error)
	// AllBranches 返回某个项目的全部分支。
	AllBranches(pid string) ([]*biz.Branch, error)
	// GetCommit 返回单个提交。
	GetCommit(pid string, sha string) (*biz.Commit, error)
	// GetCommitPipeline 返回某个提交的流水线。
	GetCommitPipeline(pid string, branch string, sha string) (*biz.Pipeline, error)
	// PipelineJobOptions 返回项目流水线的 stage/job 去重选项，供配置通过规则下拉。
	PipelineJobOptions(pid string, branch string) (stages []string, jobs []string, err error)
	// ListCommits 返回某个分支的提交列表。
	ListCommits(pid string, branch string) ([]*biz.Commit, error)
	// GetFileContentWithBranch 按分支返回文件内容。
	GetFileContentWithBranch(pid string, branch string, filename string) (string, error)
	// GetFileContentWithSha 按 sha 返回文件内容。
	GetFileContentWithSha(pid string, sha string, filename string) (string, error)
	// GetDirectoryFilesWithBranch 按分支返回目录文件列表。
	GetDirectoryFilesWithBranch(pid string, branch string, path string, recursive bool) ([]string, error)
	// GetDirectoryFilesWithSha 按 sha 返回目录文件列表。
	GetDirectoryFilesWithSha(pid string, sha string, path string, recursive bool) ([]string, error)
}

// PluginManager 管理已加载的插件。
type PluginManager interface {
	// Load 按序加载全部插件，任一失败立即返回。
	Load(PluginApp) error
	// Destroy 逆序销毁已加载的插件。
	Destroy()
	// Domain 返回域名插件。
	Domain() DomainManager
	// Ws 返回 ws 插件。
	Ws() WsSender
	// Git 返回 git 插件。
	Git() GitServer
	// Picture 返回图片插件。
	Picture() biz.PictureGetter

	// GetPlugins 返回全部已注册插件。
	GetPlugins() map[string]Plugin
}

var _ PluginManager = (*manager)(nil)

// manager 实现 PluginManager，加载时懒构建各插件。
type manager struct {
	domainFunc newFunc[DomainManager]
	wsFunc     newFunc[WsSender]
	gitFunc    newFunc[GitServer]
	picFunc    newFunc[Picture]

	domain DomainManager
	ws     WsSender
	git    GitServer
	pic    Picture

	logger mlog.Logger
}

// Load 实现 PluginManager 接口的 Load。
func (m *manager) Load(app PluginApp) (err error) {
	m.logger.Info("load plugins")
	// 任一插件加载失败：逆序销毁已加载插件，防部分初始化泄漏。
	// PluginBootstrapper 失败路径不会注册 Destroy，清理责任收口在 Load 自身。
	defer func() {
		if err != nil {
			m.rollback()
		}
	}()

	if m.git, err = m.gitFunc(app); err != nil {
		return
	}
	if m.ws, err = m.wsFunc(app); err != nil {
		return
	}
	if m.domain, err = m.domainFunc(app); err != nil {
		return
	}
	m.pic, err = m.picFunc(app)
	return
}

// rollback 按加载逆序（picture→domain→ws→git）销毁已加载插件。
// Load 部分失败与 Destroy 共用；各字段可能为 nil（未加载到该步），destroyOne 逐空守卫，
// 单个插件销毁失败只记日志，不中断其余插件回收。
func (m *manager) rollback() {
	m.destroyOne("picture", m.pic)
	m.destroyOne("domain", m.domain)
	m.destroyOne("ws", m.ws)
	m.destroyOne("git", m.git)
}

// Destroy 逆序销毁已加载的插件。
func (m *manager) Destroy() {
	m.rollback()
}

// destroyOne 销毁单个插件，失败只记日志不中断其余插件。
func (m *manager) destroyOne(name string, p Plugin) {
	if p == nil {
		return
	}
	if err := p.Destroy(); err != nil {
		m.logger.Errorf("[Plugin]: destroy %s error: %v", name, err)
	}
}

// Domain 实现 PluginManager 接口的 Domain。
func (m *manager) Domain() DomainManager {
	return m.domain
}

// Ws 实现 PluginManager 接口的 Ws。
func (m *manager) Ws() WsSender {
	return m.ws
}

// Git 实现 PluginManager 接口的 Git。
func (m *manager) Git() GitServer {
	return m.git
}

// Picture 实现 PluginManager 接口的 Picture。
func (m *manager) Picture() biz.PictureGetter {
	return m.pic
}

// GetPlugins 实现 PluginManager 接口的 GetPlugins。
func (m *manager) GetPlugins() map[string]Plugin {
	return GetPlugins()
}

// NewPluginManager 基于配置的插件构建 manager。
func NewPluginManager(cfg *config.Config, logger mlog.Logger) (PluginManager, error) {
	domain, err := GetPlugin[DomainManager](cfg.DomainManagerPlugin)
	if err != nil {
		return nil, err
	}

	ws, err := GetPlugin[WsSender](cfg.WsSenderPlugin)
	if err != nil {
		return nil, err
	}

	git, err := GetPlugin[GitServer](cfg.GitServerPlugin)
	if err != nil {
		return nil, err
	}

	pic, err := GetPlugin[Picture](cfg.PicturePlugin)
	if err != nil {
		return nil, err
	}

	ma := &manager{
		logger:     logger,
		domainFunc: domain,
		wsFunc:     ws,
		gitFunc:    git,
		picFunc:    pic,
	}
	return ma, nil
}

// GetPlugin 解析已注册插件并返回其构造器。
func GetPlugin[T Plugin](p config.Plugin) (func(app PluginApp) (T, error), error) {
	pl, ok := GetPlugins()[p.Name]
	if !ok {
		// 未注册的插件名在构造期就报错，而不是等 Initialize 阶段 nil panic。
		return nil, fmt.Errorf("plugin %q not registered, available plugins: %v", p.Name, GetPluginNames())
	}
	return func(app PluginApp) (T, error) {
		var res T
		if err := pl.Initialize(app, p.Args); err != nil {
			return res, err
		}
		typed, ok := pl.(T)
		if !ok {
			// 名字已注册但实现的接口与期望不符（配置错误），构造期报错而非类型断言 panic。
			return res, fmt.Errorf("plugin %q implements %T, expected %T", p.Name, pl, res)
		}
		return typed, nil
	}, nil
}

// Resolve 把宽入口 PluginApp 断言成插件包内声明的窄依赖视图 T。
//
// 插件在自己的包内定义 deps 接口，只声明 Initialize 真正用到的能力，例如：
//
//	type deps interface{ Logger() mlog.Logger; K8sRepo() biz.K8sRepo }
//
//	func (d *syncSecretDomainManager) Initialize(pluginApp app.PluginApp, args map[string]any) error {
//	    d.k8sRepo = app.Resolve[deps](pluginApp).K8sRepo()
//	    ...
//	}
//
// Go 类型断言检查的是 PluginApp 的动态类型（组合根 *app，实现了全部能力）而非静态接口，
// 因此 Cache/K8sRepo 等「单插件独有能力」即使不在 PluginApp 里也能断言拿到——
// 移出接口 ≠ 移出实现。组合根未实现 T 属于装配配置错误，panic 使其在 Load 期显式暴露。
func Resolve[T any](app PluginApp) T {
	v, ok := app.(T)
	if !ok {
		panic(fmt.Sprintf("app does not implement %T required by plugin", (*T)(nil)))
	}
	return v
}
