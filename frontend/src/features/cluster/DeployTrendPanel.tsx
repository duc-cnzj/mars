import { useId, useState, type MouseEvent as ReactMouseEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import {
  DEPLOY_TREND_DAYS,
  DEPLOY_TREND_RANGES,
  type DeployTrendRange,
  useDeployTrend,
} from './useDeployTrend'

/** 曲线 SVG 内部坐标系（viewBox 恒 100×40，preserveAspectRatio=none 拉伸铺满容器） */
const W = 100
const H = 40
const PAD = 3
/** x 轴刻度个数（近 30 天每约 6 天一个标签） */
const TICK_N = 6

/** 数值序列 Catmull-Rom → 三次贝塞尔平滑折线路径（与 AreaSpark 同思路） */
function smoothPath(ys: number[]): string {
  if (ys.length === 0) return ''
  const pts = ys.map((y, i) => ({ x: (i / (ys.length - 1)) * W, y }))
  let d = `M${pts[0].x.toFixed(2)},${pts[0].y.toFixed(2)}`
  for (let i = 0; i < pts.length - 1; i++) {
    const p0 = pts[i - 1] ?? pts[i]
    const p1 = pts[i]
    const p2 = pts[i + 1]
    const p3 = pts[i + 2] ?? p2
    const c1x = p1.x + (p2.x - p0.x) / 6
    const c1y = p1.y + (p2.y - p0.y) / 6
    const c2x = p2.x - (p3.x - p1.x) / 6
    const c2y = p2.y - (p3.y - p1.y) / 6
    d += ` C${c1x.toFixed(2)},${c1y.toFixed(2)} ${c2x.toFixed(2)},${c2y.toFixed(2)} ${p2.x.toFixed(2)},${p2.y.toFixed(2)}`
  }
  return d
}

/**
 * 每日部署趋势面板（集群总览区块）：每日部署次数曲线，默认近 30 天、可切 30/60/90 天窗口。
 * 形态对齐看板四卡基因：语义类样式 + 零依赖 SVG（Catmull-Rom 平滑 + 面积渐变）＋ x 轴日期刻度 +
 * hover 十字线定位采样点并出气泡（日期·次数）；顶部统计行给日均/峰值/总计三个读数。
 * 数据来自 useDeployTrend 拉 /api/admin/cluster/deploy_trend 真端点（切档即按新窗口重拉）。
 */
export function DeployTrendPanel() {
  const { t } = useTranslation()
  const id = useId()
  // 趋势窗口档位：默认 30 天，可切 30/60/90（切档即按新窗口重拉，日均随窗口长度重算）
  const [rangeDays, setRangeDays] = useState<DeployTrendRange>(DEPLOY_TREND_DAYS)
  const { counts, dates, total, dailyAvg, peak, peakIndex } = useDeployTrend(rangeDays)
  // hover 对齐到的采样点下标（null = 未悬停）
  const [hover, setHover] = useState<number | null>(null)

  const n = counts.length
  // y 轴满量程：取峰值兜底 1，杜绝除零
  const yMax = Math.max(...counts, 1)
  // 采样点 y 归一化到 [PAD, H-PAD]：c=yMax 顶格、c=0 落底
  const yOf = (i: number) => H - PAD - (counts[i] / yMax) * (H - 2 * PAD)

  // 横向网格线（25/50/75% 高度，faint 低透明度，只给读数尺度不给数字——精确值由 hover 气泡给）
  const gridYs = [0.25, 0.5, 0.75].map((f) => H - PAD - f * (H - 2 * PAD))

  const ys = counts.map((_, i) => yOf(i))
  const line = smoothPath(ys)
  const area = `${line} L${W},${H - PAD} L0,${H - PAD} Z`

  /** hover：DOM 横向位置 → 数据下标（与渲染无关） */
  const pctX = (i: number) => (n <= 1 ? 50 : (i / (n - 1)) * 100)
  const onMove = (e: ReactMouseEvent<HTMLDivElement>) => {
    if (n === 0) return
    const rect = e.currentTarget.getBoundingClientRect()
    const ratio = (e.clientX - rect.left) / (rect.width || 1)
    setHover(Math.min(n - 1, Math.max(0, Math.round(ratio * (n - 1)))))
  }

  // x 轴刻度下标：首尾 + 均匀间隔（避开重叠）
  const ticks = Array.from({ length: TICK_N }, (_, i) => Math.round((i * (n - 1)) / (TICK_N - 1)))

  // 峰值读数标注项：日均/峰值（含日期）/总计
  const metrics: { label: string; value: string; sub?: string }[] = [
    { label: t('cluster.deployTrendAvg'), value: t('cluster.deployTrendUnit', { count: dailyAvg }) },
    { label: t('cluster.deployTrendPeak'), value: `${peak}`, sub: dates[peakIndex] },
    { label: t('cluster.deployTrendTotal'), value: `${total}` },
  ]

  /** 档位切换 chip 样式：选中项主色描边 + 浅底（对齐 TopPods 的 CPU/内存切换形态） */
  const chipCls = (active: boolean) =>
    `rounded-md border px-1.5 py-0.5 text-[11px] font-mono transition-colors ${
      active
        ? 'border-primary bg-primary-soft font-medium text-primary'
        : 'border-line text-mute hover:border-primary hover:text-primary'
    }`

  return (
    <section className="rounded-lg border border-line bg-surface p-4">
      {/* 标题行：图标 + 面板名 + 时间范围档位切换（30/60/90） */}
      <div className="flex flex-wrap items-center gap-2">
        <Icon name="rocket" className="size-4 text-primary" />
        <span className="text-[13px] font-medium text-ink">{t('cluster.deployTrendTitle')}</span>
        <span className="rounded bg-bg px-1.5 py-0.5 text-[11px] text-faint">
          {t('cluster.deployTrendScope', { days: rangeDays })}
        </span>
        {/* 窗口档位切换：切档即重拉该窗口（useDeployTrend 按 days 重取），hover 复位防旧下标越界 */}
        <div className="ml-auto flex items-center gap-1">
          {DEPLOY_TREND_RANGES.map((d) => (
            <button
              key={d}
              type="button"
              aria-pressed={rangeDays === d}
              onClick={() => {
                setHover(null)
                setRangeDays(d)
              }}
              className={chipCls(rangeDays === d)}
            >
              {d}
            </button>
          ))}
        </div>
      </div>

      {/* 统计行：日均 / 峰值（含日期）/ 总计 */}
      <dl className="mt-3 flex flex-wrap items-end gap-x-8 gap-y-2">
        {metrics.map((m) => (
          <div key={m.label}>
            <dt className="text-[11px] text-faint">{m.label}</dt>
            <dd className="mt-0.5 flex items-baseline gap-1.5 font-mono text-[18px] font-semibold leading-none tabular-nums text-ink">
              {m.value}
              {m.sub && <span className="text-[11px] font-normal text-faint">{m.sub}</span>}
            </dd>
          </div>
        ))}
      </dl>

      {/* 曲线区：SVG 平滑线/面积/网格 + HTML hover 叠加（十字线/采样点/气泡） */}
      <div
        className="relative mt-2 h-40 cursor-crosshair"
        onMouseMove={onMove}
        onMouseLeave={() => setHover(null)}
      >
        <svg viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none" className="absolute inset-0 h-full w-full" aria-hidden>
          <defs>
            <linearGradient id={id} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="var(--primary)" stopOpacity="0.25" />
              <stop offset="100%" stopColor="var(--primary)" stopOpacity="0" />
            </linearGradient>
          </defs>
          {/* 网格线 + 底轴线 */}
          {gridYs.map((y) => (
            <line key={y} x1="0" x2={W} y1={y} y2={y} stroke="var(--text-faint)" strokeOpacity="0.18" strokeWidth={1} vectorEffect="non-scaling-stroke" />
          ))}
          <line x1="0" x2={W} y1={H - PAD} y2={H - PAD} stroke="var(--text-faint)" strokeOpacity="0.3" strokeWidth={1} vectorEffect="non-scaling-stroke" />
          {area && <path d={area} fill={`url(#${id})`} />}
          {line && (
            <path
              d={line}
              fill="none"
              stroke="var(--primary)"
              strokeWidth={1.5}
              strokeLinejoin="round"
              strokeLinecap="round"
              vectorEffect="non-scaling-stroke"
            />
          )}
        </svg>

        {/* hover 叠加：纵向十字线 + 采样点圆点 + 日期·次数气泡 */}
        {hover !== null && n > 0 && (
          <>
            <div
              className="pointer-events-none absolute inset-y-0 -translate-x-1/2"
              style={{ left: `${pctX(hover)}%`, width: 1, backgroundColor: 'var(--primary)', opacity: 0.4 }}
            />
            {/* 数据点标记：按采样点 y 落位（容器拉伸线性，svg 内部 y 占高比 = 容器 top% ） */}
            <div
              className="pointer-events-none absolute size-2 -translate-x-1/2 -translate-y-1/2 rounded-full"
              style={{ left: `${pctX(hover)}%`, top: `${(yOf(hover) / H) * 100}%`, backgroundColor: 'var(--primary)', boxShadow: '0 0 0 3px color-mix(in srgb, var(--primary) 25%, transparent)' }}
            />
            {/* 气泡：优先出在线条上方；曲线贴近顶缘时改到下方，避免盖住标题/出界 */}
            <div
              className="pointer-events-none absolute z-50 whitespace-nowrap rounded bg-black/85 px-1.5 py-0.5 font-mono text-[10px] text-white"
              style={{
                left: `${Math.min(92, Math.max(8, pctX(hover)))}%`,
                top: `${(yOf(hover) / H) * 100}%`,
                transform:
                  yOf(hover) / H < 0.22
                    ? 'translate(-50%, 12px)'
                    : 'translate(-50%, calc(-100% - 12px))',
              }}
            >
              {dates[hover]} · {t('cluster.deployTrendUnit', { count: counts[hover] })}
            </div>
          </>
        )}
      </div>

      {/* x 轴日期刻度 */}
      <div className="relative mt-1.5 h-4">
        {ticks.map((idx) => (
          <span
            key={idx}
            className="absolute -translate-x-1/2 text-[10px] tabular-nums text-faint"
            style={{ left: `${pctX(idx)}%` }}
          >
            {dates[idx]}
          </span>
        ))}
      </div>
    </section>
  )
}
