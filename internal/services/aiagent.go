package services

// aiagent.go 提供 AI Agent 的用户侧入口：
//   - GET  /api/ai/status        查询 AI 功能是否可用 + 当前用户 agent 项目状态
//   - POST /api/ai/agent/enable  以当前用户身份把 AI Agent 仓库部署成一个项目
//
// AI 功能的可用性由「全局唯一的 AI Agent 仓库」（repos.ai_agent）决定；所有用户的
// agent 项目统一落在写死的私有命名空间 mars-agent 下（实际 k8s 名带 ns_prefix 前缀），
// 项目名由用户邮箱可读派生（agent-<本地部分>-<哈希>）。该命名空间仅管理员/成员可见，
// 普通用户的部署因此需要 deploy.WithTrustedNamespaceAccess 这一平台内部受信标记来绕过
// 空间门禁。

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/duc-cnzj/mars/api/v6/proto/types"
	websocket_pb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/app"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/biz/schematype"
	"github.com/duc-cnzj/mars/v6/internal/deploy"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/locker"
	"github.com/duc-cnzj/mars/v6/internal/metrics"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	"github.com/gorilla/mux"
	"github.com/samber/lo"
)

const (
	// aiAgentNamespaceDesc 是自动创建 mars-agent 空间时写入的描述。
	aiAgentNamespaceDesc = "AI Agent 专用空间（平台自动管理）"

	aiStatusPath      = "/api/ai/status"
	aiAgentEnablePath = "/api/ai/agent/enable"

	// aiAgentNamespaceLockKey 串行化 mars-agent 空间的「查无 → 创建」流程：
	// NamespaceBiz.Create 是查后创建，两个用户并发首次启用会同时走到 k8s/DB 写入。
	aiAgentNamespaceLockKey = "ai-agent:namespace"
)

// errAgentNamespaceOccupied 标记「mars-agent 空间已被非系统创建者占用」这一业务冲突，
// 供 serveEnable 区分 409（可解释冲突）与 500（真实内部错误）。
var errAgentNamespaceOccupied = errors.New("ai agent namespace occupied by non-system creator")

// errAgentNamespaceInitializing 标记并发首次启用时「另一个请求正在创建空间」，
// 与占用冲突一样属于可重试的业务冲突（409）。
var errAgentNamespaceInitializing = errors.New("AI Agent 命名空间正在初始化，请稍后重试")

// aiStatusResponse 是 /api/ai/status 与 /api/ai/agent/enable 的 200 出参。
// 只保留前端实际消费的字段（零死出参）：enabled 决定入口是否可用，agent 提供
// 沙箱运行态与打开项目详情所需的 project_id。
type aiStatusResponse struct {
	// Enabled 表示 AI 功能是否可用（已配置 AI Agent 仓库且仓库处于启用状态）。
	Enabled bool `json:"enabled"`
	// Agent 是当前用户的 agent 项目状态；尚未部署时为 null。
	Agent *aiAgentStatus `json:"agent"`
}

// 沙箱运行态（aiAgentStatus.Sandbox）的枚举值：由 Deployment 副本数与就绪数派生，
// 前端据此显示「运行中 / 启动中 / 休眠中 / 尚未部署」。
const (
	agentSandboxRunning  = "running"
	agentSandboxStarting = "starting"
	agentSandboxSleeping = "sleeping"
	agentSandboxAbsent   = "absent"
)

// aiAgentStatus 是当前用户 agent 项目的状态快照：只保留前端编排所需的最小信息。
type aiAgentStatus struct {
	ProjectID int `json:"project_id"`
	// Sandbox 是沙箱运行态：running（副本就绪）/ starting（已扩容未就绪）/
	// sleeping（副本数为 0）/ absent（Deployment 不存在）。
	Sandbox string `json:"sandbox"`
}

// applyProjectFunc 抽象共享部署编排，测试可替换（生产恒为 deploy.ApplyProject）。
type applyProjectFunc func(context.Context, deploy.ApplyProjectDeps, *deploy.ApplyProjectInput) (deploy.Job, error)

// aiAgentDeps 收口 aiAgentHandler 的构造依赖。
type aiAgentDeps struct {
	AuthBiz        biz.AuthBiz
	RepoBiz        biz.RepoBiz
	NamespaceBiz   biz.NamespaceBiz
	ProjectBiz     biz.ProjectBiz
	GitBiz         biz.GitBiz
	AccessBiz      biz.AccessBiz
	AccessTokenBiz biz.AccessTokenBiz
	K8sRepo        biz.K8sRepo
	AgentActivity  biz.AgentActivity
	JobMgr         deploy.JobManager
	// Locker 与部署全局锁共用同一后端：唤醒换令牌/空间初始化都要与并发部署互斥。
	Locker locker.Locker
	// EventBiz 记录平台自动创建空间、轮换会话令牌等副作用审计。
	EventBiz biz.EventBiz
	// PluginMgr 提供 ws 插件：自建部署需要真实 PubSub 才能广播 WsReloadProjects
	// （与 gRPC WebApply 的 WebsocketSync 语义一致），不能退化成 EmptyPubSub。
	PluginMgr app.PluginManager
	Logger    mlog.Logger
}

// aiAgentHandler 提供 AI 状态查询与一键启用 agent 的 HTTP 端点，
// 路由本身套 LoginHTTP 鉴权。
type aiAgentHandler struct {
	deps   aiAgentDeps
	apply  applyProjectFunc
	authMW func(http.Handler) http.Handler
}

// newAIAgentHandler 构造 AI Agent 处理器：鉴权核心与 gRPC 拦截器、MCP、文件路由共用
// biz.Authenticate，杜绝第二套鉴权实现。
func newAIAgentHandler(deps aiAgentDeps) *aiAgentHandler {
	deps.Logger = deps.Logger.WithModule("services/aiAgentHandler")
	return &aiAgentHandler{
		deps:  deps,
		apply: deploy.ApplyProject,
		authMW: middlewares.LoginHTTP(func(ctx context.Context, token string) (context.Context, error) {
			return biz.Authenticate(ctx, deps.AuthBiz, token)
		}, deps.Logger),
	}
}

// Register 把 AI 状态与启用端点挂到 mars 网关路由上。必须在前端 SPA 兜底与
// /api/ 前缀路由之前注册，否则会被 grpc-gateway 的 PathPrefix 吞掉
// （与 /api/ai/config 同约束）。精确路径先于前缀路由注册，互不抢占。
func (h *aiAgentHandler) Register(mux *mux.Router) {
	mux.Handle(aiStatusPath, h.authMW(http.HandlerFunc(h.serveStatus))).Methods(http.MethodGet).Name("ai_status")
	mux.Handle(aiAgentEnablePath, h.authMW(http.HandlerFunc(h.serveEnable))).Methods(http.MethodPost).Name("ai_agent_enable")
}

// serveStatus 返回 AI 功能是否可用以及当前用户的 agent 项目状态。
// 未配置 AI Agent 仓库不是错误：返回 enabled=false（前端据此隐藏 AI 入口）。
func (h *aiAgentHandler) serveStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := biz.MustGetUser(ctx)

	repo, err := h.deps.RepoBiz.GetAIAgent(ctx)
	if err != nil && !errs.IsNotFound(err) {
		h.deps.Logger.ErrorCtx(ctx, "查询 AI Agent 仓库失败: ", err)
		http.Error(w, "查询 AI Agent 仓库失败", http.StatusInternalServerError)
		return
	}
	// NotFound 时 repo 为 nil（data 层 toRepo 的 nil 安全语义），buildStatus 按未配置处理。
	h.respond(ctx, w, user.Email, repo)
}

// serveEnable 以当前用户身份启用 agent：
// 校验 AI 仓库 → 确保私有命名空间 → 幂等检查项目 → 未部署则走共享部署编排。
func (h *aiAgentHandler) serveEnable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := biz.MustGetUser(ctx)
	if user.Email == "" {
		http.Error(w, "当前用户缺少 email，无法启用 AI Agent", http.StatusBadRequest)
		return
	}

	repo, err := h.deps.RepoBiz.GetAIAgent(ctx)
	if err != nil {
		if errs.IsNotFound(err) {
			http.Error(w, "AI 功能未启用：尚未配置 AI Agent 仓库", http.StatusBadRequest)
			return
		}
		h.deps.Logger.ErrorCtx(ctx, "查询 AI Agent 仓库失败: ", err)
		http.Error(w, "查询 AI Agent 仓库失败", http.StatusInternalServerError)
		return
	}
	if !repo.Enabled {
		http.Error(w, "AI Agent 仓库已禁用", http.StatusBadRequest)
		return
	}
	// 沙箱靠部署时签发的 <UserToken> 访问 mars（配置/MCP/LLM 代理）。仓库 values 未引用
	// 该变量时部署不会签发令牌，Pod 能起来但所有回连 401——在入口直接拦住，给出可操作指引。
	if !deploy.HasUserTokenTemplate(repo.GetMarsConfig().ValuesYaml) {
		http.Error(w, "AI Agent 仓库配置缺少 <.UserToken>：请将 values 中的 mars.token 设为 \"<.UserToken>\"，否则沙箱无法访问 mars", http.StatusBadRequest)
		return
	}

	ns, err := h.ensureAgentNamespace(ctx, user)
	if err != nil {
		// 空间被非系统空间占用 / 另一个请求正在初始化，都是可解释的业务冲突，
		// 其余按内部错误处理。
		if errors.Is(err, errAgentNamespaceOccupied) || errors.Is(err, errAgentNamespaceInitializing) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.deps.Logger.ErrorCtx(ctx, "准备 AI Agent 命名空间失败: ", err)
		http.Error(w, "准备 AI Agent 命名空间失败", http.StatusInternalServerError)
		return
	}

	name := biz.AgentProjectName(user.Email)
	existing, err := h.deps.ProjectBiz.FindByName(ctx, name, ns.ID)
	switch {
	case err == nil:
		// 幂等：项目已存在。仓库不符说明同名项目被别的仓库占用，拒绝静默接管。
		if existing.RepoID != repo.ID {
			http.Error(w, fmt.Sprintf("项目 %s 已被其他仓库占用（repo_id=%d）", name, existing.RepoID), http.StatusConflict)
			return
		}
		// 先记活跃再唤醒：空闲缩容拿同一把项目锁后复查活跃度，若 Touch 落在锁释放
		// 之后，缩容可能把刚唤醒的沙箱立刻关掉。Touch 必须在进入唤醒临界区之前。
		biz.TouchAgentActivity(ctx, h.deps.AgentActivity, h.deps.Logger, user.Email)
		deployed, wakeErr := h.wakeAgent(ctx, ns, existing, user)
		if wakeErr != nil {
			// 唤醒换令牌与并发部署共用项目级锁：抢锁失败说明同一项目正在被操作，
			// 属可重试冲突而不是服务端故障。
			if errors.Is(wakeErr, deploy.ErrDeployInProgress) {
				http.Error(w, wakeErr.Error(), http.StatusConflict)
				return
			}
			h.deps.Logger.ErrorCtx(ctx, "唤醒 AI Agent 沙箱失败: ", wakeErr)
			http.Error(w, "唤醒 AI Agent 沙箱失败", http.StatusInternalServerError)
			return
		}
		// Deployment 丢失（被手工删除/命名空间清理）时不能静默返回：用户会拿到
		// 一个永远起不来的沙箱。项目行还在，走一次更新部署把它补回来。
		if !deployed {
			if err := h.deployAgent(ctx, ns, repo, user, existing); err != nil {
				// 同一项目的并发部署被全局锁拒绝：可重试的业务冲突，不应报 500。
				if errors.Is(err, deploy.ErrDeployInProgress) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				h.deps.Logger.ErrorCtx(ctx, "重建 AI Agent 沙箱失败: ", err)
				http.Error(w, "重建 AI Agent 沙箱失败", http.StatusInternalServerError)
				return
			}
		}
		h.respond(ctx, w, user.Email, repo)
		return
	case !errs.IsNotFound(err):
		h.deps.Logger.ErrorCtx(ctx, "查询 AI Agent 项目失败: ", err)
		http.Error(w, "查询 AI Agent 项目失败", http.StatusInternalServerError)
		return
	}

	if err := h.deployAgent(ctx, ns, repo, user, nil); err != nil {
		// 并发启用同一用户时后到的请求会撞上部署全局锁：提示稍后重试（409），
		// 而不是让客户端以为是服务端故障。
		if errors.Is(err, deploy.ErrDeployInProgress) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.deps.Logger.ErrorCtx(ctx, "启用 AI Agent 部署失败: ", err)
		http.Error(w, "启用 AI Agent 部署失败", http.StatusInternalServerError)
		return
	}

	// 部署成功后回读项目状态（项目行由 runner 创建）。
	if _, err := h.deps.ProjectBiz.FindByName(ctx, name, ns.ID); err != nil {
		h.deps.Logger.ErrorCtx(ctx, "启用 AI Agent 后回读项目失败: ", err)
		http.Error(w, "启用成功但回读项目状态失败", http.StatusInternalServerError)
		return
	}
	biz.TouchAgentActivity(ctx, h.deps.AgentActivity, h.deps.Logger, user.Email)
	h.respond(ctx, w, user.Email, repo)
}

// deployAgent 执行 agent 项目部署：existing 为 nil 表示首次创建；非 nil 表示项目行
// 已存在但 Deployment 丢失，走更新部署重建，并原样继承分支/提交/配置/自定义元素，
// 避免把已有配置重置成仓库默认值。
// 目标空间是私有的系统空间，故带平台内部受信标记绕过空间门禁（标记只能在 deploy 包内构造）。
func (h *aiAgentHandler) deployAgent(ctx context.Context, ns *biz.Namespace, repo *biz.Repo, user *biz.UserInfo, existing *biz.Project) error {
	// 部署期间必须用真实 PubSub：runner 在开始和结束时广播 WsReloadProjects，
	// 在线客户端（管理端项目列表）据此实时刷新。EmptyPubSub 会把 ToAll 静默吞掉，
	// 导致 AI 环境初始化完成后页面仍显示旧状态（与 gRPC WebApply 的 WebsocketSync 对齐）。
	pubsub := h.deps.PluginMgr.Ws().New("", "")
	defer pubsub.Close()
	input := &deploy.JobInput{
		Type:        websocket_pb.Type_CreateProject,
		NamespaceId: int32(ns.ID),
		Name:        biz.AgentProjectName(user.Email),
		RepoID:      int32(repo.ID),
		User:        user,
		PubSub:      pubsub,
		Messager:    deploy.NoopMsger{},
	}
	if existing != nil {
		input.Type = websocket_pb.Type_UpdateProject
		input.ProjectID = int32(existing.ID)
		input.Version = lo.ToPtr(int32(existing.Version))
		input.GitBranch = existing.GitBranch
		input.GitCommit = existing.GitCommit
		input.Config = existing.Config
		input.ExtraValues = existing.FinalExtraValues
	}
	_, err := h.apply(deploy.WithTrustedNamespaceAccess(ctx), h.applyDeps(), &deploy.ApplyProjectInput{JobInput: input})
	return err
}

// wakeAgent 唤醒已缩容的沙箱：沙箱在 0 副本时换发会话令牌并缩放到 1。
// 部署时签发的是 12h 临时令牌，睡过有效期后 Pod 再起来会因拉配置 401 崩溃，
// 故唤醒必须先更新沙箱 Secret。
//
// 返回 deployed=false 表示 Deployment 不存在（项目行还在但工作负载丢了），
// 调用方应改走更新部署把沙箱重建出来，不能静默当成功。
func (h *aiAgentHandler) wakeAgent(ctx context.Context, ns *biz.Namespace, proj *biz.Project, user *biz.UserInfo) (deployed bool, err error) {
	name := biz.AgentDeploymentName(proj.Name)
	// 与部署流水线、空闲缩容共用项目级全局锁。即使沙箱仍在运行也要进锁：
	// 缩容在锁内复查活跃度，若这里不进锁直接返回，缩容可能在同一时刻把刚 Touch
	// 的沙箱关掉。进锁后「运行中」与「已缩容」的判断才与缩容决策互斥。
	// 换发令牌 + 扩容同样必须互斥，否则两次并发唤醒会各签一个令牌、
	// 后写入的 Secret 与先撤销的旧令牌交错，留下失效令牌。
	// Deployment 运行态必须在锁内读取：锁外读到的是副本数快照，可能在抢到锁前
	// 已被缩容改写；拿旧快照判定「仍在运行」会静默漏唤醒（用户看到启用成功，
	// 沙箱却仍是 0 副本）。
	release, acquired := h.deps.Locker.RenewalAcquire(
		deploy.GetSlugName(ns.ID, proj.Name), deploy.LockTTLSeconds, deploy.LockRenewalSeconds)
	if !acquired {
		return true, deploy.ErrDeployInProgress
	}
	defer release()

	dep, err := h.deps.K8sRepo.GetDeployment(ns.Name, name)
	if err != nil {
		if errs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	// Replicas 为 nil 表示 k8s 默认 1，同样视为仍在运行。
	if dep.Spec.Replicas == nil || *dep.Spec.Replicas > 0 {
		return true, nil
	}

	if err := h.refreshSessionToken(ctx, ns.Name, proj, user); err != nil {
		return true, err
	}
	if err := h.deps.K8sRepo.ScaleDeployment(ctx, ns.Name, name, 1); err != nil {
		return true, err
	}
	metrics.AgentSandboxTransitionTotal.WithLabelValues(metrics.AgentSandboxActionWake).Inc()
	return true, nil
}

// refreshSessionToken 为唤醒的沙箱换发会话令牌：签发新令牌并写入沙箱 Secret，
// 再撤销同项目的历史令牌（与部署流水线的轮换语义一致；撤销失败不阻断唤醒）。
// Secret 读/写失败时必须尽力撤销刚签发的新令牌——否则它既不在沙箱里、也不会被
// RevokeByUsageExcept 之外的路径回收，成为一张可用的孤儿凭据。
func (h *aiAgentHandler) refreshSessionToken(ctx context.Context, namespace string, proj *biz.Project, user *biz.UserInfo) error {
	token, err := h.deps.AccessTokenBiz.Grant(ctx, &biz.GrantAccessTokenInput{
		ExpireSeconds: int32(deploy.UserTokenTTL / time.Second),
		Usage:         deploy.UserTokenUsage(proj.ID),
		User:          agentTokenUser(proj, user),
	})
	if err != nil {
		return err
	}

	secretName := biz.AgentDeploymentName(proj.Name)
	secret, err := h.deps.K8sRepo.GetSecret(ctx, namespace, secretName)
	if err != nil {
		h.revokeOrphanToken(ctx, token.Token)
		return err
	}
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}
	secret.Data[biz.AgentSessionTokenSecretKey] = []byte(token.Token)
	if _, err := h.deps.K8sRepo.UpdateSecret(ctx, namespace, secretName, secret); err != nil {
		h.revokeOrphanToken(ctx, token.Token)
		return err
	}
	if err := h.deps.AccessTokenBiz.RevokeByUsageExcept(ctx, deploy.UserTokenUsage(proj.ID), token.Token); err != nil {
		h.deps.Logger.WarningCtxf(ctx, "撤销 agent 旧会话令牌失败: %v", err)
	}
	// 令牌轮换是会影响沙箱回连能力的副作用，留审计但不记录令牌原文。
	h.deps.EventBiz.AuditLogWithRequest(types.EventActionType_Update, user.Name, user.Email,
		"唤醒 AI Agent 沙箱并轮换会话令牌",
		map[string]any{"project_id": proj.ID, "project_name": proj.Name})
	return nil
}

// agentTokenUser 返回沙箱会话令牌应绑定的身份：恒为 agent 项目创建者。
// 创建者本人唤醒时沿用其完整 UserInfo（保留角色/展示名）；管理员等其他人唤醒
// 别人的沙箱时只带创建者邮箱，角色按最小权限留空——沙箱的身份归属不随「谁点了
// 一下」而变。历史项目 creator 缺失时退回唤醒者，保持旧行为。
func agentTokenUser(proj *biz.Project, requester *biz.UserInfo) *biz.UserInfo {
	if proj == nil || requester == nil || proj.Creator == "" ||
		proj.Creator == requester.Email || !biz.IsAgentProjectOf(proj.Name, proj.Creator) {
		return requester
	}
	return &biz.UserInfo{Email: proj.Creator}
}

// revokeOrphanToken 尽力撤销未能写入沙箱的新令牌；失败只记 Warning，
// 因为此时调用方已有更严重的错误需要上抛。
func (h *aiAgentHandler) revokeOrphanToken(ctx context.Context, token string) {
	if err := h.deps.AccessTokenBiz.Revoke(ctx, token); err != nil {
		h.deps.Logger.WarningCtxf(ctx, "撤销未写入沙箱的会话令牌失败: %v", err)
	}
}

// ensureAgentNamespace 确保 mars-agent 空间存在、归系统所有且为私有空成员。
// 已存在但创建者不是超级管理员时 fail-closed 报错，不接管他人空间。
// Create 语义是「查后创建」：首次并发启用会各自查到 NotFound 后同时创建，
// 故用固定键串行化；抢锁失败返回可重试冲突。
func (h *aiAgentHandler) ensureAgentNamespace(ctx context.Context, user *biz.UserInfo) (*biz.Namespace, error) {
	release, acquired := h.deps.Locker.RenewalAcquire(
		aiAgentNamespaceLockKey, deploy.LockTTLSeconds, deploy.LockRenewalSeconds)
	if !acquired {
		return nil, errAgentNamespaceInitializing
	}
	defer release()

	ns, existed, err := h.deps.NamespaceBiz.Create(ctx, biz.AgentNamespace, aiAgentNamespaceDesc, schematype.SuperAdminEmail)
	if err != nil {
		return nil, err
	}
	if !existed {
		// 平台替用户创建了命名空间，必须留痕（谁触发、创建了哪个空间）。
		h.deps.EventBiz.AuditLogWithRequest(types.EventActionType_Create, user.Name, user.Email,
			fmt.Sprintf("自动创建 AI Agent 命名空间 %s", ns.Name),
			map[string]any{"namespace_id": ns.ID, "namespace": ns.Name})
	}
	if !isSuperAdminNamespaceCreator(ns.CreatorEmail) {
		return nil, errs.WrapInvalidArgument(
			fmt.Errorf("命名空间 %s 已被 %s 创建，平台不会接管非系统空间: %w", biz.AgentNamespace, ns.CreatorEmail, errAgentNamespaceOccupied),
			"ensure ai agent namespace",
		)
	}
	// Create 不设置私有属性，这里强制私有并清空成员：空间仅管理员可见。
	return h.deps.NamespaceBiz.UpdateConfig(ctx, &biz.UpdateConfigInput{
		ID:      ns.ID,
		Private: lo.ToPtr(true),
		Emails:  []string{},
	})
}

// respond 组装并输出统一的 AI 状态响应。repo 为 nil 表示未配置 AI Agent 仓库。
func (h *aiAgentHandler) respond(ctx context.Context, w http.ResponseWriter, email string, repo *biz.Repo) {
	resp := &aiStatusResponse{}
	if repo != nil {
		resp.Enabled = repo.Enabled
	}
	agent, err := h.agentStatus(ctx, email)
	if err != nil {
		h.deps.Logger.ErrorCtx(ctx, "查询 AI Agent 项目状态失败: ", err)
		http.Error(w, "查询 AI Agent 项目状态失败", http.StatusInternalServerError)
		return
	}
	resp.Agent = agent

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.deps.Logger.ErrorCtx(ctx, "write ai status response: ", err)
	}
}

// agentStatus 查询当前用户的 agent 项目状态；空间或项目不存在时返回 nil（尚未部署）。
func (h *aiAgentHandler) agentStatus(ctx context.Context, email string) (*aiAgentStatus, error) {
	ns, err := h.deps.NamespaceBiz.FindByName(ctx, biz.AgentNamespace)
	if err != nil {
		if errs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	proj, err := h.deps.ProjectBiz.FindByName(ctx, biz.AgentProjectName(email), ns.ID)
	if err != nil {
		if errs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	sandbox, err := h.sandboxState(ns.Name, proj.Name)
	if err != nil {
		return nil, err
	}
	return &aiAgentStatus{
		ProjectID: proj.ID,
		Sandbox:   sandbox,
	}, nil
}

// sandboxState 由 Deployment 的副本数与就绪数派生沙箱运行态：
// 副本 0 = 空闲休眠；副本 >0 且已有就绪副本 = 运行中；其余（扩容中/滚动更新中）
// = 启动中；Deployment 不存在 = absent。Replicas 为 nil 表示 k8s 默认 1。
func (h *aiAgentHandler) sandboxState(namespace, projectName string) (string, error) {
	dep, err := h.deps.K8sRepo.GetDeployment(namespace, biz.AgentDeploymentName(projectName))
	if err != nil {
		if errs.IsNotFound(err) {
			return agentSandboxAbsent, nil
		}
		return "", err
	}
	if dep.Spec.Replicas != nil && *dep.Spec.Replicas == 0 {
		return agentSandboxSleeping, nil
	}
	if dep.Status.ReadyReplicas > 0 {
		return agentSandboxRunning, nil
	}
	return agentSandboxStarting, nil
}

// applyDeps 组装共享部署编排依赖；鉴权/仓库/git/版本/JobMgr 与 gRPC/WS 部署入口同源。
func (h *aiAgentHandler) applyDeps() deploy.ApplyProjectDeps {
	return deploy.ApplyProjectDeps{
		AccessBiz:  h.deps.AccessBiz,
		RepoBiz:    h.deps.RepoBiz,
		GitBiz:     h.deps.GitBiz,
		ProjectBiz: h.deps.ProjectBiz,
		JobMgr:     h.deps.JobMgr,
		Logger:     h.deps.Logger,
	}
}

// isSuperAdminNamespaceCreator 判定空间创建者是否为内置超级管理员。
// data 层 toNamespace 会把超管邮箱映射为展示名 biz.SuperAdminName，故两个值都要认。
func isSuperAdminNamespaceCreator(creator string) bool {
	return creator == schematype.SuperAdminEmail || creator == biz.SuperAdminName
}
