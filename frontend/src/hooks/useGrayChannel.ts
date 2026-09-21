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
 * cookie 存活时长（秒）。它的角色是**兜底保险丝**，不是回滚安全网——真正负责及时回退的是
 * 两处**主动对账**：启动时的 alignGrayChannel、以及到期前的 scheduleGrayChannelRenewal，
 * 都拿后端 isGray（意图）与本 cookie（事实）比对，不一致就清 cookie 并整页重载。
 * Max-Age 只保证「对账逻辑本身出问题时」不会永久卡在灰度。
 *
 * ⚠️ 为什么不改会话级：会话级 cookie 只在关标签页时消失，而标签页可能开着好几天，
 * 那样连这根保险丝都没了。短 TTL 保证最长 30 分钟必然得到一次重新对账的机会。
 */
const GRAY_COOKIE_MAX_AGE = 30 * 60

/**
 * 到期前多久触发重校验（毫秒）。
 *
 * 留余量是必须的：重校验若判定要换通道，走的是「清 cookie + 整页重载」，而这次重载请求
 * 要靠 cookie 决定落到哪个版本。等 cookie 真失效才触发，重载请求本身就已不带 cookie，
 * 会先掉到稳定版、再被启动对齐弹回灰度——白刷一次，还让用户看见一次版本闪变。
 * 提前触发则重载请求仍带 cookie，一次到位。
 */
const RENEW_BEFORE_EXPIRY_MS = 2 * 60 * 1000

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
 * 本次会话是否已排定一次整页刷新（reload 已调用、但新文档尚未接管）。
 * 供「一旦刷新就会丢失」的副作用使用——典型是登录成功 toast：整页刷新会连同 JS 内存一起
 * 清掉 toast，所以 reload 在路上的时候不要弹，留给刷新后的新文档弹（见 markLoginSuccess）。
 */
let reloadPending = false

/** 是否已有一次整页刷新在路上（见 reloadPending） */
export function isReloadPending(): boolean {
  return reloadPending
}

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
    // 一致时**续期**而不是原地返回：cookie 是一次性写入、Max-Age 单向倒计时，不续期的话
    // 任何挂得久一点的标签页都必然滑进「cookie 已过期、页面却还是灰度版」的僵尸态——
    // 此后每个懒加载 chunk 都会回落稳定版并 404。续期把窗口一起往前推，让「页面版本」与
    // 「路由判定」在整个会话期内保持一致。注意只在 isGray=true 时续：非灰度通道本就没有
    // cookie，写它就是无中生有地把自己抬进灰度。
    if (isGray) setGrayChannel(true)
    sessionStorage.removeItem(ALIGN_GUARD_KEY)
    return
  }
  if (sessionStorage.getItem(ALIGN_GUARD_KEY) === '1') return
  sessionStorage.setItem(ALIGN_GUARD_KEY, '1')
  setGrayChannel(isGray)
  reloadPending = true
  window.location.reload()
}

/** 页面内的续期定时器（模块级单例：同一文档最多一个待触发，重复预约会取消上一个） */
let renewalTimer: ReturnType<typeof setTimeout> | null = null

/**
 * 预约一次「到期前重校验」：在 cookie 失效前 RENEW_BEFORE_EXPIRY_MS 触发，重新拉取后端
 * 灰度意图并交给 alignGrayChannel 裁决——
 *   - 意图仍是灰度 → 走一致分支给 cookie 续期，用户全程无感；
 *   - 意图已撤销 → 走不一致分支清 cookie 并整页重载，干净回到稳定版。
 *
 * 为什么必须有这个定时器：alignGrayChannel 只在**文档加载时**跑一次，而 cookie 的 Max-Age
 * 是单向倒计时。一个挂满 30 分钟以上的标签页，中途没有任何时机重新对账——这正是「文档还是
 * 灰度版、chunk 请求却已回落稳定版」那类 404 的成因。本函数把那一次对账补上。
 *
 * ⚠️ 每轮对账结束都要**重新预约**（见回调里的自调用）：本函数是「单轮预约」，不是「一次性
 * 保险」。若只预约一轮，续期把 cookie 推到 T+58 分钟后就没有下一次对账了——T+58 起又回到
 * 「文档是灰度版、cookie 已过期」的僵尸态，等于只把问题推迟 28 分钟。自调用后覆盖变成
 * 连续窗口：28min 对账→续期至 58min→56min 再对账→续期至 86min……整个会话期一致。
 * 重载分支天然不续：换通道会清掉 cookie，自调用进去第一句 `!isGrayChannel()` 就返回了。
 *
 * 非灰度通道直接不预约：cookie 本就不存在，谈不上过期，也就没有僵尸页风险。
 *
 * ⚠️ 后台标签页的定时器会被浏览器节流，可能晚于 Max-Age 才触发。这是可接受的：届时 cookie
 * 已失效，alignGrayChannel 会把「cookie 为空 vs 意图仍是灰度」判为不一致，**同步重写 cookie
 * 后整页重载**，标签页重新可见时自动收敛。且 cookie 是同步写入的，从定时器触发到重载发生
 * 之间发出的 chunk 请求依然带着 cookie，不会掉到稳定版。
 *
 * @param fetchIntent 重新拉取后端灰度意图。由调用方注入（本模块不依赖 api 层），
 *                    拉取失败时抛出即可——此处刻意不做任何切换：既不能假定仍是灰度
 *                    （可能已回退），也不能假定已回退（可能只是网络抖动）。按兵不动让 cookie
 *                    自然到期，页面退化为错配态后由 preloadError 兜底整页重载收敛。
 *                    安全（绝不误切通道），代价是这一轮多刷一次页面。
 */
export function scheduleGrayChannelRenewal(fetchIntent: () => Promise<boolean>): void {
  if (renewalTimer !== null) clearTimeout(renewalTimer)
  renewalTimer = null
  if (!isGrayChannel()) return
  renewalTimer = setTimeout(() => {
    renewalTimer = null
    fetchIntent()
      .then((intent) => {
        alignGrayChannel(intent)
        // 预约下一轮：本轮续期已把 cookie 推到 T+30min，下一轮须在此之前再对一次账，
        // 否则续期只生效一轮（详见上方函数注释）。对齐判定为「换通道」时 cookie 已被清掉，
        // 自调用开头的 isGrayChannel() 判定会直接返回，不会在垂死的文档上留定时器。
        scheduleGrayChannelRenewal(fetchIntent)
      })
      .catch(() => {
        // 意图未知：不切通道（理由见上方 @param），也不续约——本轮不做裁决，cookie 到期后
        // 页面进入错配态，由 preloadError 整页重载兜底收敛。此处无 logger 依赖，静默即可。
      })
  }, GRAY_COOKIE_MAX_AGE * 1000 - RENEW_BEFORE_EXPIRY_MS)
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
