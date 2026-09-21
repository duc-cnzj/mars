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
	"github.com/duc-cnzj/mars/v6/internal/util/slice"
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
	// AdminList 返回命名空间管理列表（管理员后台）：全量空间 + 活跃度分类/统计，
	// 支持活跃度分类过滤与内存分页。
	AdminList(ctx context.Context, input *AdminListInput) ([]*AdminNamespace, *AdminLivenessStats, *pagination.Pagination, error)
	// AdminDeletedList 返回已删除（软删）命名空间列表：Restore 的配套「选谁恢复」视图，
	// 供超管在误删恢复页按行选择。支持关键词（空间名/创建者邮箱）与真 SQL 分页，
	// 按删除时间倒序。每行 Projects 仅含随空间级联删除的那批（即 Restore 会一并恢复的项目）。
	// 权限门禁在 services 层（RequireSuperAdmin）——本方法自身不做鉴权。
	AdminDeletedList(ctx context.Context, input *NamespaceDeletedListInput) ([]*Namespace, *pagination.Pagination, error)
	// ListAllNames 返回全部 mars 管理命名空间的 k8s 名称（集群看板按此过滤排行/Top Pod）。
	ListAllNames(ctx context.Context) ([]string, error)
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
	// FavoriteSort 校验输入后把 firstID 移动到 secondID 位置。
	FavoriteSort(ctx context.Context, input *FavoriteSortNamespaceInput) error
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
	// Restore 恢复被误删的命名空间（仅超管，见 services 层 Authorize 的 RequireSuperAdmin 门禁）：
	// 按展示名定位软删行 → 校验无同名在册空间 → 重建 k8s 命名空间骨架（含 docker
	// secret）→ 清 deleted_at（连带同批次级联项目）→ 派发 EventNamespaceCreated 补注入
	// TLS 证书。**不重建**任何 helm release，恢复后项目需另行重新部署。
	Restore(ctx context.Context, name string) (*Namespace, error)
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

// Restore 实现 NamespaceBiz.Restore：见接口注释。
//
// 失败回滚与 Create 对齐：只有本次调用**真正创建**了 k8s namespace 时才在 DB 写失败后删掉它
// （createdByUs 守卫），收养来的不删——同一条「只有建的人才有权删」的规则，避免并发场景下删掉
// 另一个 Restore 正在复用的对象。DB 侧已落库（deleted_at 已清）之后不再回滚：那时再删 k8s 骨架
// 只会制造「DB 在册、集群没有」的错位。
//
// 仍保留的取舍：孤儿 secret 不回收——CreateDockerSecret 成功但 DB 回写失败时，k8s 会留下一个
// DB 不记录的 docker secret。清理它需要"按 namespace 列举 secret"的能力，而 K8sRepo 端口没有
// （为此扩端口属另一件事，且删错 secret 会波及该空间下全部工作负载）。▲ 该残留只在收养路径
// 出现：自建路径的 DB 失败已被上面的回滚连带清掉（删 namespace 会带走其下 secret）。
func (n *namespaceBiz) Restore(ctx context.Context, name string) (*Namespace, error) {
	deleted, err := n.nsRepo.FindDeletedByName(ctx, name)
	if err != nil {
		return nil, err
	}
	// 同名在册空间已存在时不得恢复：两行同名的"在册"空间会让后续按名定位（部署/成员/
	// 端点汇总）出现歧义，且 k8s 侧同名 namespace 也无法存在第二份。
	if live, err := n.nsRepo.FindByName(ctx, deleted.Name); err == nil {
		return nil, errs.WrapInvalidArgument(
			fmt.Errorf("空间 %s 已存在（id=%d），请先删除或重命名后再恢复", live.Name, live.ID),
			"restore namespace",
		)
	} else if !errs.IsNotFound(err) {
		// 只有 NotFound 才说明没有同名在册空间；真实 DB 故障必须上抛，不能当作"可用"放行。
		return nil, err
	}

	// 重建 k8s 骨架：删除空间时 k8s namespace 已被物理删除（连同其下 secret），只清
	// deleted_at 会让恢复后的空间无法部署（目标 namespace 不存在）。AlreadyExists 说明
	// k8s 侧仍在（重试/人工预先建好），收养它；Terminating 则拒绝对半成品操作。
	create, err := n.k8sRepo.CreateNamespace(ctx, deleted.Name)
	// 记录 namespace 是否由本次调用真正创建（而非收养已存在的），供 DB 写失败时决定是否回滚
	// ——与 Create 的 createdByUs 同一模式：只有建的人才有权删。
	createdByUs := err == nil
	if err != nil {
		if !k8sapierrors.IsAlreadyExists(err) {
			return nil, err
		}
		found, err := n.k8sRepo.GetNamespace(ctx, deleted.Name)
		if err != nil {
			return nil, err
		}
		if found.Status.Phase == v1.NamespaceTerminating {
			return nil, ErrNamespaceTerminating
		}
		create = found
	}
	n.logger.Debug("成功恢复namespace: ", create.Name)

	// docker secret 随 namespace 一起被物理删除，必须重建并回写 DB 记录；失败只降级
	// （与 Create 一致）：私有镜像 pull 会以拉取失败浮出，本条 Error 日志是排障锚点。
	var imagePullSecrets []string
	secret, err := n.k8sRepo.CreateDockerSecret(ctx, create.Name)
	if err == nil {
		imagePullSecrets = append(imagePullSecrets, secret.Name)
	} else {
		n.logger.ErrorCtx(ctx, fmt.Sprintf("恢复 namespace %s 的 docker secret 失败", create.Name), err)
	}

	// 回滚本次自建的 k8s 骨架（收养的不动）：只在 DB 侧尚未落库时调用——一旦 deleted_at 已清，
	// DB 就已在册，再删 k8s 骨架等于制造相反方向的错位。回滚失败只记日志，不掩盖原始错误。
	rollback := func() {
		if !createdByUs {
			return
		}
		if derr := n.k8sRepo.DeleteNamespace(ctx, create.Name); derr != nil {
			n.logger.ErrorCtx(ctx, "回滚恢复失败的 namespace 时删除失败: "+create.Name, derr)
		}
	}

	// 先补 DB 骨架再清软删标记：两者之间失败会留下"k8s 已建、DB 仍软删"的可重试中间态
	// （再次调用 Restore 命中 AlreadyExists 收养路径），反之则会留下不可恢复的孤儿空间。
	if err := n.nsRepo.UpdateImagePullSecrets(ctx, deleted.ID, imagePullSecrets); err != nil {
		rollback()
		return nil, err
	}
	if err := n.nsRepo.RestoreDeleted(ctx, deleted.ID); err != nil {
		rollback()
		return nil, err
	}

	ns, err := n.nsRepo.Show(ctx, deleted.ID)
	if err != nil {
		return nil, err
	}

	// 复用创建事件：HandleInjectTlsSecret 已注册在 EventNamespaceCreated 上，派发即
	// 重新注入 TLS 证书，无需为恢复单独写一份证书注入逻辑。
	n.eventRepo.Dispatch(EventNamespaceCreated, NamespaceCreatedData{
		NsModel:  ns,
		NsK8sObj: create,
	})

	return ns, nil
}

// List 分页列出 namespace（透传 repo）。
func (n *namespaceBiz) List(ctx context.Context, input *ListNamespaceInput) ([]*Namespace, *pagination.Pagination, error) {
	return n.nsRepo.List(ctx, input)
}

// ListAllNames 返回全部 mars 管理命名空间的 k8s 名称（委托 ListAll 抽名，供集群
// 看板把排行/Top Pod 收敛到 mars 自己管理的空间）。
func (n *namespaceBiz) ListAllNames(ctx context.Context) ([]string, error) {
	namespaces, err := n.nsRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return slice.Map(namespaces, func(ns *Namespace) string { return ns.Name }), nil
}

// AdminListInput 是命名空间管理列表输入：管理员后台的搜索/私有过滤 + 活跃度分类 + 分页。
type AdminListInput struct {
	Page, PageSize int32
	Search         string
	PrivateOnly    bool
	// Liveness 活跃度分类过滤：空 = 全部，否则 active/dormant/zombie。
	Liveness string
}

// NamespaceDeletedListInput 是已删除命名空间列表输入：恢复页的搜索 + 分页。
// 无 PrivateOnly/Liveness：前者在恢复场景无价值，后者对已删除空间无意义（见
// NamespaceDeletedListPageQuery 注释）。
type NamespaceDeletedListInput struct {
	Page, PageSize int32
	// Search 关键词：模糊匹配空间名或创建者邮箱。
	Search string
}

// AdminNamespace 是命名空间管理条目：空间模型 + 最近活跃时间（空间下所有项目
// UpdatedAt 最大值）+ 活跃度分类。
type AdminNamespace struct {
	Namespace    *Namespace
	LastActiveAt time.Time
	// LivenessKind 活跃度分类（复用项目活跃度阈值：≤30 天活跃 / ≥90 天僵尸；
	// 无项目即从未活跃，归为僵尸）。
	LivenessKind LivenessKind
}

// AdminLivenessStats 是命名空间活跃度统计（基于 search 命中全量，不随分页/分类过滤裁剪）。
type AdminLivenessStats struct {
	Total, Active, Dormant, Zombie int
}

// AdminListPageQuery 是命名空间管理分页查询输入：搜索/私有过滤/分类过滤/分页全部下沉 SQL。
// Now 为分类基准时间：SQL 边界（now-31d/now-90d）与 biz 行级标记共用同一基准，杜绝边界竞态。
type AdminListPageQuery struct {
	Search      string
	PrivateOnly bool
	Liveness    string
	Page        int32
	PageSize    int32
	Now         time.Time
}

// AdminListPageResult 是命名空间管理分页查询结果：已分页命名空间（含成员/项目/收藏边）+ 全量统计。
// Count 为分类过滤后总数（未分页），Stats 为 search 命中全量统计（不随过滤/分页裁剪）。
type AdminListPageResult struct {
	Namespaces []*Namespace
	Count      int
	Stats      AdminLivenessStats
}

// NamespaceDeletedListPageQuery 是已删除命名空间分页查询输入：搜索与分页下沉 SQL。
//
// 刻意不复用 AdminListPageQuery：那张输入里的 PrivateOnly（「只看私有」在恢复场景无价值）
// 与 Liveness/Now（活跃度分类对已删除空间无意义——空间都没了，谈不上活跃）在恢复语境下
// 全部失效，共用只会让调用方传一堆恒零字段。同理结果侧也不带 AdminLivenessStats。
type NamespaceDeletedListPageQuery struct {
	// Search 关键词：模糊匹配空间名或创建者邮箱，空串不过滤。
	Search string
	// Page/PageSize 分页参数（PageSize<=0 时由 pagination 兜默认值）。
	Page, PageSize int32
}

// NamespaceDeletedListPageResult 是已删除命名空间分页查询结果：已分页软删空间（含项目边）+
// 搜索命中总数。Count 为搜索过滤后总数（未分页），驱动前端分页器。
type NamespaceDeletedListPageResult struct {
	// Namespaces 本页软删空间，按删除时间倒序（最近删除的排最前）。
	// 每行的 Projects 只含 deleted_with_namespace=true 的「随空间级联删除」批，即 Restore
	// 会实际恢复的那批，故前端展示的项目数就是恢复后的项目数。
	Namespaces []*Namespace
	// Count 搜索命中总数（不分页）。
	Count int
}

// lastActiveAt 返回命名空间最近活跃时间：其下所有项目 UpdatedAt 的最大值。
// 无项目（从未部署/从未活跃）返回零值 time.Time，服务端序列化为空串，前端展示「从未活跃」。
func lastActiveAt(ns *Namespace) time.Time {
	var latest time.Time
	for _, p := range ns.Projects {
		if p != nil && p.UpdatedAt.After(latest) {
			latest = p.UpdatedAt
		}
	}
	return latest
}

// AdminList 返回命名空间管理列表：分类过滤/统计/分页由 repo 下沉 SQL（真分页，分类键为
// 「空间下项目 UpdatedAt 最大值」的跨表聚合，SQL 侧以 EXISTS 子查询等价表达），
// biz 仅按已加载的边计算最近活跃时间与行级活跃度分类。
func (n *namespaceBiz) AdminList(ctx context.Context, input *AdminListInput) ([]*AdminNamespace, *AdminLivenessStats, *pagination.Pagination, error) {
	now := time.Now()
	page, err := n.nsRepo.ListAdminPage(ctx, &AdminListPageQuery{
		Search:      input.Search,
		PrivateOnly: input.PrivateOnly,
		Liveness:    input.Liveness,
		Page:        input.Page,
		PageSize:    input.PageSize,
		Now:         now,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	items := make([]*AdminNamespace, 0, len(page.Namespaces))
	for _, ns := range page.Namespaces {
		activeAt := lastActiveAt(ns)
		items = append(items, &AdminNamespace{
			Namespace:    ns,
			LastActiveAt: activeAt,
			LivenessKind: ClassifyLiveness(activeAt, now),
		})
	}
	return items, &page.Stats, pagination.NewPagination(input.Page, input.PageSize, page.Count), nil
}

// AdminDeletedList 返回已删除命名空间列表（仅超管恢复流程使用）。搜索/分页/项目边装配全部
// 由 repo 下沉 SQL，biz 只补分页元信息——与 AdminList 的差别是不做行级活跃度分类（空间已
// 删除，活跃度无意义），故无需遍历行、也不返回统计。
func (n *namespaceBiz) AdminDeletedList(ctx context.Context, input *NamespaceDeletedListInput) ([]*Namespace, *pagination.Pagination, error) {
	page, err := n.nsRepo.ListAdminDeletedPage(ctx, &NamespaceDeletedListPageQuery{
		Search:   input.Search,
		Page:     input.Page,
		PageSize: input.PageSize,
	})
	if err != nil {
		return nil, nil, err
	}
	return page.Namespaces, pagination.NewPagination(input.Page, input.PageSize, page.Count), nil
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

// FavoriteSort 校验输入后把 firstID 移动到 secondID 位置。
func (n *namespaceBiz) FavoriteSort(ctx context.Context, input *FavoriteSortNamespaceInput) error {
	if input == nil || input.FirstID <= 0 || input.SecondID <= 0 {
		return errs.WrapInvalidArgument(errors.New("两个空间 id 必须大于 0"), "favorite sort")
	}
	return n.nsRepo.FavoriteSort(ctx, input.UserEmail, input.FirstID, input.SecondID)
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
	// ListAdminPage 分页列出管理员视角的命名空间（search/私有过滤 + 项目边装配）：分类过滤/
	// 统计/分页全部下沉 SQL（真分页，分类键为「空间下项目 UpdatedAt 最大值」，SQL 侧以
	// EXISTS 子查询等价表达），Now 为分类基准时间。
	ListAdminPage(ctx context.Context, query *AdminListPageQuery) (*AdminListPageResult, error)
	// ListAdminDeletedPage 分页列出已软删的命名空间（仅超管恢复流程使用）：必须绕过
	// SoftDeleteMixin 拦截器才能看到软删行，搜索/分页下沉 SQL。行内项目边只装配
	// deleted_with_namespace=true 的那批（Restore 的实际恢复范围），故前端项目数即恢复数。
	ListAdminDeletedPage(ctx context.Context, query *NamespaceDeletedListPageQuery) (*NamespaceDeletedListPageResult, error)
	// Create 创建命名空间。
	Create(ctx context.Context, input *CreateNamespaceInput) (*Namespace, error)
	// Show 按 id 查询命名空间。
	Show(ctx context.Context, id int) (*Namespace, error)
	// Update 更新命名空间描述。
	Update(ctx context.Context, input *UpdateNamespaceInput) (*Namespace, error)
	// Delete 删除命名空间。
	Delete(ctx context.Context, id int) error
	// FindDeletedByName 按名称查询已软删的命名空间（仅超管恢复流程使用）。
	FindDeletedByName(ctx context.Context, name string) (*Namespace, error)
	// RestoreDeleted 恢复软删的命名空间及其同批次级联删除的项目（清空 deleted_at）。
	RestoreDeleted(ctx context.Context, id int) error
	// GetMarsNamespace 返回 mars 保留命名空间名。
	GetMarsNamespace(name string) string
	// ListAll 返回全部 namespace（cron 同步 imagePullSecrets / TLS 证书需全量遍历）。
	ListAll(ctx context.Context) ([]*Namespace, error)
	// UpdateImagePullSecrets 更新 namespace 的 imagePullSecrets 列表。
	// 两个调用方：cron 对账后回写；以及 Restore 在清 deleted_at **之前**补写 DB 骨架——后者
	// 命中的是软删行，故实现刻意不带 deleted_at IS NULL 谓词。
	UpdateImagePullSecrets(ctx context.Context, id int, secrets []string) error
	// FindByName 按名称查询命名空间。
	FindByName(ctx context.Context, name string) (*Namespace, error)
	// Favorite 设置/取消收藏命名空间。
	Favorite(ctx context.Context, input *FavoriteNamespaceInput) error
	// FavoriteSort 把 firstID 关注空间移动到 secondID 位置。
	FavoriteSort(ctx context.Context, email string, firstID, secondID int) error
	// SyncMembers 同步命名空间成员列表。
	SyncMembers(ctx context.Context, namespaceID int, memberEmails []string) (*Namespace, error)
	// UpdatePrivate 更新命名空间私有状态。
	UpdatePrivate(ctx context.Context, namespaceID int, private bool) (*Namespace, error)
	// Transfer 转移命名空间所有权。
	Transfer(ctx context.Context, id int, email string) (*Namespace, error)
	// UpdateConfig 单事务原子更新命名空间配置（描述/私有/成员/转让管理员）。
	UpdateConfig(ctx context.Context, input *UpdateConfigInput) (*Namespace, error)
}
