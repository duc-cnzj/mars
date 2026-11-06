---
name: crashloop-triage
description: CrashLoopBackOff（Pod 反复重启）的标准排查步骤——先确认 Pod 状态与原因，再看事件，最后看容器日志，逐层收敛根因。当用户报告 Pod/容器反复重启、起不来、不断重启时使用。
---

# CrashLoopBackOff 排查套路

按由外到内的顺序收敛，每一步都先拿证据再下结论。

## 运行环境边界（先对齐，少走弯路）

- 当前对话**没有 shell、没有 kubectl / helm，也没有集群凭证**：不要尝试用命令行直连 Kubernetes API，
  所有集群操作只能经 MCP 工具（工具清单以 MCP 实时暴露为准）。
- `exec_command` 只能进入**已运行**的容器；Pod 处于 Pending / ImagePullBackOff 时无容器可执行。
- 只能看到当前登录用户有权访问的命名空间，越权请求会被服务端以 PermissionDenied 拒绝。

## 第 0 步：先拿项目级结论（最快）

调用 `check_apply_status`（入参 `project_id`）。它按 mars 界面同款判据返回最近一次部署的实际运行结果，
一次调用就能判断整个项目是否失败，比逐个 Pod 探测快：

- `status` 数字枚举同 `get_project`（2=正常、3=失败）；失败/进行中时 `reason` 是人类可读描述。
- `failures` **仅在失败时出现**，每条含 `kind/workload/pod/container`、`reason`
  （CrashLoopBackOff / ImagePullBackOff / Error / OOMKilled / Evicted）、`message` 与 `logs`。
  CrashLoopBackOff 时 `logs` 是**上一次崩溃**的日志，可直接定位首条错误。
- 拿到 failures 里的 pod/container 后，可跳过第 1 步直接进第 3 步取实时日志。

## 第 1 步：确认具体 Pod 状态

若第 0 步已定位到具体 pod/container，可跳过本步。否则调用 `check_pod_status`（入参 namespace + pod）。

- 返回**只有** `{running, reason}` 两个字段：
  - `running=true` → `reason` 为空。
  - `running=false` → `reason` 是 k8s 给出的人类可读原因（如 `CrashLoopBackOff`、`ImagePullBackOff`、`Evicted`、`pod not running.`）。
  - Pod **不存在**时同样返回 `running=false`，`reason` 含 `not found` 文本（这不是工具错误，多半是 Pod 名写错或 Pod 已被替换——回去用 `list_project_containers` 重新取当前 Pod 名）。
- ⚠️ 本工具**不返回**重启次数、容器退出码、容器级状态。不要编造这些数据。

按 reason 分流：

- `CrashLoopBackOff` / `BackOff` → 继续第 2、3 步。
- `ImagePullBackOff` / `ErrImagePull` → 看第 2 步事件里 `Failed` 的 `note`：区分认证失败、
  镜像 / tag 不存在、仓库不可达。容器尚未创建，不能拉日志或 exec；镜像 tag 属项目配置、
  仓库凭据由管理员维护，先用 `get_project` 核对镜像相关配置，不要尝试在容器内排查。
- `Pending` → 看第 2 步事件：`FailedScheduling` 看资源不足 / 节点亲和 / 污点，`FailedMount`
  看 PVC / Secret 缺失。Pod 未运行不能 exec；需要事件文本时用 `get_container_logs` +
  `show_events=true`（Pending 下返回的是事件，不是容器日志）。
- 与 `OOMKilled` 相关的重复重启 → 见第 2 步 `OOMKilling`。指标是**实时采样**，OOM 重启后
  读数可能很低，不能据此排除内存问题。
- `Unhealthy` → 探针失败：对照启动日志与探针的路径 / 端口，判断是启动慢、路径写错还是
  依赖不可用。

## 第 2 步：看事件

调用 `list_namespace_events`（入参 namespace）。它返回 `{events:[{type, reason, note, object, count, last_seen, container}]}`。

- 返回的是**整个命名空间**的事件，**顺序不保证**：必须按 `object`（形如 `Pod/xxx-abc`）过滤出目标 Pod 的事件。
- `count` 是累计发生次数——它才是判断“反复重启/反复失败”的依据。
- `last_seen` 可能为空；k8s 事件默认只保留约 1 小时，事件缺失**不等于**历史上没出过问题。
- `container` 是关联字段路径（如 `spec.containers{app}`），不是容器名。

重点 reason：

- `OOMKilling` / `OOMKilled` —— 内存超过 limit 被杀；用 `get_pod_metrics` 对照实际用量（注意 `memory` 单位是 KB）。
- `Back-off restarting failed container` —— 容器反复启动失败，`count` 即重启次数。
- `FailedMount` / `FailedScheduling` / `Unhealthy` —— 挂载、调度、探针失败。

## 第 3 步：看日志

调用 `get_container_logs`（入参 namespace + pod；多容器 Pod **必须**传 container），定位第一条 FATAL / PANIC / 连接失败 / 启动参数错误。

- 返回 `{namespace, pod_name, container_name, log}`，其中 `log` **默认只包含最后 1000 行**（传 `tail_lines` 可调大/调小），不是全量。
- 超过 128 KiB 会被截断，末尾出现“...[日志过长，已截断，完整日志请在 mars 界面查看]”——看到该标记就说明日志不完整，不能把截断点当成日志结尾。
- `container_name` 只是回显入参；省略 container 时为空，不代表服务端替你选了哪个容器。
- Pod 处于 Pending 时加 `show_events=true`，`log` 里返回的是该 Pod 的**事件文本**（每行一条），不是容器日志。

容器刚崩溃、当次日志为空时：回到第 2 步用事件判断原因（OOM 看内存，启动即退看入口命令与配置），
不要因为日志为空就断言“没有错误”。

## 入口 / 服务链路不通

- `get_project_endpoints`（`project_id`）：确认是否有暴露入口。`urls` 为空数组 = 没有入口，不是工具失败；
  返回的 `cpu` / `memory` 是项目全部 Pod 的合计用量。
- `get_project_resource_tree`（`project_id`）：看 Application → Ingress/Service → Deployment 等 → ReplicaSet → Pod
  的连线（edge type=owner/selector/route），定位「入口没路由 / Service selector 没挂上工作负载 / 残留旧副本」。
  只看当前容器列表仍用 `list_project_containers`。

## 怀疑人为变更

- 部署与配置变更：`list_project_changelogs`（`project_id`，最新在前）。
- 平台操作审计：`list_operation_events`（谁在什么时候改了什么；普通用户默认只看自己的事件，管理员传 `all=true` 才展开全部）。
  注意它**不是** Kubernetes 集群事件，集群侧问题仍看第 2 步。
- 代码/CI 侧：`list_git_commits` 核对分支最新提交、`get_commit_detail` 看提交内容、
  `get_pipeline_status` 确认该提交流水线是否通过。

## 第 4 步：验证假设（需先征得用户同意）

若怀疑是网络、依赖或配置问题，可用 `exec_command` 在容器内做**最小探活**（如 `["sh","-c","nc -z host port"]`）。

- 返回 `{output, exit_code?, truncated, timed_out, start_failed, error?}`：
  - `exit_code` **仅在命令跑完且退出码非 0 时出现**；字段缺失 = 退出码 0。
  - `timed_out=true` = 超过超时（缺省 60 秒）被终止；`start_failed=true` = exec 根本没启动（如命令不存在）。
  - 退出码非 0 **不是工具错误**，要自己看 `exit_code` 与标志位。
- 每次只验证一个假设，并先向用户说明将要执行的命令，取得明确同意后再调用。
- 写工具会触发人工确认：用户批准后才会真正执行，不要在批准前假设已执行。

## 可选动作（同样需先征得用户同意）

根因定位后，按需提出（不要自动连续调用）：

- `restart_pod`：强制删除 Pod 交由控制器重建。**删除是立即的、重建是异步的**——返回 `deleted=true` 只代表删除请求被接受，需再用 `list_project_containers` / `check_pod_status` 确认新 Pod 就绪；裸 Pod 不会自动回来。
- `apply_project`：改配置/分支后重新部署。先 `get_project` 读当前配置，先 `dry_run=true` 预览，再正式部署；
  部署后用 `check_apply_status` 确认结果。
- `toggle_hibernate`：只切换项目级定时休眠开关，**不触发重部署、不会立即缩容**——不要把它当作"重启 / 停服务"的手段。

## 汇报约定

- 给出：现象 → 证据（来自哪个工具、哪个字段）→ 根因假设 → 建议动作。
- 报告中必须带上对象定位信息：空间名 + 空间 id、项目名 + 项目 id、完整 Pod 名与容器名，方便用户直接操作。
- 没有证据时不要断言根因；不确定就明确说“需要进一步验证”。
