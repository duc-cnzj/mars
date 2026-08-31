package biz

import (
	"encoding/json"
	"io"
	"sort"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/duc-cnzj/mars/api/v6/proto/mars"
	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
)

// UploadType 是上传类型分类的别名。
type UploadType = schematype.UploadType

const (
	// LocalUpload 标识文件存储在本地磁盘。
	LocalUpload UploadType = schematype.Local
	// S3Upload 标识文件存储在 S3 对象存储。
	S3Upload UploadType = schematype.S3
)

// ---------- AccessToken ----------

// AccessToken 是访问令牌的领域模型。
type AccessToken struct {
	ID         int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	Token      string
	Usage      string
	Email      string
	ExpiredAt  time.Time
	LastUsedAt *time.Time
	UserInfo   UserInfo
}

// IsExpired 判断 token 在 now 时刻是否已过期（now 晚于 ExpiredAt 视为过期）。
// 时间由调用方注入：biz 侧传注入时钟（a.timer.Now()，测试可 mock），
// transformer 侧传系统时钟（展示快照语义）。边界相等视为未过期。
func (at *AccessToken) IsExpired(now time.Time) bool {
	return now.After(at.ExpiredAt)
}

// ListAccessTokenInput 是 access token 分页列表输入。
type ListAccessTokenInput struct {
	Page, PageSize int32
	WithSoftDelete bool
	Email          string
	Search         string
	// Status 状态过滤：''=不过滤；valid=未撤销且未过期；expired=未撤销但已过期；revoked=已撤销（软删除）。
	Status string
}

// GrantAccessTokenInput 是签发 access token 的输入。
type GrantAccessTokenInput struct {
	ExpireSeconds int32
	Usage         string
	User          *UserInfo
}

// ---------- Changelog ----------

// Changelog 是项目一次部署的变更记录领域模型。
type Changelog struct {
	ID               int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	Version          int
	Username         string
	Config           string
	GitBranch        string
	GitCommit        string
	DockerImage      []string
	EnvValues        []*types.KeyValue
	ExtraValues      []*websocket_pb.ExtraValue
	FinalExtraValues []*websocket_pb.ExtraValue
	GitCommitWebURL  string
	GitCommitTitle   string
	GitCommitAuthor  string
	GitCommitDate    *time.Time
	ConfigChanged    bool
	ProjectID        int

	Project *Project
}

// CreateChangeLogInput 是创建变更记录的输入。
type CreateChangeLogInput struct {
	Username         string
	Version          int
	Config           string
	GitBranch        string
	GitCommit        string
	DockerImage      []string
	EnvValues        []*types.KeyValue
	ExtraValues      []*websocket_pb.ExtraValue
	FinalExtraValues []*websocket_pb.ExtraValue
	GitCommitWebURL  string
	GitCommitTitle   string
	GitCommitAuthor  string
	GitCommitDate    *time.Time
	ConfigChanged    bool
	ProjectID        int
}

// FindLastChangelogsByProjectIDChangeLogInput 是查询项目最近变更记录的输入。
type FindLastChangelogsByProjectIDChangeLogInput struct {
	OnlyChanged        bool
	ProjectID          int
	OrderByVersionDesc *bool
	Limit              int
}

// ---------- YamlPrettier ----------

// YamlPrettier 提供 PrettyYaml 以 YAML 文本形式呈现变更内容。
type YamlPrettier interface {
	// PrettyYaml 把变更内容序列化为 YAML 文本快照。
	PrettyYaml() string
}

// EventKey 是事件派发标识的原生 biz 类型，解耦 biz 与 internal/event 基础设施包。
type EventKey string

// String 返回事件键的字符串形式。
func (e EventKey) String() string { return string(e) }

const (
	// AuditLogEvent 是审计日志写入事件。
	AuditLogEvent EventKey = "audit_log"
	// EventNamespaceCreated 是命名空间创建事件。
	EventNamespaceCreated EventKey = "namespace_created"
	// EventNamespaceDeleted 是命名空间删除事件。
	EventNamespaceDeleted EventKey = "namespace_deleted"
	// EventProjectChanged 是项目变更事件。
	EventProjectChanged EventKey = "project_changed"
	// EventProjectDeleted 是项目删除事件。
	EventProjectDeleted EventKey = "project_deleted"
)

// ---------- Event ----------

// Event 是事件审计记录领域模型。
type Event struct {
	ID            int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	Action        types.EventActionType
	Username      string
	OperatorEmail string
	Message       string
	Old           string
	New           string
	Duration      string
	FileID        *int
	HasDiff       bool

	File *File
}

// ListEventInput 是事件分页列表输入。
// ActionTypes 是动作类型过滤的生效值集合（IN 匹配）：空 = 全部，多值 = 任一匹配；
// 由 services 层从 action_type（单值，Unknown=全部）与 action_types（多值）归一化而来。
// OperatorEmail 按操作人邮箱等值过滤：nil = 不过滤（admin 全量），非 nil = 只看该邮箱的事件。
type ListEventInput struct {
	Page, PageSize int32
	ActionTypes    []types.EventActionType
	Search         string
	OrderIDDesc    *bool
	OperatorEmail  *string
}

// NamespaceCreatedData 是 namespace 创建事件载荷：DB 模型 + k8s 对象。
type NamespaceCreatedData struct {
	NsModel  *Namespace
	NsK8sObj *corev1.Namespace
}

// NamespaceDeletedData 是 namespace 删除事件载荷。
type NamespaceDeletedData struct {
	ID int
}

// ProjectChangedData 是项目变更事件载荷。
type ProjectChangedData struct {
	ID       int
	Username string
}

// ProjectDeletedPayload 是项目删除事件载荷。
type ProjectDeletedPayload struct {
	NamespaceID int
	ProjectID   int
}

// ---------- File ----------

// File 是文件上传/拷贝记录的领域模型。
type File struct {
	ID            int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	UploadType    UploadType
	Path          string
	Size          uint64
	Username      string
	Namespace     string
	Pod           string
	Container     string
	ContainerPath string

	HumanizeSize string
}

// ListFileInput 是文件分页列表输入。
type ListFileInput struct {
	Page, PageSize int32
	OrderIDDesc    *bool
	WithSoftDelete bool
}

// CreateFileInput 是创建文件记录的输入。
type CreateFileInput struct {
	Path       string
	Username   string
	Size       uint64
	UploadType UploadType

	Namespace string
	Pod       string
	Container string
}

// UpdateFileRequest 是更新文件记录的输入。
type UpdateFileRequest struct {
	ID            int
	ContainerPath string
	Namespace     string
	Pod           string
	Container     string
}

// StreamUploadFileRequest 是流式上传文件用例的输入。
type StreamUploadFileRequest struct {
	Namespace     string
	Pod           string
	Container     string
	ContainerPath string
	Username      string
	FileName      string
	FileData      chan []byte
}

// ---------- Namespace ----------

// Favorite 是用户收藏 namespace 的关系模型。
type Favorite struct {
	ID          int
	NamespaceID int
	Email       string
}

// Member 是 namespace 成员关系模型。
type Member struct {
	ID          int
	NamespaceID int
	Email       string
}

// Namespace 是命名空间领域模型。
type Namespace struct {
	ID               int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	Name             string
	ImagePullSecrets []string
	Description      string
	Private          bool
	CreatorEmail     string

	Projects  []*Project
	Favorites []*Favorite
	Members   []*Member
}

// GetImagePullSecrets 把镜像拉取密钥名列表转换为 proto 类型列表。
func (ns *Namespace) GetImagePullSecrets() []*types.ImagePullSecret {
	var secrets = make([]*types.ImagePullSecret, 0)
	for _, s := range ns.ImagePullSecrets {
		secrets = append(secrets, &types.ImagePullSecret{Name: s})
	}
	return secrets
}

// ListNamespaceInput 是 namespace 分页列表输入。
// Search/PrivateOnly 是管理员后台管理列表的过滤维度，其余场景不传即不过滤。
type ListNamespaceInput struct {
	Page     int32
	PageSize int32
	Favorite bool
	Email    string
	Name     *string
	IsAdmin  bool
	// Search 关键词模糊匹配空间名/创建者邮箱。
	Search string
	// PrivateOnly 只看私有空间。
	PrivateOnly bool
}

// CreateNamespaceInput 是创建 namespace 的输入。
type CreateNamespaceInput struct {
	Name             string
	ImagePullSecrets []string
	Description      string
	CreatorEmail     string
}

// UpdateNamespaceInput 是更新 namespace 的输入。
type UpdateNamespaceInput struct {
	ID          int
	Description string
}

// UpdateConfigInput 是一次性原子更新命名空间配置（描述/私有/成员/转让管理员）的输入。
// 指针字段（Description/Private）区分"未传"与零值；Emails 为 nil 表示不更新成员，
// 非 nil（含空切片）表示以该名单全量同步；NewAdminEmail 空串表示不转让。
type UpdateConfigInput struct {
	ID            int
	Description   *string
	Private       *bool
	Emails        []string
	NewAdminEmail string
}

// FavoriteNamespaceInput 是设置/取消收藏的输入。
type FavoriteNamespaceInput struct {
	NamespaceID int
	UserEmail   string
	Favorite    bool
}

// FavoriteSortNamespaceInput 是移动关注列表排序的输入。
// FirstID 为被移动的空间，SecondID 为移动目标位置的参照空间（FirstID 移到 SecondID 位置）。
type FavoriteSortNamespaceInput struct {
	UserEmail string
	FirstID   int
	SecondID  int
}

// ---------- Project ----------

// Project 是项目领域模型。
type Project struct {
	ID               int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	Name             string
	GitProjectID     int
	GitBranch        string
	GitCommit        string
	Config           string
	OverrideValues   string
	DockerImage      []string
	PodSelectors     []string
	Atomic           bool
	DeployStatus     types.Deploy
	EnvValues        []*types.KeyValue
	ExtraValues      []*websocket_pb.ExtraValue
	FinalExtraValues []*websocket_pb.ExtraValue
	Version          int
	ConfigType       string
	GitCommitWebURL  string
	GitCommitTitle   string
	GitCommitAuthor  string
	GitCommitDate    *time.Time
	NamespaceID      int
	RepoID           int
	Elements         []*types.KeyValue
	Namespace        *Namespace
	Repo             *Repo
	Manifest         []string
}

// ToEventYaml 把项目关键字段排好序后转成 YAML 快照，供审计事件对比变更。
// nil receiver 返回 nil（与既有调用语义一致，测试承重）。
func (p *Project) ToEventYaml() YamlPrettier {
	if p == nil {
		return nil
	}
	sort.Slice(p.EnvValues, func(i, j int) bool {
		return p.EnvValues[i].Key < p.EnvValues[j].Key
	})
	sort.Slice(p.ExtraValues, func(i, j int) bool {
		return p.ExtraValues[i].Path < p.ExtraValues[j].Path
	})
	sort.Slice(p.FinalExtraValues, func(i, j int) bool {
		return p.FinalExtraValues[i].Path < p.FinalExtraValues[j].Path
	})

	return AnyYamlPrettier{
		"title":              p.GitCommitTitle,
		"branch":             p.GitBranch,
		"commit":             p.GitCommit,
		"atomic":             p.Atomic,
		"web_url":            p.GitCommitWebURL,
		"config":             p.Config,
		"env_values":         p.EnvValues,
		"extra_values":       p.ExtraValues,
		"final_extra_values": p.FinalExtraValues,
	}
}

// ListProjectInput 是项目分页列表输入。
type ListProjectInput struct {
	Page, PageSize     int32
	OrderByIDDesc      *bool
	OrderByVersionDesc *bool
	Name               string
	NamespaceID        int32
	RepoID             int
	// Email + IsAdmin 用于 data 层按命名空间访问谓词过滤列表：
	// 非 admin 只能看到其可访问命名空间（公开/创建者/成员）下的项目。
	Email   string
	IsAdmin bool
}

// CreateProjectInput 是创建项目的输入。
type CreateProjectInput struct {
	Name         string
	GitProjectID int
	GitBranch    string
	GitCommit    string
	Config       string
	ExtraValues  []*websocket_pb.ExtraValue
	Atomic       *bool
	ConfigType   string
	NamespaceID  int
	PodSelectors []string
	DeployStatus types.Deploy
	RepoID       int
	Creator      string
}

// UpdateProjectInput 是更新项目的输入。
type UpdateProjectInput struct {
	ID               int
	GitBranch        string
	GitCommit        string
	Config           string
	Atomic           *bool
	ConfigType       string
	PodSelectors     []string
	DockerImage      []string
	GitCommitTitle   string
	GitCommitWebURL  string
	GitCommitAuthor  string
	GitCommitDate    *time.Time
	ExtraValues      []*websocket_pb.ExtraValue
	FinalExtraValues []*websocket_pb.ExtraValue
	EnvValues        []*types.KeyValue
	OverrideValues   string
	Manifest         []string
}

// StatePod 是项目下单个 pod 的展示状态。
type StatePod struct {
	IsOld       bool
	Terminating bool
	Pending     bool
	OrderIndex  int
	Pod         *corev1.Pod
}

// ---------- Repo ----------

// Repo 是仓库领域模型。
type Repo struct {
	ID             int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	Name           string
	DefaultBranch  string
	GitProjectName string
	GitProjectID   int32
	Enabled        bool
	NeedGitRepo    bool
	MarsConfig     *mars.Config `json:"mars_config"`
	Description    string

	Projects []*Project
}

// GetMarsConfig 返回仓库的 mars 配置，nil 时回退为空配置。
func (r *Repo) GetMarsConfig() (cfg *mars.Config) {
	cfg = r.MarsConfig
	if r.MarsConfig == nil {
		cfg = &mars.Config{}
	}
	return
}

// AllRepoRequest 是全量仓库查询输入。
type AllRepoRequest struct {
	NeedGitRepo   *bool
	Enabled       *bool
	OrderByIDDesc *bool
}

// ListRepoRequest 是仓库分页列表输入。
type ListRepoRequest struct {
	Page, PageSize int32
	Enabled        *bool
	Name           string
	OrderByIDDesc  *bool
}

// CreateRepoInput 是创建仓库的输入。
type CreateRepoInput struct {
	Name         string
	Enabled      bool
	NeedGitRepo  bool
	GitProjectID *int32
	MarsConfig   *mars.Config
	Description  string
}

// UpdateRepoInput 是更新仓库的输入。
type UpdateRepoInput struct {
	ID           int32
	Name         string
	NeedGitRepo  bool
	GitProjectID *int32
	MarsConfig   *mars.Config
	Description  string
}

// CloneRepoInput 是克隆仓库的输入。
type CloneRepoInput struct {
	ID   int
	Name string
}

// ImportRepoItem 是导入仓库的单条数据：字段与导出模型对齐，
// 只携带可落库字段（id/时间戳/git 派生信息由服务端按 name 重新生成）。
type ImportRepoItem struct {
	Name         string
	Enabled      bool
	NeedGitRepo  bool
	GitProjectID *int32
	MarsConfig   *mars.Config
	Description  string
}

// ---------- Git ----------

// Status 是流水线/集群状态的字符串别名。
type Status = string

const (
	// StatusUnknown 表示流水线/集群状态未知。
	StatusUnknown Status = "unknown"
	// StatusSuccess 表示流水线/集群状态成功。
	StatusSuccess Status = "success"
	// StatusFailed 表示流水线/集群状态失败。
	StatusFailed Status = "failed"
	// StatusRunning 表示流水线/集群状态进行中。
	StatusRunning Status = "running"
	// StatusManual 表示流水线/集群存在手动触发的 job，等待人工确认。
	StatusManual Status = "manual"
)

// PictureItem 是首页图片信息。
type PictureItem struct {
	Url       string
	Copyright string
}

// OidcConfigItem 是单个 OIDC provider 的配置：provider、oauth2 配置与登出端点。
type OidcConfigItem struct {
	Provider           *oidc.Provider
	Config             oauth2.Config
	EndSessionEndpoint string
}

// OidcConfig 是 OIDC provider 名称到配置项的映射。
type OidcConfig map[string]OidcConfigItem

// Branch 是 git 分支信息。
type Branch struct {
	Name      string
	IsDefault bool
	WebURL    string
}

// GitProject 是 git 项目信息。
type GitProject struct {
	ID            int64
	Name          string
	DefaultBranch string
	WebURL        string
	Path          string
	AvatarURL     string
	Description   string
}

// Commit 是 git 提交信息。
type Commit struct {
	ID             string
	ShortID        string
	AuthorName     string
	AuthorEmail    string
	CommitterName  string
	CommitterEmail string
	Message        string
	Title          string
	WebURL         string
	CreatedAt      *time.Time
	CommittedDate  *time.Time
}

// PipelineJob 是 CI/CD 流水线单个 job 的名称、状态与所属 stage。
type PipelineJob struct {
	Name      string
	Status    Status
	StageName string
}

// Pipeline 是 git CI/CD 流水线信息。
type Pipeline struct {
	ID        int64
	ProjectID int64
	Status    Status
	Ref       string
	SHA       string
	WebURL    string
	UpdatedAt *time.Time
	CreatedAt *time.Time
	// Jobs 是流水线各 job 的名称与状态，按执行顺序排列。
	Jobs []PipelineJob
}

// ---------- K8s ----------

// ClusterStatus 是集群健康状态的字符串别名。
type ClusterStatus = string

const (
	// StatusBad 表示集群健康状态不健康。
	StatusBad ClusterStatus = "bad"
	// StatusNotGood 表示集群健康状态欠佳。
	StatusNotGood ClusterStatus = "not good"
	// StatusHealth 表示集群健康状态健康。
	StatusHealth ClusterStatus = "health"
)

// Container 标识容器终端操作的命名空间/pod/容器三元组。
type Container struct {
	Namespace string
	Pod       string
	Container string
}

// DockerConfig 是 registry 名称到登录凭据的映射。
type DockerConfig map[string]DockerConfigEntry

// DockerConfigEntry 描述单个 registry 的登录凭据。json tag 严格对齐
// Docker config.json 的字段名（小写），kubelet 解析依赖此格式——
// 缺 tag 会序列化成大写键导致镜像拉取认证失效（曾为 P0 回归）。
type DockerConfigEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth,omitempty"`
}

// DockerConfigJSON 对应 k8s DockerConfigJson 类型 Secret 的负载结构。
type DockerConfigJSON struct {
	Auths       DockerConfig      `json:"auths"`
	HttpHeaders map[string]string `json:"HttpHeaders,omitempty"`
}

// DecodeDockerConfigJSON 将 Docker config.json 字节流解析为 DockerConfigJSON，
// 供 k8s docker secret 对账（SyncImagePullSecrets）读取已有凭据。
func DecodeDockerConfigJSON(data []byte) (res DockerConfigJSON, err error) {
	err = json.Unmarshal(data, &res)
	return
}

// PodEventType 描述 Pod 生命周期事件的类别，与 data 层 informer fanout 的
// Add/Update/Delete 一一对应，作为跨层事件透传的领域枚举（不泄漏 data 内部类型）。
type PodEventType int

const (
	// PodEventAdd 表示 Pod 新增事件。
	PodEventAdd PodEventType = iota
	// PodEventUpdate 表示 Pod 更新事件。
	PodEventUpdate
	// PodEventDelete 表示 Pod 删除事件。
	PodEventDelete
)

// PodEvent 是 Pod 生命周期事件的结构化载荷，由 data 层订阅 informer 后转换而来。
// Add/Delete 事件仅 Current 有效；Update 事件携带 Old/Current 供状态变更比对。
type PodEvent struct {
	Type    PodEventType
	Old     *corev1.Pod
	Current *corev1.Pod
}

// ClusterInfo 是集群健康与资源用量汇总信息。
type ClusterInfo struct {
	Status            ClusterStatus
	FreeMemory        string
	FreeCpu           string
	FreeRequestMemory string
	FreeRequestCpu    string
	TotalMemory       string
	TotalCpu          string
	UsageMemoryRate   string
	UsageCpuRate      string
	RequestMemoryRate string
	RequestCpuRate    string
}

// CopyFromPodInput 是从 pod 拷贝文件用例的输入。
type CopyFromPodInput struct {
	Namespace string
	Pod       string
	Container string
	FilePath  string
	UserName  string
}

// CopyFileToPodInput 是把文件拷贝进 pod 用例的输入。
type CopyFileToPodInput struct {
	FileId    int64
	Namespace string
	Pod       string
	Container string
}

// ExecuteInput 是容器内执行命令的输入：stdin/stdout/stderr 接线由调用方提供。
type ExecuteInput struct {
	Stdin             io.Reader
	Stdout, Stderr    io.Writer
	TTY               bool
	Cmd               []string
	TerminalSizeQueue TerminalSizeQueue
}

// TerminalSize 是终端窗口尺寸的值对象。
type TerminalSize struct {
	Width, Height uint16
}

// TerminalSizeQueue 是终端尺寸序列端口：Next 返回下一次终端尺寸，流结束或 ctx 取消时返回
// nil。由 biz（execSizeQueue）与 transport（ptyHandler）实现，data 层在 k8s exec 出口
// 适配为 client-go 的 remotecommand.TerminalSizeQueue，隔离基础设施类型。
type TerminalSizeQueue interface {
	// Next 返回下一次终端尺寸，流结束或 ctx 取消时返回 nil。
	Next() *TerminalSize
}

// ExecExitError 是容器内命令以非零退出码结束时的领域错误，携带退出码与输出消息。
// 由 data 层在 k8s exec 出口把 client-go 的退出码错误翻译而来，隔离基础设施类型。
type ExecExitError struct {
	Code    int
	Message string
}

// Error 返回退出错误消息。
func (e *ExecExitError) Error() string { return e.Message }

// LogFn 是部署过程的日志回调函数。
type LogFn func(format string, v ...any)

// WrapLogFn 是携带容器列表上下文的日志回调函数。
type WrapLogFn func(container []*websocket_pb.Container, format string, v ...any)

// UnWrap 返回不带容器上下文的日志函数（容器列表传 nil）。
func (l WrapLogFn) UnWrap() func(format string, v ...any) {
	return func(format string, v ...any) {
		l(nil, format, v...)
	}
}
