package application

import (
	"context"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/api/v4/websocket"
	"github.com/duc-cnzj/mars/v4/internal/config"
	"github.com/duc-cnzj/mars/v4/internal/mlog"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
)

type Plugin interface {
	// Name plugin name.
	Name() string
	// Initialize init plugin.
	Initialize(app App, args map[string]any) error
	// Destroy plugin.
	Destroy() error
}

var (
	mu        sync.RWMutex
	pluginSet = make(map[string]Plugin)
)

type newFunc[T Plugin] func(App) (T, error)

// GetPlugins get all registered
func GetPlugins() map[string]Plugin {
	mu.RLock()
	defer mu.RUnlock()
	return pluginSet
}

// RegisterPlugin register plugin.
func RegisterPlugin(name string, pluginInterface Plugin) {
	mu.Lock()
	defer mu.Unlock()
	pluginSet[name] = pluginInterface
}

type WsSender interface {
	Plugin

	// New PubSub
	New(uid, id string) PubSub
}

type WebsocketMessage interface {
	proto.Message
	GetMetadata() *websocket.Metadata
}

type PubSub interface {
	ProjectPodEventSubscriber
	ProjectPodEventPublisher

	Info() any
	Uid() string
	ID() string
	ToSelf(WebsocketMessage) error
	ToAll(WebsocketMessage) error
	ToOthers(WebsocketMessage) error
	Subscribe() <-chan []byte
	Close() error
}

type ProjectPodEventSubscriber interface {
	Join(projectID int64) error
	Leave(nsID int64, projectID int64) error
	Run(ctx context.Context) error
}

type ProjectPodEventPublisher interface {
	Publish(nsID int64, pod *corev1.Pod) error
}

type PictureItem struct {
	Url       string
	Copyright string
}

type Picture interface {
	Plugin

	// Get picture.
	Get(ctx context.Context, random bool) (*PictureItem, error)
}

type DomainManager interface {
	Plugin

	// GetDomainByIndex domainSuffix: test.com, project: mars, namespace: default index: 0,1,2..., preOccupiedLen: 预占用的长度
	GetDomainByIndex(projectName, namespace string, index, preOccupiedLen int) string

	// GetDomain domainSuffix: test.com, project: mars, namespace: production, preOccupiedLen: 预占用的长度
	GetDomain(projectName, namespace string, preOccupiedLen int) string

	// GetCertSecretName 获取 HTTPS 证书对应的 secret
	GetCertSecretName(projectName string, index int) string

	// GetClusterIssuer CertManager 要用
	GetClusterIssuer() string

	// GetCerts 在 namespace 创建的时候注入证书信息
	GetCerts() (name, key, crt string)
}

type Status = string

const (
	StatusUnknown Status = "unknown"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
	StatusRunning Status = "running"
)

type Project interface {
	GetID() int64
	GetName() string
	GetDefaultBranch() string
	GetPath() string
	GetWebURL() string
	GetAvatarURL() string
	GetDescription() string
}

type Branch interface {
	GetName() string
	IsDefault() bool
	GetWebURL() string
}

type Pipeline interface {
	GetID() int64
	GetProjectID() int64
	GetStatus() Status
	GetRef() string
	GetSHA() string
	GetWebURL() string
	GetUpdatedAt() *time.Time
	GetCreatedAt() *time.Time
}

type Commit interface {
	GetID() string
	GetShortID() string
	GetTitle() string
	GetCommittedDate() *time.Time
	GetAuthorName() string
	GetAuthorEmail() string
	GetCommitterName() string
	GetCommitterEmail() string
	GetCreatedAt() *time.Time
	GetMessage() string
	GetProjectID() int64
	GetWebURL() string
}

type paginate interface {
	Page() int
	PageSize() int
	HasMore() bool
	NextPage() int
}

type ListProjectResponse interface {
	paginate
	GetItems() []Project
}

type ListBranchResponse interface {
	paginate
	GetItems() []Branch
}

type GitServer interface {
	Plugin

	GetProject(pid string) (Project, error)
	ListProjects(page, pageSize int) (ListProjectResponse, error)
	AllProjects() ([]Project, error)

	ListBranches(pid string, page, pageSize int) (ListBranchResponse, error)
	AllBranches(pid string) ([]Branch, error)

	GetCommit(pid string, sha string) (Commit, error)
	GetCommitPipeline(pid string, branch string, sha string) (Pipeline, error)
	ListCommits(pid string, branch string) ([]Commit, error)

	GetFileContentWithBranch(pid string, branch string, filename string) (string, error)
	GetFileContentWithSha(pid string, sha string, filename string) (string, error)

	GetDirectoryFilesWithBranch(pid string, branch string, path string, recursive bool) ([]string, error)
	GetDirectoryFilesWithSha(pid string, sha string, path string, recursive bool) ([]string, error)
}

type PluginManger interface {
	Load(App) error
	Domain() DomainManager
	Ws() WsSender
	Git() GitServer
	Picture() Picture

	GetPlugins() map[string]Plugin
}

var _ PluginManger = (*manager)(nil)

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

func (m *manager) Load(app App) (err error) {
	m.logger.Info("load plugins")
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

func (m *manager) Domain() DomainManager {
	return m.domain
}

func (m *manager) Ws() WsSender {
	return m.ws
}

func (m *manager) Git() GitServer {
	return m.git
}

func (m *manager) Picture() Picture {
	return m.pic
}

func (m *manager) GetPlugins() map[string]Plugin {
	return GetPlugins()
}

func NewPluginManager(cfg *config.Config, logger mlog.Logger) (PluginManger, error) {
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

func GetPlugin[T Plugin](p config.Plugin) (func(app App) (T, error), error) {
	pl := GetPlugins()[p.Name]
	return func(app App) (T, error) {
		var res T
		if err := pl.Initialize(app, p.Args); err != nil {
			return res, err
		}
		return pl.(T), nil
	}, nil
}
