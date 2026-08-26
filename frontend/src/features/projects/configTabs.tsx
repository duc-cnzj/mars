import type { KeyboardEvent, ReactNode } from 'react'
import { cn } from '@/lib/utils'

/**
 * 底部「配置文件 / 各 TextArea」tab 条的方向键导航（roving tabindex，对齐原 shadcn Tabs 的键盘行为）：
 * ←/→ 循环移动焦点并激活（聚焦即切换），Home/End 跳首尾。事件在按钮上触发后冒泡到 tablist
 * 容器统一处理；target 可能是按钮内 span，向上找 [role=tab] 定位当前项。
 * 配置更新 TabEdit 与创建项目 CreateProjectModal 共用——两页底部 tab 交互必须一致，抽共享防漂移。
 */
export function handleTablistKeyDown(e: KeyboardEvent<HTMLDivElement>) {
  const tabs = Array.from(e.currentTarget.querySelectorAll<HTMLButtonElement>('[role="tab"]'))
  if (tabs.length === 0) return
  const curEl = (e.target as HTMLElement).closest<HTMLButtonElement>('[role="tab"]')
  const cur = curEl ? tabs.indexOf(curEl) : -1
  let next = cur
  switch (e.key) {
    case 'ArrowRight':
      next = cur < 0 ? 0 : (cur + 1) % tabs.length
      break
    case 'ArrowLeft':
      next = cur < 0 ? tabs.length - 1 : (cur - 1 + tabs.length) % tabs.length
      break
    case 'Home':
      next = 0
      break
    case 'End':
      next = tabs.length - 1
      break
    default:
      return
  }
  e.preventDefault()
  tabs[next].focus()
  tabs[next].click()
}

/**
 * 底部配置 tab 条的单按钮（pill 胶囊）：尺寸对齐部署按钮（Button size=xs：h-6 + text-xs），
 * 选中态主题色柔色变体（primary-soft 底 + primary 文字），比实心填充轻、比无色默认态鲜明。
 * TabEdit 与创建项目 CreateProjectModal 共用——两页底部 tab 交互与视觉必须一致。
 * 长标题由外部包 span truncate + max-w 截断；按钮自身 shrink-0，容器不足时横向滚动不挤压。
 * roving tabindex：仅选中项可 Tab 聚焦（tabIndex=0），其余 -1，方向键在容器 onKeyDown 统一移动。
 */
export function BottomTabButton({
  active,
  onClick,
  children,
  title,
}: {
  active: boolean
  onClick: () => void
  children: ReactNode
  /** 完整标题（长文本 tab 的 tooltip） */
  title?: string
}) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      tabIndex={active ? 0 : -1}
      title={title}
      onClick={onClick}
      className={cn(
        'flex shrink-0 items-center gap-1.5 whitespace-nowrap transition-colors',
        'h-6 rounded-md px-2 text-xs font-medium',
        active
          ? 'bg-primary/20 text-primary'
          : 'text-foreground/60 hover:text-foreground',
      )}
    >
      {children}
    </button>
  )
}
