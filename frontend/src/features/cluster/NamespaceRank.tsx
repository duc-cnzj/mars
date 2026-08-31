import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { fmtCpuMilli, fmtMem, type NamespaceMetric } from './board'

/** 排行展示条数（对齐服务端 Top N 语义） */
const RANK_LIMIT = 8

/** 排序维度 → 数据字段（CPU 毫核 / 内存字节） */
const SORT_FIELD = { cpu: 'cpuMilli', mem: 'memoryBytes' } as const

/** 排序维度类型：cpu（毫核）/ mem（字节） */
type SortKey = keyof typeof SORT_FIELD

/**
 * 命名空间资源排行：支持按 CPU / 内存维度切换排序（默认 CPU），
 * 条形图相对当前维度最大值归一化，副列展示另一维度 + Pod 数。
 * 排行前列即资源大户——管理员判断「哪个空间最耗资源」的第一抓手。
 */
export function NamespaceRank({ namespaces }: { namespaces: NamespaceMetric[] }) {
  const { t } = useTranslation()
  // 排序维度：默认 CPU，可切内存（资源大户判断 CPU/内存两维度都要看）
  const [sortKey, setSortKey] = useState<SortKey>('cpu')
  const field = SORT_FIELD[sortKey]
  const top = [...namespaces].sort((a, b) => b[field] - a[field]).slice(0, RANK_LIMIT)
  // 以最大值为 1 兜底：空列表 → Math.max(1)=1；全 0 数据（集群无指标）→ Math.max(1,0,...)=1，
  // 避免 0/0=NaN 的 width:"NaN%" 无效样式（条消失）
  const maxMain = Math.max(1, ...top.map((n) => n[field]))

  const chipCls = (active: boolean) =>
    `rounded-md border px-1.5 py-0.5 text-[11px] transition-colors ${
      active
        ? 'border-primary bg-primary-soft font-medium text-primary'
        : 'border-line text-mute hover:border-primary hover:text-primary'
    }`

  return (
    <section className="rounded-lg border border-line bg-surface">
      <header className="flex items-center gap-2 border-b border-line px-4 py-3">
        <span className="text-[13px] font-semibold text-ink">{t('cluster.namespaceTitle')}</span>
        <span className="font-mono text-[11px] text-faint">TOP {top.length}</span>
        {/* 排序维度切换：CPU / 内存 */}
        <div className="ml-auto flex items-center gap-1">
          <button type="button" onClick={() => setSortKey('cpu')} className={chipCls(sortKey === 'cpu')}>
            {t('cluster.nsSortCpu')}
          </button>
          <button type="button" onClick={() => setSortKey('mem')} className={chipCls(sortKey === 'mem')}>
            {t('cluster.nsSortMem')}
          </button>
        </div>
      </header>

      {/* 空间过多时固定最大高度内滚动（表头常驻；对齐 NodeTable 的 max-h-[400px] 滚动模式） */}
      <div className="max-h-[400px] space-y-2.5 overflow-y-auto px-4 py-3">
        {top.map((ns, i) => {
          const pct = (ns[field] / maxMain) * 100
          return (
            <div key={ns.namespace} className="flex items-center gap-3">
              {/* 名次 */}
              <span className="w-4 shrink-0 text-right font-mono text-[11px] tabular-nums text-faint">
                {i + 1}
              </span>
              {/* 命名空间 + 当前排序维度的用量条 */}
              <div className="min-w-0 flex-1">
                <div className="mb-1 flex items-baseline justify-between gap-2">
                  <span className="truncate font-mono text-[12px] text-ink">{ns.namespace}</span>
                  <span className="shrink-0 font-mono text-[11px] tabular-nums text-mute">
                    {sortKey === 'cpu' ? fmtCpuMilli(ns.cpuMilli) : fmtMem(ns.memoryBytes)}
                  </span>
                </div>
                <div className="h-1.5 w-full overflow-hidden rounded-full bg-raised">
                  <div
                    className="h-full rounded-full bg-primary/80 transition-all duration-500"
                    style={{ width: `${pct}%` }}
                  />
                </div>
              </div>
              {/* 另一维度 + Pod 数 */}
              <div className="hidden shrink-0 text-right text-[11px] text-faint sm:block">
                <div className="font-mono tabular-nums text-mute">
                  {sortKey === 'cpu' ? fmtMem(ns.memoryBytes) : fmtCpuMilli(ns.cpuMilli)}
                </div>
                <div>{t('cluster.podCount', { count: ns.podCount })}</div>
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}
