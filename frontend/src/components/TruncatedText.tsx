import { useEffect, useRef, useState } from 'react'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/shadcn/tooltip'
import { cn } from '@/lib/utils'

/**
 * 单行截断文本：超出容器宽度时悬停显示全文（shadcn Tooltip——原生 title 有延迟且无样式，用户感知不到）。
 *
 * 恒走 Tooltip 分支、open 受控（truncated && hover，onOpenChange 置空）：不按 truncated 在裸文本 ↔
 * Tooltip 之间切换 DOM，也不把 onOpenChange 接真实 setState（Radix 受控开合逻辑会和鼠标事件互斗，
 * 悬停一开就关）——对齐 ProjectRow / Elements FieldLabel 的既有模式。
 *
 * 截断判定 scrollWidth > clientWidth（+1px 缓冲防亚像素抖动），ResizeObserver 跟随容器/文字变化重算；
 * ref 必须挂在真正发生 truncate 的 span 上——量在别处恒不溢出、永远判成「没省略」。
 */
export function TruncatedText({ text, className }: { text: string; className?: string }) {
  const ref = useRef<HTMLSpanElement>(null)
  const [truncated, setTruncated] = useState(false)
  const [hover, setHover] = useState(false)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const measure = () => setTruncated(el.scrollWidth > el.clientWidth + 1)
    measure()
    const ro = new ResizeObserver(measure)
    ro.observe(el)
    return () => ro.disconnect()
  }, [text])

  return (
    <TooltipProvider delayDuration={100}>
      <Tooltip open={truncated && hover} onOpenChange={() => {}}>
        <TooltipTrigger asChild>
          <span
            ref={ref}
            onMouseEnter={() => setHover(true)}
            onMouseLeave={() => setHover(false)}
            className={cn('truncate', className)}
          >
            {text}
          </span>
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-[min(320px,80vw)] break-words">
          {text}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
