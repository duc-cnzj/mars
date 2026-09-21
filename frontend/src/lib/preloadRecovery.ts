/**
 * 构建产物版本错配的兜底恢复。
 *
 * 背景：静态资源按内容哈希命名，且**编译期 embed 进 Go 二进制**（后端 `frontend/embed.go`
 * 的 `//go:embed build/*`，挂 `/resources/`）；而灰度（canary）分流是 nginx **逐请求按
 * `mars_gray` cookie** 判定的。两者的生命周期并不绑定，于是下面任一条路径都会产出错配——
 * 文档来自 A 版，某个懒加载 chunk 的请求却落到 B 版：
 *   - 灰度 cookie 过了 Max-Age 自然失效，页面还开着（文档仍是灰度版，chunk 请求回落稳定版）；
 *   - CDN / 中间层缓存了旧 index.html，引用的 hash 在新版 Pod 上已不存在；
 *   - 运维用 `-H "mars_gray: never"` 临时把通道压回稳定版，而页面外壳还是灰度版；
 *   - 标签页挂在那儿跨过了一次发布。
 * B 版二进制里没有那个文件 → `http.FileServer` 返回 404 → 浏览器抛
 * `Failed to fetch dynamically imported module`，页面卡在半渲染态。
 *
 * 兜底逻辑：接住 Vite 派发的 `vite:preloadError`，整页重载。重载会让**文档与后续资源在
 * 同一次请求上下文里**重新对齐到「当前 cookie 实际指向的那个版本」，从而自愈。
 *
 * 守卫：每个标签页最多自动重载一次。若某个 chunk 在服务端「真的」缺失，无守卫会退化成
 * 无限重载，比一次 404 更糟。守卫在应用挂载后**延时**解除（见 schedulePreloadGuardRelease），
 * 语义与 useGrayChannel 的 ALIGN_GUARD_KEY 一致——「本次版本已自洽，允许未来某次错配
 * 再自愈一次」。
 */

/**
 * 本标签页是否已因 chunk 加载失败自动重载过。
 * 用 sessionStorage 而非模块级变量：守卫必须**跨整页重载存活**，否则重载后守卫归零，
 * 又会再刷一次，形成死循环——这与 ALIGN_GUARD_KEY 选 sessionStorage 的理由完全相同。
 */
const PRELOAD_RELOAD_GUARD = 'mars_preload_reloaded'

/**
 * 守卫解除前的等待时长（毫秒）。
 *
 * ⚠️ 不能改成「应用一挂载就解除」：App 的首次提交**早于**本文档里懒加载 chunk 的失败落地——
 * <Suspense> 边界会先带着 fallback 提交，effect 随即执行，而 chunk 请求要等网络往返才知道结果。
 * 若在提交那一刻解除，等于在「重试还没发生」时就把守卫作废：服务端真的缺这个 chunk 时，
 * 新文档一挂载 → 守卫清空 → 同一路由再失败 → 再整页重载，退化成无限重载（正是本模块要防的场景）。
 *
 * 延时 10s 这个量级的作用：覆盖「重载后那次重试的失败」落地所需时间（正常 404 亚秒级、
 * 慢响应数秒），同时不把守卫永久焊死——健康存活 10s 后解除，未来的另一次版本错配仍能立即自愈。
 */
const GUARD_RELEASE_DELAY_MS = 10_000

/**
 * 安装 chunk 加载失败兜底。必须在任何懒加载 import 之前调用（main.tsx 入口处）。
 *
 * 刻意**不调用** `event.preventDefault()`：Vite 的派发实现是
 * `window.dispatchEvent(e); if (!e.defaultPrevented) throw payload`。
 * 一旦 preventDefault，动态 import 的 promise 会**正常 resolve 成 undefined**，
 * undefined 流进 React.lazy 再被读 `.default` 时抛出更难定位的 TypeError——等于把
 * 一句清晰的「加载失败」换成面目全非的错误。既然无论如何都要整页重载，不吞更诚实。
 */
export function installPreloadRecovery(): void {
  window.addEventListener('vite:preloadError', () => {
    if (sessionStorage.getItem(PRELOAD_RELOAD_GUARD) === '1') return
    sessionStorage.setItem(PRELOAD_RELOAD_GUARD, '1')
    window.location.reload()
  })
}

/**
 * 预约解除重载守卫，由 App 顶层 effect 在挂载时调用（见常量注释：解除必须延时，
 * 不能与「挂载成功」同时发生，否则守卫会在重试失败落地前失效）。
 *
 * 解除后守卫归零，好让后续（可能几十分钟之后、cookie 换通道时）的另一次错配还能再自愈一次。
 */
export function schedulePreloadGuardRelease(): void {
  window.setTimeout(() => {
    sessionStorage.removeItem(PRELOAD_RELOAD_GUARD)
  }, GUARD_RELEASE_DELAY_MS)
}
