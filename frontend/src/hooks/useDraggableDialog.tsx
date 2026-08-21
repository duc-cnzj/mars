import { useCallback, useEffect, useRef, useState, type CSSProperties } from 'react'

/**
 * 可拖拽弹窗 Hook（增强版，兼容旧 API）：
 * - 拖拽：把手（DialogTitle 标题栏）pointerdown 后跟随鼠标移动，松开定位
 * - 缩放：四边/四角把手（getResizeHandleProps(dir)）拖拽改变宽高，对边锚定
 * - 双击最大化/还原：dragHandleProps 上双击切换全屏
 * - z-index 层级管理：指针按下时把当前弹窗置顶（跨多个弹窗共享计数器）
 * - 窗口 resize 后 clamp 防溢出
 * 未拖动前保持居中；拖动后通过 inline style 覆盖 shadcn DialogContent 的居中定位类
 * （left-[50%] top-[50%] translate-x/y-[-50%]），portal / 焦点圈定 / ESC 仍由 Radix 负责。
 * 弹窗容器用 closest('[data-slot="dialog-content"]') 定位，兼容同时打开多个弹窗。
 */

/** 跨实例共享的 z-index 计数器：从 shadcn 的 z-50 之上开始累加，实现置顶 */
let zCounter = 51

/** 供非可拖拽但需盖过可拖拽弹窗的浮层（如强杀确认框）分配 z-index */
export function nextZIndex() {
  return ++zCounter
}

const MIN_W = 320
const MIN_H = 240

const clamp = (min: number, max: number, v: number) => Math.max(min, Math.min(max, v))

/** 缩放方向：八方向（四边 + 四角），dir 串里带 e/w 决定宽度哪边跟手，带 n/s 决定高度哪边跟手 */
export type ResizeDir = 'n' | 's' | 'e' | 'w' | 'ne' | 'nw' | 'se' | 'sw'

interface Rect {
  x: number
  y: number
  width: number
  height: number
  /** 最大化前的位置快照，用于双击还原 */
  prev?: Rect
}

/** 窗口尺寸变化后把弹窗 clamp 回可视区域（防溢出） */
function clampToWindow(rect: Rect): Rect {
  const ww = window.innerWidth
  const wh = window.innerHeight
  const width = clamp(MIN_W, ww - rect.x, rect.width)
  const height = clamp(MIN_H, wh - rect.y, rect.height)
  const x = clamp(0, ww - width, rect.x)
  const y = clamp(0, wh - height, rect.y)
  return { x, y, width, height, prev: rect.prev }
}

export function useDraggableDialog(onResize?: () => void) {
  const [rect, setRect] = useState<Rect | null>(null)
  const [zIndex, setZIndex] = useState(() => ++zCounter)
  // 拖拽/缩放把手按下时缓存起始信息 + 目标元素引用；拖动期间直接改 DOM style，
  // 避免每次 pointermove 都 setRect 触发整个弹窗子树重渲染（日志/xterm 很重 → 延迟高、跟手不准）
  const dragRef = useRef<{ startX: number; startY: number; orig: Rect; el: HTMLDivElement } | null>(null)
  const resizeRef = useRef<{ startX: number; startY: number; orig: Rect; dir: ResizeDir; el: HTMLDivElement } | null>(null)

  const bringToFront = useCallback(() => setZIndex(() => ++zCounter), [])

  // 尺寸变化信号：仅当宽/高真的变化时通知宿主（纯拖动位移不算），
  // 用于让 shell 终端等子组件在弹窗缩放/最大化/还原后 refit（对齐旧版 DraggableModal onResize）
  const onResizeRef = useRef(onResize)
  useEffect(() => {
    onResizeRef.current = onResize
  }, [onResize])
  const prevSizeRef = useRef('')
  useEffect(() => {
    if (!rect) return
    const size = `${rect.width}x${rect.height}`
    if (size !== prevSizeRef.current) {
      prevSizeRef.current = size
      onResizeRef.current?.()
    }
  }, [rect])

  /** 标题栏拖拽移动：拖动中直接改元素 style（即时响应，无 React 重渲染），松开时提交一次 rect */
  const handlePointerDown = useCallback((e: React.PointerEvent) => {
    if (e.button !== 0) return
    // 命中可选中文本区（标题上的项目名/命名空间，标记 data-no-drag）：
    // 让位给原生文本选择，不 preventDefault、不启动拖拽（否则无法选中复制）
    if ((e.target as HTMLElement).closest?.('[data-no-drag]')) return
    const el = e.currentTarget.closest('[data-slot="dialog-content"]') as HTMLDivElement | null
    if (!el) return
    e.preventDefault()
    bringToFront()
    // 同 resize：关掉 transition:all 0.2s，否则拖动位置/尺寸以 200ms 动画跟手（滞后）
    el.style.transition = 'none'
    const r = el.getBoundingClientRect()
    dragRef.current = {
      startX: e.clientX,
      startY: e.clientY,
      orig: { x: r.left, y: r.top, width: r.width, height: r.height },
      el,
    }
    const move = (ev: PointerEvent) => {
      const d = dragRef.current
      if (!d) return
      const ww = window.innerWidth
      const wh = window.innerHeight
      const x = clamp(0, ww - d.orig.width, d.orig.x + (ev.clientX - d.startX))
      const y = clamp(0, wh - d.orig.height, d.orig.y + (ev.clientY - d.startY))
      // 直接改 inline style：位置即时生效；清掉居中 translate/transform。
      // 同时把宽高写死：maxWidth 一旦置 none，shadcn 的 w-full 会立刻把弹窗撑满整窗
      //（max-w-[calc(100%-2rem)]/sm:max-w-5xl 被 maxWidth:none 顶掉），拖拽会瞬间跳变。
      d.el.style.left = `${x}px`
      d.el.style.top = `${y}px`
      d.el.style.width = `${d.orig.width}px`
      d.el.style.height = `${d.orig.height}px`
      d.el.style.transform = 'none'
      d.el.style.translate = 'none'
      d.el.style.maxWidth = 'none'
    }
    const up = () => {
      const d = dragRef.current
      dragRef.current = null
      window.removeEventListener('pointermove', move)
      window.removeEventListener('pointerup', up)
      if (!d || !d.el.isConnected) return
      // 结束时按真实位置提交 state（供最大化还原 / 窗口 resize 钳制读取），只重渲染一次。
      // 必须保留 prev：双击最大化/还原时，双击的前两拍 pointerup 也会走到这里提交 rect，
      // 若不带 prev 会把最大化前的位置快照丢掉，还原时 prev.prev 取不到 → 回落到左上角。
      const rr = d.el.getBoundingClientRect()
      setRect((r) => ({ x: rr.left, y: rr.top, width: rr.width, height: rr.height, prev: r?.prev }))
    }
    window.addEventListener('pointermove', move)
    window.addEventListener('pointerup', up)
  }, [bringToFront])

  /** 四边/四角缩放：拖动中直接改元素 style（无重渲染），松开时提交一次 rect */
  const handleResizePointerDown = useCallback(
    (e: React.PointerEvent, dir: ResizeDir) => {
      if (e.button !== 0) return
      const el = e.currentTarget.closest('[data-slot="dialog-content"]') as HTMLDivElement | null
      if (!el) return
      e.preventDefault()
      e.stopPropagation()
      bringToFront()
      // 关键：弹窗基座带 transition:all 0.2s（shadcn duration-200），
      // 拖拽/缩放时 width/left/translate 会以 200ms 过渡动画跑——宽度从 w-full(100%)
      // 过渡到目标值出现「跳到近全屏再缩回」的跳变，且每步滞后 200ms 跟不上手。
      // 手势开始时整段关掉过渡（入口/出口是 animation 不是 transition，不受影响）。
      el.style.transition = 'none'
      const r = el.getBoundingClientRect()
      resizeRef.current = {
        startX: e.clientX,
        startY: e.clientY,
        orig: { x: r.left, y: r.top, width: r.width, height: r.height },
        dir,
        el,
      }
      const move = (ev: PointerEvent) => {
        const r = resizeRef.current
        if (!r) return
        const ww = window.innerWidth
        const wh = window.innerHeight
        const dx = ev.clientX - r.startX
        const dy = ev.clientY - r.startY
        // 水平：dir 带 e → 右缘跟光标（左缘锚定，宽 = orig + dx）；
        //      带 w → 左缘跟光标（右缘锚定，宽 = orig - dx，left 同步补回）。
        // 垂直：带 s → 底缘跟光标；带 n → 顶缘跟光标（底缘锚定）。
        // 角落 dir 同时带水平+垂直，两轴一起算。对边锚定保证拖哪边只有哪边动，
        // 且首次 resize 时不会因 CSS 居中（top/left 50% + translate -50%）从四边向中间伸缩。
        let left = r.orig.x
        let width = r.orig.width
        if (r.dir.includes('e')) {
          width = clamp(MIN_W, ww - left, r.orig.width + dx)
        } else if (r.dir.includes('w')) {
          width = clamp(MIN_W, r.orig.x + r.orig.width, r.orig.width - dx)
          left = r.orig.x + r.orig.width - width
        }
        let top = r.orig.y
        let height = r.orig.height
        if (r.dir.includes('s')) {
          height = clamp(MIN_H, wh - top, r.orig.height + dy)
        } else if (r.dir.includes('n')) {
          height = clamp(MIN_H, r.orig.y + r.orig.height, r.orig.height - dy)
          top = r.orig.y + r.orig.height - height
        }
        // 直接改 style，即时生效：left/top 锚定 + 清掉居中 translate/transform，
        // max-width 清掉（否则被 shadcn 的 max-w-[calc(100%-2rem)] 卡住撑不开）
        r.el.style.left = `${left}px`
        r.el.style.top = `${top}px`
        r.el.style.transform = 'none'
        r.el.style.translate = 'none'
        r.el.style.width = `${width}px`
        r.el.style.height = `${height}px`
        r.el.style.maxWidth = 'none'
      }
      const up = () => {
        const r = resizeRef.current
        resizeRef.current = null
        window.removeEventListener('pointermove', move)
        window.removeEventListener('pointerup', up)
        if (!r || !r.el.isConnected) return
        const rr = r.el.getBoundingClientRect()
        setRect((cur) => ({ x: rr.left, y: rr.top, width: rr.width, height: rr.height, prev: cur?.prev }))
      }
      window.addEventListener('pointermove', move)
      window.addEventListener('pointerup', up)
    },
    [bringToFront],
  )

  /** 双击标题栏 / 卡片边缘：全屏 / 还原（整个 header 均可双击，包括项目名文本区） */
  const toggleMaximize = useCallback((e: React.MouseEvent) => {
    const el = (e.currentTarget as HTMLElement).closest(
      '[data-slot="dialog-content"]',
    ) as HTMLDivElement | null
    const ww = window.innerWidth
    const wh = window.innerHeight
    setRect((prev) => {
      if (prev) {
        const isMax = prev.width >= ww - 1 && prev.height >= wh - 1
        if (isMax) {
          // 还原到最大化前的位置尺寸
          const back = prev.prev ?? { x: 0, y: 0, width: 720, height: 560 }
          return { x: back.x, y: back.y, width: back.width, height: back.height }
        }
        return {
          ...prev,
          prev: { x: prev.x, y: prev.y, width: prev.width, height: prev.height },
          x: 0,
          y: 0,
          width: ww,
          height: wh,
        }
      }
      // 尚未拖动过（仍由 CSS 居中）：从 DOM 取当前盒取快照
      const r = el?.getBoundingClientRect()
      const base = {
        x: r?.left ?? 0,
        y: r?.top ?? 0,
        width: r?.width ?? 720,
        height: r?.height ?? 560,
      }
      return { ...base, prev: base, x: 0, y: 0, width: ww, height: wh }
    })
  }, [])

  /** 弹窗内容任意位置 pointer down 也置顶（跨多个弹窗） */
  const handleContentPointerDown = useCallback(
    (e: React.PointerEvent) => {
      // React 合成事件沿组件树传播而非 DOM 树：嵌套在本 content 里的子 Dialog
      // （如 ConfigHistory 历史弹窗，Radix portal 到 body）在 DOM 上是本节点的兄弟，
      // 但内部 pointerdown 仍会冒到这里的 capture。按真实 DOM 落点判定——只有按下
      // 落在本弹窗内容内部才置顶，否则会把宿主弹窗顶到子弹窗之上、盖住子弹窗。
      if (!e.currentTarget.contains(e.target as Node)) return
      bringToFront()
    },
    [bringToFront],
  )

  // 窗口 resize 后 clamp 防溢出
  useEffect(() => {
    const onResize = () => setRect((prev) => (prev ? clampToWindow(prev) : prev))
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  const isMaximized =
    rect !== null && rect.width >= window.innerWidth - 1 && rect.height >= window.innerHeight - 1

  // Tailwind v4 的 translate-x/y-[-50%] 用现代 CSS `translate` 属性（非 transform），
  // 拖动后必须同时清空 transform 与 translate，否则居中位移会叠加半个弹窗尺寸。
  const contentStyle: CSSProperties | undefined = rect
    ? {
        left: rect.x,
        top: rect.y,
        width: rect.width,
        height: rect.height,
        maxWidth: 'none',
        transform: 'none',
        translate: 'none',
        zIndex,
      }
    : { zIndex }

  /** 按方向取 resize 把手事件：四边/四角把手都挂这个，方向决定哪边跟手、哪边锚定 */
  const getResizeHandleProps = useCallback(
    (dir: ResizeDir) => ({
      onPointerDown: (e: React.PointerEvent) => handleResizePointerDown(e, dir),
    }),
    [handleResizePointerDown],
  )

  return {
    contentStyle,
    contentProps: { onPointerDownCapture: handleContentPointerDown },
    dragHandleProps: { onPointerDown: handlePointerDown, onDoubleClick: toggleMaximize },
    getResizeHandleProps,
    bringToFront,
    isMaximized,
  }
}
