package biz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/duc-cnzj/mars/v6/internal/errs"
)

// 访问谓词领域化：加载实体 + 判定权限 + 映射领域错误的完整决策收进 biz，
// transport 只负责取当前用户与打日志。
//
// 全部访问判定收进 AccessBiz 方法，依赖（nsRepo/projBiz）由 receiver 持有，
// 当前用户直接走本包 MustGetUser（见 context.go），方法签名只留业务参数，
// 无任何顶层自由函数：判定逻辑全部收进本接口方法。
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
	// AI Agent 自助例外：mars-agent 空间刻意私有（仅管理员可见），但普通用户需要
	// 查看/操作自己的沙箱项目，故当前用户邮箱派生的 agent 项目额外放行。
	RequireProjectAccess(ctx context.Context, id int) (*Project, error)
	// RequirePodAccess 是 Pod 级访问门卫：按命名空间名与 Pod 名校验当前用户
	// 可访问性，可访问返回命名空间，否则 errs.ErrorPermissionDenied。
	// 放行规则：先按命名空间 ACL（公开空间任意登录用户；私有空间 admin/创建者/成员）；
	// 再叠加 AI Agent 自助例外——mars-agent 空间内只放行 Pod 名前缀等于当前用户
	// 派生 Deployment 名的 Pod（即用户自己的沙箱），其他用户的 Pod 一律拒绝。
	// 供容器日志/终端/文件拷贝/强制删除等以 Pod 为操作对象的敏感方法开头调用；
	// 命名空间维度的操作（如列出整个空间的事件）不得使用本方法。
	RequirePodAccess(ctx context.Context, namespace, pod string) (*Namespace, error)
	// RequireNamespaceOwner 是命名空间所有权门卫：校验当前用户是否为命名空间
	// owner，仅 owner 放行（admin 例外），否则 errs.ErrorPermissionDenied。
	// 只允许 owner 执行的变更（转让/改描述/删除/改私有/同步成员）统一走这里。
	RequireNamespaceOwner(ctx context.Context, ns *Namespace) error
	// RequireAdmin 是 admin 门禁：fullMethodName 精确命中 allowlist 时放行，
	// 否则要求当前用户为 admin。event/file/repo 三个服务的 Authorize 共用。
	// 命中统一精确匹配（不混用 Contains/EqualFold），防止豁免条件意外放行。
	RequireAdmin(ctx context.Context, fullMethodName string, allowlist ...string) (context.Context, error)
	// RequireSuperAdmin 是方法级超管门禁：fullMethodName 精确命中 superOnly 时要求
	// 当前用户为内置超级管理员，否则 errs.ErrorPermissionDenied；未命中放行。
	// 与 RequireAdmin 的 allowlist 极性相反（那里是豁免名单，这里是强制名单）——
	// 供"比 admin 更严"的方法（如 Restore）在 Authorize 中收紧阈值。
	RequireSuperAdmin(ctx context.Context, fullMethodName string, superOnly ...string) error
	// RequireFileAccess 是文件访问门卫：校验当前用户是否为文件所有者（Username
	// 匹配）或 admin，否则 errs.ErrorPermissionDenied。文件可能含部署配置/执行记录等
	// 敏感内容，只允许所有者或 admin 下载，防止枚举文件 ID 拖库。
	// 消费方：HTTP 下载、gRPC ShowRecords（回放会话）、gRPC CopyToPod（按 FileId
	// 拷入 pod——不校验归属则任意登录用户可传他人 FileId 读走，构成跨租户 IDOR）；
	// 其余 gRPC 文件管理（列表/磁盘信息/删除）仍走 Authorize 的 admin 门禁（RequireAdmin）。
	RequireFileAccess(ctx context.Context, fil *File) error
	// CanAccessNamespace 是纯布尔谓词：判定当前用户能否访问命名空间
	// （admin/创建者/成员/公开空间放行），不映射错误。
	// 供 IsExists 等"不可访问视同不存在"的静默场景直接调用——错误化会把静默
	// 隐藏存在性变成显式 403，暴露私有命名空间存在性（IDOR 侧信道）。
	CanAccessNamespace(ctx context.Context, ns *Namespace) bool
}

// accessBiz 是 AccessBiz 的默认实现：持有实体加载 repo。
// 用户提取直接走本包 MustGetUser（见 context.go）——访问判定仅服务已鉴权请求，
// ctx 必有用户。
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
		return nil, errs.WrapPermissionDenied(errs.ErrorPermissionDenied, fmt.Sprintf("访问命名空间 %s", namespace))
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
		return nil, errs.WrapPermissionDenied(errs.ErrorPermissionDenied, fmt.Sprintf("访问命名空间 #%d", id))
	}
	return ns, nil
}

// RequireProjectAccess 是项目访问门卫：加载项目并校验当前用户对【项目所属
// 命名空间】的可访问性，可访问返回项目，否则 errs.ErrorPermissionDenied。
// 实现约定：项目加载失败返回原始错误；命名空间级校验委托 RequireNamespaceAccessByID；
// 命名空间拒绝后再判定 AI Agent 自助例外（当前用户自己的 agent 项目）。
// 错误日志由消费方（services 层）统一打印。
func (a *accessBiz) RequireProjectAccess(ctx context.Context, id int) (*Project, error) {
	proj, err := a.projBiz.Show(ctx, id)
	if err != nil {
		return nil, err
	}
	_, nserr := a.RequireNamespaceAccessByID(ctx, proj.NamespaceID)
	if nserr == nil {
		return proj, nil
	}
	// 仅当底层确实是权限拒绝时才追加项目上下文；命名空间加载失败（DB/网络）
	// 等不确定错误原样上抛，不得被误分类为 PermissionDenied。
	if !errors.Is(nserr, errs.ErrorPermissionDenied) {
		return nil, nserr
	}
	own, ownErr := a.isOwnAgentProject(ctx, proj)
	if ownErr != nil {
		return nil, ownErr
	}
	if !own {
		return nil, errs.WrapPermissionDenied(nserr, fmt.Sprintf("访问项目 #%d", id))
	}
	return proj, nil
}

// RequirePodAccess 是 Pod 级访问门卫：按命名空间名加载命名空间后，先走通用
// 命名空间 ACL，拒绝时再判定 AI Agent 自助例外（自己的沙箱 Pod）。
// 实现约定：命名空间加载失败返回原始错误；无权访问返回 errs.ErrorPermissionDenied。
// 错误日志由消费方（services 层）统一打印。
func (a *accessBiz) RequirePodAccess(ctx context.Context, namespace, pod string) (*Namespace, error) {
	ns, err := a.nsRepo.FindByName(ctx, namespace)
	if err != nil {
		return nil, err
	}
	if a.CanAccessNamespace(ctx, ns) {
		return ns, nil
	}
	own, ownErr := a.isOwnAgentPod(ctx, ns, pod)
	if ownErr != nil {
		return nil, ownErr
	}
	if !own {
		return nil, errs.WrapPermissionDenied(errs.ErrorPermissionDenied,
			fmt.Sprintf("访问命名空间 %s 的 Pod %s", namespace, pod))
	}
	return ns, nil
}

// isOwnAgentProject 判定项目是否属于当前用户的自助 AI 沙箱：项目位于
// mars-agent 专用空间，且项目名等于当前用户邮箱派生的 AgentProjectName。
// 邮箱为空时恒为 false——空邮箱无法唯一标识用户，放行会共享同一个沙箱项目。
func (a *accessBiz) isOwnAgentProject(ctx context.Context, proj *Project) (bool, error) {
	email := MustGetUser(ctx).Email
	if proj == nil || email == "" {
		return false, nil
	}
	inAgentNS, err := a.isAgentNamespaceID(ctx, proj.NamespaceID)
	if err != nil {
		return false, err
	}
	return inAgentNS && proj.Name == AgentProjectName(email), nil
}

// isOwnAgentPod 判定 Pod 是否属于当前用户的自助 AI 沙箱：命名空间必须是
// mars-agent 专用空间，且 Pod 名以该用户派生 Deployment 名加 "-" 为前缀
// （Pod 名 = Deployment 名 + ReplicaSet 哈希 + Pod 哈希）。
// 邮箱为空或 Pod 名为空时恒为 false。
func (a *accessBiz) isOwnAgentPod(ctx context.Context, ns *Namespace, pod string) (bool, error) {
	email := MustGetUser(ctx).Email
	if ns == nil || pod == "" || email == "" {
		return false, nil
	}
	inAgentNS, err := a.isAgentNamespaceID(ctx, ns.ID)
	if err != nil {
		return false, err
	}
	if !inAgentNS {
		return false, nil
	}
	prefix := AgentDeploymentName(AgentProjectName(email)) + "-"
	return strings.HasPrefix(pod, prefix), nil
}

// isAgentNamespaceID 判定命名空间 ID 是否为 AI Agent 专用空间。
// 逻辑名 AgentNamespace 经 data 层补 ns_prefix 后才能定位实际空间，故必须按
// 加载出的 ID 比较，不能拿命名空间名字面值与 AgentNamespace 直接对比。
// 空间尚不存在（用户未启用过 AI）视为非 agent 空间，不产生错误。
func (a *accessBiz) isAgentNamespaceID(ctx context.Context, id int) (bool, error) {
	ns, err := a.nsRepo.FindByName(ctx, AgentNamespace)
	if err != nil {
		if errs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return ns != nil && ns.ID == id, nil
}

// RequireNamespaceOwner 是命名空间所有权门卫：校验当前用户是否为命名空间 owner，
// 仅 owner 放行（admin 例外），否则 errs.ErrorPermissionDenied。
// 判定失败（含 ns 为 nil）同样拒绝——无法判定所有权时安全默认是拒绝；
// 越权拒绝的错误日志由消费方（services 层）统一打印。
func (a *accessBiz) RequireNamespaceOwner(ctx context.Context, ns *Namespace) error {
	user := MustGetUser(ctx)
	if ns == nil || (!user.IsAdmin() && ns.CreatorEmail != user.Email) {
		return errs.WrapPermissionDenied(errs.ErrorPermissionDenied, "变更命名空间（仅 owner 可操作）")
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
		return nil, errs.WrapPermissionDenied(errs.ErrorPermissionDenied, "管理员操作")
	}
	return ctx, nil
}

// RequireSuperAdmin 是方法级超管门禁：fullMethodName 精确命中 superOnly 时要求当前
// 用户为内置超级管理员（身份判据见 UserInfo.IsSuperAdmin），否则 errs.ErrorPermissionDenied；
// 未命中直接放行，把判据留给调用方（通常紧跟其后的 RequireAdmin）。
//
// ⚠️ 语义与 RequireAdmin **极性相反**：RequireAdmin 的可变参数是"豁免 admin 校验"的白名单
// （命中即放宽到用户级），本方法的可变参数是"必须超管"的强制名单（命中即收紧到超管）。
// 需求方向不同故极性不同，不要按 RequireAdmin 的直觉套用。
//
// 为什么需要它：RequireAdmin 的 allowlist 只能表达"比 admin 更宽"，无法表达"比 admin 更严"，
// 而恢复被误删资源（Restore）会重建集群侧骨架（k8s namespace + docker secret），设计上收归
// 超管。阈值收口于此处，避免同一权威判据在各服务手写后各自漂移。
func (a *accessBiz) RequireSuperAdmin(ctx context.Context, fullMethodName string, superOnly ...string) error {
	for _, name := range superOnly {
		if fullMethodName == name {
			if !MustGetUser(ctx).IsSuperAdmin() {
				return errs.WrapPermissionDenied(errs.ErrorPermissionDenied, "超级管理员操作")
			}
			return nil
		}
	}
	return nil
}

// RequireFileAccess 是文件访问门卫：校验当前用户是否为文件所有者（Username
// 匹配）或 admin，否则 errs.ErrorPermissionDenied。
func (a *accessBiz) RequireFileAccess(ctx context.Context, fil *File) error {
	user := MustGetUser(ctx)
	if fil.Username == user.Name || user.IsAdmin() {
		return nil
	}
	return errs.WrapPermissionDenied(errs.ErrorPermissionDenied, fmt.Sprintf("访问文件 #%d", fil.ID))
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
