---
name: mars-resource-model
description: 解释 mars 平台的资源模型——命名空间(Namespace)、项目(Project)、Pod、容器(Container)之间的从属关系，以及用户口中的"项目""容器"分别对应哪个对象；并给出 MCP 工具的入参寻址方式与返回值语义（工具清单以 MCP 实时暴露为准）。当需要把用户的自然语言指代落到具体对象、不确定某个操作用哪个工具或如何解读返回值时使用。
---

# mars 资源模型

mars 是一个面向 Kubernetes 的应用管理平台，资源层级自上而下为：

1. **Namespace（命名空间）** —— 权限的最小作用域。用户能访问哪些资源，取决于其是否属于该命名空间。
   命名空间分私有（仅创建者与成员可见）与公开两类。
2. **Project（项目）** —— 隶属于某个命名空间，是用户日常看到的"一个应用 / 一套部署"。
   项目下有若干**副本（replicas）**，每个副本对应一个运行中的 Pod。
3. **Pod** —— 实际运行的 Kubernetes Pod，由项目部署的 workload（通常是 Deployment）创建。
4. **Container（容器）** —— Pod 内的容器。同一个 Pod 可能有多个容器（主容器 + sidecar）。

## 排查时的对应关系

- 用户说"我的项目重启了" → 先定位到 Project，再查它的 Pod 列表。
- 用户说"某个容器报错" → 需要**同时**确定 Pod 名与容器名，二者缺一不可。
- 用户说"为什么一直重启" → 常见是 Pod 处于 CrashLoopBackOff，参见 crashloop-triage skill。

## 工具寻址方式（ID 与名字不要混用）

| 工具 | 入参 | 关键返回 |
|---|---|---|
| `list_namespaces` | `page` / `page_size` / `name`（模糊：空间名或空间内项目名） | `{items, count, page, page_size}`；items 内嵌该空间的项目概要（不含 config） |
| `get_current_user` | 无参数 | `{id, email, name, roles, is_admin, is_super_admin}`；返回**当前对话登录用户**的身份与角色 |
| `list_projects` | `namespace`（必填）+ `name`（模糊） | `{projects, total, page, page_size}`；项目 `id` 供后续工具使用 |
| `get_project` | `project_id` | 完整项目模型（直接回吐，未包裹）；含 `hibernate_enabled` / `hibernate_state` / `hibernate_window_names` |
| `list_hibernate_windows` | 无参数 | `{items:[{id,name,start,end,description,repo_names}]}`；`start`/`end` 为 HH:MM，`start>end` 表示跨零点 |
| `list_project_containers` | `project_id` | `{items:[{namespace,pod,container,is_old,terminating,pending,ready}]}` |
| `check_apply_status` | `project_id` | `{status, reason, containers, failures}`；status 数字枚举同 `get_project`；failures 失败时含 kind/workload/pod/container、reason、message、logs（CrashLoopBackOff 时是**上一次崩溃**的日志） |
| `get_project_endpoints` | `project_id` | `{urls:[{name,url,type,svc_name,ingress_name}], cpu, memory}`；没有暴露入口时 `urls` 为空数组（不是错误） |
| `get_project_resource_tree` | `project_id` | `{status, nodes:[{id,kind,name,namespace,status,labels,old}], edges:[{id,type,source,target}]}`；kind/edge type 见工具描述 |
| `check_pod_status` | `namespace` + `pod` | `{running, reason}` |
| `list_namespace_events` | `namespace` | `{events:[{type,reason,note,object,count,last_seen,container}]}` |
| `list_operation_events` | `page` / `page_size` / `search` / `all`（`all` 仅管理员生效） | `{page,page_size,items:[{id,action,username,message,old,new,has_diff,event_at,...}]}`；普通用户只能看到自己的事件 |
| `get_container_logs` | `namespace` + `pod` + `container`（多容器必填）+ `show_events` + `tail_lines`（默认 1000） | `{namespace,pod_name,container_name,log}` |
| `get_pod_metrics` | `namespace` + `pod` | `{cpu, memory, humanize_cpu, humanize_memory, time, length}` |
| `list_project_changelogs` | `project_id` + `limit` + `only_changed` | `{items:[...]}` |
| `list_git_branches` | `project_id` + `include_all`（默认只返回可部署分支白名单） | `{items:[{value,label,type,branch,git_project_id}]}`；`value` 即 `apply_project.git_branch` |
| `list_git_commits` | `project_id` + `branch` + `limit`（缺省 20、上限 100） | 同上结构；`value` 是完整 commit SHA |
| `get_commit_detail` | `project_id` + `branch` + `commit`（完整 SHA） | `{id,short_id,title,author_*,committer_*,web_url,message,committed_date,...}` |
| `get_pipeline_status` | `project_id` + `branch` + `commit`（完整 SHA） | `{status, web_url, jobs:[{name,status,stage_name}]}`；status=success\|failed\|running\|manual\|unknown |
| `apply_project`（写） | `namespace` + `name` 定位项目；`config` / `extra_values` / `git_*` / `dry_run` | `{yaml_files, project, dry_run}`；dry_run 时 project 为空 |
| `toggle_hibernate`（写） | `project_id` + `enabled` + `window_names`（启用时至少一个、须属仓库允许集合） | `{enabled, state, window_names}`；**只改开关，不触发重部署** |
| `restart_pod`（写） | `namespace` + `pod` | `{deleted, namespace, pod, message}` |
| `exec_command`（写） | `namespace` + `pod` + `container`（可选） + `command` + `timeout_seconds` | `{output, exit_code?, truncated, timed_out, start_failed, error?}` |

- `namespace` 一律用 `list_namespaces` 返回的 `name`（界面名，免 k8s 前缀）。
- `pod` / `container` 一律先用 `list_project_containers` 取**当前真实存在**的名字（Pod 名带随机后缀，不能拼）。

## 返回值语义（最容易误读的五点）

1. **零值字段会直接缺失**：所有出参都是带 `omitempty` 的标准 JSON（snake_case）。
   字段不存在 = 空串 / 0 / false / 空数组，不是"未知"。例如 `ready=false` 时 `ready` 字段根本不存在。
2. **枚举是数字**：`deploy_status` 0=StatusUnknown、1=StatusDeploying、2=StatusDeployed、3=StatusFailed。
   **唯一例外**：`list_projects` 的 `deploy_status` 是字符串（`StatusDeployed` 等）。
3. **数量单位**：`get_pod_metrics.cpu` 是毫核（250 = 0.25 核）；`memory` 是 KB（1000 字节，已取整）；
   需要展示时优先用 `humanize_cpu` / `humanize_memory`。
4. **成功 ≠ 完成**：`apply_project` 与 `restart_pod` 都是异步生效——返回成功只代表请求被接受。
   部署/重启后**优先**用 `check_apply_status`（一次拿到 status、reason 与失败容器清单）确认结果，
   需要看具体 Pod 或完整日志时再用 `list_project_containers` / `check_pod_status` / `get_container_logs`。
   任何工具失败都会以 `isError=true` 回吐原因（NotFound / PermissionDenied / 参数校验失败等）。
5. **轮询节奏**：部署后**先直接**调 `check_apply_status`，不要先 `sleep` 固定的大段时长（如 30–45s）。
   通常 10–20s 内就能拿到结果；若 `status=1`（Deploying），按 3–5 秒间隔再次查询，
   总上限约 20–30 秒；也可同时用 `list_project_containers` 看是否出现非 `is_old` 的新 Pod。
   超时仍无结论就如实说明"仍在部署"，不要编造成功。

## 配置模型（改配置前必读）

用户**不能**直接改 helm 的 values.yaml——mars 只开放两条改配置的通道，两者都是**片段**，
由服务端按仓库配置里声明的**落点路径**合并进 values 后再渲染部署：

1. **config**（mars 配置片段）—— 在项目配置页填写的 YAML 片段。
2. **自定义参数（custom values）** —— 仓库后台声明的字段，如 replicas、镜像 tag 这类
   radio / 数值输入项；用户只填值，不控制它落到 values 的哪个键。

helm 的 values.yaml 本身是**仓库级**的、由管理员在后台维护，没有任何工具能直接编辑它。

推论：用户想改的东西若既不在 config 片段里、也不是声明过的自定义参数，说明 mars 没有把它
开放给用户改——应引导用户联系仓库管理员在后台补一个自定义字段，**不要**把 helm 的原始 key
硬塞进 config（片段会被合到落点子树下，顶层键会落到错误层级，看起来"设上了"其实没生效）。

`get_project` 返回的 `final_extra_values` 是当前生效的自定义参数列表：其中 `path` 就是
`apply_project.extra_values` 的键。改配置的正确顺序是：先 `get_project` 读出 config 与
final_extra_values → 再 `apply_project` 只传要改的字段（config 是整段替换，extra_values 是
按键覆盖）→ 正式部署前可先 `dry_run=true` 预览。

## 回滚（用变更记录还原配置）

1. `list_project_changelogs`（`project_id`，最新在前）找到故障前最后一个正常版本；
2. 该记录的 `config` 与 `final_extra_values` 就是当时的配置状态；
3. 需要回退代码版本时：用 `list_git_branches` 确认分支可部署 → `list_git_commits` 取目标提交的
   完整 SHA → `get_commit_detail` 核对提交内容 → `get_pipeline_status` 确认该提交 CI 已通过；
4. `apply_project` 传 `config`（整段替换）+ `extra_values`（按 path 覆盖，未列出的沿用现状），
   需要同时回退分支/提交时再传 `git_branch` / `git_commit`；
5. 正式部署前先 `dry_run=true` 核对渲染结果，部署后用 `check_apply_status` 确认
   （需要看具体 Pod 时再补 `list_project_containers` / `check_pod_status`）。

注意：变更记录里的 `config` 是**当次配置片段**，不是完整 helm values；`date` 是中文相对时间，
解析时间用 `created_at`（RFC3339）。`config` / `git_branch` 无法通过工具清空（需到 mars 界面操作）。

## 汇报约定（必须带 ID）

向用户展示任何资源时，必须**同时给出 ID 与名称**——ID 是用户后续操作的寻址入口，只写名字用户无法准确定位：

| 资源 | 展示格式 | 后续工具用什么寻址 |
|---|---|---|
| 命名空间 | `空间名（空间 id=123）` | `namespace` = 空间名（不是空间 id） |
| 项目 | `项目名（项目 id=456）` | `project_id` = 项目 id |
| Pod | 完整 Pod 名（带随机后缀） | `pod` = Pod 名 |
| 容器 | 容器名 | `container` = 容器名 |
| 变更记录 | `版本 N（记录 id=...）` | 仅用于展示/对账 |

- 不要只写名称，也不要用「某个项目」「那个空间」这类模糊指代。
- 列表为空时也要说明筛选条件与所在空间（名称 + id）。
- 一次列出多个对象时，逐个带 ID，便于用户直接指代「第几个 / 哪个 id」。

## 权限边界（重要）

- 一切操作都在**命名空间**作用域内鉴权；mars 会用当前登录用户的身份复核权限。
- 你只能操作用户有权访问的命名空间下的对象，越权请求会被服务端拒绝（`PermissionDenied`）。

## 工具使用约定

- 先用列取类工具（`list_namespaces` / `list_projects` / `list_project_containers`）确定对象，再执行动作类工具。
- 动作类工具（`apply_project` / `toggle_hibernate` / `restart_pod` / `exec_command`）会真正改动环境。
  写工具会触发人工确认：调用前必须先向用户说明"将要做什么、影响哪个环境"，取得明确同意；
  工具调用会暂停并等待用户批准，不要在批准前假设已执行。`toggle_hibernate` 只切开关、不触发重部署；
  `restart_pod` 是立即删除、重建异步。
- 判断「入口通不通 / Service 有没有挂到工作负载 / 是否残留旧副本」用 `get_project_resource_tree`；
  需要访问地址与项目总用量用 `get_project_endpoints`；追查「谁改过配置」用 `list_operation_events`。
