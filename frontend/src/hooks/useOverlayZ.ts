import { useLayoutEffect, useState } from 'react'
import { nextZIndex } from '@/lib/zIndex'

/**
 * 顶层浮层 z-index：`open` 变 true 时取一个共享 z（nextZIndex），保证盖过可拖拽宿主弹窗。
 *
 * **为什么必须 useLayoutEffect**：宿主（useDraggableDialog）的 z 从 51 起、且每次指针交互
 * bringToFront 都会递增；而浮层在**挂载**时算的那一次 z 会随后过期。若用 useEffect（被动副作用，
 * 跑在浏览器绘制之后），打开的首帧用的还是旧 z —— 实测成员弹窗首帧 z=57 而宿主 z=61，浮层先绘制在
 * **下层**、1~2 帧后才跳到 64 置顶，表现为「点了没反应，顿一下才冒出来」。useLayoutEffect 在 DOM
 * 变更后、绘制前同步执行，首帧即置顶。
 *
 * **什么时候该用**：`open` 由「Radix 触发器之外」的东西置 true —— 普通 Button 的 onClick、
 * 或挂载即打开（`{cond && <Modal open />}`）。这些场景 onOpenChange 不会触发，只能在 effect 里兜。
 *
 * **什么时候不该用**：浮层由 Radix 的 PopoverTrigger / TooltipTrigger 打开时，onOpenChange 在点击/
 * 悬停事件里**同步**触发，在同一批次就能定好 z，首帧天然正确。这类站点请继续用
 * `useState(0)` + `onOpenChange={(o) => { if (o) setZ(nextZIndex()); setOpen(o) }}`
 * —— 初始 0 是有意的：挂载时不消耗共享计数器（SearchableSelect 这类实例多、
 * 多数从不打开的组件，用本 hook 会在挂载时白白 bump 全局 zCounter）。
 */

export function useOverlayZ(open: boolean): number {
  const [z, setZ] = useState(() => nextZIndex())
  useLayoutEffect(() => {
    if (open) setZ(nextZIndex())
  }, [open])
  return z
}
