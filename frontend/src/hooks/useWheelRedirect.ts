import { useCallback, useRef } from 'react'

/** 元素在纵向是否可滚动承接（overflow-y 可滚且内容超出） */
function scrollableY(el: HTMLElement): boolean {
  const s = getComputedStyle(el)
  return (
    (s.overflowY === 'auto' || s.overflowY === 'scroll' || s.overflowY === 'overlay') &&
    el.scrollHeight > el.clientHeight
  )
}

/**
 * 从内容区根往下找「主滚动条」：自身可滚用自身，否则按文档顺序找第一个可滚子元素。
 * 详情弹窗按 Tab 滚动承接不同（TabEdit 自滚 / TabLog 由 wrapper 滚 / Shell 是 xterm 滚动），
 * 运行时动态探测最贴近内容区的那条滚动条，转发才不落空。
 */
function findScrollTarget(root: HTMLElement): HTMLElement | null {
  if (scrollableY(root)) return root
  const stack: HTMLElement[] = []
  for (let i = root.children.length - 1; i >= 0; i--) {
    const child = root.children[i]
    if (child instanceof HTMLElement) stack.push(child)
  }
  while (stack.length) {
    const node = stack.pop()!
    if (scrollableY(node)) return node
    for (let i = node.children.length - 1; i >= 0; i--) {
      const child = node.children[i]
      if (child instanceof HTMLElement) stack.push(child)
    }
  }
  return null
}

/**
 * 整块 dialog 区域的滚轮重定向：指针停在 dialog 内任意位置滚动，都滚动 dialog 内部滚动条，
 * 而不是穿透滚主页面。
 *
 * 非被动 wheel 监听挂在整个 DialogContent 节点上（dialogRef 挂载）。处理时：
 *  - 光标下（e.target 到 dialog 之间）已有可滚动承接（内容区 / CodeMirror / 日志面板等）→
 *    交给原生滚动，不拦截（配合内容区 overscroll-contain，滚到边界也不链回主页面）；
 *  - 否则（标题 / 吸顶头 / Tab 栏 / 空白边距等不可滚动区）→ 把 deltaY 转发到内容区主滚动条
 *    （findScrollTarget 动态探测）并 preventDefault。
 *
 * React onWheel 在 root 容器以 passive 注册、preventDefault 无效，须用 addEventListener 非被动
 * 监听。用 ref callback 挂/收而非 useEffect([open])：组件随卡片常驻挂载，而 Radix 在 open 翻转
 * 后的后续提交才渲染 DialogContent，useEffect([open]) 在 open 置位那一帧跑时节点尚未挂载、
 * deps 不再变则永不重跑；ref callback 在节点 mount/unmount 时天然触发，与提交时序无关。
 *
 * 用法（DialogContent 需已 forwardRef）：
 *   const { dialogRef, contentRef } = useWheelRedirect()
 *   <DialogContent ref={dialogRef} ...>
 *     ...
 *     <div ref={contentRef} className="...overflow-y-auto overscroll-contain ...">...</div>
 *   </DialogContent>
 */
export function useWheelRedirect() {
  const dialogNodeRef = useRef<HTMLDivElement | null>(null)
  const contentRef = useRef<HTMLDivElement | null>(null)

  const onWheel = useCallback((e: WheelEvent) => {
    const dialog = dialogNodeRef.current
    const content = contentRef.current
    if (!dialog || !content || e.deltaY === 0) return
    // 光标下已有可滚动承接 → 原生滚动，不拦截
    let el: HTMLElement | null = e.target instanceof HTMLElement ? e.target : null
    while (el && el !== dialog) {
      if (scrollableY(el)) return
      el = el.parentElement
    }
    // 无承接（标题/吸顶头/Tab 栏/边距）→ 转发到内容区主滚动条并 preventDefault，不滚主页面
    const target = findScrollTarget(content)
    if (!target) return
    target.scrollTop += e.deltaY
    e.preventDefault()
  }, [])

  const dialogRef = useCallback(
    (node: HTMLDivElement | null) => {
      if (dialogNodeRef.current) dialogNodeRef.current.removeEventListener('wheel', onWheel)
      dialogNodeRef.current = node
      if (node) node.addEventListener('wheel', onWheel, { passive: false })
    },
    [onWheel],
  )

  return { dialogRef, contentRef }
}
