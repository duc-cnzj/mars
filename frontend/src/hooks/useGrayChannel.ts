import { useSyncExternalStore } from 'react'

/**
 * 灰度发布通道——底栏「抢先体验」标记的唯一判据。
 *
 * 判据是**本浏览器携带的灰度路由 cookie**，不是 /api/auth/info 的 isGray。两者回答的是
 * 不同问题：后端字段是「谁被标了灰度」的意图，cookie 才是「这台浏览器实际会被怎么分流」。
 * 二者会脱节（后台拨了开关但你还没刷新 → 意图已变、cookie 未变），标记只认后者，这样它
 * 永远不说谎——它显示的一定是当下真正生效的通道。
 *
 * 唯一例外：chart 同时挂了 canary-by-header，运维可给请求带 `mars_gray: never` 头把流量强制
 * 压回稳定版（ingress 口径 header 位阶高于 cookie），此时 cookie 说灰度、实际走稳定版，标记
 * 会偏高一档。浏览器无法给文档请求附加自定义 header，该通道对面板自身不可达，只可能由网关/
 * 代理注入触发；面板观测不到，故不做这层校正。
 *
 * 判定语义与 nginx-ingress 的 canary 严格对齐（注解 canary-by-cookie: mars_gray）：只有
 * 值恰为 `always` 才路由到灰度（`never` 强制排除，其余值一律忽略并回落到 canary-weight）。
 * 所以这里也只在 `=== 'always'` 时算灰度——若只看「键存不存在」，`mars_gray=never` 会被
 * 误判成灰度，标记就又反过来撒谎了。
 */

/** 灰度路由 cookie 名：必须与 ingress 注解 canary-by-cookie 的值完全一致 */
const GRAY_COOKIE = 'mars_gray'

/** 灰度路由 cookie 值：ingress 只认这一个字面量 */
const GRAY_COOKIE_VALUE = 'always'

/**
 * cookie 存活时长（秒）。取短 TTL 而非会话级：过期即自愈，避免「后台已取消灰度、浏览器
 * 还挂着旧 cookie」这种单向卡死——这是灰度回滚方向唯一的安全网。
 */
const GRAY_COOKIE_MAX_AGE = 30 * 60

// 模块级 listener 集合：本模块写入 cookie 后主动广播，让所有订阅方（底栏）立即重渲染。
// 对齐 useVersion 的模块级单例形制。cookie 本身没有 change 事件，外部改动（devtools /
// 服务端下发）观测不到，只能靠刷新页面重新读取——这是浏览器 API 的硬边界，不是取舍。
const listeners = new Set<() => void>()
const emit = () => {
  for (const l of listeners) l()
}

/** 从 document.cookie 取指定 cookie 的值，不存在返回 null（同名只取第一个，与浏览器行为一致） */
function readCookie(name: string): string | null {
  for (const part of document.cookie.split(';')) {
    const eq = part.indexOf('=')
    if (eq < 0) continue
    if (part.slice(0, eq).trim() === name) return decodeURIComponent(part.slice(eq + 1).trim())
  }
  return null
}

/**
 * 当前浏览器是否处于灰度通道（与 ingress 判定同义）。
 * Path 固定 `/`：canary 要切的是文档与静态资源请求，挂在 /api 上就切不动 bundle。
 * SameSite=Lax：同源请求足够，且天然规避「跨站请求被捎带」；本地 http 下不设 Secure。
 */
function isGrayChannel(): boolean {
  return readCookie(GRAY_COOKIE) === GRAY_COOKIE_VALUE
}

/** 下发/清除灰度路由 cookie：enabled=false 时以 Max-Age=0 立即失效，并广播一次 */
export function setGrayChannel(enabled: boolean): void {
  const base = 'Path=/; SameSite=Lax'
  document.cookie = enabled
    ? `${GRAY_COOKIE}=${GRAY_COOKIE_VALUE}; ${base}; Max-Age=${GRAY_COOKIE_MAX_AGE}`
    : `${GRAY_COOKIE}=; ${base}; Max-Age=0`
  emit()
}

/**
 * sessionStorage 守卫键：本标签页是否已因「对齐」刷过一次。
 * 必须用 sessionStorage 而非模块级变量——模块状态会被 reload 重置，而这里要防的恰恰是
 * 「reload 之后又判定不一致 → 再刷」的死循环，守卫本身必须跨 reload 存活。
 */
const ALIGN_GUARD_KEY = 'mars_gray_aligned'

/**
 * 启动对齐：把后端 isGray（灰度「意图」）落成本浏览器的灰度路由 cookie（灰度「事实」），
 * 两者不一致时硬刷新一次，让新 cookie 真正生效。这是「意图 → 事实」之间唯一的那根线。
 *
 * 为什么必须刷新：cookie 只对**下一次文档请求**生效，而当前 bundle 已经加载完了——写
 * cookie 不会让已渲染的页面换版本，只能 reload 重新向 ingress 要一次文档。
 *
 * 为什么需要守卫：若 cookie 因任何原因写不进去（隐私模式 / ITP / 策略拦截），reload 后重新
 * 读到「仍不一致」就会再刷一次，形成无限刷新。守卫保证每个标签页最多自动刷一次；一旦读到
 * 已对齐即清除守卫，好让后续真正的开关变更（超管拨开关 → 意图变化）还能再刷。
 */
export function alignGrayChannel(isGray: boolean): void {
  if (isGrayChannel() === isGray) {
    sessionStorage.removeItem(ALIGN_GUARD_KEY)
    return
  }
  if (sessionStorage.getItem(ALIGN_GUARD_KEY) === '1') return
  sessionStorage.setItem(ALIGN_GUARD_KEY, '1')
  setGrayChannel(isGray)
  window.location.reload()
}

/**
 * 订阅灰度通道状态：读 cookie 判当前发布通道，供底栏标记渲染。
 * 返回布尔原始值，useSyncExternalStore 以 Object.is 比较，不会触发重渲染死循环。
 */
export function useGrayChannel(): boolean {
  return useSyncExternalStore((cb) => {
    listeners.add(cb)
    return () => {
      listeners.delete(cb)
    }
  }, isGrayChannel)
}
