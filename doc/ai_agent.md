# AI 对话说明（容器排查助手）

> 本文是 mars AI 能力的**权威说明**：服务端 AI 对话（`internal/services/aichat/*`，
> 依赖 `internal/services/mcp` 与内嵌 skills）。
> **维护约定**：启用条件、对话接口与 SSE 事件、写工具确认票据、`/api/ai/*` 下游接口
> 任一变更时，必须同步更新本文；MCP 工具清单与鉴权细节以 [mcp.md](mcp.md) 为准，
> 权限判定契约见 [access_control.md](access_control.md)。

## 1. 它是什么

AI 对话是 mars 内置的容器排查助手，**完全运行在 mars 服务端**：

- 浏览器把完整对话历史随请求发来，mars 在服务端编排「上游模型 ↔ MCP 工具」循环，
  以 SSE 逐帧回吐模型文本、工具调用与工具结果；
- 工具能力与 `/mcp` 同源：只读工具直接执行，写工具必须经过人工确认；
- 身份就是当前登录用户，每次工具调用都在进程内 MCP 会话里按该用户身份执行，
  越权由既有 RBAC 拒绝；
- 上游 `api_key` 只存在于服务端，前端永远拿不到；
- **不需要沙箱、不需要 Kubernetes、不落任何服务端会话状态**：对话历史保存在浏览器，
  确认票据由 HMAC 签名自证，多副本部署天然安全。

```text
浏览器（对话面板）
   │  POST /api/ai/chat（携带历史与可选的关注范围）
   ▼
mars 服务端 ──▶ 上游 OpenAI 兼容接口（流式）
   │                ▲
   │  SSE 增量 ◀────┘
   ├── 只读工具 ──▶ 进程内 MCP 会话（当前用户身份）
   ├── 写工具   ──▶ 暂停并下发 need_confirm（HMAC 票据）
   │                    │
   │  POST /api/ai/chat/confirm（票据 + 已推进的 messages）
   ▼                    ▼
继续同一轮对话（批准则执行，拒绝则回填工具结果）
```

## 2. 启用条件与配置

对话功能只依赖三项 AI 配置，齐全即可用：

```yaml
ai:
  # 模型 id（原样发给上游，例如 deepseek-chat）。
  model: "deepseek-chat"
  # OpenAI 兼容上游地址。
  base_url: "https://api.deepseek.com/v1"
  # 上游密钥，仅在服务端转发时注入请求头，绝不下发浏览器。
  api_key: "sk-xxxx"
```

后台「系统设置 → AI 助手」只读展示上述三项（`ai.api_key` 掩码），仅**内置超级管理员**
可见；修改请改 `config.yaml` 后重启。任一为空时 `/api/ai/chat/status` 返回
`enabled=false`，聊天接口返回 503。

## 3. 接口一览

| 方法与路径 | 鉴权 | 说明 |
|---|---|---|
| `GET /api/ai/chat/status` | 登录用户 | `{enabled, model}`；未配置时 `enabled=false` |
| `POST /api/ai/chat` | 登录用户 | 发起一轮对话，响应为 `text/event-stream` |
| `POST /api/ai/chat/confirm` | 登录用户 | 提交写工具确认结果并继续同一轮对话（同样是 SSE） |
| `POST /api/ai/usage` | 登录用户 | 上报使用漏斗事件（`entry_click` / `chat_first_token` / `chat_done` / `chat_error`）；未知事件 400；204 |
| `/mcp` | 登录用户 | 标准 MCP（Streamable HTTP），第三方客户端可直接接入，见 [mcp.md](mcp.md) |

请求体、消息条数、单条 `content` 与工具参数**不设服务端大小上限**，超长内容原样透传
（模型上下文超限由上游自行报错）；`system` 角色禁止由客户端传入（system prompt 只能由
服务端注入），`role=tool` 必须携带 `tool_call_id`，`tool_calls` 必须有 `id` 与
`function.name`。

请求体可携带 `locale`（前端当前界面语言，如 `zh-CN` / `en`）：服务端据此在 system
prompt 里要求模型用同一语言回答（工具名、命令、资源名与错误原文保持原样）。未传或
未知语言按简体中文处理；`/api/ai/chat/confirm` 必须一并回传——续跑时会重新拼装
system prompt，语言偏好不能丢。

请求体还可携带 `bindings`（`[{namespace, project}]`）：场景入口带来的默认关注范围，
非空时写进 system prompt；它只是上下文提示，不改变任何权限判定。

### 3.1 SSE 事件

| event | data | 说明 |
|---|---|---|
| `meta` | `{model}` | 流开始 |
| `delta` | `{text}` | 模型增量文本 |
| `tool_call` | `{id,name,arguments,read_only}` | 开始执行工具 |
| `tool_result` | `{id,name,ok,content}` | 工具结果（内容原样透传，不做服务端截断） |
| `need_confirm` | `{ticket,id,name,arguments,summary,messages}` | 遇到写工具暂停；`messages` 是服务端已推进的完整对话状态，前端原样回传 |
| `error` | `{message}` | 终止帧；上游原始错误只进服务端日志，不回传 |
| `done` | `{rounds,elapsed_ms}` | 正常结束 |

心跳：等待上游首 token 或工具执行期间每 15s 写一帧 `: ping`；响应头带
`Cache-Control: no-cache` 与 `X-Accel-Buffering: no`，保证网关不缓冲、逐帧到达浏览器。

## 4. 工具与写工具确认

- 只读工具（MCP `readOnlyHint=true`）直接执行；`load_skill` 是本地工具，按需加载
  内嵌的 `SKILL.md` 正文（目录在 system prompt 里，正文不预塞上下文）。
- 写工具（`apply_project` / `toggle_hibernate` / `restart_pod` / `exec_command` /
  `update_repo_custom_fields`，按当前身份裁剪，system prompt 只列本次会话可见的那些）
  立即暂停并下发 `need_confirm`。票据是 HMAC-SHA256
  签名的无状态凭据，绑定 `email + tool_call_id + tool_name + 参数哈希 + 过期时间
  （5 分钟）`；密钥由 `ai.api_key` 派生，每个副本一致。
- 前端把 `need_confirm.messages` 与 `ticket` 原样回传 `/api/ai/chat/confirm`。
  服务端只校验并执行**票据指向的那一个调用**：签名 → 有效期 → 用户身份 →
  对话中该调用的名称/参数哈希 → 同批前序调用都已有工具结果。客户端历史里其他
  `tool_calls` 永不执行，防止伪造调用绕过确认。
- 一张票据只授权一个调用：批准/拒绝后继续处理后续调用，遇到下一个写工具会再次
  下发 `need_confirm`；拒绝时把「用户拒绝执行该操作」作为工具结果回填给模型。
- 单次请求最多 30 轮「模型 → 工具」，超过即返回 `error`，不编造完成。

## 5. 前端（mars-ui）

前端通过 `/api/ai/chat/status` 决定是否展示入口。对话内容保存在浏览器
localStorage（按登录用户 email 分键），**不写入 mars 数据库**；刷新页面后可继续历史
对话，服务端无状态。本地按会话分段保存（每段带标题、最后更新时间与当时的绑定）：
「新对话」另起一段并保留上一段，「历史对话」面板可切回旧会话并恢复其消息与绑定，
每段会话的删除都需二次确认。本地记录有清理机制：超过 7 天未更新（以最后更新时间为准）
的会话在读取时直接清除；每个用户最多保留 30 段会话，单段最多 200 条消息
（从最近的 user 消息边界截断，避免把工具调用对切散）；空会话不落盘。

面板顶部的「当前绑定」决定 system prompt 里的默认关注范围。**切换绑定会另开一段新会话**
（旧会话保留在历史里）：旧消息是按旧绑定范围生成的，若继续沿用会被模型当作既定上下文，
导致后续操作仍落在旧空间/项目上。生成中或等待写工具确认时绑定条禁用，不能中途换绑。

发送区上方提供「免确认」勾选框：开关状态按登录用户持久化到 localStorage，刷新或重新
打开面板后保持上次选择。开启后前端收到 `need_confirm` 会自动回传 `approved=true`
继续执行，不再逐条弹确认条。该模式只改变交互节奏，**不绕过服务端票据校验**——签名、
有效期、用户身份与参数哈希仍逐项验证，客户端历史里的其他 `tool_calls` 依旧不会执行。

「停止」使用 `AbortController` 中断 SSE，已生成的文本保留，用户可继续输入。

## 6. 权限与身份

- 每次工具调用都继承当前登录用户身份，与 `/mcp` 使用同一套工具裁剪与 `Require*`
  门卫；管理员/普通用户看到的工具集不同，越权调用返回 `PermissionDenied`；
- 客户端历史里的 `tool_calls` **永不执行**，只作为上下文，防止伪造调用绕过确认；
- 写工具必须通过票据校验，票据绑定用户邮箱，不可跨用户复用；
- 工具结果原样透传，不做服务端截断（上下文超限由上游模型自行报错）；
- 未登录 → 401；AI 未配置 → 503；请求结构不合法（非法 JSON / `system` 角色 /
  缺 `tool_call_id` 等）→ 400。

## 7. 与 MCP 的关系

服务端对话与 `/mcp` 共用同一份工具定义、身份裁剪与 instructions（资源模型、
排查口径、返回值约定）。会话通过进程内 MCP transport 建立，工具 handler 拿到的
就是当前请求的用户上下文，因此两条链路的权限语义完全一致。

第三方 MCP 客户端（opencode、Claude Code、Cursor 等）仍可带 mars bearer token
直连 `/mcp`，接入方式见 [mcp.md](mcp.md)。

## 8. 可观测性

| 指标 | 类型 | 含义 |
|---|---|---|
| `mars_ai_chat_requests_total{result}` | Counter | 服务端对话请求数：`result=ok`（含正常结束与暂停待确认）/ `error` |
| `mars_ai_chat_first_token_seconds` | Histogram | 从请求开始到上游首个 token 增量的耗时 |
| `mars_ai_chat_tool_calls_total{tool,result}` | Counter | 工具调用次数；`tool` 取工具名，`result=ok/error` |
| `mars_ai_usage_event_total{event}` | Counter | 前端上报的使用漏斗事件（`entry_click` / `chat_first_token` / `chat_done` / `chat_error`；标签取白名单，基数有界） |

日志约定：不打印对话内容与工具参数（可能含敏感信息）；上游非 2xx 只读前 4 KiB
记日志，向客户端只回通用错误。

## 9. 故障排查

| 现象 | 可能原因与处理 |
|---|---|
| 入口不显示 / `/api/ai/chat` 返回 503 | `ai.base_url` / `ai.api_key` / `ai.model` 未配齐；补齐后重启 |
| `/api/ai/chat/confirm` 返回 400「对话状态与确认票据不匹配」 | 前端没有原样回传 `need_confirm.messages`，或票据超过 5 分钟过期；重新发起该轮对话 |
| `/api/ai/chat` 返回「工具调用轮数超过上限」 | 模型在工具间循环；引导用户缩小问题范围，或检查工具是否持续报错 |
| 工具调用返回 `PermissionDenied` | 当前用户的 RBAC 不允许该操作；与 `/mcp` 行为一致，不可绕过 |
| SSE 长时间无输出 | 检查网关是否缓冲（需 `X-Accel-Buffering: no`）；服务端每 15s 有 `: ping` 心跳 |
| 换副本后仍能继续对话 | 设计如此：服务端无状态，历史随请求携带，票据自证有效 |

## 10. 相关文档

- MCP 端点与工具清单：[mcp.md](mcp.md)
- 权限判定完整契约：[access_control.md](access_control.md)
- OpenAPI：[openapi.yaml](openapi.yaml)
