import { useEffect, useRef, useState } from 'react'
import { nextZIndex } from '@/lib/zIndex'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/shadcn/tooltip'

interface EndpointUrlProps {
  url: string
  /** 是否为可点击链接（不经 http 前缀判断，由调用方判定） */
  isLink?: boolean
  /** 应用到内层 <a>/<span> 的额外样式类（各调用处字体需求不同：font-mono / text-[12px] 等） */
  className?: string
}

/**
 * 端点 URL 展示组件：外层 span 配 min-w-0 flex-1 truncate 受父 flex 行约束，
 * 内层 <a>/<span> 配受控 Tooltip（截断 + 悬浮才显示完整 URL）。
 * controlled 模式：open={truncated && hover} + onOpenChange={() => {}}，
 * z-index 用 nextZIndex() 避让弹窗。
 * 未截断时裸渲染（零 Tooltip 组件树开销）。
 */
export function EndpointUrl({ url, isLink, className = '' }: EndpointUrlProps) {
  const ref = useRef<HTMLSpanElement>(null)
  const [truncated, setTruncated] = useState(false)
  const [hover, setHover] = useState(false)
  const [tipZ, setTipZ] = useState(50)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const measure = () => setTruncated(el.scrollWidth > el.clientWidth + 1)
    measure()
    const ro = new ResizeObserver(measure)
    ro.observe(el)
    return () => ro.disconnect()
  }, [url])

  const linkCls = `text-primary hover:underline ${className}`.trim()
  const textCls = `text-ink ${className}`.trim()

  const inner = isLink ? (
    <a
      href={url}
      target="_blank"
      rel="noreferrer"
      translate="no"
      className={linkCls}
      onMouseEnter={() => { setTipZ(nextZIndex()); setHover(true) }}
      onMouseLeave={() => setHover(false)}
    >
      {url}
    </a>
  ) : (
    <span
      translate="no"
      className={textCls}
      onMouseEnter={() => { setTipZ(nextZIndex()); setHover(true) }}
      onMouseLeave={() => setHover(false)}
    >
      {url}
    </span>
  )

  // 未截断 + 未悬浮：裸渲染，省掉 TooltipProvider/Tooltip 组件树
  if (!truncated && !hover) {
    return <span ref={ref} className="min-w-0 flex-1 truncate">{inner}</span>
  }

  return (
    <span ref={ref} className="min-w-0 flex-1 truncate">
      <TooltipProvider delayDuration={100}>
        <Tooltip open={truncated && hover} onOpenChange={() => {}}>
          <TooltipTrigger asChild>{inner}</TooltipTrigger>
          <TooltipContent side="top" style={{ zIndex: tipZ }} className="text-wrap text-[12px] leading-relaxed">
            {url}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </span>
  )
}