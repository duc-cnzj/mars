package biz

import (
	"context"

	"github.com/duc-cnzj/mars/v6/internal/errs"
)

// 访问谓词领域化：加载实体 + 判定权限 + 映射领域错误的完整决策收进 biz，
// transport 只负责取当前用户与打日志。
//
// 全部访问判定收进 AccessBiz 方法，依赖（nsRepo/projBiz）由 receiver 持有，
// 当前用户直接走本包 MustGetUser（见 context.go），方法签名只留业务参数，
// 无任何顶层自由函数。历史上有三个自由函数因
// "跨包消费/纯谓词"暂留：CheckNsAccessByID 被 deploy 复用、CheckFileAccess 被
// httphandler 直接调用、NsCanAccess 被 services 侧 IsExists 等场景当纯谓词调——
// 三者的消费方现已统一改为持有/临时构造 AccessBiz 实例，自由函数全部清除，
// 判定逻辑收进本接口方法（命名随之从 Check*/NsCanAccess 对齐为 Require*/Can*）。
//
// 命名规约：报错门卫统一 Require* 前缀（必须有权，否则返回 errs.ErrorPermissionDenied）；
// 纯布尔谓词用 Can* 前缀（不报错，供"不可访问视同不存在"的静默场景）。
// 每个方法注释写明「门卫对象 + 放行规则 + 返回语义」，让调用方一眼看出守什么、
// 放谁。CanAccessNamespace 保留布尔语义：IsExists 场景需静默隐藏存在性，错误化
// 组合会把隐藏变成显式 403，暴露私有命名空间存在性（IDOR 侧信道）。
//
// 用户提取统一走本包 MustGetUser：访问判定仅服务已鉴权请求，ctx 必有用户，
// 缺用户即编程错误（panic），杜绝"拿不到用户被当作非管理员悄悄放行"。
// 免登录的 publicMethods 白名单方法不触达任何 AccessBiz 判定，panic 语义安全。
//
// 日志职责：本包不打印任何错误日志——判定失败（越权/加载失败）由消费方
// （services 层）拿到错误后用 logError 统一打印，遵守「错误日志只在 services
// 层打印」的项目规范，避免 biz 层双留痕。

// AccessBiz 是命名空间/项目/文件访问控制的业务层服务接口：供各 gRPC/HTTP 服务
// 在敏感传输方法开头调用，收口"加载实体 + 判定权限 + 错误映射"。日志由消费方
// （services 层）统一打印，本接口不产生任何日志副作用。
// 契约：每个敏感传输方法的开头都应以 AccessBiz 方法起步，避免漏写越权校验。
type AccessBiz interface {
	// RequireNamespaceAccessByName 是命名空间访问门卫：按 k8s 命名空间名加载并
	// 校验当前用户可访问性，可访问返回命名空间，否则 errs.ErrorPermissionDenied。
	// 放行规则：公开空间任意登录用户；私有空间仅 admin / 创建者 / 成员。
	// 供容器操作、指标查询等按名字定位命名空间的敏感方法开头调用。
	RequireNamespaceAccessByName(ctx context.Context, namespace string) (*Namespace, error)
	// RequireNamespaceAccessByID 是命名空间访问门卫：按命名空间 ID 加载并校验
	// 当前用户可访问性，可访问返回命名空间，否则 errs.ErrorPermissionDenied。
	// 放行规则同上（按 ID 定位）。deploy 的部署编排与 services 各命名空间级
	// 方法共用本实现。
	RequireNamespaceAccessByID(ctx context.Context, id int) (*Namespace, error)
	// RequireProjectAccess 是项目访问门卫：加载项目并校验当前用户对【项目所属
	// 命名空间】的可访问性，可访问返回项目，否则 errs.ErrorPermissionDenied。
	// 安全意义：私有命名空间的项目携带完整部署配置与环境变量，漏掉校验等于
	// 任意登录用户可枚举 ID 读取敏感数据。
	RequireProjectAccess(ctx context.Context, id int) (*Project, error)
	// RequireNamespaceOwner 是命名空间所有权门卫：校验当前用户是否为命名空间
	// owner，仅 owner 放行（admin 例外），否则 errs.ErrorPermissionDenied。
	// 只允许 owner 执行的变更（转让/改描述/删除/改私有/同步成员）统一走这里。
	RequireNamespaceOwner(ctx context.Context, ns *Namespace) error
	// RequireAdmin 是 admin 门禁：fullMethodName 精确命中 allowlist 时放行，
	// 否则要求当前用户为 admin。event/file/repo 三个服务的 Authorize 共用。
	// 命中统一精确匹配（不混用 Contains/EqualFold），防止豁免条件意外放行。
	RequireAdmin(ctx context.Context, fullMethodName string, allowlist ...string) (context.Context, error)
	// RequireFileAccess 是文件访问门卫：校验当前用户是否为文件所有者（Username
	// 匹配）或 admin，否则 errs.ErrorPermissionDenied。文件可能含部署配置/执行记录等
	// 敏感内容，只允许所有者或 admin 下载，防止枚举文件 ID 拖库。
	// 注意与 gRPC file 服务 admin 门禁（RequireAdmin）是两条独立规则：HTTP 下载
	// 放行所有者，gRPC 文件管理仅限 admin。
	RequireFileAccess(ctx context.Context, fil *File) error
	// CanAccessNamespace 是纯布尔谓词：判定当前用户能否访问命名空间
	// （admin/创建者/成员/公开空间放行），不映射错误。
	// 供 IsExists 等"不可访问视同不存在"的静默场景直接调用——错误化会把静默
	// 隐藏存在性变成显式 403，暴露私有命名空间存在性（IDOR 侧信道）。
	CanAccessNamespace(ctx context.Context, ns *Namespace) bool
}

// accessBiz 是 AccessBiz 的默认实现：持有实体加载 repo。
// 用户提取直接走本包 MustGetUser（原 auth 包已并入 biz，见 context.go）——
// 访问判定仅服务已鉴权请求，ctx 必有用户，不再需要传输层注入 getUser 回调。
// 本实现不持有 logger：判定失败的错误日志由消费方（services 层）统一打印。
type accessBiz struct {
	nsRepo  NamespaceBiz
	projBiz ProjectBiz
}

// NewAccessBiz 构造访问控制服务：nsRepo/projBiz 提供实体加载。
// nsRepo/projBiz 可为 nil 懒加载——仅当对应方法被调用时才会用到，例如
// file/repo/event 三个服务只调用 RequireAdmin（不触达实体加载）时传 nil 即可。
// 判定失败的错误日志由消费方（services 层）用 logError 统一打印，本构造器
// 不再接收 logger（日志职责已上移，避免 biz 层打印）。
func NewAccessBiz(nsRepo NamespaceBiz, projBiz ProjectBiz) AccessBiz {
	return &accessBiz{nsRepo: nsRepo, projBiz: projBiz}
}

// RequireNamespaceAccessByName 是命名空间访问门卫：按 k8s 命名空间名加载并校验
// 当前用户可访问性，可访问返回命名空间，否则 errs.ErrorPermissionDenied。
// 实现约定：加载失败返回原始错误；判定失败返回 errs.ErrorPermissionDenied。
// 两类错误的日志均由消费方（services 层）统一打印。
func (a *accessBiz) RequireNamespaceAccessByName(ctx context.Context, namespace string) (*Namespace, error) {
	ns, err := a.nsRepo.FindByName(ctx, namespace)
	if err != nil {
		return nil, err
	}
	if !a.CanAccessNamespace(ctx, ns) {
		return nil, errs.ErrorPermissionDenied
	}
	return ns, nil
}

// RequireNamespaceAccessByID 是命名空间访问门卫：按命名空间 ID 加载并校验当前
// 用户可访问性，可访问返回命名空间，否则 errs.ErrorPermissionDenied。
// 判定逻辑收进本方法后成为共享访问路径：deploy 与 services
// （metrics/endpoint/namespace）四个调用方共用同一实现。
func (a *accessBiz) RequireNamespaceAccessByID(ctx context.Context, id int) (*Namespace, error) {
	ns, err := a.nsRepo.Show(ctx, id)
	if err != nil {
		return nil, err
	}
	if !a.CanAccessNamespace(ctx, ns) {
		return nil, errs.ErrorPermissionDenied
	}
	return ns, nil
}

// RequireProjectAccess 是项目访问门卫：加载项目并校验当前用户对【项目所属
// 命名空间】的可访问性，可访问返回项目，否则 errs.ErrorPermissionDenied。
// 实现约定：项目加载失败返回原始错误；命名空间级校验委托 RequireNamespaceAccessByID。
// 错误日志由消费方（services 层）统一打印。
func (a *accessBiz) RequireProjectAccess(ctx context.Context, id int) (*Project, error) {
	proj, err := a.projBiz.Show(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, nserr := a.RequireNamespaceAccessByID(ctx, proj.NamespaceID); nserr != nil {
		return nil, nserr
	}
	return proj, nil
}

// RequireNamespaceOwner 是命名空间所有权门卫：校验当前用户是否为命名空间 owner，
// 仅 owner 放行（admin 例外），否则 errs.ErrorPermissionDenied。
// 判定失败（含 ns 为 nil）同样拒绝——无法判定所有权时安全默认是拒绝；
// 越权拒绝的错误日志由消费方（services 层）统一打印。
func (a *accessBiz) RequireNamespaceOwner(ctx context.Context, ns *Namespace) error {
	user := MustGetUser(ctx)
	if ns == nil || (!user.IsAdmin() && ns.CreatorEmail != user.Email) {
		return errs.ErrorPermissionDenied
	}
	return nil
}

// RequireAdmin 是 admin 门禁：fullMethodName 精确命中 allowlist 时放行，否则
// 要求当前用户为 admin。event/file/repo 三个服务的 Authorize 共用同一判据，
// 命中统一为精确匹配（不混用 Contains/EqualFold），防止豁免条件意外放行。
func (a *accessBiz) RequireAdmin(ctx context.Context, fullMethodName string, allowlist ...string) (context.Context, error) {
	for _, name := range allowlist {
		if fullMethodName == name {
			return ctx, nil
		}
	}
	if !MustGetUser(ctx).IsAdmin() {
		return nil, errs.ErrorPermissionDenied
	}
	return ctx, nil
}

// RequireFileAccess 是文件访问门卫：校验当前用户是否为文件所有者（Username
// 匹配）或 admin，否则 errs.ErrorPermissionDenied。
// 原先的自由函数由 httphandler 直接调用，现 httphandler 已注入 accessBiz，
// 判定逻辑收进本方法。
func (a *accessBiz) RequireFileAccess(ctx context.Context, fil *File) error {
	user := MustGetUser(ctx)
	if fil.Username == user.Name || user.IsAdmin() {
		return nil
	}
	return errs.ErrorPermissionDenied
}

// CanAccessNamespace 是纯布尔谓词：判定当前用户能否访问命名空间
// （admin/创建者/成员/公开空间放行），不映射错误。user 由 MustGetUser 从
// 上下文提取。供 IsExists 等"不可访问视同不存在"的静默场景直接调用——错误化
// 组合会把静默隐藏存在性变成显式 403，暴露私有命名空间存在性（IDOR 侧信道）。
func (a *accessBiz) CanAccessNamespace(ctx context.Context, ns *Namespace) bool {
	if ns == nil {
		return false
	}
	user := MustGetUser(ctx)
	if user.IsAdmin() {
		return true
	}
	if !ns.Private {
		return true
	}
	if ns.CreatorEmail == user.Email {
		return true
	}
	for _, m := range ns.Members {
		if m.Email == user.Email {
			return true
		}
	}
	return false
}
