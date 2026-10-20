package mcp

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	changelogpb "github.com/duc-cnzj/mars/api/v6/proto/changelog"
	containerpb "github.com/duc-cnzj/mars/api/v6/proto/container"
	metricspb "github.com/duc-cnzj/mars/api/v6/proto/metrics"
	namespacepb "github.com/duc-cnzj/mars/api/v6/proto/namespace"
	projectpb "github.com/duc-cnzj/mars/api/v6/proto/project"
	websocketpb "github.com/duc-cnzj/mars/api/v6/proto/websocket"
	"github.com/duc-cnzj/mars/v6/internal/biz"
	"github.com/duc-cnzj/mars/v6/internal/errs"
	"github.com/duc-cnzj/mars/v6/internal/util/pagination"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// maxLogBytes 是 get_container_logs 单次返回日志的上限：容器日志常见数十 MB，
// 原样塞进 LLM 上下文会直接撑爆 token 预算，故在服务端截断并显式标注已被截断，
// 让模型知道还有更多内容（可提示用户去 mars 界面看完整日志）。
const maxLogBytes = 128 * 1024

// logTruncatedNotice 追加在被截断的日志末尾。截断后日志已不完整，必须让模型
// 明确知道「后面还有内容」，否则它会把残缺日志当成全部事实下结论。
const logTruncatedNotice = "\n\n...[日志过长，已截断，完整日志请在 mars 界面查看]"

// readOnlyAnnotations 是只读工具的通用注解：显式告诉客户端这些工具不修改任何环境，
// 客户端可据此免去「是否允许执行」的交互确认（MCP ToolAnnotations 语义）。
var readOnlyAnnotations = &sdk.ToolAnnotations{ReadOnlyHint: true}

// writeAnnotations 是写工具的注解，零值即写语义：ReadOnlyHint 是 bool 且带 omitempty，
// 零值 false 序列化后即「缺失」→ 按 MCP 规范视为非只读 → 客户端在执行前弹人工确认；
// DestructiveHint 是 *bool，nil 时规范默认 true，正合「本工具有破坏性」的事实。
// 显式声明这个变量（而非在每处写字面量）是为了让「写工具不打只读标」成为可见契约：
// 谁把某个写工具的 Annotations 换成 readOnlyAnnotations，这里就是唯一的审查抓手。
var writeAnnotations = &sdk.ToolAnnotations{}

// toolset 持有工具实现所需的依赖。每个工具都是它的方法，注册时以方法值绑定，
// 这样单个工具函数只关心「入参 → 调用 → 出参」，不需要层层传依赖。
//
// 依赖口径：**平台已有对应 gRPC 服务接口的能力，一律复用该接口**（*Svc），
// 这样门卫、审计、出参 JSON 形状都与 mars UI 完全同源；只有平台没有现成接口的
// 能力才直连 biz——当前只剩三处：list_projects（无按空间+项目名过滤的接口）、
// list_namespace_events（EventSvc 的 List 是操作审计日志，不是 k8s 事件）、
// exec_command（services 的 ExecOnce 是 streaming RPC，MCP 侧构造不出其 server）。
type toolset struct {
	projectBiz   biz.ProjectBiz
	containerBiz biz.ContainerBiz
	k8sBiz       biz.K8sBiz
	accessBiz    biz.AccessBiz

	namespaceSvc namespacepb.NamespaceServer
	projectSvc   projectpb.ProjectServer
	containerSvc containerpb.ContainerServer
	metricsSvc   metricspb.MetricsServer
	changelogSvc changelogpb.ChangelogServer
}

// registerTools 把全部排查工具注册到 MCP server 上：前半是只读诊断工具（免客户端确认），
// 后半是写工具（不打 ReadOnlyHint，由客户端弹人工确认）。新增工具请同步
// mcp.go 的 serverInstructions，避免模型不知道该怎么用。
func registerTools(srv *sdk.Server, deps Deps) {
	ts := &toolset{
		projectBiz:   deps.ProjectBiz,
		containerBiz: deps.ContainerBiz,
		k8sBiz:       deps.K8sBiz,
		accessBiz:    deps.AccessBiz,
		namespaceSvc: deps.NamespaceSvc,
		projectSvc:   deps.ProjectSvc,
		containerSvc: deps.ContainerSvc,
		metricsSvc:   deps.MetricsSvc,
		changelogSvc: deps.ChangelogSvc,
	}

	sdk.AddTool(srv, &sdk.Tool{
		Name: "list_namespaces",
		Description: "列出当前 mars 用户可访问的命名空间（项目空间，也是权限边界；私有空间仅管理员/创建者/成员可见）。" +
			"返回 {items, count, page, page_size}：" +
			"items 每项含 id、name、description、private（false 时字段缺失）、creator_email、favorite、members[{id,email}]，以及 projects（该空间下的项目概要：id/name/deploy_status 等，**不含** config/repo）；" +
			"count 为命中的空间总数。" +
			"注意：deploy_status 是数字枚举——0=StatusUnknown、1=StatusDeploying、2=StatusDeployed、3=StatusFailed。" +
			"name 入参是模糊过滤：同时匹配空间名与该空间下的项目名。" +
			"后续所有工具入参的 namespace 都取这里的 name（界面名，免 k8s 前缀，服务端会幂等补全）。" +
			"items[].id 是空间 ID、items[].projects[].id 是项目 ID——向用户展示结果时必须带上这些 ID。",
		Annotations: readOnlyAnnotations,
	}, ts.listNamespaces)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "get_current_user",
		Description: "返回当前 MCP 调用者的 mars 身份与权限属性：{id, email, name, roles, is_admin, is_super_admin}。" +
			"用于确认「我现在以谁的身份操作」——尤其在 AI 沙箱里：沙箱令牌锚定的是 agent 项目创建者，管理员进入别人的沙箱时这里返回的仍是创建者而非管理员。" +
			"所有工具的实际权限都受这个身份约束；跨权限访问会返回 PermissionDenied，不要试图绕过。" +
			"注意：roles 是生效角色（后台手动升降级已生效），is_admin 等价于 roles 含 mars_admin。",
		Annotations: readOnlyAnnotations,
	}, ts.getCurrentUser)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "list_projects",
		Description: "列出指定命名空间下当前用户可见的项目（应用），支持按项目名模糊过滤与分页；需要空间名时先用 list_namespaces。" +
			"返回 {projects, total, page, page_size}：" +
			"projects 每项含 id（get_project / list_project_containers 的入参；向用户展示时必须带上）、name、deploy_status、version、git_branch、git_commit、updated_by（最后一次人为操作人）；" +
			"total 为命中总数，分页缺省 page=1、page_size=15。" +
			"注意：本工具的 deploy_status 是**字符串**（StatusUnknown / StatusDeploying / StatusDeployed / StatusFailed），与 get_project 返回的**数字枚举**表示不同。" +
			"projects 为空数组表示该空间下没有匹配项目（不是错误）。",
		Annotations: readOnlyAnnotations,
	}, ts.listProjects)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "get_project",
		Description: "按项目 ID 返回完整项目模型（与 mars 界面的项目详情同源），是改配置前必须先读的依据。关键字段：" +
			"config（该项目当前存储的 mars 配置片段，**不是** helm values.yaml；改配置时要把改好的整段传给 apply_project）；" +
			"final_extra_values（[{path,value}]，当前生效的自定义参数，其中 path 就是 apply_project.extra_values 的键）；" +
			"version（部署版本号/乐观锁）、deploy_status（数字枚举：0=StatusUnknown、1=StatusDeploying、2=StatusDeployed、3=StatusFailed）、" +
			"git_branch、git_commit、docker_image（镜像列表）、pod_selectors、env_values、repo_id/repo（仓库绑定，不可经工具修改）、namespace、hibernate_*（定时休眠）。" +
			"注意：所有工具出参都是带 omitempty 的 JSON——零值字段会直接缺失（缺失=空串/0/false/空数组），不要把“字段不存在”当成“未知”。" +
			"项目不存在或无权限时返回工具错误。",
		Annotations: readOnlyAnnotations,
	}, ts.getProject)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "list_project_containers",
		Description: "列出指定项目当前实际的 Pod 与容器，是 check_pod_status / get_container_logs / exec_command 取 pod、container 入参的唯一来源。" +
			"返回 {items:[{namespace, pod, container, is_old, terminating, pending, ready}]}：" +
			"is_old=true 表示滚动更新中即将下线的旧 Pod；terminating=true 表示正在删除；pending=true 表示尚未调度/启动；ready=true 表示容器已就绪（布尔字段为 false 时缺失）。" +
			"优先操作 ready=true 且 is_old、terminating 均缺失（false）的条目；items 为空数组表示项目当前没有运行中的 Pod（未部署或全部异常）。" +
			"同一 Pod 可能有多个容器，调用日志/exec 时把对应 container 一并传入。",
		Annotations: readOnlyAnnotations,
	}, ts.listProjectContainers)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "check_pod_status",
		Description: "查询指定 Pod 是否处于 Running。返回 {running, reason}：" +
			"running=true 时 reason 为空；running=false 时 reason 是 k8s 给出的人类可读原因（如 CrashLoopBackOff、ImagePullBackOff、Evicted，或 'pod not running.'）；" +
			"Pod **不存在**时同样返回 running=false，reason 中含 not found 文本——这是正常的否定结果，不是工具错误。" +
			"限制：只有 running 布尔与原因，**不含**重启次数、容器退出码、容器级状态；这些要结合 list_namespace_events 与 get_container_logs 判断，不要凭空推断重启次数或退出码。",
		Annotations: readOnlyAnnotations,
	}, ts.checkPodStatus)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "list_namespace_events",
		Description: "列出命名空间内的 Kubernetes 事件（注意：不是 mars 的操作审计日志），用于定位调度失败、镜像拉取失败、OOMKilled、CrashLoopBackOff、探针失败等集群侧问题；Pod 起不来时优先看这里。" +
			"返回 {events:[{type, reason, note, object, count, last_seen, container}]}：" +
			"type=Normal|Warning；reason 如 Scheduled/FailedScheduling/BackOff/OOMKilling/FailedMount/Unhealthy；note 为事件正文；" +
			"object='Kind/Name'（用它筛出目标 Pod 的事件）；count 为累计发生次数（越大越说明问题反复出现）；last_seen 为 RFC3339（可能为空）；" +
			"container 是事件关联对象的字段路径（如 spec.containers{app}，**不是**容器名，可能为空）。" +
			"返回的是整个命名空间的事件且顺序不保证——需自行按 object 过滤。k8s 事件默认只保留约 1 小时，事件缺失不代表历史上没出过问题。",
		Annotations: readOnlyAnnotations,
	}, ts.listNamespaceEvents)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "get_container_logs",
		Description: "读取容器日志（默认只取**最后 1000 行**，不是全量；可用 tail_lines 调整行数）。返回 {namespace, pod_name, container_name, log}：" +
			"log 是日志正文；超过 128 KiB 会被截断，并在末尾追加“...[日志过长，已截断，完整日志请在 mars 界面查看]”——看到该标记说明后面还有内容，不能把截断点当成日志结尾。" +
			"show_events=true 且 Pod 处于 Pending 时，log 返回的是该 Pod 的**事件文本**（每行一条），不是容器日志。" +
			"注意：① 多容器 Pod **必须**显式传 container（服务端不会自动挑选主容器，省略会报错；单容器 Pod 可省略）；" +
			"② container_name 只是回显入参 container，省略时为空，不代表服务端实际选了哪个容器；" +
			"③ Pod 不存在、或 Pending 且未要求事件时返回工具错误。",
		Annotations: readOnlyAnnotations,
	}, ts.getContainerLogs)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "get_pod_metrics",
		Description: "查询单个 Pod 的 CPU/内存实时用量（来自 metrics-server），用于判断资源不足、内存泄漏等性能问题。" +
			"返回 {cpu, memory, humanize_cpu, humanize_memory, time, length}：" +
			"cpu 单位是**毫核**（250 表示 0.25 核）；memory 单位是 **KB（1000 字节，已取整）**；" +
			"humanize_cpu / humanize_memory 是可直接展示的字符串（如 '250 m'、'123 MiB'）；time 是采样时刻（HH:MM:SS）；" +
			"length 固定为该服务流式采样的时间窗样本数（30），一次性查询可忽略。" +
			"Pod 不在运行或不存在时返回 NotFound 错误（reason 带原因）；Pod 在运行但 metrics-server 暂无数据时返回错误文本含 'pod metrics unavailable' 的错误——这是采样暂缺，不是 Pod 消失，可稍后重试。",
		Annotations: readOnlyAnnotations,
	}, ts.getPodMetrics)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "list_project_changelogs",
		Description: "列出项目最近的部署变更记录（服务端固定取最近 15 条，再按 limit 截断；按版本号倒序，最新在前），用于把故障时间点与代码/配置变更对上号。" +
			"返回 {items:[...]}，每条关键字段：version（部署版本号）、username（操作人）、config_changed（本次 config 是否真的变化）、" +
			"date（中文相对时间，由服务端模板生成，如 '3 天以前'；不是固定格式，不要对其做字符串解析）与 created_at（RFC3339，解析时间请用它）、git_branch / git_commit / git_commit_title / git_commit_author / git_commit_date、" +
			"config（当次部署的 mars 配置片段）、final_extra_values（当次生效的自定义参数）、docker_image（逗号拼接的**字符串**）、env_values。" +
			"only_changed=true 只返回 config 确实变化的记录；limit 缺省 10（超过 15 也只能拿到 15 条）。该接口不返回 manifest（清单）。",
		Annotations: readOnlyAnnotations,
	}, ts.listProjectChangelogs)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "apply_project",
		Description: "[写操作] **更新已有项目的部署配置**（不创建新项目）。用 namespace + name 定位已存在的项目；项目不存在直接报错，请先用 list_projects / get_project 确认项目已存在。项目名、所属空间、仓库绑定都**不可通过本工具修改**（仓库绑定在项目创建时即固化）。" +
			"**改配置只有两条通道**：① config——mars 配置片段；② 自定义参数（custom values）——仓库后台声明的自定义元素（replicas、镜像 tag 等），用 extra_values 指定。两者都是**片段**，由服务端按仓库配置的落点路径合并进 helm values 后再渲染；**没有直接编辑 helm values.yaml 的入口**（helm values.yaml 属仓库级、由管理员在后台维护）。所以想改的东西若既不在 config 片段里、也不是声明过的自定义参数，说明 mars 没把它开放给用户改——引导用户找仓库管理员加自定义字段，不要往 config 里硬塞 helm 的原始键（会落到错误的层级）。" +
			"**增量更新语义**：除 version 外其余入参不传即沿用项目当前值——只想换分支就只传 git_branch，不必把 config 再抄一遍；要改配置就传 config（**整段替换**，不是键级合并，改一个键也要带上整段）；要改自定义参数就只在 extra_values 里传要改的那几个（key 用 get_project 返回的 final_extra_values 里的 path），未列出的元素沿用项目现状、不会被重置回仓库默认。version 不传则由工具自动取当前版本（乐观锁，等价于以线上现状为基准），显式传则可检出期间的并发部署。" +
			"局限：config / git_branch 这类字符串字段无法用本工具清空（没传与传空串无法区分，需清空请到 mars 界面）；extra_values 不受此限（键存在即设定，传空串就是置空）。" +
			"**返回值** {yaml_files, project, dry_run}：yaml_files 是本次渲染出的 Kubernetes 清单；" +
			"dry_run=true 时只渲染、不写库不部署，project 为空；dry_run=false 时 project 为部署后的项目模型（含新 version 与 deploy_status）——" +
			"**返回成功不等于 Pod 已就绪**：部署是异步的，deploy_status 可能仍是 StatusDeploying，需再用 list_project_containers / check_pod_status / get_container_logs 验证。" +
			"**失败语义**：项目不存在 → NotFound 错误（不会创建）；version 与线上不一致 → 并发部署冲突错误；无空间权限 → PermissionDenied。" +
			"强烈建议先 dry_run=true 预览渲染出的 YAML 再正式部署——正式部署会真实改动线上环境。",
		Annotations: writeAnnotations,
	}, ts.applyProject)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "restart_pod",
		Description: "[写操作] 重启 Pod：等价 kubectl delete pod --force --grace-period=0——Pod 被**立即**删除（不给优雅退出时间），再由其工作负载控制器（Deployment/StatefulSet 等）重建。" +
			"返回 {deleted, namespace, pod, message}：deleted=true 只表示删除请求已被 apiserver 接受，**重建是异步的**，需再用 list_project_containers / check_pod_status 确认新 Pod 就绪。" +
			"风险：裸 Pod（无控制器）删除后不会自动回来；正在处理的请求会立即中断。pod 入参要传具体 Pod 名（先用 list_project_containers 获取），不是项目名。",
		Annotations: writeAnnotations,
	}, ts.restartPod)

	sdk.AddTool(srv, &sdk.Tool{
		Name: "exec_command",
		Description: "[写操作] 在容器内执行**一次性**命令（非交互式，无 TTY），返回结构化结果而不是流。" +
			"返回 {output, exit_code?, truncated, timed_out, start_failed, error?}：" +
			"output 是 stdout/stderr 按到达顺序合并的文本，MCP 侧最多回吐约 128 KiB；" +
			"exit_code 仅在命令真正跑完且退出码**非 0** 时出现——字段缺失表示退出码为 0（若同时有 truncated/timed_out/start_failed 标志，则以标志位为准）；" +
			"truncated=true 表示输出超限被强制终止（服务端上限 5 MiB）或被截断到 128 KiB；timed_out=true 表示超过 timeout_seconds（缺省 60 秒）被强制终止；" +
			"start_failed=true 表示容器 exec 未能启动（如命令不存在）；error 只在这三种异常时给出说明。" +
			"**退出码非 0 不是工具错误**（isError=false），必须自行检查 exit_code 与各标志位。" +
			"需要管道、&&、重定向等 shell 特性时把 command 写成 [\"sh\",\"-c\",\"<完整命令>\"]；Pod 未运行或容器名不存在时返回工具错误。命令与结果会写入 mars 审计。",
		Annotations: writeAnnotations,
	}, ts.execCommand)
}

// ---------- list_namespaces ----------

// listNamespacesIn 是 list_namespaces 的入参。
type listNamespacesIn struct {
	Page     int32  `json:"page,omitempty" jsonschema:"页码，从 1 开始，缺省 1"`
	PageSize int32  `json:"page_size,omitempty" jsonschema:"每页条数，缺省 15"`
	Name     string `json:"name,omitempty" jsonschema:"模糊过滤（可选）：匹配空间名，或该空间下任一项目名（用户常记得项目名而记不住空间名）"`
}

// listNamespaces 列出当前用户可见的命名空间：直接复用 NamespaceSvc.List，由它按
// token 注入的用户身份做可见性过滤（非管理员只看公开/创建者/成员空间），出参即
// mars 界面的命名空间模型 JSON（含嵌套项目），MCP 侧不再另造投影。
func (t *toolset) listNamespaces(ctx context.Context, _ *sdk.CallToolRequest, in listNamespacesIn) (*sdk.CallToolResult, any, error) {
	req := &namespacepb.ListRequest{Page: &in.Page, PageSize: &in.PageSize}
	if in.Name != "" {
		req.Name = &in.Name
	}
	resp, err := t.namespaceSvc.List(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	return nil, resp, nil
}

// ---------- get_current_user ----------

// getCurrentUserIn 是 get_current_user 的入参：无参数，空对象。
type getCurrentUserIn struct{}

// getCurrentUserOut 是 get_current_user 的出参：令牌绑定的身份与权限属性。
type getCurrentUserOut struct {
	ID           string   `json:"id,omitempty"`
	Email        string   `json:"email"`
	Name         string   `json:"name,omitempty"`
	Roles        []string `json:"roles"`
	IsAdmin      bool     `json:"is_admin"`
	IsSuperAdmin bool     `json:"is_super_admin"`
}

// getCurrentUser 返回当前调用者的身份。用户由鉴权中间件注入 ctx（biz.Authenticate），
// 与所有工具的权限判定同源；ctx 缺用户属鉴权链路故障，直接上抛错误而不是返回空身份。
func (t *toolset) getCurrentUser(ctx context.Context, _ *sdk.CallToolRequest, _ getCurrentUserIn) (*sdk.CallToolResult, any, error) {
	user, err := biz.GetUser(ctx)
	if err != nil {
		return nil, nil, err
	}
	// 普通用户没有角色时也要回吐空数组而不是 null：调用方（LLM）按「roles 是数组」
	// 消费，null 会被部分客户端当成缺失或类型错误。
	roles := user.Roles
	if roles == nil {
		roles = []string{}
	}
	return nil, getCurrentUserOut{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Roles:        roles,
		IsAdmin:      user.IsAdmin(),
		IsSuperAdmin: user.IsSuperAdmin(),
	}, nil
}

// ---------- list_projects ----------

// listProjectsIn 是 list_projects 的入参。
type listProjectsIn struct {
	Namespace string `json:"namespace" jsonschema:"命名空间名称（必填）"`
	Name      string `json:"name,omitempty" jsonschema:"按项目名模糊过滤（可选，大小写不敏感）"`
	Page      int32  `json:"page,omitempty" jsonschema:"页码，从 1 开始，缺省 1"`
	PageSize  int32  `json:"page_size,omitempty" jsonschema:"每页条数，缺省 15"`
}

// projectSummary 是项目的精简视图。它是本工具集里唯一的「手工投影」——因为平台
// 没有「按空间名 + 项目名过滤」的现成接口（ProjectSvc.List 只支持按仓库过滤、
// NamespaceSvc 返回的是空间嵌套项目），只能直连 biz 并自行收敛字段。
type projectSummary struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	DeployStatus string `json:"deploy_status"`
	Version      int    `json:"version"`
	GitBranch    string `json:"git_branch"`
	GitCommit    string `json:"git_commit"`
	UpdatedBy    string `json:"updated_by"`
}

// listProjectsOut 是 list_projects 的出参。
type listProjectsOut struct {
	Projects []projectSummary `json:"projects"`
	Total    int32            `json:"total"`
	Page     int32            `json:"page"`
	PageSize int32            `json:"page_size"`
}

// listProjects 列出命名空间下的项目：先经 AccessBiz 校验空间可访问（不可访问返回
// PermissionDenied），再用解析出的空间 ID 拉项目列表——data 层按邮箱/管理员谓词
// 二次过滤，双重保证不越权。
func (t *toolset) listProjects(ctx context.Context, _ *sdk.CallToolRequest, in listProjectsIn) (*sdk.CallToolResult, any, error) {
	ns, err := t.accessBiz.RequireNamespaceAccessByName(ctx, in.Namespace)
	if err != nil {
		return nil, nil, err
	}
	user := biz.MustGetUser(ctx)
	page, size := pagination.InitByDefault(&in.Page, &in.PageSize)

	projects, pag, err := t.projectBiz.List(ctx, &biz.ListProjectInput{
		Page:        page,
		PageSize:    size,
		Name:        in.Name,
		NamespaceID: int32(ns.ID),
		Email:       user.Email,
		IsAdmin:     user.IsAdmin(),
	})
	if err != nil {
		return nil, nil, err
	}

	out := listProjectsOut{
		Projects: make([]projectSummary, 0, len(projects)),
		Total:    pag.Count,
		Page:     pag.Page,
		PageSize: pag.PageSize,
	}
	for _, p := range projects {
		out.Projects = append(out.Projects, toProjectSummary(p))
	}
	return nil, out, nil
}

// ---------- get_project ----------

// getProjectIn 是 get_project 的入参。
type getProjectIn struct {
	ProjectID int `json:"project_id" jsonschema:"项目 ID（必填）"`
}

// getProject 返回项目详情：直接复用 ProjectSvc.Show 的返回值（与 mars 界面同源的
// 项目模型 JSON），门卫（RequireProjectAccess）与休眠态补全都在该服务内部完成——
// 项目携带完整部署配置，是 ID 枚举越权的高危面，故绝不自行绕过服务去读。
// 出参含 config——它是「要改哪些键」的唯一依据，不吐出来模型就只能盲猜。
func (t *toolset) getProject(ctx context.Context, _ *sdk.CallToolRequest, in getProjectIn) (*sdk.CallToolResult, any, error) {
	resp, err := t.projectSvc.Show(ctx, &projectpb.ShowRequest{Id: int32(in.ProjectID)})
	if err != nil {
		return nil, nil, err
	}
	return nil, resp.GetItem(), nil
}

// ---------- list_project_containers ----------

// listProjectContainersIn 是 list_project_containers 的入参。
type listProjectContainersIn struct {
	ProjectID int `json:"project_id" jsonschema:"项目 ID（必填）"`
}

// listProjectContainers 返回项目当前实际运行的容器：复用 ProjectSvc.AllContainers
// （项目级门卫与 Pod 匹配逻辑都在服务内部），出参即 StateContainer 模型 JSON，
// 是后续 check_pod_status / get_container_logs 的取参来源。
func (t *toolset) listProjectContainers(ctx context.Context, _ *sdk.CallToolRequest, in listProjectContainersIn) (*sdk.CallToolResult, any, error) {
	resp, err := t.projectSvc.AllContainers(ctx, &projectpb.AllContainersRequest{Id: int32(in.ProjectID)})
	if err != nil {
		return nil, nil, err
	}
	return nil, resp, nil
}

// ---------- check_pod_status ----------

// checkPodStatusIn 是 check_pod_status 的入参。
type checkPodStatusIn struct {
	Namespace string `json:"namespace" jsonschema:"命名空间名称（必填）"`
	Pod       string `json:"pod" jsonschema:"Pod 名称（必填）"`
}

// checkPodStatus 判断 Pod 是否运行：复用 ContainerSvc.IsPodRunning（命名空间门卫
// 与 k8s 读取都在服务内部），出参即 {running, reason}。
func (t *toolset) checkPodStatus(ctx context.Context, _ *sdk.CallToolRequest, in checkPodStatusIn) (*sdk.CallToolResult, any, error) {
	resp, err := t.containerSvc.IsPodRunning(ctx, &containerpb.IsPodRunningRequest{
		Namespace: in.Namespace,
		Pod:       in.Pod,
	})
	if err != nil {
		return nil, nil, err
	}
	return nil, resp, nil
}

// ---------- list_namespace_events ----------

// listNamespaceEventsIn 是 list_namespace_events 的入参。
type listNamespaceEventsIn struct {
	Namespace string `json:"namespace" jsonschema:"命名空间名称（必填）"`
}

// eventBrief 是 Kubernetes 事件的精简视图。k8s 事件没有对应的 mars 服务接口
// （EventSvc 的 List 查的是平台操作审计日志，不是集群事件），故只有这里直连 biz。
type eventBrief struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Note      string `json:"note"`
	Object    string `json:"object"`
	Count     int32  `json:"count"`
	LastSeen  string `json:"last_seen"`
	Container string `json:"container,omitempty"`
}

// listNamespaceEventsOut 是 list_namespace_events 的出参。
type listNamespaceEventsOut struct {
	Events []eventBrief `json:"events"`
}

// listNamespaceEvents 返回命名空间内的 k8s 事件，是「Pod 为什么起不来」的第一手证据：
// 调度失败、拉镜像失败、探针失败、OOMKilled 等都只在这里出现。经命名空间门卫后读取。
func (t *toolset) listNamespaceEvents(ctx context.Context, _ *sdk.CallToolRequest, in listNamespaceEventsIn) (*sdk.CallToolResult, any, error) {
	if _, err := t.accessBiz.RequireNamespaceAccessByName(ctx, in.Namespace); err != nil {
		return nil, nil, err
	}
	events, err := t.k8sBiz.ListEvents(in.Namespace)
	if err != nil {
		return nil, nil, err
	}
	out := listNamespaceEventsOut{Events: make([]eventBrief, 0, len(events))}
	for _, ev := range events {
		out.Events = append(out.Events, eventBrief{
			Type:      ev.Type,
			Reason:    ev.Reason,
			Note:      ev.Note,
			Object:    strings.TrimPrefix(ev.Regarding.Kind+"/"+ev.Regarding.Name, "/"),
			Count:     ev.DeprecatedCount,
			LastSeen:  formatTime(ev.DeprecatedLastTimestamp),
			Container: ev.Regarding.FieldPath,
		})
	}
	return nil, out, nil
}

// ---------- get_container_logs ----------

// getContainerLogsIn 是 get_container_logs 的入参。
type getContainerLogsIn struct {
	Namespace  string `json:"namespace" jsonschema:"命名空间名称（必填）"`
	Pod        string `json:"pod" jsonschema:"Pod 名称（必填）"`
	Container  string `json:"container,omitempty" jsonschema:"容器名称：多容器 Pod 必填（服务端不会自动选主容器，省略会报错）；单容器 Pod 可省略"`
	ShowEvents bool   `json:"show_events,omitempty" jsonschema:"仅在 Pod 处于 Pending 时有意义：为 true 时返回该 Pod 的事件文本而非容器日志"`
	TailLines  int64  `json:"tail_lines,omitempty" jsonschema:"日志尾部行数（可选，缺省 1000；0 等同缺省，负数被服务端拒绝）。仍需注意 128 KiB 字节截断上限"`
}

// getContainerLogs 拉取容器日志：复用 ContainerSvc.ContainerLog（命名空间门卫、
// Pending 事件聚合、pod 未运行 404、尾部行数策略（默认 1000，由 tail_lines 覆盖）
// 都在服务/biz 内）。
// MCP 侧额外补两处契约，使工具描述所述行为真正成立：
//   - tail_lines 负数直接拒绝（biz 会把 <=0 一律当默认值，gRPC 入口的 proto 校验
//     在 MCP 直调路径上不生效）；
//   - 省略 container 时只允许单容器 Pod（gRPC 入口靠 proto 的 required 强制 container，
//     同一校验在 MCP 路径不生效；这里显式复刻，避免把 k8s API 的原始报错透给模型）。
//
// 唯一在 MCP 侧追加的出参处理是字节上限截断——1000 行仍可能是数十 MB，直接回吐会撑爆
// LLM 上下文；截断后在被截断的日志串尾部追加提示语（出参保持原始模型形状不变）。
func (t *toolset) getContainerLogs(ctx context.Context, _ *sdk.CallToolRequest, in getContainerLogsIn) (*sdk.CallToolResult, any, error) {
	if in.TailLines < 0 {
		return nil, nil, errs.InvalidArgument("tail_lines 不能为负数：省略或传 0 表示默认 1000 行")
	}
	if in.Container == "" {
		// 先过 Pod 级门卫再读 Pod：GetPod 的错误信息会区分「不存在」与其他失败，
		// 不能让它先于鉴权执行，否则构成跨租户 Pod 存在性侧信道。
		if _, err := t.accessBiz.RequirePodAccess(ctx, in.Namespace, in.Pod); err != nil {
			return nil, nil, err
		}
		pod, err := t.k8sBiz.GetPod(in.Namespace, in.Pod)
		if err != nil {
			return nil, nil, err
		}
		if len(pod.Spec.Containers) > 1 {
			names := make([]string, 0, len(pod.Spec.Containers))
			for _, c := range pod.Spec.Containers {
				names = append(names, c.Name)
			}
			return nil, nil, errs.InvalidArgument(fmt.Sprintf(
				"Pod %s/%s 有多个容器（%s），必须显式指定 container", in.Namespace, in.Pod, strings.Join(names, ", ")))
		}
	}
	resp, err := t.containerSvc.ContainerLog(ctx, &containerpb.LogRequest{
		Namespace:  in.Namespace,
		Pod:        in.Pod,
		Container:  in.Container,
		ShowEvents: in.ShowEvents,
		TailLines:  in.TailLines,
	})
	if err != nil {
		return nil, nil, err
	}
	if text, truncated := truncateLog(resp.GetLog()); truncated {
		resp.Log = text + logTruncatedNotice
	}
	return nil, resp, nil
}

// ---------- get_pod_metrics ----------

// getPodMetricsIn 是 get_pod_metrics 的入参。
type getPodMetricsIn struct {
	Namespace string `json:"namespace" jsonschema:"命名空间名称（必填）"`
	Pod       string `json:"pod" jsonschema:"Pod 名称（必填）"`
}

// getPodMetrics 查询 Pod 的 CPU/内存实时用量：复用 MetricsSvc.TopPod（命名空间门卫
// 与采样分类都在服务/biz 内部），出参含原始 float 与 humanize 后的字符串两种形态。
func (t *toolset) getPodMetrics(ctx context.Context, _ *sdk.CallToolRequest, in getPodMetricsIn) (*sdk.CallToolResult, any, error) {
	resp, err := t.metricsSvc.TopPod(ctx, &metricspb.TopPodRequest{
		Namespace: in.Namespace,
		Pod:       in.Pod,
	})
	if err != nil {
		return nil, nil, err
	}
	return nil, resp, nil
}

// ---------- list_project_changelogs ----------

// listProjectChangelogsIn 是 list_project_changelogs 的入参。
type listProjectChangelogsIn struct {
	ProjectID   int  `json:"project_id" jsonschema:"项目 ID（必填）"`
	Limit       int  `json:"limit,omitempty" jsonschema:"最多返回几条，缺省 10；服务端固定只取最近 15 条，传更大的值也拿不到更多"`
	OnlyChanged bool `json:"only_changed,omitempty" jsonschema:"只返回 config 真正发生变化的记录（config_changed=true）"`
}

// listProjectChangelogs 返回项目最近若干次部署记录：把「故障出现的时间点」和
// 「这期间谁改了什么」对上号，是排查「刚发版就挂了」类问题的最短路径。
// 复用 ChangelogSvc.FindLastChangelogsByProjectID（项目级门卫在服务内部，它固定取
// 最近 15 条），limit 只是在此基础上再收窄，故截断在 MCP 侧就地做。
func (t *toolset) listProjectChangelogs(ctx context.Context, _ *sdk.CallToolRequest, in listProjectChangelogsIn) (*sdk.CallToolResult, any, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	resp, err := t.changelogSvc.FindLastChangelogsByProjectID(ctx, &changelogpb.FindLastChangelogsByProjectIDRequest{
		ProjectId:   int32(in.ProjectID),
		OnlyChanged: in.OnlyChanged,
	})
	if err != nil {
		return nil, nil, err
	}
	if len(resp.GetItems()) > limit {
		resp.Items = resp.Items[:limit]
	}
	return nil, resp, nil
}

// ---------- apply_project ----------

// applyProjectIn 是 apply_project 的入参。
//
// **只支持更新已有项目**：namespace + name 是**定位器**，不是可改字段——本工具不改项目名、
// 不改所属空间、不改仓库绑定（项目一旦创建，其 repo_id 不可变更）。项目不存在直接报错，不创建。
//
// 语义是**请求级增量更新**：除 version 外其余字段不传即「沿用项目当前值」，见
// loadCurrentProjectForUpdate；自定义参数（仓库 mars.Config.elements 声明的 replicas 之类）
// 要改哪个就在 extra_values 里只传哪个，其余仍沿用现状。
//
// 局限：Go 的字符串零值让「没传」与「传空串」不可区分，故本工具无法把 config/分支等**字符串字段**
// 清空——需清空请去 mars 界面操作。extra_values 不受此限：它是 map，「键存在」即显式设定，
// 传空串就是真的置空。
type applyProjectIn struct {
	Namespace string `json:"namespace" jsonschema:"命名空间名称（必填；仅用于定位项目，本工具不改项目的空间归属）"`
	Name      string `json:"name" jsonschema:"项目名称（必填；仅用于定位项目，本工具不改项目名）"`
	GitBranch string `json:"git_branch,omitempty" jsonschema:"部署所用 Git 分支（可选；不传沿用项目当前分支）"`
	GitCommit string `json:"git_commit,omitempty" jsonschema:"部署所用 Git 提交（可选；不传沿用项目当前提交）"`
	Config    string `json:"config,omitempty" jsonschema:"部署配置片段（mars 配置格式，不是 helm values.yaml：服务端按仓库配置的落点路径把该片段合并进 values；不传沿用项目当前配置）"`
	// ExtraValues 是自定义元素值的增量设置：key = 元素 path（如 web->replicas），value = 值。
	// 用 map 而非数组：键存在即「显式设定」（含空串），避开字符串零值分不清「没传/传空」的问题。
	ExtraValues map[string]string `json:"extra_values,omitempty" jsonschema:"自定义参数（custom values）的增量设置：key 为元素 path（形如 web->replicas），value 为新值；只需传要改的那几个，未列出的元素沿用项目当前值。key 必须是仓库后台声明过的元素 path"`
	Version     int32             `json:"version,omitempty" jsonschema:"项目版本号（乐观锁；不传自动取项目当前版本；显式传则以其为准，可检出并发变更）"`
	DryRun      bool              `json:"dry_run,omitempty" jsonschema:"仅渲染 YAML 不真正部署（建议先 true 预览）"`
}

// applyProject 更新已有项目：门卫与审计都由 ProjectSvc.WebApplyByName 内部完成
// （按空间名校验访问权限 → 解析仓库 → 渲染 → dry-run 记 DryRun、真实部署记 Update），
// 工具侧只做「读出当前值补齐未传字段 + 入参搬运」，不重复编排、也不自行判权限——两套门卫必然漂移。
func (t *toolset) applyProject(ctx context.Context, _ *sdk.CallToolRequest, in applyProjectIn) (*sdk.CallToolResult, any, error) {
	repoID, inherited, err := t.loadCurrentProjectForUpdate(ctx, &in)
	if err != nil {
		return nil, nil, err
	}

	req := &projectpb.WebApplyByNameRequest{
		Namespace:   in.Namespace,
		Name:        in.Name,
		GitBranch:   in.GitBranch,
		GitCommit:   in.GitCommit,
		Config:      in.Config,
		ExtraValues: mergeExtraValues(inherited, in.ExtraValues),
		DryRun:      in.DryRun,
	}
	// repo_id 恒取自项目当前绑定（本工具不支持改绑定）：WebApplyByName 无论走不走更新分支都要先
	// 解析出仓库，不传就只能按 name 精确匹配仓库名，项目名与仓库名不一致时会误报 404。
	// 0 与「未传」在 proto 里等价（都退化为按 name 匹配），故无需为 0 单独留 nil。
	req.RepoId = &repoID
	if in.Version > 0 {
		version := in.Version
		req.Version = &version
	}

	resp, err := t.projectSvc.WebApplyByName(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	return nil, resp, nil
}

// loadCurrentProjectForUpdate 读出项目当前状态并就地补齐入参里未显式提供的字段，
// 使 apply 成为纯粹的**增量更新**；同时返回项目当前绑定的仓库 id 与已落库的自定义元素值
// （finalExtraValues），供调用方原样回传。
//
// 为什么必须先读：部署链路把空值当「重置」——git_branch 空落**仓库默认分支**、git_commit 空落
// 该分支**最新提交**、config 空会把项目配置**清空**。不读就写，等于「只想改一个字段却顺手把
// 分支/提交/配置重置」。所以要改哪个字段就只传哪个，其余由这里从当前项目补上。
//
// 为什么要连自定义元素值一起继承：extra_values 空同样被 ElementsLoader 当「全用仓库默认值」，
// 于是元素会先落进 useDefaultMap 再被持久化（loader.go + runner.go），agent 只想改 config
// 却把 replicas 之类的自定义值悄悄抹回仓库默认。项目行里的 finalExtraValues 就是权威现状，
// 原样回传即「保持不动」——与 config/分支的继承语义一致。
//
// 为什么返回 repo_id 而不是让入参传：本工具**只更新、不创建**，项目的仓库绑定在创建时即固化、
// 不可变更，故 repo_id 只能取自当前项目。且 WebApplyByName 无论走不走更新分支都要先 resolveRepo，
// repo_id 缺失时会退化为按项目名精确匹配仓库名，项目名与仓库名不一致的情况下会直接 404。
//
// 为什么 NotFound 必须报错：本工具语义是「更新已有项目」，项目不存在就是无对象可更新——必须
// 明确失败，绝不能悄悄走成创建（那会让「改错项目名」变成「凭空多出一个项目」）。非 NotFound
// （DB 抖动等）原样上抛，不伪装成「项目不存在」。
func (t *toolset) loadCurrentProjectForUpdate(ctx context.Context, in *applyProjectIn) (int32, []*websocketpb.ExtraValue, error) {
	resp, err := t.projectSvc.ShowByName(ctx, &projectpb.ShowByNameRequest{Namespace: in.Namespace, Name: in.Name})
	if err != nil {
		if errs.IsNotFound(err) {
			return 0, nil, errs.NotFound(fmt.Sprintf("项目 %s/%s 不存在：apply_project 只支持更新已有项目，不支持创建", in.Namespace, in.Name))
		}
		return 0, nil, err
	}
	current := resp.GetItem()
	// 调用方显式给的值一律优先——只有「没给」才继承（<=0 / "" 即未指定）。
	if in.GitBranch == "" {
		in.GitBranch = current.GetGitBranch()
	}
	if in.GitCommit == "" {
		in.GitCommit = current.GetGitCommit()
	}
	if in.Config == "" {
		in.Config = current.GetConfig()
	}
	if in.Version <= 0 {
		in.Version = current.GetVersion()
	}
	return current.GetRepoId(), current.GetFinalExtraValues(), nil
}

// mergeExtraValues 把调用方用 extra_values 显式指定的元素值合并到「继承来的当前值」之上，
// 得到最终下发给部署链路的完整元素值集合。
//
// 为什么是「合并」而不是「入参直接替换」：ElementsLoader 对**没出现在请求里的**元素一律落回
// **仓库默认值**（并写进 useDefaultMap 被持久化），所以只传要改的那一个元素、其余不补，就会把
// 其他元素全部重置回仓库默认——正是 loadCurrentProjectForUpdate 要防的那个坑的翻版。以现状为底、
// 用入参覆盖对应 path，才能做到「改一个元素、其余原地不动」。
//
// 为什么入参是 map 而不是数组：「键存在」本身就是显式的「我要设它」，于是传空串 = 真的置空，
// 不会被 Go 的零值语义误判成「没传」——元素值本来就允许空串，这个歧义必须消掉。
//
// 为什么不复用 base 里的 *ExtraValue 指针：base 是 ShowByName 响应里项目模型的切片，就地改写
// 会污染上游响应（同一响应可能被别处引用）；description 也无需搬运，后端落库时会按当时的元素
// 定义重新固化（见 runner.go 的 extraValueDescriptions）。
func mergeExtraValues(base []*websocketpb.ExtraValue, override map[string]string) []*websocketpb.ExtraValue {
	if len(override) == 0 {
		return base
	}

	merged := make([]*websocketpb.ExtraValue, 0, len(base)+len(override))
	applied := make(map[string]struct{}, len(override))
	for _, v := range base {
		value := v.GetValue()
		if overrideValue, ok := override[v.GetPath()]; ok {
			value = overrideValue
			applied[v.GetPath()] = struct{}{}
		}
		merged = append(merged, &websocketpb.ExtraValue{Path: v.GetPath(), Value: value})
	}

	// 现状里没有的 path（新增元素）按字典序补在末尾：override 是 map，遍历顺序不定，
	// 不排序的话下发顺序随机，测试会 flaky、diff 也读不出规律。
	added := make([]string, 0, len(override))
	for path := range override {
		if _, ok := applied[path]; !ok {
			added = append(added, path)
		}
	}
	sort.Strings(added)
	for _, path := range added {
		merged = append(merged, &websocketpb.ExtraValue{Path: path, Value: override[path]})
	}
	return merged
}

// ---------- restart_pod ----------

// restartPodIn 是 restart_pod 的入参。
type restartPodIn struct {
	Namespace string `json:"namespace" jsonschema:"命名空间名称（必填）"`
	Pod       string `json:"pod" jsonschema:"Pod 名称（必填，先用 list_project_containers 拿准确名字）"`
}

// restartPod 重启 Pod：复用 ContainerSvc.ForceDeletePod（其内已有命名空间门卫与
// ForceDeletePod 审计）。grace_period_seconds=0 等价 kubectl delete --force ——
// Pod 被立即删除后由工作负载控制器重建，这才是「重启」；不是 k8s 原生 API 的能力。
func (t *toolset) restartPod(ctx context.Context, _ *sdk.CallToolRequest, in restartPodIn) (*sdk.CallToolResult, any, error) {
	resp, err := t.containerSvc.ForceDeletePod(ctx, &containerpb.ForceDeletePodRequest{
		Namespace:          in.Namespace,
		Pod:                in.Pod,
		GracePeriodSeconds: 0,
	})
	if err != nil {
		return nil, nil, err
	}
	return nil, resp, nil
}

// ---------- exec_command ----------

// execCommandIn 是 exec_command 的入参。
type execCommandIn struct {
	Namespace      string   `json:"namespace" jsonschema:"命名空间名称（必填）"`
	Pod            string   `json:"pod" jsonschema:"Pod 名称（必填）"`
	Container      string   `json:"container,omitempty" jsonschema:"容器名称（可选；省略时取 Pod 主容器：优先 kubectl.kubernetes.io/default-container 注解，否则第一个容器）"`
	Command        []string `json:"command" jsonschema:"命令与参数（必填，如 [\"ls\",\"-l\"]；需 shell 特性请用 [\"sh\",\"-c\",\"...\"]）"`
	TimeoutSeconds int64    `json:"timeout_seconds,omitempty" jsonschema:"最大执行秒数（可选，缺省 60 秒；超时强制终止）"`
}

// execCommandOut 是 exec_command 的出参：把 biz 错误帧的三类「执行结果」显式拆成布尔位，
// 让模型能区分「命令正常退出但返回码非 0」与「命令压根没跑起来/被服务端掐了」。
type execCommandOut struct {
	Output      string `json:"output"`
	Truncated   bool   `json:"truncated"`
	TimedOut    bool   `json:"timed_out"`
	StartFailed bool   `json:"start_failed"`
	ExitCode    *int32 `json:"exit_code,omitempty"`
	Error       string `json:"error,omitempty"`
}

// execCommand 在容器内执行一次性命令：命名空间门卫通过后交给 ContainerBiz.ExecOnce
// （内部做 Pod 运行态校验、默认容器解析、超时/输出上限强制终止，并把命令与结果落审计）。
// 门卫用 toolset 内的 accessBiz 而非 ExecOnce 的 services 包装——后者是 streaming RPC，
// 其 server 类型 MCP 侧无法构造；这与 list_namespace_events 等既有工具的「先门卫再调
// biz」同模式。
func (t *toolset) execCommand(ctx context.Context, _ *sdk.CallToolRequest, in execCommandIn) (*sdk.CallToolResult, execCommandOut, error) {
	if _, err := t.accessBiz.RequirePodAccess(ctx, in.Namespace, in.Pod); err != nil {
		return nil, execCommandOut{}, err
	}
	stream := &execOnceStream{ctx: ctx}
	err := t.containerBiz.ExecOnce(ctx, stream, biz.MustGetUser(ctx), &biz.ExecOnceInput{
		Namespace:      in.Namespace,
		Pod:            in.Pod,
		Container:      in.Container,
		Command:        in.Command,
		TimeoutSeconds: in.TimeoutSeconds,
	})
	if err != nil {
		return nil, execCommandOut{}, err
	}
	return nil, stream.outcome(), nil
}

// execOnceStream 把 biz.ExecOnce 的流式输出收进内存，供 MCP 一次性返回（MCP 是请求/响应模型，
// 没有流式回吐通道）。Send 会被两个 goroutine 并发调用——biz 的逐字节输出循环与主流程的
// 结果错误帧（退出码/超时/截断）——故所有字段都由 mu 保护。
type execOnceStream struct {
	ctx    context.Context
	mu     sync.Mutex
	buf    bytes.Buffer
	code   int64
	msg    string
	hasErr bool
}

// Context 满足 biz.ExecOnceStream：biz 用它做日志与取消传播。
func (s *execOnceStream) Context() context.Context { return s.ctx }

// Send 满足 biz.ExecOnceStream：数据帧逐字节追加到缓冲，错误帧记录码与消息。
// 永不返回 error——返回 error 会让 biz 的输出循环提前退出、丢掉后续输出。
func (s *execOnceStream) Send(resp *containerpb.ExecResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg := resp.GetMessage(); len(msg) > 0 {
		s.buf.Write(msg)
		return nil
	}
	if e := resp.GetError(); e != nil {
		s.code, s.msg, s.hasErr = e.GetCode(), e.GetMessage(), true
	}
	return nil
}

// outcome 把收集到的输出与错误帧折算成结构化出参。错误码语义与 biz 一致（见
// biz.ExecOnceTruncatedCode/ExecOnceExecFailedCode/ExecOnceTimeoutCode）：-1 输出截断、
// -2 exec 启动失败、-3 执行超时，其余落在 0-255 的即真实退出码。
func (s *execOnceStream) outcome() execCommandOut {
	s.mu.Lock()
	defer s.mu.Unlock()

	logText, cut := truncateLog(s.buf.String())
	out := execCommandOut{Output: logText, Truncated: cut}
	if !s.hasErr {
		return out
	}

	switch s.code {
	case biz.ExecOnceTruncatedCode:
		out.Truncated = true
	case biz.ExecOnceTimeoutCode:
		out.TimedOut = true
	case biz.ExecOnceExecFailedCode:
		out.StartFailed = true
	default:
		if s.code >= 0 && s.code <= 255 {
			code := int32(s.code)
			out.ExitCode = &code
		}
	}
	if out.Truncated || out.TimedOut || out.StartFailed {
		out.Error = s.msg
	}
	return out
}

// ---------- 公共小工具 ----------

// toProjectSummary 把 biz.Project 投影成精简视图，仅供 list_projects 使用
// （get_project 直接回吐 ProjectSvc.Show 的完整项目模型，不做投影）。
func toProjectSummary(p *biz.Project) projectSummary {
	return projectSummary{
		ID:           p.ID,
		Name:         p.Name,
		DeployStatus: p.DeployStatus.String(),
		Version:      p.Version,
		GitBranch:    p.GitBranch,
		GitCommit:    p.GitCommit,
		UpdatedBy:    p.UpdatedBy,
	}
}

// truncateLog 按 maxLogBytes 截断日志，返回截断后的文本与是否发生截断。
// 单字节边界截断可能在末尾留下半个 UTF-8 字符，故回退到最后一个完整 rune 边界。
func truncateLog(content string) (string, bool) {
	if len(content) <= maxLogBytes {
		return content, false
	}
	cut := content[:maxLogBytes]
	for len(cut) > 0 && !isValidUTF8Boundary(cut) {
		cut = cut[:len(cut)-1]
	}
	return cut, true
}

// isValidUTF8Boundary 判断切片末尾是否落在合法的 UTF-8 字符边界上：
// 末字节若为 ASCII 或 UTF-8 起始字节（非 10xxxxxx 续字节）即为边界。
func isValidUTF8Boundary(s string) bool {
	return s[len(s)-1]&0xC0 != 0x80
}

// formatTime 把可能为零值的 metav1.Time 格式化为 RFC3339，零值返回空串——
// 事件列表里 deprecated 时间字段可能缺失，直接格式化会输出 "0001-01-01T00:00:00Z" 干扰阅读。
func formatTime(t metav1.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
