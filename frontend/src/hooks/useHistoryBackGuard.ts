import { useEffect, useRef } from 'react'

/**
 * history.state 上标记「后退哨兵记录」的键。
 * 哨兵＝弹窗打开时压入的一条与当前页**同 URL** 的历史记录：用户后退/右滑只会退到
 * 这条记录上（pathname 不变 → `<Routes>` 匹配到同一元素 → React 不重挂），
 * 而不是退到上一个真实页面，于是弹窗连同命令行 xterm/WS 会话原样保留。
 */
const GUARD_KEY = '__marsBackGuard'

/** 根元素上的样式开关类：配合 index.css 掐掉触控板双指右滑的「页面间滑动」手势 */
const ROOT_CLASS = 'back-guard-on'

/** 当前记录是不是本 hook 压的哨兵 */
function isSentinel(state: unknown): boolean {
  return !!state && (state as Record<string, unknown>)[GUARD_KEY] === true
}

/**
 * 压哨兵。必须原样带上已有的 history.state：
 * react-router 把 {usr,key,idx} 存在 history.state 里，并用 idx 做 POP 位移计算
 * （见 react-router `createBrowserHistory` 的 getIndex/handlePop）。压一个空 state 会让
 * idx 丢失 → delta 变 NaN → 路由自身的索引记账被带偏。原样展开既保住 usr/key，
 * 也让这条 POP 的 delta 归零，路由把它当「位置没动」处理。
 */
function pushSentinel(url: string) {
  window.history.pushState(
    { ...(window.history.state ?? {}), [GUARD_KEY]: true },
    '',
    url,
  )
}

/**
 * 弹窗打开期间的浏览器后退守卫：后退按钮 / ⌘← / Alt← / 触控板双指右滑一律被吞掉，
 * 页面与弹窗纹丝不动（关闭仍只能走弹窗自己的 X，与弹窗三条 dismiss 路径全拦的定稿一致）。
 *
 * 为什么必须「打开时就压哨兵」，不能等 popstate 到了再补救：
 * react-router 的 popstate 监听在本 hook 之前注册，POP 到达时它已排队一次目标路由的重渲；
 * 等我们拿到事件再 `history.go(1)` 回去，Workbench 早已被卸载、弹窗连同 xterm/WS 会话
 * 一起销毁（命令中断）。只有让后退的落点本身就是「同 URL 记录」，才能既不离开页面、
 * 又不触发重挂。
 *
 * 用法（调用点要声明在「URL 同步」effect **之前**，原因见下）：
 *   useHistoryBackGuard(open, urlRef)
 *
 * @param active 是否启用守卫（为真时压哨兵，转假时摘哨兵）
 * @param canonicalUrlRef 当前状态下 URL 的规范值（由调用方的 URL 同步 effect 写入）。
 *   摘哨兵/补压哨兵要用它而不是 `location.href`——后退落点是上一条记录，它的
 *   query（如 `?open=`、页码）是「打开弹窗那一刻」的旧值，直接读会写脏。
 *   收 ref 对象而非取值函数：ref 引用天然稳定，不会让本 effect 每次渲染都重跑。
 *   类型写成 `{ current: string }` 而非 `RefObject<string>`：后者在 React 18 类型里
 *   current 可空，会把「必有值」的这层约定丢掉。
 */
export function useHistoryBackGuard(
  active: boolean,
  canonicalUrlRef: { current: string },
): void {
  // 渲染期同步最新 active，供 cleanup 里延后的微任务判断「摘哨兵前是否已重新武装」
  const activeRef = useRef(active)
  activeRef.current = active

  useEffect(() => {
    if (!active) return

    activeRef.current = true
    document.documentElement.classList.add(ROOT_CLASS)
    // 幂等压栈：StrictMode 双跑、连续开合弹窗时可能已站在哨兵记录上，不重复压
    if (!isSentinel(window.history.state)) pushSentinel(window.location.href)

    const onPop = () => {
      // 退到了非哨兵记录上（用户按了后退/右滑）→ 立刻补压哨兵，等量抵消这次后退。
      // 此处不做任何路由/状态处理：落点 pathname 与当前一致，React 只是以相同路由重渲一次，
      // 弹窗、终端会话、Tab 状态全部原地保留。
      if (!isSentinel(window.history.state)) pushSentinel(canonicalUrlRef.current)
    }
    window.addEventListener('popstate', onPop)

    return () => {
      activeRef.current = false
      window.removeEventListener('popstate', onPop)
      document.documentElement.classList.remove(ROOT_CLASS)
      // 摘哨兵延到微任务：避免 `history.back()` 与「关闭弹窗后立刻重新打开」的 pushState
      // 争抢同一栈位（同一 tick 内两者次序在规范里未定义）。期间若已重新武装则不再摘。
      queueMicrotask(() => {
        if (activeRef.current || !isSentinel(window.history.state)) return
        // back() 的落点是「打开弹窗那一刻」的记录：弹窗开着时若翻过页/换过 Tab，
        // 它的 query 是旧的 → 在这条记录自己的 pop 上把 URL 重写回当前规范值。
        const onUnwindPop = () => {
          window.removeEventListener('popstate', onUnwindPop)
          window.history.replaceState(window.history.state, '', canonicalUrlRef.current)
        }
        window.addEventListener('popstate', onUnwindPop)
        window.history.back()
      })
    }
  }, [active, canonicalUrlRef])
}
