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

type PluginManger interface {
	Load(App) error
	Domain() DomainManager
	Ws() WsSender
	Git() GitServer
	Picture() Picture
	Destroy()

	GetPlugins() map[string]Plugin
}

var _ PluginManger = (*manager)(nil)

type newFunc[T Plugin] func(App) (T, error)

type manager struct {
	domainFunc newFunc[DomainManager]
	wsFunc     newFunc[WsSender]
	gitFunc    newFunc[GitServer]
	picFunc    newFunc[Picture]

	domain DomainManager
	ws     WsSender
	git    GitServer
	pic    Picture

	logger      mlog.Logger
	destroyFunc []func()
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

func (m *manager) Destroy() {
	for _, f := range m.destroyFunc {
		f()
	}
}

func NewPluginManager(cfg *config.Config, logger mlog.Logger) (PluginManger, func(), error) {
	var destroyFunc []func()
	domain, f, err := GetPlugin[DomainManager](cfg.DomainManagerPlugin)
	if err != nil {
		return nil, nil, err
	}
	destroyFunc = append(destroyFunc, f)

	ws, f, err := GetPlugin[WsSender](cfg.WsSenderPlugin)
	if err != nil {
		return nil, nil, err
	}
	destroyFunc = append(destroyFunc, f)

	git, f, err := GetPlugin[GitServer](cfg.GitServerPlugin)
	if err != nil {
		return nil, nil, err
	}
	destroyFunc = append(destroyFunc, f)

	pic, f, err := GetPlugin[Picture](cfg.PicturePlugin)
	if err != nil {
		return nil, nil, err
	}
	destroyFunc = append(destroyFunc, f)

	ma := &manager{
		logger:      logger,
		domainFunc:  domain,
		wsFunc:      ws,
		gitFunc:     git,
		picFunc:     pic,
		destroyFunc: destroyFunc,
	}
	return ma, func() { ma.Destroy() }, nil
}

func GetPlugin[T Plugin](p config.Plugin) (func(app App) (T, error), func(), error) {
	pl := GetPlugins()[p.Name]
	var inited bool
	return func(app App) (T, error) {
			var res T
			if err := pl.Initialize(app, p.Args); err != nil {
				return res, err
			}
			inited = true
			return pl.(T), nil
		}, func() {
			if inited {
				pl.Destroy()
			}
		}, nil
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

type EmptyPubSub struct{}

func (e *EmptyPubSub) Join(projectID int64) error {
	return nil
}

func (e *EmptyPubSub) Leave(nsID int64, projectID int64) error {
	return nil
}

func (e *EmptyPubSub) Run(ctx context.Context) error {
	return nil
}

func (e *EmptyPubSub) Publish(nsID int64, pod *corev1.Pod) error {
	return nil
}

func (e *EmptyPubSub) Info() any {
	return nil
}

func (e *EmptyPubSub) Uid() string {
	return ""
}

func (e *EmptyPubSub) ID() string {
	return ""
}

func (e *EmptyPubSub) ToSelf(message WebsocketMessage) error {
	return nil
}

func (e *EmptyPubSub) ToAll(message WebsocketMessage) error {
	return nil
}

func (e *EmptyPubSub) ToOthers(message WebsocketMessage) error {
	return nil
}

func (e *EmptyPubSub) Subscribe() <-chan []byte {
	return nil
}

func (e *EmptyPubSub) Close() error {
	return nil
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

var (
	ListCommitsCacheSeconds       int = 10
	AllBranchesCacheSeconds       int = 60 * 2
	AllProjectsCacheSeconds       int = 60 * 5
	GetFileContentCacheSeconds    int = 0
	GetDirectoryFilesCacheSeconds int = 0

	GetCommitCacheSeconds int = 60 * 60
)

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

type GitCacheServer interface {
	ReCacheAllProjects() error
	ReCacheAllBranches(pid string) error
}

//
//// gitServerCache
//// 用来缓存一些耗时比较久的请求
//type gitServerCache struct {
//	s     GitServer
//	cache cache.Cache
//}
//
//func (g *gitServerCache) Name() string {
//	return ""
//}
//
//func (g *gitServerCache) Initialize(args map[string]any) error {
//	return nil
//}
//
//func (g *gitServerCache) Destroy() error {
//	return nil
//}
//
//func newGitServerCache(s GitServer) *gitServerCache {
//	return &gitServerCache{s: s}
//}
//
//func (g *gitServerCache) GetProject(pid string) (Project, error) {
//	return g.s.GetProject(pid)
//}
//
//func (g *gitServerCache) ListProjects(page, pageSize int) (ListProjectResponse, error) {
//	return g.s.ListProjects(page, pageSize)
//}
//
//func CacheKeyAllProjects() cache.CacheKey {
//	return cache.NewKey("AllProjects")
//}
//
//func (g *gitServerCache) ReCacheAllProjects() error {
//	data, err := g.allProjectsWithoutCache()
//	if err != nil {
//		return err
//	}
//
//	return g.cache.SetWithTTL(CacheKeyAllProjects(), data, AllProjectsCacheSeconds)
//}
//
//func (g *gitServerCache) AllProjects() ([]Project, error) {
//	remember, err := g.cache.Remember(CacheKeyAllProjects(), AllProjectsCacheSeconds, g.allProjectsWithoutCache)
//	if err != nil {
//		return nil, err
//	}
//	var res []*project
//	json.Unmarshal(remember, &res)
//	var all = make([]Project, 0, len(res))
//	for _, re := range res {
//		all = append(all, re)
//	}
//	return all, nil
//}
//
//func (g *gitServerCache) allProjectsWithoutCache() ([]byte, error) {
//	projects, err := g.s.AllProjects()
//	if err != nil {
//		return nil, err
//	}
//	var all = make([]Project, 0, len(projects))
//	for _, projectInterface := range projects {
//		all = append(all, &project{
//			ID:            projectInterface.GetID(),
//			Name:          projectInterface.GetName(),
//			DefaultBranch: projectInterface.GetDefaultBranch(),
//			Path:          projectInterface.GetPath(),
//			WebUrl:        projectInterface.GetWebURL(),
//			AvatarUrl:     projectInterface.GetAvatarURL(),
//			Description:   projectInterface.GetDescription(),
//		})
//	}
//	marshal, _ := json.Marshal(all)
//	return marshal, nil
//}
//
//func (g *gitServerCache) ListBranches(pid string, page, pageSize int) (ListBranchResponse, error) {
//	return g.s.ListBranches(pid, page, pageSize)
//}
//
//func CacheKeyAllBranches[T ~string | ~int | ~int64](pid T) cache.CacheKey {
//	return cache.NewKey("AllBranches-%v", pid)
//}
//
//func (g *gitServerCache) ReCacheAllBranches(pid string) error {
//	data, err := g.allBranchWithoutCache(pid)()
//	if err != nil {
//		return err
//	}
//	return g.cache.SetWithTTL(CacheKeyAllBranches(pid), data, AllBranchesCacheSeconds)
//}
//
//func (g *gitServerCache) AllBranches(pid string) ([]Branch, error) {
//	remember, err := g.cache.Remember(CacheKeyAllBranches(pid), AllBranchesCacheSeconds, g.allBranchWithoutCache(pid))
//	if err != nil {
//		return nil, err
//	}
//	var res []*branch
//	json.Unmarshal(remember, &res)
//	// Why? 为什么我不能直接返回 res，奇怪的 go 语法
//	var all = make([]Branch, 0, len(res))
//	for _, b := range res {
//		all = append(all, b)
//	}
//	return all, nil
//}
//
//func (g *gitServerCache) allBranchWithoutCache(pid string) func() ([]byte, error) {
//	return func() ([]byte, error) {
//		b, err := g.s.AllBranches(pid)
//		if err != nil {
//			return nil, err
//		}
//		var all = make([]Branch, 0, len(b))
//		for _, branchInterface := range b {
//			all = append(all, &branch{
//				Name:    branchInterface.GetName(),
//				Default: branchInterface.IsDefault(),
//				WebUrl:  branchInterface.GetWebURL(),
//			})
//		}
//
//		marshal, _ := json.Marshal(all)
//		return marshal, nil
//	}
//}
//
//func (g *gitServerCache) GetCommit(pid string, sha string) (Commit, error) {
//	remember, err := g.cache.Remember(cache.NewKey("GetCommit:%s-%s", pid, sha), GetCommitCacheSeconds, func() ([]byte, error) {
//		c, err := g.s.GetCommit(pid, sha)
//		if err != nil {
//			return nil, err
//		}
//		result := &commit{
//			ID:             c.GetID(),
//			ShortID:        c.GetShortID(),
//			Title:          c.GetTitle(),
//			CommittedDate:  c.GetCommittedDate(),
//			AuthorName:     c.GetAuthorName(),
//			AuthorEmail:    c.GetAuthorEmail(),
//			CommitterName:  c.GetCommitterName(),
//			CommitterEmail: c.GetCommitterEmail(),
//			CreatedAt:      c.GetCreatedAt(),
//			Message:        c.GetMessage(),
//			ProjectID:      c.GetProjectID(),
//			WebURL:         c.GetWebURL(),
//		}
//		marshal, _ := json.Marshal(result)
//		return marshal, nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	msg := &commit{}
//	_ = json.Unmarshal(remember, msg)
//	return msg, nil
//}
//
//func (g *gitServerCache) GetCommitPipeline(pid string, branch string, sha string) (Pipeline, error) {
//	return g.s.GetCommitPipeline(pid, branch, sha)
//}
//
//func (g *gitServerCache) ListCommits(pid string, branch string) ([]Commit, error) {
//	remember, err := g.cache.Remember(cache.NewKey("ListCommits:%s-%s", pid, branch), ListCommitsCacheSeconds, func() ([]byte, error) {
//		commits, err := g.s.ListCommits(pid, branch)
//		if err != nil {
//			return nil, err
//		}
//		var result = make([]Commit, 0, len(commits))
//		for _, commitInterface := range commits {
//			result = append(result, &commit{
//				ID:             commitInterface.GetID(),
//				ShortID:        commitInterface.GetShortID(),
//				Title:          commitInterface.GetTitle(),
//				CommittedDate:  commitInterface.GetCommittedDate(),
//				AuthorName:     commitInterface.GetAuthorName(),
//				AuthorEmail:    commitInterface.GetAuthorEmail(),
//				CommitterName:  commitInterface.GetCommitterName(),
//				CommitterEmail: commitInterface.GetCommitterEmail(),
//				CreatedAt:      commitInterface.GetCreatedAt(),
//				Message:        commitInterface.GetMessage(),
//				ProjectID:      commitInterface.GetProjectID(),
//				WebURL:         commitInterface.GetWebURL(),
//			})
//		}
//		marshal, _ := json.Marshal(result)
//		return marshal, nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	var res []*commit
//	json.Unmarshal(remember, &res)
//	var lists = make([]Commit, 0, len(res))
//
//	for _, re := range res {
//		lists = append(lists, re)
//	}
//	return lists, nil
//}
//
//func (g *gitServerCache) GetFileContentWithBranch(pid string, branch string, filename string) (string, error) {
//	remember, err := g.cache.Remember(cache.NewKey("GetFileContentWithBranch-%s-%s-%s", pid, branch, filename), GetFileContentCacheSeconds, func() ([]byte, error) {
//		content, err := g.s.GetFileContentWithBranch(pid, branch, filename)
//		if err != nil {
//			return nil, err
//		}
//		return []byte(content), nil
//	})
//	if err != nil {
//		return "", err
//	}
//	return string(remember), nil
//}
//
//func (g *gitServerCache) GetFileContentWithSha(pid string, sha string, filename string) (string, error) {
//	remember, err := g.cache.Remember(cache.NewKey("GetFileContentWithSha-%s-%s-%s", pid, sha, filename), GetFileContentCacheSeconds, func() ([]byte, error) {
//		content, err := g.s.GetFileContentWithSha(pid, sha, filename)
//		if err != nil {
//			return nil, err
//		}
//		return []byte(content), nil
//	})
//	if err != nil {
//		return "", err
//	}
//	return string(remember), nil
//}
//
//func (g *gitServerCache) GetDirectoryFilesWithBranch(pid string, branch string, path string, recursive bool) ([]string, error) {
//	remember, err := g.cache.Remember(cache.NewKey("GetDirectoryFilesWithBranch-%s-%s-%s-%v", pid, branch, path, recursive), GetDirectoryFilesCacheSeconds, func() ([]byte, error) {
//		withBranch, err := g.s.GetDirectoryFilesWithBranch(pid, branch, path, recursive)
//		if err != nil {
//			return nil, err
//		}
//		marshal, _ := json.Marshal(withBranch)
//		return marshal, nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	var res []string
//	json.Unmarshal(remember, &res)
//	return res, nil
//}
//
//func (g *gitServerCache) GetDirectoryFilesWithSha(pid string, sha string, path string, recursive bool) ([]string, error) {
//	remember, err := g.cache.Remember(cache.NewKey("GetDirectoryFilesWithSha-%s-%s-%s-%v", pid, sha, path, recursive), GetDirectoryFilesCacheSeconds, func() ([]byte, error) {
//		withBranch, err := g.s.GetDirectoryFilesWithSha(pid, sha, path, recursive)
//		if err != nil {
//			return nil, err
//		}
//		marshal, _ := json.Marshal(withBranch)
//		return marshal, nil
//	})
//	if err != nil {
//		return nil, err
//	}
//	var res []string
//	json.Unmarshal(remember, &res)
//	return res, nil
//}

type project struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	Path          string `json:"path"`
	WebUrl        string `json:"web_url"`
	AvatarUrl     string `json:"avatar_url"`
	Description   string `json:"description"`
}

func (p *project) GetID() int64 {
	return p.ID
}

func (p *project) GetName() string {
	return p.Name
}

func (p *project) GetDefaultBranch() string {
	return p.DefaultBranch
}

func (p *project) GetPath() string {
	return p.Path
}

func (p *project) GetWebURL() string {
	return p.WebUrl
}

func (p *project) GetAvatarURL() string {
	return p.AvatarUrl
}

func (p *project) GetDescription() string {
	return p.Description
}

type branch struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
	WebUrl  string `json:"web_url"`
}

func (b *branch) GetName() string {
	return b.Name
}

func (b *branch) IsDefault() bool {
	return b.Default
}

func (b *branch) GetWebURL() string {
	return b.WebUrl
}

type commit struct {
	ID             string     `json:"id"`
	ShortID        string     `json:"short_id"`
	Title          string     `json:"title"`
	CommittedDate  *time.Time `json:"committed_date"`
	AuthorName     string     `json:"author_name"`
	AuthorEmail    string     `json:"author_email"`
	CommitterName  string     `json:"committer_name"`
	CommitterEmail string     `json:"committer_email"`
	CreatedAt      *time.Time `json:"created_at"`
	Message        string     `json:"message"`
	ProjectID      int64      `json:"project_id"`
	WebURL         string     `json:"web_url"`
}

func (c *commit) GetID() string {
	return c.ID
}

func (c *commit) GetShortID() string {
	return c.ShortID
}

func (c *commit) GetTitle() string {
	return c.Title
}

func (c *commit) GetCommittedDate() *time.Time {
	return c.CommittedDate
}

func (c *commit) GetAuthorName() string {
	return c.AuthorName
}

func (c *commit) GetAuthorEmail() string {
	return c.AuthorEmail
}

func (c *commit) GetCommitterName() string {
	return c.CommitterName
}

func (c *commit) GetCommitterEmail() string {
	return c.CommitterEmail
}

func (c *commit) GetCreatedAt() *time.Time {
	return c.CreatedAt
}

func (c *commit) GetMessage() string {
	return c.Message
}

func (c *commit) GetProjectID() int64 {
	return c.ProjectID
}

func (c *commit) GetWebURL() string {
	return c.WebURL
}
