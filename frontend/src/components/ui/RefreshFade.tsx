import { Children, cloneElement, isValidElement, useMemo, type CSSProperties, type ReactNode } from 'react'

/** 错峰封顶：超过 10 项不再增加延迟（对齐 Repos/Events 的行入场节奏） */
const MAX_STAGGER = 10
const STAGGER_MS = 30

/** 渐入 class 常量（引用稳定：行 memo 浅比较可命中「父重渲但行未变」跳过重渲） */
const FADE_CLASS = 'animate-list-in'

/** 可注入 className/style 的子元素类型（本组件只对真正的元素节点做 cloneElement 注入） */
type FadeChild = React.ReactElement<{ className?: string; style?: CSSProperties }>

/**
 * 数据重取渐入容器：key=version 在每次取数完成后重挂载列表，逐行注入
 * animate-list-in + 错峰延迟，复刻 Repos/Events 的「重新获取数据时内容渐入」体验。
 * - 兼容 keep-last-frame 刷新：无需 skeleton 翻转，数据到位后整表重播一次淡入上浮，不闪断
 * - 用 cloneElement 注入 className/style，不包裹 DOM——不破坏子元素自身的
 *   border-b last:border-b-0 / divide-y / grid 布局语义
 * - style 引用按 (行数, version) 缓存：父级重渲但行未变时引用稳定，行 memo 浅比较通过不重渲；
 *   version bump 时新 style → 整表重播渐入。注入的 className/style 须由行组件转发到根元素才生效
 * - ⚠️ 轮询页勿「每帧 bump version」：会闪屏。轮询页只对手动刷新 bump version
 *   （见 ResourceBoard：fetchSnapshot 轮询静默、refresh 成功才 +version），
 *   进入/手动刷新各重播一次渐入，轮询帧静默不重播
 */
export function RefreshFade({ version, className, children }: {
  version: number
  className?: string
  children: ReactNode
}) {
  const count = Children.count(children)
  // 行数 / version 变化才重建 style 数组：父重渲但数据未变时复用同一批 style 对象 → 行 memo 生效
  const styles = useMemo(
    () => Array.from({ length: count }, (_, i) => ({ animationDelay: `${Math.min(i, MAX_STAGGER) * STAGGER_MS}ms` })),
    [count, version],
  )
  return (
    <div key={version} className={className}>
      {Children.map(children, (child, i) => {
        if (!isValidElement(child)) return child
        const el = child as FadeChild
        return cloneElement(el, {
          className: el.props.className ? `${el.props.className} ${FADE_CLASS}` : FADE_CLASS,
          style: el.props.style ? { ...el.props.style, ...styles[i] } : styles[i],
        })
      })}
    </div>
  )
}
