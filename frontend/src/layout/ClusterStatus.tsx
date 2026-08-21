import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../api/client'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/shadcn/popover'
import { useWebsocket } from '../realtime/useWebsocket'
import type { components } from '../api/schema'

type ClusterInfo = components['schemas']['websocket.ClusterInfo']

/** 状态 → 语义类映射：状态色固定（绿/黄/红），不随主题变化，与旧版配色一致 */
const STATUS_CLASS: Record<string, { dot: string; text: string }> = {
  health: { dot: 'bg-[#6ee7b7] shadow-[0_0_6px_#a7f3d0]', text: 'text-[#6ee7b7]' },
  // 闪烁强调状态，reduced-motion 下关闭（黄/红静态色仍可辨识状态）
  'not good': { dot: 'bg-[#f59e0b] animate-pulse motion-reduce:animate-none shadow-[0_0_6px_#fbbf24]', text: 'text-[#f59e0b]' },
  bad: { dot: 'bg-[#f87171] animate-pulse motion-reduce:animate-none shadow-[0_0_6px_#fca5a5]', text: 'text-[#f87171]' },
}

/**
 * 集群状态灯（还原旧版 ClusterInfo）：health 绿 / not good 黄闪烁 / bad 红闪烁，
 * 点击展开资源余量详情。数据源双通道——WS ClusterInfoSync 实时帧优先（后端每 15s 推送，
 * 状态灯不再「加载后永远 stale」），GET /api/cluster_info 作为即时种子兜底
 * （WS 首帧要等推送周期，首屏先拉一版立即亮灯；WS 未连接/失败时仍能展示）。
 * 用 Popover（键盘可聚焦、触屏可点）而非纯 CSS hover，避免详情对键盘/触屏不可达。
 * 请求在途且无任何数据时渲染灰色呼吸点占位（骨架屏，避免布局跳动）；失败/未知状态才隐藏。
 */
export function ClusterStatus() {
  const { t } = useTranslation()
  // WS 实时集群信息（后端每 15s 推送 ClusterInfoSync），优先展示
  const { clusterInfo: wsClusterInfo } = useWebsocket()
  const [info, setInfo] = useState<ClusterInfo | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let alive = true
    setLoading(true)
    api
      .GET('/api/cluster_info')
      .then(({ data }) => {
        if (alive) setInfo(data?.item ?? null)
      })
      .finally(() => {
        if (alive) setLoading(false)
      })
    return () => {
      alive = false
    }
  }, [])

  // WS 数据优先，REST 兜底：WS 首帧要等 15s 推送，REST 先拉一版立即亮灯；WS 未连接/失败时仍能展示
  const active = wsClusterInfo ?? info

  // 请求在途且无任何数据：灰色呼吸点占位（骨架屏），包裹结构与加载后完全一致，位置不跳动
  if (loading && !active) {
    return (
      <div className="flex h-8 items-center" aria-busy>
        <button
          type="button"
          className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-primary-foreground"
          aria-label={t('header.clusterStatus')}
        >
          <span className="h-2.5 w-2.5 animate-pulse rounded-full bg-primary-foreground/25 motion-reduce:animate-none" />
        </button>
      </div>
    )
  }
  // 无数据 / 未知状态 → 不渲染，避免悬挂红叉
  if (!active) return null
  const cls = STATUS_CLASS[active.status]
  if (!cls) return null

  const rows: [string, string][] = [
    [t('header.cpuFree'), active.freeRequestCpu],
    [t('header.memoryFree'), active.freeRequestMemory],
    [t('header.cpuRate'), active.requestCpuRate],
    [t('header.memoryRate'), active.requestMemoryRate],
    [t('header.cpuTotal'), active.totalCpu],
    [t('header.memoryTotal'), active.totalMemory],
  ]

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="flex h-8 items-center gap-2 rounded-lg px-2 py-1.5 text-primary-foreground"
          aria-label={t('header.clusterStatus')}
        >
          <span className={`h-2.5 w-2.5 rounded-full motion-reduce:animate-none ${cls.dot}`} />
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" sideOffset={8} className="w-56 p-3">
        <div className="mb-1 flex items-center gap-1.5">
          <span className={`h-2 w-2 rounded-full ${cls.dot}`} />
          <span className="text-[12px] font-medium text-ink">{t('header.clusterStatus')}</span>
          <span className={`ml-auto font-mono text-[11px] ${cls.text}`}>{active.status}</span>
        </div>
        <dl className="space-y-0.5">
          {rows.map(([label, value]) => (
            <div key={label} className="flex items-center justify-between text-[11px]">
              <dt className="text-faint">{label}</dt>
              <dd className="font-mono text-mute">{value}</dd>
            </div>
          ))}
        </dl>
      </PopoverContent>
    </Popover>
  )
}
