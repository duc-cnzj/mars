// Package mcp 把 mars 的「容器排查」能力以标准 MCP（Model Context Protocol）
// 暴露出去，供 mars 沙箱侧的 opencode agent 以及任意第三方 MCP 客户端经 mars
// 网关调用。
//
// 传输：MCP Streamable HTTP（MCP 规范 2025-06-18 起的标准远程传输），挂在 /mcp。
// 该端点是「标准」MCP：任何支持 Streamable HTTP 的客户端（opencode、Claude Code、
// Cursor 等）带上 mars bearer token 即可连接，不依赖 mars 私有协议。
//
// 鉴权：与 gRPC 拦截器、文件路由共用 middlewares.LoginHTTP + biz.Authenticate——
// 调用方必须在 Authorization 头携带 mars token；令牌校验通过后把用户注入 ctx，
// 各工具内部再按命名空间/项目做 RBAC 判定（与 gRPC 服务同一套 AccessBiz）。
//
// 出参口径：**能复用平台 gRPC 服务接口的能力一律复用**（NamespaceSvc/ProjectSvc/
// ContainerSvc/MetricsSvc/ChangelogSvc/GitSvc/RepoSvc/EventSvc/HibernateSvc），工具只做
// 入参搬运，出参直接回吐服务返回的模型 JSON——与 mars 界面完全同源，且不声明 MCP
// outputSchema（tools/list 保持精简，模型拿到的是原始 JSON）。只有平台没有对应接口的
// 三个能力直连 biz 并自行收敛字段。
//
// 安全边界：**读工具只读，写工具交人工确认**。
//
//   - 读工具（list_*/get_*/check_*）打 ReadOnlyHint=true，客户端可免确认直接执行；
//   - 写工具（apply_project/toggle_hibernate/restart_pod/exec_command）**刻意不打** ReadOnlyHint，
//     按 MCP 规范即「非只读」，客户端（opencode/Claude Code/Cursor）会在真正发出
//     tools/call 前弹人工确认——这是本批采纳的审批机制（最轻量，零 DB 改动）。
//   - 不提供 delete_project 等销毁类能力：写工具集已含改名/改配置/删 Pod 的破坏面，
//     再叠加不可逆的资源销毁会让「客户端确认」这道闸门不足以兜底，故不纳入。
//
// 残余风险：配置宽松或恶意的客户端可以无视 hint 直接调用。因写操作仍受 mars RBAC
// 约束（token 只能操作其有权空间），风险被限制在「调用者本就有权做的变更」；
// 要更强的保证需升级为「服务端两阶段 dry-run→confirm」或审批单，本批不做。
package mcp

import (
	"context"
	"net/http"
	"time"

	changelogpb "github.com/duc-cnzj/mars/api/v6/proto/changelog"
	containerpb "github.com/duc-cnzj/mars/api/v6/proto/container"
	eventpb "github.com/duc-cnzj/mars/api/v6/proto/event"
	gitpb "github.com/duc-cnzj/mars/api/v6/proto/git"
	hibernatepb "github.com/duc-cnzj/mars/api/v6/proto/hibernate"
	metricspb "github.com/duc-cnzj/mars/api/v6/proto/metrics"
	namespacepb "github.com/duc-cnzj/mars/api/v6/proto/namespace"
	projectpb "github.com/duc-cnzj/mars/api/v6/proto/project"
	repopb "github.com/duc-cnzj/mars/api/v6/proto/repo"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/mlog"
	"github.com/duc-cnzj/mars/v6/internal/server/middlewares"
	marsVersion "github.com/duc-cnzj/mars/v6/internal/version"
	"github.com/gorilla/mux"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Path 是 MCP 端点在 mars 网关上的挂载路径。沙箱侧 mars-agent 生成的 opencode
// 配置按 /mcp 拼接 MCP 地址（见 agent/config.go），改动此处需同步该约定。
const Path = "/mcp"

// sessionTimeout 是空闲 MCP 会话的回收时长：客户端长时间不再发请求即关闭会话，
// 避免第三方客户端忘记 DELETE 时会话在服务端无限累积。
//
// 注意：go-sdk 未提供会话数上限，空闲回收前会话表仍可增长；/mcp 要求 bearer 鉴权，
// 累积速度受已认证账号数 × 并发会话数约束，属接受的残余风险。切 Stateless 会拒绝
// GET SSE、可能破坏部分客户端，故保留有状态模式。
const sessionTimeout = 30 * time.Minute

// serverInstructions 是 MCP initialize 响应里的 instructions 字段：客户端（含其
// 背后的 LLM）在建立会话时即可看到，用来交代 mars 的资源模型、工具用法与安全约定。
// 这是「进入容器后给模型的提示词」的服务端一半，另一半由沙箱镜像的 AGENTS.md 承载。
//
// 分层约定：本段只写**跨工具**的信息（资源模型、返回值通用约定、推荐排查路径、配置模型
// 的边界），单个工具的入参/出参/语义一律写在 tools.go 的 Tool.Description 里，两处不复述
// 同一段话——否则改一处忘另一处，模型会收到互相漂移的两份说明。
const serverInstructions = `你是 mars 平台（Kubernetes 应用管理平台）的容器排查助手，通过本 MCP 服务器的工具读取集群状态。

资源模型：命名空间（namespace/空间）→ 项目（project/应用）→ Pod → 容器（container）。
- 命名空间是权限边界：公开空间人人可见，私有空间仅管理员、创建者与成员可见。
- 一个项目是一个持续存在的部署配置，每次 apply 版本号递增；可能对应多个 Pod（滚动更新时新旧 Pod 并存）。

目标项目定位（先对齐，再动手）：
- 用户没点名项目、或点名的名字有多个候选时，禁止自行挑一个：先调 list_namespaces（name 可模糊匹配
  空间名或项目名）或 list_projects 列出候选，把「空间名（空间 id）／项目名（项目 id）／部署状态」
  列给用户，请其确认后再继续；只有候选唯一时才能直接采用。
- 用户只说「我的项目」且可见项目不止一个时，同样先列候选再追问，不要默认最近部署的那个。
- 用户确认目标项目后，本次会话后续默认沿用该项目；用户明确切换时才更换。
- 写操作（apply_project / toggle_hibernate / restart_pod / exec_command）目标不唯一时一律停止并追问——
  宁可多问一句，不可改错项目。读操作可先做跨项目汇总（如整体健康概览），但下钻到具体项目、
  拉日志或执行命令前仍需完成上面的对齐。
- 用户用简称/部分名称指代时，先用上面的 name 模糊过滤定位；命中多个就走候选列表流程，不要猜。

工具分层（每个工具的完整入参、出参与语义以它自己的 description 为准，本说明不复述）：
- 只读诊断（ReadOnlyHint，客户端免确认）：get_current_user、list_namespaces、list_projects、
  get_project、list_hibernate_windows、list_project_containers、check_apply_status、
  get_project_endpoints、get_project_resource_tree、check_pod_status、list_namespace_events、
  list_operation_events、get_container_logs、get_pod_metrics、list_project_changelogs、
  list_git_branches、list_git_commits、get_commit_detail、get_pipeline_status。
- 写操作（客户端会弹人工确认）：apply_project、toggle_hibernate、restart_pod、exec_command。
- 身份确认：get_current_user 返回本次调用绑定的 mars 身份。AI 沙箱令牌锚定的是 agent
  项目创建者——管理员进入别人的沙箱时，这里也只会看到创建者，跨权限操作一律被拒绝。

返回值通用约定（不理解这四条最容易误读）：
1. 所有出参都是标准 JSON（snake_case），且服务端模型带 omitempty——**零值字段直接缺失**：
   缺字段 = 空串/0/false/空数组，不是“未知”。例如 StateContainer 的 ready=false 时字段不存在。
2. 枚举在模型 JSON 里是**数字**：deploy_status 0=StatusUnknown、1=StatusDeploying、
   2=StatusDeployed、3=StatusFailed。**唯一例外**是 list_projects，它的 deploy_status 是字符串
   （StatusUnknown/StatusDeploying/StatusDeployed/StatusFailed）。
3. 入参寻址方式各不相同：list_namespaces 用 name（界面名，免 k8s 前缀）；list_projects 用
   namespace + 可选 name；get_project / list_project_containers / check_apply_status /
   get_project_endpoints / get_project_resource_tree / list_project_changelogs /
   list_git_branches / list_git_commits / get_commit_detail / get_pipeline_status /
   toggle_hibernate 用 project_id；check_pod_status / get_container_logs / get_pod_metrics /
   restart_pod / exec_command 用 namespace + pod（+ container）。不要混用 ID 与名字。
4. 工具失败会以 isError=true 回吐原因（NotFound / PermissionDenied / 参数校验失败等）；
   成功返回也不代表异步动作已完成（apply 的部署、restart 的重建都要再做一次状态确认）。

推荐排查路径：
1. 先用 list_projects / get_project 定位目标项目，用 list_project_containers 拿到它当前的 Pod 与容器名。
2. 部署/发版后先调 check_apply_status 拿整体判定、失败原因与上一次崩溃日志；再看
   get_project_endpoints 确认入口与资源用量，get_project_resource_tree 确认 Ingress/Service/工作负载链路。
3. 用 check_pod_status 判断单个 Pod 是否运行；非运行时读其 reason（reason 含 not found 说明 Pod 名错了或 Pod 已消失）。
4. 用 list_namespace_events 看 Pod 事件（调度失败、拉镜像失败、OOMKilled、CrashLoopBackOff 等），
   按返回里的 object 字段过滤出目标 Pod；count 是累计次数。mars 平台操作审计是另一个工具
   list_operation_events，别把两者混为一谈。
5. 用 get_container_logs 拉容器日志定位应用层错误（默认尾部 1000 行，可用 tail_lines 调整；多容器 Pod 必须传 container；
   日志末尾出现“已截断”提示说明日志不完整）；如需资源画像用 get_pod_metrics。
6. 需要把线上版本与代码/发布对上号时：list_project_changelogs 看最近部署记录，list_git_commits 取提交，
   get_commit_detail 看提交说明，get_pipeline_status 确认 CI 结论（CI 成功 ≠ 已部署，部署结果仍看 check_apply_status）。

配置模型（细节见 get_project / apply_project 的工具描述）：
- 用户能改的只有两块：① config——mars 配置片段；② 自定义参数（custom values）——仓库后台声明的字段。
  两者都按仓库声明的落点路径合并进 helm values 后再渲染，所以**用户看到的 config 不是 helm values.yaml 本身**。
- 自定义字段的权威清单在 get_project 返回的 repo.mars_config.elements：path 是 extra_values 的
  唯一合法键；type 决定取值约束（Switch=bool、InputNumber=整数、Select/Radio 必须落在
  select_values 内，其余类型透传）。仓库级副本字段与休眠窗配置也在这份 mars_config 里。
- helm values.yaml 属仓库级、由管理员维护，**没有任何工具能直接编辑它**。要改的东西若不在上述两块里，
  说明 mars 没开放给用户，应引导用户联系仓库管理员补自定义字段，不要在 config 里硬塞 helm 原始键。

改配置的固定顺序（先读后写，不可跳步）：
- 第 1 步 get_project 读出当前 config 与 final_extra_values。前者是判断"该改哪个键"的唯一依据，
  后者给出可改的自定义参数清单（其 path 就是 extra_values 的 key）。不读就写等于闭着眼睛改线上配置。
- 第 2 步再调 apply_project（**只更新、不创建**：项目必须已存在）。只传要改的东西，未传的字段
  （分支/提交/配置）与 extra_values 里没列出的自定义元素，工具都会自动沿用项目当前值，
  因此**不需要**每次把完整配置搬上来。

写工具（会真正改动环境，每次调用都会触发客户端的用户确认弹窗）：
- apply_project（更新已有项目的部署/配置）、toggle_hibernate（切换项目定时休眠开关，不触发重部署）、
  restart_pod（强制删除 Pod 交控制器重建）、exec_command（容器内一次性命令）。
  可改范围、增量语义、返回值与失败含义见各工具 description。

汇报与展示约定（向用户输出内容时必须遵守）：
- 凡是列出或定位到资源，**必须同时给出 ID 与名称**，不能只写名称：
  空间 → 「空间名（空间 id=123）」；项目 → 「项目名（项目 id=456）」；
  Pod → 完整 Pod 名（带随机后缀）；容器 → 容器名；变更记录 → 「版本 N（记录 id=...）」。
- ID 是用户后续操作的入口：项目类工具一律用 project_id，空间类工具用空间 name（不是空间 id）。
  把两者都展示出来，用户才能直接在 mars 界面或下一次指令里准确指代。
- 不要用"某个项目""那个空间"这类模糊指代，也不要把 ID 省略成"已找到"。
- 列表结果即使为空，也要说明筛选条件与所在空间（名称 + id）。

约束：
- 所有工具都受当前 token 的权限约束，跨权限访问返回 PermissionDenied，不要尝试绕过。
- 写操作不可逆（apply 会改线上部署、restart 会中断正在处理的请求、exec 可能改动容器内数据）。
  调用前先向用户说明「将要做什么、影响哪个环境」，拿到确认后再调；不要连续自动调用多个写工具。
- 工具的返回即最终结果：失败会以 isError=true 明确回吐原因，不要假设已成功。`

// Deps 收口 MCP 服务端的构造依赖，由 wire 按字段注入。不引入数据层。
//
// *Svc 是 services 层的 gRPC 服务接口（定义在 api proto 包，不在 internal 下，
// 故不引入 import 环）：**平台已有对应接口的能力一律复用它**——门卫、审计、
// 出参 JSON 形状全部与 mars UI 同源，避免 MCP 侧再造一套迟早漂移的投影。
// GitSvc/RepoSvc 服务的是分支/提交/pipeline 查询；这两个服务本身没有项目级 ACL
// （Repo.Show 仅要求登录、Git 服务无 Authorize），故工具入口必须先过 RequireProjectAccess。
// EventSvc.List 自带「按 ctx 用户收敛本人事件」的归属过滤；HibernateSvc 的 ListWindows
// 已在 Authorize 白名单里对登录用户放行，两者的读写边界都由服务自己兜住。
// 剩余 biz 依赖只服务三个没有现成接口的能力：list_projects（无按空间+项目名过滤的
// 接口）、list_namespace_events（EventSvc.List 是操作审计日志，不是 k8s 事件）、
// exec_command（services 的 ExecOnce 是 streaming RPC，MCP 侧构造不出其 server）。
type Deps struct {
	ProjectBiz    biz.ProjectBiz
	ContainerBiz  biz.ContainerBiz
	K8sBiz        biz.K8sBiz
	AccessBiz     biz.AccessBiz
	AuthBiz       biz.AuthBiz
	AgentActivity biz.AgentActivity
	NamespaceSvc  namespacepb.NamespaceServer
	ProjectSvc    projectpb.ProjectServer
	ContainerSvc  containerpb.ContainerServer
	MetricsSvc    metricspb.MetricsServer
	ChangelogSvc  changelogpb.ChangelogServer
	GitSvc        gitpb.GitServer
	RepoSvc       repopb.RepoServer
	EventSvc      eventpb.EventServer
	HibernateSvc  hibernatepb.HibernateServer
	Logger        mlog.Logger
}

// Server 是 MCP 服务端门面：持有已套好鉴权中间件的 http.Handler，注册到 mars
// 网关的 /mcp 路径上。第三方 MCP 客户端与沙箱 agent 走的是同一个入口。
type Server struct {
	handler http.Handler
	logger  mlog.Logger
}

// NewServer 构造标准 MCP 服务端：注册只读工具集 → 用 Streamable HTTP 传输承载
// （官方 go-sdk 实现，自带 initialize/tools/list/tools/call 与 SSE 会话管理）→
// 外层套 LoginHTTP 完成 token 校验与用户注入；鉴权核心与 gRPC、文件路由共用
// biz.Authenticate，杜绝第二套鉴权实现。
func NewServer(deps Deps) *Server {
	logger := deps.Logger.WithModule("services/mcp")

	srv := sdk.NewServer(
		&sdk.Implementation{Name: "mars", Version: marsVersion.GetVersion().Version},
		&sdk.ServerOptions{Instructions: serverInstructions},
	)
	registerTools(srv, deps)

	streamable := sdk.NewStreamableHTTPHandler(
		// 全部会话共用同一个 MCP server：工具集是静态的，按请求身份做区分的职责
		// 由工具内部读取 ctx 中的用户完成，不在传输层分叉。
		func(*http.Request) *sdk.Server { return srv },
		&sdk.StreamableHTTPOptions{SessionTimeout: sessionTimeout},
	)

	auth := middlewares.LoginHTTP(func(ctx context.Context, token string) (context.Context, error) {
		return biz.Authenticate(ctx, deps.AuthBiz, token)
	}, logger)

	// 套在鉴权之内：每个已鉴权的 MCP 请求都刷新沙箱活跃度，工具调用期间
	// 沙箱不会被空闲关闭任务缩容。刷新失败只记日志，不阻断 MCP 请求。
	withActivity := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, err := biz.GetUser(r.Context()); err == nil {
			biz.TouchAgentActivity(r.Context(), deps.AgentActivity, logger, user.Email)
		}
		streamable.ServeHTTP(w, r)
	})

	return &Server{handler: auth(withActivity), logger: logger}
}

// Register 把 MCP 端点挂到 mars 网关路由上。必须在前端 SPA 兜底路由之前调用，
// 否则 /mcp 会被兜底吞成 index.html（与 /api、/ws 的注册顺序约束一致，
// 见 internal/server/http.go 的 initServer）。
func (s *Server) Register(mux *mux.Router) {
	mux.Path(Path).Handler(s.handler).Name("mcp")
}
