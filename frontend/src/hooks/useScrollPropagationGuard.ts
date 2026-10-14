import { useCallback, useRef } from 'react'

/**
 * 截断滚轮/触摸滚动事件向 document 的传播，保住「Portal 弹层内滚动容器」自身的滚动。
 *
 * 底层逻辑：Radix Portal 把弹层挂到 body，落在宿主 modal Dialog/Sheet 的 react-remove-scroll
 * shard 之外；该库的 document 级兜底（shard 外事件无队列条目 → 落 `!noIsolation` 分支，见
 * react-remove-scroll SideEffect）会把弹层内一切 wheel/touchmove 一律 preventDefault，弹层里的
 * 列表 / 预览文本因此滚不动（Radix #1159）。在滚动容器上原生截断传播，事件到不了 document，
 * 宿主锁就吞不掉弹层自身的滚动；非 modal 宿主下也只是让滚轮到边界后不再链回主页面，行为可接受。
 *
 * 只对「默认非模态的 Popover」必要——Select / DropdownMenu 等自带 RemoveScroll 的 primitive
 * 会自我保护，不中此招。
 *
 * 用法：返回一个 ref 回调，直接挂到弹层内的滚动容器上——
 *   const bindScrollGuard = useScrollPropagationGuard<HTMLDivElement>()
 *   <div ref={bindScrollGuard} className="overflow-y-auto">…</div>
 *
 * 用 ref 回调而非 useEffect：Popover 关闭即卸载内容、打开后经 Presence 后续提交才渲染，
 * useEffect 抓不到这段挂载时序（同 useWheelRedirect 的理由）。
 */
export function useScrollPropagationGuard<T extends HTMLElement = HTMLElement>() {
  const nodeRef = useRef<T | null>(null)
  return useCallback((node: T | null) => {
    if (nodeRef.current) {
      nodeRef.current.removeEventListener('wheel', stopScrollPropagation)
      nodeRef.current.removeEventListener('touchmove', stopScrollPropagation)
    }
    nodeRef.current = node
    if (!node) return
    node.addEventListener('wheel', stopScrollPropagation, { passive: true })
    node.addEventListener('touchmove', stopScrollPropagation, { passive: true })
  }, [])
}

/** 截断 wheel/touchmove 向 document 的传播（被动监听即可：只需拦住冒泡，不需 preventDefault）。 */
const stopScrollPropagation = (e: Event) => e.stopPropagation()
