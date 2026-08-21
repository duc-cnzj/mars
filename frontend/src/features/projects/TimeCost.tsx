import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Icon } from '../../components/icons'
import { formatSeconds } from '@/lib/format'

/** 部署耗时计时：running 期间以 100ms 刷新，结束后定格。颜色随耗时渐变（<10s 绿 / <30s 黄 / <50s 粉 / <70s 橙 / 其余深红），与旧版 TimeCost 一致。 */
export function TimeCost({ running }: { running: boolean }) {
  const { t } = useTranslation()
  const [startAt, setStartAt] = useState<number | null>(null)
  const [now, setNow] = useState(() => Date.now())
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    if (running) {
      setStartAt(Date.now())
      setNow(Date.now())
      timerRef.current = setInterval(() => setNow(Date.now()), 100)
    } else if (timerRef.current) {
      clearInterval(timerRef.current)
      timerRef.current = null
    }
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
      timerRef.current = null
    }
  }, [running])

  const seconds = startAt ? (now - startAt) / 1000 : 0
  // 语义色阶：<10s 绿 / <30s 黄 / 其余红（走 ok/warn/err token 随主题换肤，不再硬编码 hex）
  const tone = seconds < 10 ? 'text-ok' : seconds < 30 ? 'text-warn' : 'text-err'

  return (
    <span
      className={`flex items-center gap-1 font-mono text-[11px] tabular-nums ${tone}`}
      title={t('events.duration')}
    >
      <Icon name="clock" className="text-[12px]" />
      {formatSeconds(seconds)}s
    </span>
  )
}
