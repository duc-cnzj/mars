import { useTranslation } from 'react-i18next'
import { fmtCpuMilli, fmtMem, type PodMetric } from './board'

/** 排行展示条数（对齐服务端 Top N 语义） */
const POD_LIMIT = 10

/** Top Pod 排行维度：cpu（毫核）/ mem（字节），与 useResourceBoard 的 topSort 状态同源 */
type SortKey = 'cpu' | 'mem'

/**
 * 全集群 Top Pod 排行：按 CPU 或内存用量降序（topSort 受控，切换后由父级重新拉取
 * 后端对应维度的 TopN——前端对已拉取列表按同维度降序重排 + 截断前 10：重排幂等，
 * 仅兜底截断，保证「内存 Top 10」是真实的内存大户）。
 * 定位「资源占用大户一目了然」，据此定位异常突刺（如 CI 跑满、某服务内存泄漏）。
 */
export function TopPods({ pods, topSort, onTopSortChange }: {
  pods: PodMetric[]
  topSort: SortKey
  onTopSortChange: (s: SortKey) => void
}) {
  const { t } = useTranslation()
  const top = [...pods]
    .sort((a, b) => (topSort === 'mem' ? b.memoryBytes - a.memoryBytes : b.cpuMilli - a.cpuMilli))
    .slice(0, POD_LIMIT)

  const chipCls = (active: boolean) =>
    `rounded-md border px-1.5 py-0.5 text-[11px] transition-colors ${
      active
        ? 'border-primary bg-primary-soft font-medium text-primary'
        : 'border-line text-mute hover:border-primary hover:text-primary'
    }`

  return (
    <section className="rounded-lg border border-line bg-surface">
      <header className="flex items-center gap-2 border-b border-line px-4 py-3">
        <span className="text-[13px] font-semibold text-ink">{t('cluster.topPodsTitle')}</span>
        <span className="font-mono text-[11px] text-faint">TOP {top.length}</span>
        {/* 排行维度切换：CPU / 内存（切换即重新拉取后端对应维度 TopN） */}
        <div className="ml-auto flex items-center gap-1">
          <button type="button" onClick={() => onTopSortChange('cpu')} className={chipCls(topSort === 'cpu')}>
            {t('cluster.topSortCpu')}
          </button>
          <button type="button" onClick={() => onTopSortChange('mem')} className={chipCls(topSort === 'mem')}>
            {t('cluster.topSortMem')}
          </button>
        </div>
      </header>

      <div className="divide-y divide-line">
        {/* 表头（常驻，不随行区滚动） */}
        <div className="grid grid-cols-[1fr_auto_auto] items-center gap-3 px-4 py-2 text-[11px] text-faint">
          <span>{t('cluster.podCol')}</span>
          <span className="w-20 text-right">{t('cluster.cpuShort')}</span>
          <span className="w-24 text-right">{t('cluster.memShort')}</span>
        </div>
        {/* 行区：Pod 过多时固定最大高度内滚动（对齐 NodeTable 的 max-h-[400px] 滚动模式） */}
        <div className="max-h-[400px] divide-y divide-line overflow-y-auto">
          {top.map((p) => (
            <div
              key={`${p.namespace}/${p.pod}`}
              className="grid grid-cols-[1fr_auto_auto] items-center gap-3 px-4 py-2.5"
            >
              <div className="min-w-0">
                <div className="truncate font-mono text-[12px] text-ink" title={p.pod}>
                  {p.pod}
                </div>
                <div className="truncate font-mono text-[10px] text-faint">{p.namespace}</div>
              </div>
              <div className="w-20 shrink-0 text-right font-mono text-[11px] tabular-nums text-mute">
                {fmtCpuMilli(p.cpuMilli)}
              </div>
              <div className="w-24 shrink-0 text-right font-mono text-[11px] tabular-nums text-mute">
                {fmtMem(p.memoryBytes)}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
