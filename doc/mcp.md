
# MCP 接入说明（AI 容器排查能力）

> 本文是 mars MCP 端点的**权威接入文档**，由 `internal/services/mcp/` 的实现反推归纳。
> **维护约定**：新增/删除/改名工具、调整鉴权或权限裁剪、改变返回约定时，必须同步更新本文，
> 并与 `internal/services/mcp/mcp.go` 的 `serverInstructions` 保持一致（两处不复述同一段话，
> 但事实不能互相矛盾）。权限判定的完整契约见 [access_control.md](access_control.md)。

## 1. 概述

mars 把「容器排查 + 部署配置」能力以标准 **MCP（Model Context Protocol）** 暴露出去：

- **端点**：`/mcp`（与 Web、gRPC、`/api` 同一个 mars 网关地址）
- **传输**：Streamable HTTP（MCP 规范 2025-06-18 起的标准远程传输），**Stateless 模式**
- **客户端**：任何支持 Streamable HTTP 的 MCP 客户端——opencode、Claude Code、Cursor 等；
  沙箱内的 `opencode` agent 由 `mars-agent` 自动注入配置，无需手工接入
- **鉴权**：mars token（`Authorization: Bearer <token>`），与 gRPC 拦截器、文件路由共用
  `biz.Authenticate`；工具内部再按命名空间/项目做 RBAC（同一套 `AccessBiz`）
- **能力范围**：命名空间/项目/Pod/容器/事件/日志/指标/变更记录/Git 分支与提交/CI 流水线查询，
  以及部署配置更新、定时休眠开关、Pod 重启、容器内一次性命令；管理员另有仓库级自定义字段读写

它不是 mars 私有协议：换成任何 MCP 客户端，只要带上有效 token 就能用。

## 2. 快速接入

> 下文命令中的 `https://mars.example.com` 是**示例网关入口**，请替换为实际部署域名
> （本地可用 `http://127.0.0.1:4000`）。

### 2.1 取得 token

两种 token 都可用（都支持可选的 `Bearer ` 前缀，大小写不敏感）：

| 方式 | 接口 | 有效期 | 适用场景 |
|---|---|---|---|
| 登录 JWT | `POST /api/auth/login`，body `{"username":"...","password":"..."}` | 12 小时 | 临时调试 |
| 访问令牌 | `POST /api/access_tokens`，body `{"expire_seconds":2592000,"usage":"mcp"}` | 自定义 | 长期客户端/CI |

```bash
# 登录取 JWT（响应 {token, expires_in}）
curl -sS https://mars.example.com/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"123456"}'
```

> 访问令牌同样在 mars 界面的令牌管理里维护，可随时吊销；令牌的权限就是**签发时那个用户的权限**。

### 2.2 客户端配置

**Claude Code（CLI）**

```bash
claude mcp add --transport http mars https://mars.example.com/mcp \
  --header "Authorization: Bearer $MARS_TOKEN"
```

**opencode（`opencode.json`）**

```json
{
  "mcp": {
    "mars": {
      "type": "remote",
      "url": "https://mars.example.com/mcp",
      "enabled": true,
      "headers": { "Authorization": "Bearer {env:MARS_TOKEN}" }
    }
  }
}
```

**Cursor（`~/.cursor/mcp.json`）**

```json
{
  "mcpServers": {
    "mars": {
      "url": "https://mars.example.com/mcp",
      "headers": { "Authorization": "Bearer ${env:MARS_TOKEN}" }
    }
  }
}
```

> AI 沙箱内由 `mars-agent` 自动生成 opencode 的 MCP 配置（地址按 `/mcp` 拼接，token 为沙箱令牌），
> 不需要手工配置。沙箱令牌锚定 **agent 项目创建者**——管理员进入别人的沙箱时，
> `get_current_user` 看到的也是创建者，跨权限操作一律被拒绝。沙箱的启用、空闲缩容与
> 唤醒机制见 [ai_agent.md](ai_agent.md)。

### 2.3 手工冒烟（curl）

```bash
# initialize（Stateless 模式不下发 session id，后续请求无需回传）
curl -sS https://mars.example.com/mcp \
  -H "Authorization: Bearer $MARS_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"curl","version":"0"}}}'

# 列工具
curl -sS https://mars.example.com/mcp \
  -H "Authorization: Bearer $MARS_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'

# 调用工具
curl -sS https://mars.example.com/mcp \
  -H "Authorization: Bearer $MARS_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_current_user","arguments":{}}}'
```

`Accept` 必须同时包含 `application/json` 与 `text/event-stream`，否则服务端按规范拒绝。
响应可能是 JSON，也可能是 SSE（`event: message` + `data: {...}`），两者都是标准行为。

## 3. 鉴权与权限模型

### 3.1 三层校验

| 层 | 位置 | 作用 |
|---|---|---|
| ① 传输层登录 | `middlewares.LoginHTTP` + `biz.Authenticate` | 校验 `Authorization` 头，注入用户；失败 401 |
| ② 工具集裁剪 | `NewServer` 的两个预构建 `sdk.Server` 实例 | 管理员连接注册全量工具；普通用户连接不注册管理员专属工具，`tools/list` 从连接那刻就不同 |
| ③ 工具内 RBAC | `AccessBiz.Require*` / `Can*` | 项目/空间的真实权限边界，跨权限返回 `PermissionDenied` |

要点：

- **Stateless 下没有跨请求身份缓存**：`getServer` 每个请求都会被调用，按当前请求 ctx 里的用户
  选择 server 实例，所以不存在「initialize 是管理员、后续请求降权」的错配。
- **管理员专属工具双保险**：普通用户的 server 上根本没有这张工具表（`tools/call` 会被 SDK 按
  未知工具拒绝），handler 内还有 `requireAdmin` 纵深防御（MCP 在进程内直调 services，
  gRPC 的 `Authorize` / `Validator` 拦截器不会执行）。
- **身份即权限上限**：所有工具都受当前 token 的权限约束，不要尝试绕过；工具失败会以
  `isError=true` 明确回吐原因。

### 3.2 管理员 vs 普通用户

| 连接身份 | 可见工具数 | 差异 |
|---|---|---|
| 普通用户 | 23 | 19 只读 + 4 写 |
| 管理员 | 26 | 额外 3 个仓库管理工具（`list_repos`、`get_repo_custom_fields`、`update_repo_custom_fields`） |

`initialize` 的 `instructions` 也按身份下发：普通用户只拿到通用排查说明，管理员额外追加
仓库管理能力说明——避免普通用户的模型被告知「有一批你根本看不到的工具」。

## 4. 工具清单

### 4.1 只读诊断（`ReadOnlyHint=true`，客户端可免确认）

| 工具 | 寻址入参 | 用途 |
|---|---|---|
| `get_current_user` | 无 | 返回调用者身份 `{id,email,name,roles,is_admin,is_super_admin}` |
| `list_namespaces` | `name`（模糊，匹配空间名或项目名）+ 分页 | 可见命名空间及空间下项目概要 |
| `list_projects` | `namespace`（必填）+ `name`（模糊）+ 分页 | 空间下项目列表 |
| `get_project` | `project_id` | 完整项目模型（config、final_extra_values、repo.mars_config、休眠等） |
| `list_hibernate_windows` | 无 | 全局定时休眠时间窗定义 |
| `list_project_containers` | `project_id` | 项目当前 Pod/容器（`check_pod_status`/日志/exec 的入参来源） |
| `check_apply_status` | `project_id` | 最近一次部署结果 + 失败原因 + 崩溃日志 |
| `get_project_endpoints` | `project_id` | 访问地址（Ingress/Service/HTTPRoute）+ CPU/内存合计 |
| `get_project_resource_tree` | `project_id` | 资源拓扑树（Application→Ingress/Service→工作负载→RS→Pod） |
| `check_pod_status` | `namespace` + `pod` | 单个 Pod 是否 Running 及原因 |
| `list_namespace_events` | `namespace` | **Kubernetes 事件**（调度/拉镜像/OOM/探针失败等） |
| `list_operation_events` | 分页 + `search` + `all` | **mars 操作审计**（谁在何时改了什么） |
| `get_container_logs` | `namespace` + `pod` + `container`（多容器必填）+ `tail_lines` + `show_events` | 容器日志（默认尾部 1000 行，上限 128 KiB） |
| `get_pod_metrics` | `namespace` + `pod` | 单 Pod CPU（毫核）/内存（KB）实时用量 |
| `list_project_changelogs` | `project_id` + `limit` + `only_changed` | 最近部署变更记录（服务端固定最近 15 条） |
| `list_git_branches` | `project_id` + `include_all` | 仓库分支（默认只返回可部署白名单） |
| `list_git_commits` | `project_id` + `branch` + `limit` | 分支提交历史（不含作者/message/URL） |
| `get_commit_detail` | `project_id` + `branch` + `commit`（完整 SHA） | 提交完整详情（作者、说明、GitLab 链接） |
| `get_pipeline_status` | `project_id` + `branch` + `commit` | CI 流水线状态（按仓库的通过规则判定） |

### 4.2 写操作（不打 `ReadOnlyHint`，客户端会弹人工确认）

| 工具 | 寻址入参 | 语义与风险 |
|---|---|---|
| `apply_project` | `namespace` + `name`（定位已存在项目） | **只更新、不创建**。改 `git_branch`/`git_commit`/`config`/`extra_values`，支持 `dry_run`；`version` 为乐观锁；部署异步，返回成功 ≠ Pod 就绪 |
| `toggle_hibernate` | `project_id` + `enabled` + `window_names` | 只改定时休眠开关与所选时间窗，**不触发重部署**；`enabled=true` 时窗名至少一个且必须落在仓库允许集合内 |
| `restart_pod` | `namespace` + `pod` | 等价 `kubectl delete pod --force --grace-period=0`，由控制器重建；裸 Pod 不会自动回来；重建异步 |
| `exec_command` | `namespace` + `pod` + `container` + `command` + `timeout_seconds` | 容器内一次性命令（无 TTY）；需要 shell 特性请用 `["sh","-c","..."]`；退出码非 0 **不是**工具错误，需自行检查 `exit_code` 与 `truncated`/`timed_out`/`start_failed` |

写操作审批机制：**依赖 MCP 客户端的确认弹窗**（工具声明为非只读），服务端不做两阶段确认，
也没有审批单。配置宽松或恶意的客户端可以无视 hint 直接调用——但写操作仍受 mars RBAC 约束，
风险被限制在「调用者本就有权做的变更」。不提供 `delete_project` 等不可逆销毁能力。

### 4.3 管理员专属（仅管理员连接可见）

| 工具 | 类型 | 寻址入参 | 说明 |
|---|---|---|---|
| `list_repos` | 只读 | 分页 + `name`/`enabled`/`ai_agent` 过滤 | 仓库概要，**不含** `mars_config`；用于拿 `repo_id` |
| `get_repo_custom_fields` | 只读 | `repo_id` | 仓库配置：`values_yaml`、分支白名单、分区、休眠、自定义字段定义 `elements` |
| `update_repo_custom_fields` | 写 | `repo_id` + `upsert_elements` / `remove_paths` | 增量新增/整条覆盖/删除仓库级自定义字段；**只改 `elements`**，其余配置原样保留 |

注意点：

- `values_yaml` 可能含密钥/账号等敏感配置，**仅在管理员连接可见**，请勿转发或写入外部记录。
- 写库前工具会自校验元素定义（类型、默认值、`select_values`、`group` 长度），因为
  `ValidatorUnaryServerInterceptor` 在进程内调用链上不生效；非法即整次调用失败、配置不变。
- 删除或改坏被休眠配置 `replica_paths` 引用的字段会直接报错（事务不落库）。
- 改定义不会自动重部署：之后的重部署/定时唤醒会按**新定义**重新校验存量项目，
  改 type 或收窄 `select_values` 可能让不兼容项目部署失败；删除字段会让项目值在下次部署时被丢弃。
- 工具先读后写且无版本校验，**并发编辑同一仓库时后写者覆盖先写者**，请避免并发。

## 5. 返回值通用约定（最容易误读的四条）

1. **零值字段直接缺失**：服务端模型带 `omitempty`，缺字段 = 空串/0/false/空数组，**不是「未知」**。
   例：`StateContainer` 的 `ready=false` 时该字段不存在。
2. **枚举在模型 JSON 里是数字**：`deploy_status` `0=StatusUnknown`、`1=StatusDeploying`、
   `2=StatusDeployed`、`3=StatusFailed`。**唯一例外**是 `list_projects`，它的
   `deploy_status` 是字符串。
3. **入参寻址方式不同**：`list_namespaces` 用空间名；`list_projects` 用 `namespace`（+可选 `name`）；
   项目类工具一律用 `project_id`；Pod/容器类工具用 `namespace + pod (+容器)`。不要混用 ID 与名字。
4. **工具失败以 `isError=true` 回吐原因**；成功也不代表异步动作已完成（部署/重启都要再确认状态）。

补充：出参**不声明 `outputSchema`**（`tools/list` 保持精简），模型拿到的是与 mars 界面同源的原始 JSON。
所有能复用平台 gRPC 服务接口的能力都直接复用（门卫、审计、出参形状同源），只有三处直连 biz：
`list_projects`、`list_namespace_events`、`exec_command`。

## 6. 推荐排查路径

1. `list_projects` / `get_project` 定位项目 → `list_project_containers` 拿当前 Pod 与容器名。
2. 部署/发版后：`check_apply_status` 拿整体判定与失败原因 → `get_project_endpoints` 看入口与用量
   → `get_project_resource_tree` 看 Ingress/Service/工作负载链路。
3. `check_pod_status` 判断单 Pod 是否运行；`reason` 含 `not found` 说明 Pod 名错了或已消失。
4. `list_namespace_events` 看集群侧事件，按返回的 `object` 字段过滤目标 Pod，`count` 是累计次数。
   mars 平台操作审计是另一个工具 `list_operation_events`，别混用。
5. `get_container_logs` 定位应用层错误（多容器 Pod 必须传 `container`；看到「已截断」说明日志不完整）；
   `get_pod_metrics` 看资源画像。
6. 把线上版本与代码/发布对上号：`list_project_changelogs` → `list_git_commits` → `get_commit_detail`
   → `get_pipeline_status`（**CI 成功 ≠ 已部署**，部署结果仍看 `check_apply_status`）。

## 7. 配置模型与改配置

用户能通过 MCP 改的只有两块：

1. **`config`**：mars 配置片段（不是 helm values.yaml）；
2. **自定义参数（custom values）**：仓库后台声明的自定义元素，通过 `extra_values` 指定。

两者都由服务端按仓库配置的落点路径合并进 helm values 后再渲染，所以**用户看到的 config 不是
values.yaml 本身**。helm values.yaml 属仓库级、由管理员维护，**没有任何工具能直接编辑它**——
需要改的东西若不在上述两块里，说明 mars 没开放，应引导用户联系仓库管理员补自定义字段。

**固定顺序（先读后写，不可跳步）**：

1. `get_project` 读出当前 `config` 与 `final_extra_values`（后者给出可改元素清单，`path` 就是
   `extra_values` 的键，`repo.mars_config.elements` 给出类型约束）；
2. `apply_project` **只传要改的字段**——未传的分支/提交/配置、未列出的自定义元素都会自动沿用
   项目当前值。`config` 是**整段替换**（改一个键也要带整段）；`extra_values` 是**按 key 的增量**。

已知限制：`config`/`git_branch` 这类字符串字段无法用工具清空（没传与传空串无法区分）；
`extra_values` 不受此限（传空串即置空）。

## 8. 部署与实现说明

| 关注点 | 结论 |
|---|---|
| 多副本 | **必须 Stateless**：网关是多副本且 Service 无会话粘滞，有状态模式会把会话表留在单 pod 内存，跨副本请求返回 404 `session not found`（客户端常误报成 SSE 400） |
| 不支持的能力 | Stateless 下不提供 GET SSE 推送流（go-sdk 按规范返回 405，主流客户端视为「服务端不推送」继续工作）与 server→client 请求；本工具集不依赖它们 |
| 沙箱活跃度 | 每个已鉴权 MCP 请求都会刷新沙箱活跃度，工具调用期间不会被空闲关闭任务缩容；刷新失败只记日志，不阻断请求 |
| 路由顺序 | `/mcp` 必须在前端 SPA 兜底路由 `/{any:.*}` **之前**注册，否则会被吞成 `index.html`（与 `/api`、`/ws` 同一约束） |
| 日志上限 | `get_container_logs` 单次最多 128 KiB 并追加截断提示；`exec_command` 服务端上限 5 MiB、MCP 侧回吐约 128 KiB |

代码地图：

| 文件 | 内容 |
|---|---|
| `internal/services/mcp/mcp.go` | 包说明、`serverInstructions`、`Deps`、`NewServer`（双实例 + Stateless + 鉴权）、`Register` |
| `internal/services/mcp/tools.go` | 公共工具注册与实现（19 只读 + 4 写）、`requireAdmin`、管理员工具注册 |
| `internal/services/mcp/repos.go` | `list_repos` 入参 |
| `internal/services/mcp/custom_fields.go` | 仓库自定义字段读写、元素类型/默认值校验与投影 |
| `internal/services/wire.go` | `MCP` 依赖注入（services 层 gRPC 服务接口 + 少量 biz） |
| `internal/services/httphandler.go` | `RegisterMCPRoute` 定义（委托给 `mcp.Server`） |
| `internal/server/http.go` | 调用 `RegisterMCPRoute(router)`，必须在前端 SPA 兜底路由之前 |

### 新增工具的检查清单

1. 在 `tools.go` 的 `registerTools` 里 `sdk.AddTool`，写明 `Description`（入参/出参/语义/失败含义）；
   只读工具必须用 `readOnlyAnnotations`，写工具用 `writeAnnotations`。
2. 同步 `mcp.go` 的 `serverInstructions`（跨工具信息，如推荐路径、分层），不要与工具 description 复述同一段。
3. 管理员专属工具放 `registerAdminTools`，handler 入口先过 `requireAdmin`，并同步 `adminToolInstructions`。
4. 能复用 services 层 gRPC 接口就别直连 biz；直连 biz 必须在包注释里说明原因。
5. 补单测并更新本文的工具清单与权限表。

## 9. 已知限制与残余风险

- **客户端确认不是安全边界**：写操作的审批完全依赖 MCP 客户端的 hint 行为；要更强保证需升级为
  服务端两阶段 dry-run→confirm 或审批单，本批不做。
- **危险操作面**：`apply_project` 改线上部署、`restart_pod` 立即中断请求、`exec_command` 可改动容器内数据。
  调用前应向用户说明「将要做什么、影响哪个环境」。
- **数据时效**：k8s 事件默认只保留约 1 小时，事件缺失不代表历史上没出过问题；metrics-server
  采样暂缺时返回 `pod metrics unavailable`（不是 Pod 消失），可稍后重试。
- **日志/输出截断**：见第 8 节；截断标记意味着内容不完整，不能把截断点当结尾。
- **相对时间是展示串**：`event_at` / `date` 等字段是服务端生成的中文相对时间（如「3 天以前」），
  不要做字符串解析；解析时间请用 `created_at`（RFC3339）。
