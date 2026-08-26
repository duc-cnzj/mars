# 权限对照表（访问控制契约）

> 本表是 mars 全部传输接口（gRPC + HTTP）权限要求的权威清单，由 `internal/services/` 各服务实现反推归纳。
> **维护约定**：改动任一服务的鉴权逻辑（登录白名单 / Authorize 门禁 / AccessBiz 调用）时，必须同步更新本文档，防止契约与实现漂移。

## 1. 权限判定模型（6 级）

| 等级 | 判定 | 说明 |
|---|---|---|
| 🆓 公开 | 命中 `biz.IsPublicMethod` 白名单（登录拦截器放行） | 无需任何凭证即可调用 |
| 🔑 登录即可 | gRPC 登录拦截器校验 Bearer token | 任意有效 token 用户 |
| 🛡️ 命名空间级 | `RequireNamespaceAccessByName` / `RequireNamespaceAccessByID` / `RequireProjectAccess` | 公开空间任意登录；私有空间仅 admin / 创建者 / 成员 |
| 🏠 owner 专属 | `RequireNamespaceOwner` | 仅命名空间创建者 |
| ⭐ admin 专属 | `RequireAdmin`（Authorize 门禁） | 仅 admin 角色；fullMethodName 精确命中 allowlist 时豁免 |
| 📄 文件所有者/admin | `RequireFileAccess` | `fil.Username == user.Name` 或 admin |

## 2. 判定载体（AccessBiz 方法 → 权限等级）

| AccessBiz 方法 | 等级 | 语义 |
|---|---|---|
| `RequireNamespaceAccessByName(ctx, namespace)` | 🛡️ | 按 k8s 命名空间名定位 + 校验可访问性，返回命名空间 |
| `RequireNamespaceAccessByID(ctx, id)` | 🛡️ | 按命名空间 ID 定位 + 校验可访问性，返回命名空间 |
| `RequireProjectAccess(ctx, id)` | 🛡️ | 取项目 + 校验其所属命名空间可访问性，返回项目 |
| `RequireNamespaceOwner(ctx, ns)` | 🏠 | 校验当前用户是否为命名空间创建者 |
| `RequireAdmin(ctx, fullMethodName, allowlist...)` | ⭐ | allowlist 精确命中放行，否则要求 admin |
| `RequireFileAccess(ctx, fil)` | 📄 | 当前用户为文件所有者（Username 匹配）或 admin（独立于 gRPC admin 门禁，用于 HTTP 下载与 gRPC ShowRecords 回放） |
| `CanAccessNamespace(ctx, ns)` | 🛡️（布尔） | 纯布尔谓词：admin / 创建者 / 成员 / 公开空间放行，不映射错误；供 IsExists「不可访问视同不存在」的静默场景 |

> AccessBiz 由 `internal/biz/access.go` 定义，经 `NewAccessBiz(nsRepo, projBiz)` 构造（wire 组装）。用户提取在方法内部直接走 `biz.MustGetUser`（原 `internal/auth` 包已并入 `biz/context.go`），不再由传输层注入 `getUser` 回调——ctx 无用户即编程错误（panic），免登录白名单方法不触达任何判定。全部访问判定均已收进 AccessBiz 方法，无顶层自由函数；deploy 等非传输层消费方按 job 临时构造 AccessBiz 复用同一判定。

### 2.1 判定载体大白话解读

> 判定载体 = 把「取当前登录用户 → 加载目标实体 → 按权限等级判定 → 越权打日志 → 返回错误」整套收口成函数。传输层不在方法体里手写 if 判断，开头调一个 AccessBiz 方法即可，形成统一门卫。

| 方法 | 大白话 | 通过后是否顺带返回实体 |
|---|---|---|
| `RequireNamespaceAccessByName` | 这个叫 X 的空间，你能进吗？能进就把空间对象顺带还你 | ✅ 空间对象 |
| `RequireNamespaceAccessByID` | 这个报 ID 的空间，你能进吗？判定逻辑同上，只是按 ID 定位 | ✅ 空间对象 |
| `RequireProjectAccess` | 这个项目能看吗？先查它挂在哪个命名空间下，再按命名空间规则判。项目携带部署配置/环境变量，必须查归属防枚举 ID 拖库 | ✅ 项目对象 |
| `RequireNamespaceOwner` | 你是这个空间的房主（创建者）或 admin 吗？只管 yes/no，不返回对象 | ❌ |
| `RequireAdmin` | 大门守卫：白名单精确命中就放行，否则必须是 admin（抄错一个字符都不算豁免） | ❌ |
| `RequireFileAccess` | 文件上传者是你，或你是 admin？HTTP 下载直接对象比对；gRPC ShowRecords 回放先 `GetByID` 查库取文件元数据再比对 | ❌ |
| `CanAccessNamespace` | 这个空间你能不能进？只回答能/不能，不报错——「进不去」当作「不存在」静默隐藏，不暴露存在性 | ❌ |

记忆法：1/2/3 负责「能看某块资源吗」且通过后带实体；4 负责「你能动这块地吗」；5 是全局 admin 大闸；6 是文件所有者轻量闸（HTTP 下载 / ShowRecords 回放）；7 是 1/2/3 的纯布尔底座（能进=1、不能=0，不报错）。§3 鉴权链正是三层叠用：登录拦截器 → Authorize（5）→ 方法内（1/2/3/4/6）。

## 3. gRPC 鉴权链（三层叠加）

每个 gRPC 方法按序经过（`internal/server/grpc.go` + `internal/server/middlewares/login.go` + `interceptor.go`）：

1. **登录拦截器**（`middlewares.LoginUnaryServerInterceptor(authFn, logger)` / Stream 版 `LoginStreamServerInterceptor(authFn, logger)`）：命中 `biz.IsPublicMethod` 白名单（与 §4.1 免登录清单逐行对应，白名单归属 biz 层）的公开方法直接放行；其余方法要求 Bearer token，校验通过后把用户注入上下文；认证失败打 `[auth audit]` Warning 审计日志（401 兜底）。
2. **Authorize 门禁**（`AuthUnaryServerInterceptor`）：服务实现 `Authorize` 接口（file/repo）→ 自动调用 `Authorize(ctx, fullMethodName)`，内部走 `RequireAdmin`。
3. **方法内访问控制**：各服务方法体开头调用 AccessBiz 的 Require*/Can* 方法（命名空间/项目/owner 级）。

## 4. 各服务方法 → 权限对照

### 4.1 免登录服务（🆓）

| 服务 | 方法 | 权限 | 依据 |
|---|---|---|---|
| auth | Login / Settings / Exchange | 🆓 | biz.IsPublicMethod 白名单 |
| cluster | ClusterInfo | 🆓 | biz.IsPublicMethod 白名单 |
| picture | Background | 🆓 | biz.IsPublicMethod 白名单 |
| version | Version | 🆓 | biz.IsPublicMethod 白名单 |

### 4.2 admin 门禁服务（⭐，Authorize → RequireAdmin）

| 服务 | 方法 | 权限 | 依据 |
|---|---|---|---|
| file | MaxUploadSize | 🔑 | allowlist `/file.File/MaxUploadSize` |
| file | ShowRecords | 📄 `RequireFileAccess`（所有者/admin） | file.go:79 方法体 `GetByID` + `RequireFileAccess`；allowlist file.go:135 放行后做所有者判定 |
| file | List / DiskInfo / Delete | ⭐ | file.go:135（Authorize 门禁） |
| repo | List / Show | 🔑 | allowlist `/repo.Repo/List` `/Show` |
| repo | Create / Update / ToggleEnabled / Delete / Clone / Import / Export / ExportOne | ⭐ | repo.go:357（Authorize 门禁） |

### 4.3 命名空间/项目级访问控制（🛡️ / 🏠）

| 服务 | 方法 | 权限 | 依据 |
|---|---|---|---|
| namespace | Show | 🛡️ `RequireNamespaceAccessByID` | namespace.go:165 |
| namespace | Transfer / Delete / UpdatePrivate / SyncMembers | 🏠 `RequireNamespaceOwner` | namespace.go:55 `showNsAndCheckOwner` |
| namespace | List / Create / Favorite / IsExists | 🔑 | List 按 user 过滤；IsExists 私有空间视同不存在；Create+IgnoreIfExists 命中无权访问空间返 403（namespace.go:139） |
| namespace | **UpdateDesc** | 🔑 登录即可 | ⚠️ 无 owner/访问校验，见 §6.1 |
| project | Show / MemoryCpuAndEndpoints / Delete / Version / AllContainers | 🛡️ `RequireProjectAccess` | project.go:206 |
| project | WebApply / Apply | 🛡️ `RequireNamespaceAccessByID` | deploy/apply.go:62（私有空间成员即可部署） |
| project | List | 🔑 | data 层按命名空间访问谓词过滤 |
| container | IsPodRunning / IsPodExists / ContainerLog / CopyToPod / StreamCopyToPod / StreamContainerLog / Exec / ExecOnce | 🛡️ `RequireNamespaceAccessByName` | container.go |
| endpoint | InNamespace | 🛡️ `RequireNamespaceAccessByID` | endpoint.go:42 |
| endpoint | InProject | 🛡️ `RequireProjectAccess` | endpoint.go:55 |
| changelog | FindLastChangelogsByProjectID | 🛡️ `RequireProjectAccess` | changelog.go:49（携带完整 Config/EnvValues） |
| metrics | TopPod / StreamTopPod | 🛡️ `RequireNamespaceAccessByName` | metrics.go:64 / 84 |
| metrics | CpuMemoryInProject | 🛡️ `RequireProjectAccess` | metrics.go:127 |
| metrics | CpuMemoryInNamespace | 🛡️ `RequireNamespaceAccessByID` | metrics.go:143 |

### 4.4 登录即可（🔑，无资源级控制）

| 服务 | 方法 | 权限 | 依据 |
|---|---|---|---|
| auth | Info | 🔑 | 登录拦截器验签后注入 ctx 用户，本方法 `MustGetUser` 取用户做映射（无用户=编程错误 panic，白名单不含 Info） |
| accessToken | List | 🔑 | 按 Email 只返回本人 token（access_token.go:55） |
| accessToken | Grant / Lease / Revoke | 🔑 | 按 token 值操作，"持有即有权"（见 §6.3） |
| git | AllRepos / ProjectOptions / BranchOptions / CommitOptions / Commit / PipelineInfo / GetChartValuesYaml | 🔑 | git 信息不绑定命名空间（全局仓库） |
| event | List / Show | 🔑 登录即可（普通用户仅见本人） | event.go:54 按操作人邮箱（operator_email）归属过滤：admin 全量，普通用户只看自己的事件；Show 越权访问返回 404（视同不存在，防审计日志 id 枚举） |

## 5. HTTP 端点（httphandler.go / file_handler.go / swagger_handler.go）

| 端点 | 权限 | 依据 |
|---|---|---|
| POST /api/files（上传） | 🔑 authenticated | file_handler.go:62（authHandler 包装） |
| GET /api/download_file/{id} | 📄 `RequireFileAccess` | file_handler.go:104 |
| POST /api/copy_from_pod | 🛡️ `RequireNamespaceAccessByName` | file_handler.go:144 |
| /api/ws_info、/ws | 🔑 authenticated | httphandler.go:59（RegisterWsRoute） |
| /doc/swagger.json、/docs/ | 🆓 公开（仅 HttpCache） | swagger_handler.go:26 |

> gRPC file 服务的 admin 门禁与文件所有者判定是**两条独立规则**：HTTP 下载与 gRPC ShowRecords 共用 `RequireFileAccess` 放行文件所有者/admin，其余 gRPC 文件管理（列表/磁盘信息/删除）仍走 Authorize admin 门禁。ShowRecords 先 allowlist 过 admin 门禁，再在方法体内 `RequireFileAccess` 做所有者判定——两层叠加，普通用户仅能回放自己的会话。

## 6. 审计观察（蓝军视角）

### 6.1 namespace.UpdateDesc —— 越权改写面（潜在缺口）
`UpdateDesc` 改任意命名空间描述（含私有空间）无需 owner/访问校验，与 Transfer/Delete/UpdatePrivate/SyncMembers 四个 owner 变更入口的收口不一致。若描述承载敏感信息，即构成越权改写。**建议**：统一走 `showNsAndCheckOwner`。

### 6.2 project.Apply/WebApply —— 部署门槛 = 命名空间可访问性
私有空间**任何成员**都可发起部署（helm 安装/资源变更）。部署是高风险副作用，如需收紧为 owner/admin，需在 `deploy/apply.go:62` 把 `RequireNamespaceAccessByID` 换成 owner 判定。

### 6.3 accessToken.Revoke/Lease —— 不校验 token 归属
知道 token 值即可撤销/续租（`accessTokenBiz.Revoke` 直接 `repo.Revoke`）。token 是机密凭证（非可枚举 ID），"持有即有权"语义成立，属合理设计；但该约定必须钉死在文档/注释里，防止未来误加归属校验导致自助撤销失效。

### 6.4 namespace.Create+IgnoreIfExists —— 幂等放行的访问门禁（已修复）

`IgnoreIfExists=true` 命中已存在命名空间时原本原样返回该空间完整对象——`preCheckNs` 来自全局 `FindByName`（不感知权限），若命中的是私有空间且调用者无权访问，描述/成员邮箱/创建者等元数据会泄露给无权限用户。已修复（services/namespace.go:139）：放行前先 `n.access.CanAccessNamespace` 校验，无权访问直接 403，与 IsExists"私有空间视同不存在"的隐藏语义对齐。

**残留（可接受）**：`IgnoreIfExists=false` 的 AlreadyExists 响应仍向调用者暴露同名空间"存在"这一事实（存在性预言机）。因 k8s 命名空间全局唯一，此事实在集群层本就公开，无法彻底消除；代价是 IsExists 的存在性隐藏被 Create 部分绕过。**保持不变量**：Create 的 FindByName 预查必须是全局查（不按权限过滤）——若改成权限过滤，无权限用户会被引导去 k8s 创建同名空间，撞 k8s 全局唯一性后反推"存在"反而更糟，且幂等部署会误判重复创建。
