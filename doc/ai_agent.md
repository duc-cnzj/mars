# AI Agent 说明（容器排查助手）

> 本文是 mars AI Agent（沙箱助手）的**权威说明**，由 `internal/services/aiagent.go`、
> `aiconfig.go`、`llmproxy.go`、`internal/biz/aiagent.go`、`internal/cronjob/cron_tasks.go`
> 的实现反推归纳。**维护约定**：启用流程、沙箱命名、令牌语义、空闲缩容策略、
> 下游接口（`/api/ai/*`、`/llm/v1/*`）任一变更时，必须同步更新本文；
> MCP 工具清单与鉴权细节以 [mcp.md](mcp.md) 为准，权限判定契约见 [access_control.md](access_control.md)。

## 1. 它是什么

AI Agent 是 mars 内置的「每人一个」容器排查助手：

- 用户点击顶栏星芒图标，mars 以**普通项目**的形式为当前用户部署一个沙箱；
- 沙箱内运行 `mars-agent`（opencode + `mai` 交互命令），既能对话，也能通过 MCP
  查询/操作 mars 里的命名空间、项目、Pod、日志、指标，甚至发起部署更新；
- 模型请求统一走 mars 的 `/llm/v1` 代理，**上游地址与 API Key 只存在于 mars**，
  沙箱永远拿不到；
- 沙箱空闲时自动缩容到 0，再次点击图标时唤醒并轮换会话令牌。

沙箱不是旁路系统：它复用正常的项目部署流水线（helm 渲染 → k8s 安装）、正常的
项目/Pod 权限门卫、正常的 MCP 与 WebSocket 通道。AI 只是多了一层「自动创建 +
自动唤醒 + 空闲回收」的编排。

```text
用户点击顶栏 AI
        │
        ▼
mars 网关 ── POST /api/ai/agent/enable ──▶ 共享部署编排（普通 project 流水线）
        │                                        │
        │                                        ▼
        │                              私有空间 mars-agent
        │                                └─ 沙箱项目 agent-<local>-<hash>
        │                                        │ 1 副本（空闲缩容到 0）
        │             ┌──────────────────────────┴─────────────────────────┐
        │             ▼                                                    ▼
        │      /llm/v1（模型代理，注入上游 Key）                  /mcp（工具能力）
        │      /api/ai/config（运行时配置）                       /ws（终端/日志）
        └──────────────────────────────────────────────────────────────────
```

## 2. 组成

| 组件 | 说明 |
|---|---|
| AI Agent 仓库 | 仓库管理里的一个特殊标记（`ai_agent`）。全局至多一个，启用后 AI 功能才可用；未配置或已禁用时顶栏入口点击即提示 |
| `mars-agent` 空间 | 逻辑名 `mars-agent`（实际 k8s 名带 `ns_prefix`），**私有**且由平台以超管身份创建；普通用户不可见，仅管理员可在空间列表中看到 |
| 用户沙箱项目 | 项目名由用户邮箱派生：`agent-<邮箱本地部分净化>-<邮箱 sha256 前 12 位>`；每个用户至多一个 |
| `mars-agent` 镜像 | 沙箱工作负载，内置 `opencode` 常驻服务与 `mai` 交互命令；就绪由 Deployment 的 readinessProbe（opencode `/health`）决定 |
| `/api/ai/config` | 沙箱启动时拉取的运行时配置（当前只有模型 id），mars 是唯一事实来源 |
| `/llm/v1/*` | OpenAI 兼容的模型代理；沙箱只认识 mars，不感知上游厂商 |
| `/mcp` | 容器排查与部署配置工具集，沙箱内由 `mars-agent` 自动注入 opencode 配置，见 [mcp.md](mcp.md) |
| 空闲缩容任务 | cron `agent_idle_reconcile`（每分钟），把超过 `ai.idle_timeout` 未活动的沙箱缩容到 0 |

## 3. 前置条件与配置

### 3.1 环境要求

- mars 需运行在 Kubernetes 环境（配置了 `kubeconfig`，或 Pod 内注入了
  `KUBERNETES_SERVICE_HOST/PORT`），否则无法部署/缩容沙箱；
- 空闲自动关闭还要求缓存驱动为 `memory` 或 `db`（`noCache` 下活跃度永远读不到，
  任务不会注册，避免把运行中的沙箱误判为空闲）。

### 3.2 标记 AI Agent 仓库（管理员）

在仓库管理中把某个仓库的 **「AI Agent 仓库」** 开关打开。沙箱侧组件（agent 源码、
Helm chart、镜像构建）统一在独立仓库维护：

| 内容 | 地址 |
|---|---|
| 仓库（源码 / chart / 镜像构建） | [mars-ai-agent](https://github.com/duc-cnzj/mars-ai-agent) |
| Helm chart | 仓库内 `charts/mars-agent/`（chart 名 `mars-agent`，单 release = 单 Deployment） |
| 默认镜像 | `registry.cn-hangzhou.aliyuncs.com/duc-cnzj/mars-agent:<tag>`；Docker Hub `docker.io/duccnzj/mars-agent:<tag>`（`<tag>` 取 chart 的 `image.tag`，默认 `latest`） |

> 仓库名、Go module 路径与 git 地址一致（`duc-cnzj/mars-ai-agent`）；chart、镜像与
> 二进制名是 `mars-agent`。

约定：

- 全局只能有一个：标记新仓库会顶替旧标记；关闭开关即清除自身标记；
- 该仓库必须处于**启用**状态，否则 AI 功能视为不可用；
- 仓库的 `mars_config.values.yaml` **必须引用 `<.UserToken>`**，通常写法：

```yaml
# 沙箱回连 mars 的令牌由部署流水线注入；未引用时入口会直接 400 拦截，
# 因为沙箱能起来但访问 mars（配置/MCP/LLM 代理）全部 401，属于无报错的软失败。
mars:
  token: "<.UserToken>"
```

启用入口会把这一项作为硬校验，缺失时返回可操作的提示。

### 3.3 AI 配置（`config.yaml`）

```yaml
ai:
  # 裸模型 id；agent 会写成 mars/<id>。留空时 /api/ai/config 返回 503，沙箱无法启动。
  model: "deepseek-chat"
  # OpenAI 兼容上游地址（如 DeepSeek）。沙箱统一请求 <gateway>/llm/v1，由 mars 转发。
  # 该地址不会下发给沙箱；留空或非法时 LLM 代理返回 503。
  base_url: "https://api.deepseek.com/v1"
  # 上游密钥，仅在 mars 转发时注入请求头，绝不下发沙箱。
  api_key: "sk-xxxx"
  # 空闲自动关闭阈值：默认 30m；负数表示关闭该能力；零值（未配置）按默认处理。
  idle_timeout: 30m
```

后台「系统设置 → AI 助手」只读展示上述四项（`ai.api_key` 掩码），仅**内置超级管理员**
可见；修改请改 `config.yaml` 后重启。这样设计的原因：沙箱 chart 只携带网关地址与
会话令牌，模型参数在 agent 启动时经 `/api/ai/config` 拉取，避免同一份配置在 mars 与
每个沙箱的 values 里各维护一份。

> 注意：令牌签发后不续期（12h）。默认 30m 空闲缩容会在令牌过期前先关闭沙箱，
> 唤醒时再轮换；如果把 `idle_timeout` 关掉或调到 12h 以上，长时间运行的沙箱会在
> 12h 后无法回连 mars。此时沙箱没有休眠，点击 AI 入口只会打开已运行的沙箱、
> 不会换令牌——需要重新部署该项目（或恢复空闲缩容，让它睡下后再唤醒）。

## 4. 用户使用流程

### 4.1 顶栏入口

顶栏的 AI 星芒按钮在功能启用时常驻；`GET /api/ai/status` 明确返回 `enabled=false`
时隐藏入口（查询失败保持可见，点击时再给出明确提示）。点击后的编排（对应
`frontend/src/features/ai/AIAgentLauncher.tsx`）：

1. `GET /api/ai/status`：仓库未配置/已禁用（`enabled=false`）→ 提示并结束；
2. 沙箱 `sleeping` / `absent` 或尚未部署 → `POST /api/ai/agent/enable`（后端幂等，
   内部会扩容到 1 或重建丢失的 Deployment）；
3. 轮询 `/api/ai/status`，直到 `sandbox=running`（就绪探针通过）；间隔 1.5s，
   超时 3 分钟（首次启动可能包含镜像拉取）；
4. 拉取项目详情，复用普通项目的详情弹窗——但只开放**命令行**一个 Tab，
   并自动在命令行里执行 `mai`，用户不用再手敲。

弹窗打开且页面可见时，前端每分钟调一次 `POST /api/ai/activity` 续活跃度：
用户正在看日志/思考时不会被空闲任务误缩容；后台标签页不续活，把资源让给
真正在用的人。停止续活后仍按 `ai.idle_timeout` 正常休眠。

为把冷启动等待挪到用户看不见的地方，布局挂载时还会对「最近 7 天用过且当前
`sleeping`」的沙箱做一次后台预热（`POST /api/ai/agent/enable`），同一浏览器
24 小时内最多一次；从未用过的用户不会被被动创建沙箱。

沙箱运行态（`agent.sandbox`）：

| 值 | 含义 |
|---|---|
| `running` | 副本数 > 0 且已有就绪副本，可进入命令行 |
| `starting` | 已扩容但尚无就绪副本（扩容中/滚动更新中） |
| `sleeping` | 副本数为 0，空闲休眠，需要唤醒 |
| `absent` | Deployment 不存在（项目行在、工作负载丢失），启用时会走更新部署重建 |

整个 `agent` 字段为 `null` 表示当前用户尚未部署沙箱。

### 4.2 场景入口（项目卡片 / 项目详情 / 容器日志）

入口分两类：项目卡片上是一枚常驻的轻量 AI 图标，详情弹窗与日志 Tab 只在需要时给提示条：

| 位置 | 条件 | 按钮 | 交给 AI 的上下文 |
|---|---|---|---|
| 工作台项目卡片 | AI 功能已启用（不区分部署状态） | 「问 AI」图标（点击不打开详情弹窗） | **只绑定**空间名 + 项目名作为默认上下文，不代发问题 |
| 项目详情弹窗 | `deployStatus=StatusFailed` | 「让 AI 诊断」（红色提示条） | 空间名 + 项目名，排查部署失败原因 |
| 容器日志 Tab | 普通项目（非 AI 沙箱） | 「让 AI 分析」 | 项目名 + 项目 id；无可用容器时同样展示 |

所有入口复用同一套唤醒编排，命令行都只自动执行交互式 `mai`（不使用 `mai run`）。

**项目卡片入口只绑定、不提问**：点击后命令行执行 `mai --bind '<空间>/<项目>'`，
`mars-agent` 复用当前目录最近会话（没有则新建），用 `noReply + synthetic` 注入一段
隐藏的绑定上下文（TUI 里看不到额外消息、也不触发模型回复），再 attach 同一会话。
用户进入 TUI 后自己输入第一条消息即可，不必再说明「问的是哪个项目」；模型在问题
未明确指定项目时默认按该绑定作答。绑定常驻 server 不可达时无法注入，`mai` 告警后
仍按普通 TUI 启动，不阻断使用。

**场景问题（详情/日志入口）**作为 `mai` 的位置参数随初始命令一起下发
（`mai '<问题>'`，空白折叠为单行、POSIX 单引号转义），由 `mars-agent` 在沙箱内完成
投递：常驻 server 可达时把问题写入当前目录最近更新的会话（没有则新建）并 attach
该会话，用户进入 TUI 时问题已发送、回复在流式输出；server 不可达时退回
`opencode --prompt`，由内核自己填充并提交。问题走 HTTP 请求体而不是终端按键注入，
因此不存在抢在 TUI 就绪前丢失、或同帧回车被当成粘贴内容的问题。问题只携带结构化的
定位信息，要求 AI 用 mars MCP 工具自查最近事件、容器日志与项目配置，不把原始日志
塞进输入。

> AI 页面（复用项目详情弹窗）标题栏下方常驻一条**绑定条**，显示当前绑定的范围，
> 支持两种粒度**多选**：整个「空间」（`--bind '<空间>'`，含该空间全部项目）或
> 「空间/项目」（`--bind '<空间>/<项目>'`），也可清除绑定；确认后命令行以多个 `--bind`
> 重建会话（历史对话仍复用同一会话），让新的绑定上下文生效，且不会重发进入时的首条问题。
> Pod 不参与绑定：滚动更新后 Pod 名会变，绑定到项目/空间、由 AI 通过 MCP 工具
> （`list_project_containers` / `check_pod_status` 等）现查更稳。

> 绑定条下方还有一条**常见问题**（三条固定文案）：「帮我部署到最新代码」「为什么容器
> 一直未就绪」「分析容器日志是否有错误」。点击只把完整问题**复制到剪贴板**（toast 提示
> 已复制），由用户自行粘贴到 `mai` 输入框发送——不重建终端会话、不模拟按键，也不发起
> 任何网络请求；按钮文案与复制内容都走 i18n，切换语言不影响语义。
>
> 复制出的第一条是**写操作**问题：用户发送后 AI 通过 MCP `apply_project` 部署最新提交，
> 由 MCP 客户端确认后执行，权限按当前用户的项目 ACL 校验；其余两条只读。

> 安全边界：场景问题作为 `mai` 的参数传入，不经过 shell 解释；即使 `mai` 启动
> 失败停留在 shell 提示符，问题文本也不会被当作命令执行。

### 4.3 接口一览

| 方法与路径 | 鉴权 | 说明 |
|---|---|---|
| `GET /api/ai/status` | 登录用户 | 返回 `enabled` 与 `agent`（含 `project_id` / `sandbox`），未部署时 `agent` 为 null |
| `POST /api/ai/agent/enable` | 登录用户 | 启用/唤醒当前用户的沙箱；幂等；返回同 status |
| `GET /api/ai/config` | 登录用户（沙箱令牌） | 返回 `{model}`；未配置 `ai.model` 时 503 |
| `POST /api/ai/usage` | 登录用户 | 上报使用漏斗事件（`entry_click` / `enable_start` 等白名单）；未知事件 400；204 |
| `POST /api/ai/activity` | 登录用户 | 刷新当前用户沙箱活跃度（页面可见时前端每分钟一次）；204 |
| `GET/POST /llm/v1/*` | 登录用户（沙箱令牌） | OpenAI 兼容代理，转发到 `ai.base_url` |
| `/mcp` | 登录用户（沙箱令牌） | MCP Streamable HTTP，工具集按身份裁剪 |

`/api/ai/agent/enable` 的可解释失败：

| 状态 | 场景 |
|---|---|
| 400 | 当前用户缺少 email；未配置/已禁用 AI Agent 仓库；values 未引用 `<.UserToken>` |
| 409 | 同一项目部署进行中（`ErrDeployInProgress`）；`mars-agent` 空间被非系统创建者占用；并发首次启用时另一个请求正在初始化空间 |
| 500 | 仓库/项目/空间查询失败、部署失败等真实内部错误 |

`POST /api/ai/agent/enable` 需要用户邮箱作为身份锚点（缺失 email 时返回 400）；
`GET /api/ai/status` 在邮箱为空时按「尚未部署」返回（`agent=null`）。

## 5. 权限与身份

- **空间级 API 不放开**：`mars-agent` 是私有空间，普通用户调命名空间级接口仍被拒绝
  （不因为「是我的沙箱」而外溢到空间维度）；
- **项目/Pod 级自助例外**：按项目 ID、项目名（空间名 + 项目名）、Pod 名寻址的入口，
  在空间 ACL 拒绝后，仅放行**当前用户邮箱派生的那个沙箱项目/其 Pod**；他人项目、
  他人 Pod 一律 `PermissionDenied`。按「空间名 + 项目名」寻址时，项目不存在也统一
  返回 403，不通过 403/404 的差异泄露私有空间内是否存在该项目；
- **令牌身份恒为创建者**：部署/唤醒签发的会话令牌绑定沙箱项目创建者，而不是触发
  这次操作的请求人。管理员进入别人的沙箱重新部署/唤醒，令牌仍是创建者身份，
  不会把管理员权限带进沙箱；
- 沙箱令牌是 12h 的 access token，usage 为 `deploy-token:<project_id>`（按项目轮换，
  只保留最新一份）；部署或唤醒时签发新令牌并写入沙箱 Secret，成功后撤销旧令牌；
  唤醒时若 Secret 写入失败，会尽力撤销刚签发的新令牌，避免留下孤儿凭据。

## 6. LLM 代理与 MCP

### 6.1 LLM 代理

沙箱只知道 `<gateway>/llm/v1`，由 mars 转发到 `ai.base_url`：

- 入口校验：AI Agent 仓库已配置且启用 + 当前用户的沙箱项目存在，否则 403；
  仓库/空间/项目查询本身失败时返回 500；
- 转发时**剥离沙箱的 `Authorization`**，换成 `Bearer <ai.api_key>`；同时剥离 `Cookie`，
  补齐标准 `X-Forwarded-*` 头——沙箱令牌绝不透传给上游；
- 使用 `httputil.ReverseProxy`，`FlushInterval=-1`，天然支持 SSE 流式响应；
  等待上游响应头超时 60s，响应体读取不受限；
- 每个通过准入的请求都会刷新沙箱活跃度，用户对话期间不会被空闲任务缩容。

### 6.2 MCP

沙箱内 opencode 的 MCP 配置由 `mars-agent` 自动生成（地址为 `<gateway>/mcp`，
令牌为沙箱令牌），无需手工接入。工具清单、只读/写分类、连接身份裁剪与错误帧约定
见 [mcp.md](mcp.md)。每次 MCP 调用同样刷新沙箱活跃度。

## 7. 空闲缩容与唤醒

- 任务 `agent_idle_reconcile` **每分钟**执行；仅在 k8s 环境、`idle_timeout > 0`、
  缓存驱动为 `memory`/`db` 时注册；
- 活跃度由 `LLM 代理 / MCP 调用 / /api/ai/config 拉取 / 启用唤醒 / /api/ai/activity
  心跳` 刷新（TTL = `ai.idle_timeout`）；心跳仅在弹窗打开且页面可见时发送，
  仅查看日志（后台标签页）不会刷新；
- 活跃窗口过期且副本数 > 0 的沙箱被缩容到 0；缩容与唤醒/部署**共用项目级锁**，
  缩容前在锁内二次复查活跃度，宁可漏关也不误关正在使用的沙箱；
- 唤醒流程：点击 AI → `enable` → 抢同一把项目锁 → 判断 Deployment 运行态 →
  已缩容则**先换发令牌写入 Secret，再扩容到 1**（否则睡过 12h 的沙箱会带着过期
  令牌启动、拉配置 401）；Deployment 丢失则按已有项目配置走更新部署重建，
  继承分支/提交/配置/自定义元素，不会重置成仓库默认值；
- 活跃度查询/缩容失败只记日志并跳过本轮，下一轮重试。

## 8. 可观测性

| 指标 | 类型 | 含义 |
|---|---|---|
| `mars_agent_sandbox_transition_total{action}` | Counter | 沙箱状态迁移次数：`action=sleep`（空闲缩容）/ `action=wake`（唤醒扩容） |
| `mars_agent_llm_proxy_requests_total` | Counter | 通过准入并实际转发的 LLM 代理请求数（被拒绝的不计） |
| `mars_agent_usage_event_total{event}` | Counter | AI 使用漏斗事件数（`entry_click` / `status_disabled` / `enable_start` / `enable_ok` / `enable_fail` / `sandbox_ready` / `sandbox_open`；标签取白名单，基数有界） |
| `mars_agent_open_latency_seconds` | Histogram | 从点击入口到沙箱就绪的耗时（仅 `sandbox_ready` 事件带合法时长时记录） |

审计事件（`AuditLogWithRequest`）：

| 动作 | 操作人 | 内容 |
|---|---|---|
| 自动创建 `mars-agent` 空间 | 触发启用的用户 | 空间 ID / 名称 |
| 唤醒并轮换会话令牌 | 触发唤醒的用户 | 项目 ID / 名称（不记录令牌原文） |
| 空闲缩容到 0 | `system` | 项目 ID / 名称 / 创建者 |

## 9. 故障排查

| 现象 | 可能原因与处理 |
|---|---|
| 点击 AI 提示「AI 功能未启用」 | 未标记 AI Agent 仓库，或该仓库被禁用；到仓库管理打开标记并确保仓库启用 |
| 启用返回 400「缺少 `<.UserToken>`」 | AI Agent 仓库的 values 未写 `mars.token: "<.UserToken>"`；补上后重试 |
| 启用返回 409「命名空间正在初始化」/ 部署进行中 | 并发首次启用或同一项目正在部署；稍后重试即可 |
| 启用返回 409「命名空间已被非系统创建者占用」 | 已存在同名 `mars-agent` 空间但不是平台创建；平台不会接管，需管理员确认后改名/清理再重试 |
| 启用返回 409「项目已被其他仓库占用」 | 该用户的沙箱项目绑定在旧 AI Agent 仓库；管理员切换仓库后，需清理旧项目或恢复原仓库标记 |
| 一直停在 `starting` | 看沙箱 Pod 日志：镜像拉取失败、readinessProbe 未通过，或 `ai.model` 未配置导致 agent 拉配置失败 |
| 沙箱内访问 mars 全部 401 | 令牌过期（超过 12h）或 values 未引用 `<.UserToken>`；沙箱已休眠时重新点击 AI 入口会唤醒并轮换令牌，持续运行超过 12h 的沙箱需要重新部署 |
| `/llm/v1` 返回 503 | 未配置/非法的 `ai.base_url`；返回 403 则是 AI 未启用或当前用户尚未部署沙箱（先调 `enable`） |
| 沙箱总是被关闭 | 超过 `ai.idle_timeout` 无 LLM/MCP/配置请求；调大阈值或设为负数关闭（注意 12h 令牌上限） |
| 普通用户看不到 `mars-agent` 空间 | 设计如此：空间仅管理员可见，普通用户通过顶栏入口按项目 ID 打开自己的沙箱 |

## 10. 相关文档

- 沙箱侧 agent（源码 / chart / 镜像）：[mars-ai-agent](https://github.com/duc-cnzj/mars-ai-agent)
- MCP 端点与工具清单：[mcp.md](mcp.md)
- 权限判定完整契约（含 AI 沙箱自助例外）：[access_control.md](access_control.md)
- 配置项说明：[config_example.yaml](../config_example.yaml) 的 `ai` 段
