import type { ReactNode } from 'react'
import { Badge } from './shadcn/badge'

export type Tone = 'ok' | 'warn' | 'err' | 'info' | 'accent' | 'mute'

/** tone → 文字/底色（语义 token，随主题换肤） */
const tones: Record<Tone, string> = {
  ok: 'text-ok bg-ok-soft',
  warn: 'text-warn bg-warn-soft',
  err: 'text-err bg-err-soft',
  info: 'text-info bg-info-soft',
  accent: 'text-primary bg-primary-soft',
  mute: 'text-muted-foreground bg-muted',
}

/** tone → 前置圆点色 */
const dots: Record<Tone, string> = {
  ok: 'bg-ok',
  warn: 'bg-warn',
  err: 'bg-err',
  info: 'bg-info',
  accent: 'bg-primary',
  mute: 'bg-muted-foreground',
}

/**
 * 语义状态标签：圆点 + 彩色底。基于 shadcn Badge 渲染，
 * tone 是应用域的状态语义（ok/warn/err/info/accent/mute），走语义 token 换肤。
 */
export function Tag({
  tone = 'ok',
  children,
  dot = true,
  className = '',
}: {
  tone?: Tone
  children: ReactNode
  dot?: boolean
  className?: string
}) {
  return (
    <Badge className={`gap-1.5 px-2 py-0.5 text-[11px] font-medium ${tones[tone]} ${className}`}>
      {dot && <span className={`h-1.5 w-1.5 rounded-full ${dots[tone]}`} />}
      {children}
    </Badge>
  )
}
