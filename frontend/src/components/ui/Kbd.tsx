import { cn } from '@/lib/utils'
import type { ReactNode } from 'react'

/** 键盘按键提示（⌘K / Ctrl K 等）。className 可覆盖字号/对齐，如放大 ⌘ 符号 */
export function Kbd({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <kbd
      className={cn(
        'rounded border border-line bg-surface px-1.5 py-0.5 font-mono text-[10px] leading-none text-mute',
        className,
      )}
    >
      {children}
    </kbd>
  )
}
