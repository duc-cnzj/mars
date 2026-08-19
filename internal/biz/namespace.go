package biz

// namespace.go 定义命名空间生命周期用例层（NamespaceBiz）：
// Create/Delete 的编排（GetMarsNamespace 预查、k8s 建删、docker secret 降级、
// 并发卸载 release、轮询确认删除、事件派发）从 services 传输层下沉到 biz。
// 协议映射收敛到 internal/errs：Terminating 已是携带 AlreadyExists status 的领域错误
// （ErrNamespaceTerminating），transport 直接透传。transport 只保留鉴权
// （RequireNamespaceOwner）、幂等策略（IgnoreIfExists 放行/exists 拒绝，
// 状态码由 errs.AlreadyExists 提供）与审计日志（AuditLogWithRequest 需携带
// proto request，留在传输层）。

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	v1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
)

// ErrNamespaceTerminating 是 Create 遇 k8s namespace 处于 Terminating 时的领域错误：
// 已携带 gRPC AlreadyExists status（errs.AlreadyExists 构造），transport 直接透传
// 即可让调用方拿到 409，无需在传输层再映射状态码。
var ErrNamespaceTerminating = errs.AlreadyExists("该名称空间正在删除中")

// NamespaceDeleteTimeout / NamespacePollInterval 是 Delete 等待 k8s namespace 真正删除的
// 轮询参数：从 services 包级私有 var 提升为导出，供 services 测试（package services 下
// 无法触碰 biz 私有符号）覆盖为快速值，避免为覆盖 timer/ticker 分支支付真实 5s 墙钟。
var (
	NamespaceDeleteTimeout = 5 * time.Second
	NamespacePollInterval  = 500 * time.Millisecond
)

// NamespaceBiz 是 namespace 域唯一业务接口：CRUD 数据门面（校验 + 透传 repo）与
// 生命周期用例编排（Create/Delete 走 k8s/helm/event）合一。透传版 Create/Delete
// 已由编排版覆盖，接口收敛为 9 个查询/变更门面 + 2 个编排方法。
type NamespaceBiz interface {
	// List 分页列出 namespace。
	List(ctx context.Context, input *ListNamespaceInput) ([]*Namespace, *pagination.Pagination, error)
	// Show 按 id 查询 namespace。
	Show(ctx context.Context, id int) (*Namespace, error)
	// Update 校验输入后更新 namespace。
	Update(ctx context.Context, input *UpdateNamespaceInput) (*Namespace, error)
	// FindByName 按名称查询 namespace。
	FindByName(ctx context.Context, name string) (*Namespace, error)
	// GetMarsNamespace 返回 mars 保留命名空间名。
	GetMarsNamespace(name string) string
	// Favorite 校验输入后设置/取消收藏。
	Favorite(ctx context.Context, input *FavoriteNamespaceInput) error
	// SyncMembers 校验 id 后同步成员列表。
	SyncMembers(ctx context.Context, namespaceID int, memberEmails []string) (*Namespace, error)
	// UpdatePrivate 校验 id 后更新私有状态。
	UpdatePrivate(ctx context.Context, namespaceID int, private bool) (*Namespace, error)
	// Transfer 校验 id 与 email 后转移所有权。
	Transfer(ctx context.Context, id int, email string) (*Namespace, error)
	// UpdateConfig 校验 id 后单事务原子更新命名空间配置（描述/私有/成员/转让管理员）。
	UpdateConfig(ctx context.Context, input *UpdateConfigInput) (*Namespace, error)
	// Create 创建命名空间：GetMarsNamespace → FindByName 预查 → k8s 建（含收养已存在）
	// → docker secret（失败降级）→ DB 记录（本请求新建的失败回滚）→ 自动关注 → 派发
	// EventNamespaceCreated。返回 exists=true 表示命名空间已存在（未建新记录），
	// 调用方按 IgnoreIfExists 策略决定放行或返回 AlreadyExists。
	Create(ctx context.Context, namespace, description, creatorEmail string) (*Namespace, bool, error)
	// Delete 删除命名空间：并发卸载全部 helm release → 删 secrets → 删 k8s ns（非
	// NotFound 中止）→ 删 DB → 轮询确认删除 → 派发 EventNamespaceDeleted。
	// 返回被删除的项目名列表（供审计日志）。ns 为调用方已鉴权（RequireNamespaceOwner）的 Show 结果。
	Delete(ctx context.Context, ns *Namespace) ([]string, error)
}

var _ NamespaceBiz = (*namespaceBiz)(nil)

// namespaceBiz 是 NamespaceBiz 的生产实现：以命名空间/项目/k8s/helm/事件仓库编排创建
// 与删除，并持有 logger 记录非致命路径（docker secret 降级、轮询瞬态错误、回滚失败）。
type namespaceBiz struct {
	logger     mlog.Logger
	nsRepo     NamespaceRepo
	k8sRepo    K8sRepo
	helmerRepo HelmerRepo
	eventRepo  EventRepo
}

// NewNamespaceBiz 构造命名空间用例实现，依赖由 wire 注入。
func NewNamespaceBiz(logger mlog.Logger, nsRepo NamespaceRepo, k8sRepo K8sRepo, helmerRepo HelmerRepo, eventRepo EventRepo) NamespaceBiz {
	return &namespaceBiz{
		logger:     logger.WithModule("biz/namespace"),
		nsRepo:     nsRepo,
		k8sRepo:    k8sRepo,
		helmerRepo: helmerRepo,
		eventRepo:  eventRepo,
	}
}

// Create 实现 NamespaceBiz.Create：见接口注释。
func (n *namespaceBiz) Create(ctx context.Context, namespace, description, creatorEmail string) (*Namespace, bool, error) {
	nsName := n.nsRepo.GetMarsNamespace(namespace)
	preCheckNs, err := n.nsRepo.FindByName(ctx, nsName)
	if err != nil {
		// 只有 NotFound 才说明名称空间可创建；其余错误（如 DB 故障）必须上抛，
		// 不能误判为"不存在"后继续走 k8s 创建流程。
		if !errs.IsNotFound(err) {
			return nil, false, err
		}
	} else {
		// 已存在：不建新记录，原样返回交给调用方按 IgnoreIfExists 策略决策。
		return preCheckNs, true, nil
	}

	create, err := n.k8sRepo.CreateNamespace(ctx, nsName)
	// 记录 namespace 是否由本请求真正创建（而非收养已存在的），
	// 用于 DB 记录失败时回滚清理，避免 k8s 里留下孤儿资源。
	createdByUs := err == nil
	if err != nil {
		if !k8sapierrors.IsAlreadyExists(err) {
			return nil, false, err
		}
		found, err := n.k8sRepo.GetNamespace(ctx, nsName)
		if err != nil {
			return nil, false, err
		}
		if found.Status.Phase == v1.NamespaceTerminating {
			return nil, false, ErrNamespaceTerminating
		}
		// 收养已存在的 k8s namespace，用 found 而不是失败的 create（后者是零值，Name 为空）
		create = found
	}
	n.logger.Debug("成功创建namespace: ", create.Name)

	var imagePullSecrets []string
	secret, err := n.k8sRepo.CreateDockerSecret(ctx, create.Name)
	if err == nil {
		imagePullSecrets = append(imagePullSecrets, secret.Name)
	} else {
		// CreateDockerSecret 失败只可能是 k8s API 错误（RBAC/网络/配额），
		// 属于真实基建问题——namespace 创建继续（降级），但必须 Error 级可见，
		// 否则后续私有镜像 pull 失败会以"不透明的拉取失败"浮出，排障无抓手。
		n.logger.ErrorCtx(ctx, fmt.Sprintf("创建 namespace %s 的 docker secret 失败", create.Name), err)
	}

	ns, err := n.nsRepo.Create(ctx, &CreateNamespaceInput{
		Name:             create.Name,
		ImagePullSecrets: imagePullSecrets,
		Description:      description,
		CreatorEmail:     creatorEmail,
	})
	if err != nil {
		// DB 记录创建失败，回滚本次刚创建的 k8s namespace（收养的不删），
		// 避免 k8s 里留下无 DB 记录的孤儿 namespace。
		if createdByUs {
			if derr := n.k8sRepo.DeleteNamespace(ctx, create.Name); derr != nil {
				n.logger.ErrorCtx(ctx, "删除刚创建的 namespace 失败: "+create.Name, derr)
			}
		}
		return nil, false, err
	}
	if err := n.nsRepo.Favorite(ctx, &FavoriteNamespaceInput{
		NamespaceID: ns.ID,
		UserEmail:   creatorEmail,
		Favorite:    true,
	}); err != nil {
		// 创建成功但自动关注失败：namespace 已可用，记录错误但不阻断创建结果。
		n.logger.ErrorCtx(ctx, "创建 namespace 后自动关注失败: "+ns.Name, err)
	}

	n.eventRepo.Dispatch(EventNamespaceCreated, NamespaceCreatedData{
		NsModel:  ns,
		NsK8sObj: create,
	})

	return ns, false, nil
}

// Delete 实现 NamespaceBiz.Delete：见接口注释。
func (n *namespaceBiz) Delete(ctx context.Context, ns *Namespace) ([]string, error) {
	var deletedProjectNames []string
	wg := sync.WaitGroup{}
	wg.Add(len(ns.Projects))
	for _, project := range ns.Projects {
		deletedProjectNames = append(deletedProjectNames, project.Name)
		go func(releaseName, namespace string) {
			defer wg.Done()
			defer n.logger.HandlePanic("namespaceBiz.Delete")
			n.logger.Debugf("delete release %s namespace %s", releaseName, namespace)
			if err := n.helmerRepo.Uninstall(releaseName, namespace, n.logger.Debugf); err != nil {
				n.logger.ErrorCtx(ctx, fmt.Sprintf("卸载 release %s 于 namespace %s 失败", releaseName, namespace), err)
				return
			}
		}(project.Name, ns.Name)
	}
	wg.Wait()
	for _, secret := range ns.ImagePullSecrets {
		n.logger.DebugCtxf(ctx, "delete ns %s secret %s", ns.Name, secret)
		if err := n.k8sRepo.DeleteSecret(ctx, ns.Name, secret); err != nil {
			n.logger.ErrorCtx(ctx, "删除 namespace secret 出现错误: ", err)
		}
	}
	if err := n.k8sRepo.DeleteNamespace(ctx, ns.Name); err != nil && !k8sapierrors.IsNotFound(err) {
		// F18 同类：k8s namespace 未真正删除（非 NotFound）时不得继续删 DB 记录，
		// 否则会留下孤儿 namespace，且下方轮询超时后会误发 NamespaceDeleted 事件。
		// NotFound 视为"已删除干净"，正常继续。
		return nil, err
	}
	if err := n.nsRepo.Delete(ctx, ns.ID); err != nil {
		return nil, err
	}

	timer := time.NewTimer(NamespaceDeleteTimeout)
	defer timer.Stop()

	ticker := time.NewTicker(NamespacePollInterval)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			if _, err := n.k8sRepo.GetNamespace(ctx, ns.Name); err != nil {
				// 只有 NotFound 才说明 namespace 已删除。瞬态错误（API 抖动/鉴权/网络）
				// 状态未知，不能当作已删除提前 break——否则会误发 NamespaceDeleted 事件，
				// 记录后继续轮询，直到超时兜底。
				if k8sapierrors.IsNotFound(err) {
					n.logger.Debug(err)
					break loop
				}
				n.logger.ErrorCtx(ctx, "等待 namespace 删除时查询失败: "+ns.Name, err)
			}
		case <-timer.C:
			break loop
		}
	}

	n.eventRepo.Dispatch(EventNamespaceDeleted, NamespaceDeletedData{ID: ns.ID})

	return deletedProjectNames, nil
}

// List 分页列出 namespace（透传 repo）。
func (n *namespaceBiz) List(ctx context.Context, input *ListNamespaceInput) ([]*Namespace, *pagination.Pagination, error) {
	return n.nsRepo.List(ctx, input)
}

// Show 按 id 查询 namespace（透传 repo）。
func (n *namespaceBiz) Show(ctx context.Context, id int) (*Namespace, error) {
	return n.nsRepo.Show(ctx, id)
}

// Update 校验输入后更新 namespace。
func (n *namespaceBiz) Update(ctx context.Context, input *UpdateNamespaceInput) (*Namespace, error) {
	if input == nil || input.ID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("namespace 不能为空或 id 不能小于等于 0"), "update namespace")
	}
	return n.nsRepo.Update(ctx, input)
}

// GetMarsNamespace 获取 mars 命名空间名（透传 repo）。
func (n *namespaceBiz) GetMarsNamespace(name string) string {
	return n.nsRepo.GetMarsNamespace(name)
}

// FindByName 按名称查询 namespace（透传 repo）。
func (n *namespaceBiz) FindByName(ctx context.Context, name string) (*Namespace, error) {
	return n.nsRepo.FindByName(ctx, name)
}

// Favorite 校验输入后设置/取消收藏。
func (n *namespaceBiz) Favorite(ctx context.Context, input *FavoriteNamespaceInput) error {
	if input == nil || input.NamespaceID <= 0 {
		return errs.WrapInvalidArgument(errors.New("namespace 不能为空或 id 不能小于等于 0"), "favorite namespace")
	}
	return n.nsRepo.Favorite(ctx, input)
}

// SyncMembers 校验 namespace id 后同步成员。
func (n *namespaceBiz) SyncMembers(ctx context.Context, namespaceID int, memberEmails []string) (*Namespace, error) {
	if namespaceID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("namespace id 不能小于等于 0"), "sync members")
	}
	return n.nsRepo.SyncMembers(ctx, namespaceID, memberEmails)
}

// UpdatePrivate 校验 namespace id 后更新私有状态。
func (n *namespaceBiz) UpdatePrivate(ctx context.Context, namespaceID int, private bool) (*Namespace, error) {
	if namespaceID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("namespace id 不能小于等于 0"), "update private")
	}
	return n.nsRepo.UpdatePrivate(ctx, namespaceID, private)
}

// UpdateConfig 校验 id 后单事务原子更新 namespace 配置（描述/私有/成员/转让管理员）。
func (n *namespaceBiz) UpdateConfig(ctx context.Context, input *UpdateConfigInput) (*Namespace, error) {
	if input.ID <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("namespace id 不能小于等于 0"), "update config")
	}
	return n.nsRepo.UpdateConfig(ctx, input)
}

// Transfer 校验 id 与 email 后转移 namespace 所有权。
func (n *namespaceBiz) Transfer(ctx context.Context, id int, email string) (*Namespace, error) {
	if id <= 0 {
		return nil, errs.WrapInvalidArgument(errors.New("namespace id 不能小于等于 0"), "transfer namespace")
	}
	if email == "" {
		return nil, errs.WrapInvalidArgument(errors.New("transfer email 不能为空"), "transfer namespace")
	}
	return n.nsRepo.Transfer(ctx, id, email)
}

// NamespaceRepo 是命名空间仓库端口。
type NamespaceRepo interface {
	// List 分页列出命名空间（可按收藏/名称/访问权限过滤）。
	List(ctx context.Context, input *ListNamespaceInput) ([]*Namespace, *pagination.Pagination, error)
	// Create 创建命名空间。
	Create(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error)
	// Show 按 id 查询命名空间。
	Show(ctx context.Context, id int) (*Namespace, error)
	// Update 更新命名空间描述。
	Update(ctx context.Context, input *UpdateNamespaceInput) (*Namespace, error)
	// Delete 删除命名空间。
	Delete(ctx context.Context, id int) error
	// GetMarsNamespace 返回 mars 保留命名空间名。
	GetMarsNamespace(name string) string
	// ListAll 返回全部 namespace（cron 同步 imagePullSecrets / TLS 证书需全量遍历）。
	ListAll(ctx context.Context) ([]*Namespace, error)
	// UpdateImagePullSecrets 更新 namespace 的 imagePullSecrets 列表（cron 对账后回写）。
	UpdateImagePullSecrets(ctx context.Context, id int, secrets []string) error
	// FindByName 按名称查询命名空间。
	FindByName(ctx context.Context, name string) (*Namespace, error)
	// Favorite 设置/取消收藏命名空间。
	Favorite(ctx context.Context, input *FavoriteNamespaceInput) error
	// SyncMembers 同步命名空间成员列表。
	SyncMembers(ctx context.Context, namespaceID int, memberEmails []string) (*Namespace, error)
	// UpdatePrivate 更新命名空间私有状态。
	UpdatePrivate(ctx context.Context, namespaceID int, private bool) (*Namespace, error)
	// Transfer 转移命名空间所有权。
	Transfer(ctx context.Context, id int, email string) (*Namespace, error)
	// UpdateConfig 单事务原子更新命名空间配置（描述/私有/成员/转让管理员）。
	UpdateConfig(ctx context.Context, input *UpdateConfigInput) (*Namespace, error)
}
