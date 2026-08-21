import type { Tone } from './Tag'

const dots: Record<Tone, string> = {
  ok: 'bg-ok',
  warn: 'bg-warn',
  err: 'bg-err',
  info: 'bg-info',
  accent: 'bg-primary',
  mute: 'bg-faint',
}

/** 语义状态圆点 */
export function StatusDot({ tone = 'mute' }: { tone?: Tone }) {
  return <span className={`inline-block h-2 w-2 shrink-0 rounded-full ${dots[tone]}`} />
}
