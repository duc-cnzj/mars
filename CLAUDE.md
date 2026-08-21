## 是否达到S级别评判标准(如果没有达到S级别，你需要列出需要修改的点，并且询问用户是否需要修改)

1. 可读性、可测试性、覆盖 100%
2. 零死代码
3. 各个函数是否有注释，并且注释正确， 每个函数都要有注释（中文注释）
4. 断言级边界全过
5. 目录/文件是否有存在的必要
6. 函数命名、文件命名是否正确
7. 结构是否合理，是否存在多余无意义的代码，是否可以简化代码实现相同的功能
8. .go文件放在当前目录是否合理，例如明明不属于当前目录的文件，却被放在当前目录下

## 错误处理设计规范（三层行为规则）

### 原则：错误协议映射（底层错误 → http/grpc status）全部收敛在 `internal/errs`（单一事实来源），biz/data/services 都不触碰字面 HTTP/gRPC 码，统一通过 errs 语义构造器表达与判定错误

### 核心判定：按错误实际类型归类，不按操作意图硬编码

- **不确定错误**（查询/更新/外部 API 调用的返回错误，既可能是"记录不存在"也可能是"DB 断开/网络抖动"）→ data 边界统一用 `errs.Wrap`（自动归类：ent.NotFound/k8s apierrors.NotFound→404、ent.ValidationError→400、ent.ConstraintError→409、其余→500），禁止用 `WrapNotFound` 通杀——否则 DB 抖动会被误报成 404"记录不存在"
- **确定语义错误**（代码刚显式构造的校验失败如 `errors.New("repo 名称已经存在")`、认证失败）→ 用语义构造器 `WrapInvalidArgument`/`WrapUnauthenticated`（biz 层同理）

### 各层行为约束

1. **data 层**

   - 只在 data 层出口（repo 方法边界 + 生命周期方法 InitDB/InitS3/InitK8s/Migrate）用 `errs.Wrap(err, msg)` 包裹一次不确定错误，携带业务上下文（如 "query access token"）——data 只表达领域语义，协议码由 errs 构造器按底层错误**实际类型**自动归类，data 不触碰字面 HTTP code；`errs.Wrap` 是 ent+k8s 感知的自动归类构造器：ent.NotFound / k8s apierrors.NotFound → NotFound(404)、ent.ValidationError → InvalidArgument(400)、ent.ConstraintError → AlreadyExists(409)、已带 status 码的错误保留原码、其余无法归类的底层失败（DB 连接异常/k8s API 调用失败/上传存储故障）才落默认 Internal/500——保证统一 Wrap 包裹查询/更新操作时不会把 ent.NotFound/k8s NotFound 误映射成 500，也不会把网络故障误判成 404；仅当错误是代码刚显式构造的确定语义（如 `errors.New("repo 名称已经存在")` 校验失败）才用 `errs.WrapInvalidArgument(err, msg)` 语义构造器
   - 内部函数透传错误，禁止层层包裹（每条链路只留一份堆栈，避免堆栈爆炸与性能损耗）
   - 领域错误（NotFound/权限等）不被技术性 wrap 破坏语义
   - ⚠️ 禁止裸 `pkg/errors.Wrap(status.Error(...), msg)` 绕过语义构造器——协议码映射必须收口 errs 的 Wrap\*/消息构造器（注：当前 grpc-go 的 `status.FromError` 已能穿透 wrap 链，但码映射散落仍是坏味道）

2. **biz 层**

   - 业务语义表达层：领域错误统一用 `errs` 语义构造器——`errs.WrapNotFound`/`errs.WrapInvalidArgument`/`errs.WrapUnauthenticated` 包裹底层错误，`errs.NotFound`/`errs.InvalidArgument`/`errs.Unauthenticated`/`errs.AlreadyExists` 直接构造消息错误，`errs.ErrorPermissionDenied` 作为权限 sentinel；判定用 `errs.IsNotFound`，不直接依赖 ent 错误类型
   - 生产代码禁止 import `google.golang.org/grpc/status|codes`（`internal/biz` 零 grpc import 是可验证验收项）；协议映射收口在 errs 内部（`grpcStatusError` 模式不外露）
   - data 层直接复用 errs 语义构造器（`internal/data/errors.go` 委托层已删），禁止在业务逻辑里重新造轮子

3. **services/transport 层**
   - 透传 biz 返回的错误，不重新造 status 码；非 NotFound 直接返回原始错误，避免丢失原始错误码
   - 判定协议语义用 `errs.IsNotFound`/`errs.ErrorPermissionDenied`；散落的 `status.Errorf`/`status.Error` 应逐步收口回 errs 构造器

### 验收

- `grep "google.golang.org/grpc/status|codes" internal/biz/*.go`（非 test）应为 0
- 全仓 `grep "status.Error"`/`status.Errorf` 在 services 层应为 0 或仅有已注释说明的例外
- 每个 data 层 repo 出口与生命周期方法（InitDB/InitS3/InitK8s/Migrate）的错误都经过 `errs.Wrap*`/`errs.Wrap` 构造器且带中文上下文
- 错误链上各层无重复堆栈

## 日志打印规范（错误日志统一到最上层打印）

### 原则：错误日志只在最上层（services 层）打印，data/biz 层不出现显式错误日志打印代码

- 底层只「生产错误」（返回/上抛），最上层（services）「消费错误」（统一打日志）
- 上下文通过 `errs.Wrap*` 语义构造器的 wrap 链携带，堆栈由 services 层 `logError` → `errstack` 打出——「不打印」不等于「不留痕」

### 各层行为约束

1. **data 层**：禁止 `logger.Error*` 显式打印后 `return err` 的双留痕（一个错误 2 条日志）；只返回错误
2. **biz 层**：禁止 `ErrorCtx(ctx, err); return nil, err` 双留痕（门卫类等）；`status.Errorf` 是构造错误非日志，不在禁止范围
3. **services 层**：统一用 `logError`（`logger.ErrorCtx + return err`）收敛，是错误日志的唯一出口

### 必须保留的例外（技术约束，不违反本规则）

1. **异步 goroutine / 事件监听 / informer 回调**里的错误无法 return 给上层（如 PodEventListener、HandleProjectChanged、informer 回调），必须原地打日志，否则错误无声丢失
2. **无法向调用方返回错误的初始化/边缘路径**（如 `logger.Error(err); return` 且无 err 返回值），若无法改造为冒泡返回，保留原地打印

### 验收

- data/biz 层 `grep "logger.Error"`/`.ErrorCtx` 仅剩上述两类例外（异步边缘 + 无法上抛），双留痕（打印后 return err）为 0
- services 层错误日志收敛到 `logError` 单一出口
- 同一错误在整条链上只出现一条日志

## 条件分支规范（if-else 治理）

### 原则：guard clause 优先——异常/错误/边界路径提前 `return`/`continue`/`break`，主路径不缩进；禁止 else-if 链与冗余 else

### 各层行为约束

1. **guard clause 优先**：错误/边界条件用 `if cond { return/continue/break }` 提前退出，主路径不包进 else

   ```go
   // 好：错误提前返回，主路径不缩进
   if err != nil {
       return err
   }
   doMain()

   // 差：else 包住主路径
   if err != nil {
       return err
   } else {
       doMain()
   }
   ```

2. **禁止 `else if` 链**：同一条件的多段判断用 `switch`（或逐步 return 重排），不叠 else-if

   ```go
   // 好：switch 表达多段判断
   switch {
   case seconds <= -1:
       return "<invalid>"
   case seconds <= 0:
       return "0秒"
   case seconds < 60*2:
       return fmt.Sprintf("%d秒", seconds)
   default:
       // ...
   }

   // 差：else-if 链
   if seconds <= -1 {
       return "<invalid>"
   } else if seconds <= 0 {
       return "0秒"
   } else if /* ... */ {}
   ```

3. **if 分支以 `return`/`continue`/`break`/`panic` 结尾时禁止冗余 else**：else 分支提为独立语句，happy path 不缩进

4. **二分 if/else 允许**：if 尾无提前返回的纯二分支可保留 else，但优先考虑 early-return

### 豁免（不可手改的生成代码）

- `internal/data/ent/` 等工具生成代码（重新生成即覆盖），不适用本规则，禁止手改
