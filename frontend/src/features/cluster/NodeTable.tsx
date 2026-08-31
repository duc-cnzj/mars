import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Tag, type Tone } from '@/components/ui'
import { fmtCpuCore, fmtMem, usageRate, type NodeMetric } from './board'

/** 节点调度状态 → 语义 tone（Ready 绿 / SchedulingDisabled 黄 / NotReady 红） */
const NODE_STATUS_TONE: Record<NodeMetric['status'], Tone> = {
  Ready: 'ok',
  NotReady: 'err',
  SchedulingDisabled: 'warn',
}

/** 使用率 → 填充色（<60 绿 / <85 黄 / ≥85 红） */
const usageColor = (pct: number): string =>
  pct >= 85 ? 'var(--err)' : pct >= 60 ? 'var(--warn)' : 'var(--ok)'

/** 轻量使用率进度条（shadcn Progress 的 indicator 色是写死的 primary，按用量着色需自绘） */
function UsageBar({ pct }: { pct: number }) {
  return (
    <div className="h-1.5 w-full overflow-hidden rounded-full bg-raised">
      <div
        className="h-full rounded-full transition-all duration-500"
        style={{ width: `${Math.min(100, pct)}%`, backgroundColor: usageColor(pct) }}
      />
    </div>
  )
}

/** 节点资源表：每节点 CPU/内存「用量 / 容量 + 用量着色进度条」，状态用 Tag 表达 */
export function NodeTable({ nodes }: { nodes: NodeMetric[] }) {
  const { t } = useTranslation()

  return (
    <section className="rounded-lg border border-line bg-surface">
      <header className="flex items-center gap-2 border-b border-line px-4 py-3">
        <span className="text-[13px] font-semibold text-ink">{t('cluster.nodeTitle')}</span>
        <span className="font-mono text-[11px] text-faint">{nodes.length}</span>
      </header>

      {/* 节点过多时固定最大高度内滚动（表头常驻，省掉整页无限拉长） */}
      <div className="max-h-[400px] divide-y divide-line overflow-y-auto">
        {nodes.map((n) => {
          // 复用 usageRate 的 cap>0 守卫：容量缺失/0 时返回 0，避免 Infinity% 满红条
          const cpuPct = usageRate(n.cpuUsage, n.cpuCapacity)
          const memPct = usageRate(n.memUsage, n.memCapacity)
          return (
            <div key={n.name} className="grid grid-cols-1 items-center gap-2 px-4 py-2.5 sm:grid-cols-3 lg:grid-cols-4">
              {/* 节点名 + 角色 */}
              <div className="min-w-0">
                <div className="truncate font-mono text-[12px] text-ink">{n.name}</div>
                <div className="mt-0.5 flex items-center gap-1.5">
                  <Tag tone={NODE_STATUS_TONE[n.status]} dot={false} className="px-1.5">
                    {n.status}
                  </Tag>
                  <span className="font-mono text-[10px] uppercase text-faint">{n.role}</span>
                </div>
              </div>

              {/* CPU：用量/容量 + 进度条 */}
              <div className="min-w-0">
                <div className="flex items-center justify-between text-[11px]">
                  <span className="text-faint">{t('cluster.cpuShort')}</span>
                  <span className="font-mono tabular-nums text-mute">
                    {fmtCpuCore(n.cpuUsage)} / {fmtCpuCore(n.cpuCapacity)}
                  </span>
                </div>
                <div className="mt-1">
                  <UsageBar pct={cpuPct} />
                </div>
              </div>

              {/* 内存：用量/容量 + 进度条 */}
              <div className="min-w-0">
                <div className="flex items-center justify-between text-[11px]">
                  <span className="text-faint">{t('cluster.memShort')}</span>
                  <span className="font-mono tabular-nums text-mute">
                    {fmtMem(n.memUsage)} / {fmtMem(n.memCapacity)}
                  </span>
                </div>
                <div className="mt-1">
                  <UsageBar pct={memPct} />
                </div>
              </div>

              {/* 请求量（调度视角） */}
              <div className="hidden min-w-0 text-right lg:block">
                <div className="flex items-center justify-end gap-3 text-[11px]">
                  <span className="text-faint">{t('cluster.reqCpu')}</span>
                  <span className={cn('font-mono tabular-nums', cpuPct >= 60 ? 'text-warn' : 'text-mute')}>
                    {fmtCpuCore(n.cpuRequest)}
                  </span>
                </div>
                <div className="mt-0.5 flex items-center justify-end gap-3 text-[11px]">
                  <span className="text-faint">{t('cluster.reqMem')}</span>
                  <span className={cn('font-mono tabular-nums', memPct >= 60 ? 'text-warn' : 'text-mute')}>
                    {fmtMem(n.memRequest)}
                  </span>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}
