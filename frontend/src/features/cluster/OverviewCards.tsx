import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { AreaSpark } from '@/components/charts/AreaSpark'
import { Progress } from '@/components/ui'
import { Tag, type Tone } from '@/components/ui'
import type { ClusterOverview, ClusterStatus } from './board'

/** 集群健康状态 → 语义 tone（health 绿 / not good 黄 / bad 红，对齐 ClusterStatus 色系） */
const STATUS_TONE: Record<ClusterStatus, Tone> = {
  health: 'ok',
  'not good': 'warn',
  bad: 'err',
}

/** 状态 → 词条键（as const 保留字面量类型，过 i18next 资源键校验） */
const STATUS_KEY = {
  health: 'cluster.statusHealth',
  'not good': 'cluster.statusNotGood',
  bad: 'cluster.statusBad',
} as const

/**
 * 集群总览卡片组：健康状态 / CPU / 内存 / 请求率 四张卡。
 * CPU/内存卡用 AreaSpark 展示实时趋势（真实轮询驱动），请求率卡用进度条表达占比。
 */
export function OverviewCards({
  overview,
  namespaceCount,
  podCount,
  cpuTrend,
  memTrend,
}: {
  overview: ClusterOverview
  namespaceCount: number
  podCount: number
  cpuTrend: number[]
  memTrend: number[]
}) {
  const { t } = useTranslation()
  const nodeTotal = overview.nodeTotal

  return (
    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
      {/* 集群健康状态 */}
      <section className="rounded-lg border border-line bg-surface p-4">
        <div className="flex items-center gap-2 text-[12px] text-faint">
          <Icon name="gauge" className="size-4" />
          {t('cluster.clusterStatus')}
        </div>
        <div className="mt-3 flex items-center gap-2">
          <Tag tone={STATUS_TONE[overview.status]}>{t(STATUS_KEY[overview.status])}</Tag>
        </div>
        <dl className="mt-4 space-y-1.5 text-[12px]">
          <div className="flex items-center justify-between">
            <dt className="text-faint">{t('cluster.nodeReady')}</dt>
            <dd className="font-mono text-mute">{overview.nodeReady}/{nodeTotal}</dd>
          </div>
          <div className="flex items-center justify-between">
            <dt className="text-faint">{t('cluster.namespaceCount')}</dt>
            <dd className="font-mono text-mute">{namespaceCount}</dd>
          </div>
          <div className="flex items-center justify-between">
            <dt className="text-faint">{t('cluster.podCountLabel')}</dt>
            <dd className="font-mono text-mute">{podCount}</dd>
          </div>
        </dl>
      </section>

      {/* CPU 使用率 */}
      <section className="rounded-lg border border-line bg-surface p-4">
        <div className="flex items-center gap-2 text-[12px] text-faint">
          <Icon name="cpu" className="size-4" />
          {t('cluster.cpuUsage')}
        </div>
        <div className="mt-1 font-mono text-[26px] font-semibold leading-none tabular-nums text-ink">
          {overview.usageCpuRate.toFixed(1)}%
        </div>
        <div className="mt-2">
          <AreaSpark
            label={t('cluster.cpuRateTrend')}
            value=""
            points={cpuTrend}
            color="var(--chart-1)"
            height={10}
          />
        </div>
        <div className="mt-2 flex items-center justify-between text-[11px] text-faint">
          <span>
            {t('cluster.used')} <span className="font-mono text-mute">{overview.usedCpu}</span>
          </span>
          <span className="font-mono text-mute">{overview.totalCpu}</span>
        </div>
      </section>

      {/* 内存使用率 */}
      <section className="rounded-lg border border-line bg-surface p-4">
        <div className="flex items-center gap-2 text-[12px] text-faint">
          <Icon name="memory" className="size-4" />
          {t('cluster.memoryUsage')}
        </div>
        <div className="mt-1 font-mono text-[26px] font-semibold leading-none tabular-nums text-ink">
          {overview.usageMemRate.toFixed(1)}%
        </div>
        <div className="mt-2">
          <AreaSpark
            label={t('cluster.memRateTrend')}
            value=""
            points={memTrend}
            color="var(--ok)"
            height={10}
          />
        </div>
        <div className="mt-2 flex items-center justify-between text-[11px] text-faint">
          <span>
            {t('cluster.used')} <span className="font-mono text-mute">{overview.usedMemory}</span>
          </span>
          <span className="font-mono text-mute">{overview.totalMemory}</span>
        </div>
      </section>

      {/* 资源请求率（决定调度与状态） */}
      <section className="rounded-lg border border-line bg-surface p-4">
        <div className="flex items-center gap-2 text-[12px] text-faint">
          <Icon name="boxes" className="size-4" />
          {t('cluster.requestRate')}
        </div>
        <div className="mt-3 space-y-3">
          <div>
            <div className="mb-1 flex items-center justify-between text-[11px]">
              <span className="text-faint">{t('cluster.requestCpu')}</span>
              <span className="font-mono tabular-nums text-mute">{overview.requestCpuRate.toFixed(1)}%</span>
            </div>
            <Progress value={overview.requestCpuRate} className="h-1.5 bg-primary/15" />
          </div>
          <div>
            <div className="mb-1 flex items-center justify-between text-[11px]">
              <span className="text-faint">{t('cluster.requestMemory')}</span>
              <span className="font-mono tabular-nums text-mute">{overview.requestMemRate.toFixed(1)}%</span>
            </div>
            <Progress value={overview.requestMemRate} className="h-1.5 bg-primary/15" />
          </div>
        </div>
      </section>
    </div>
  )
}
