import { useCallback, useEffect, useState, type CSSProperties } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { Icon } from '@/components/Icons'
import { Empty, RefreshFade, SkeletonGrid, Tag } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { api } from '@/api/client'
import type { components } from '@/api/schema'
import { fmtCpuMilli, fmtMem } from './board'

type ResourceProjectDto = components['schemas']['cluster.ResourceProject']
type ResourceNamespaceDto = components['schemas']['cluster.ResourceNamespace']
type ResourceWorkloadDto = components['schemas']['cluster.ResourceProjectWorkload']

/** 空间资源展示模型：后端 int64 字段经 JSON 序列化为字符串，统一转 number 再渲染/排序。 */
interface ResourceProjectView {
  name: string
  podCount: number
  cpuRequestMilli: number
  cpuUsageMilli: number
  memRequestBytes: number
  memUsageBytes: number
  workloads: ResourceWorkloadView[]
}

/** 项目内工作负载展示模型（Deployment/StatefulSet/DaemonSet 细分聚合） */
interface ResourceWorkloadView {
  kind: string
  name: string
  podCount: number
  cpuRequestMilli: number
  cpuUsageMilli: number
  memRequestBytes: number
  memUsageBytes: number
}

interface ResourceNamespaceView {
  name: string
  podCount: number
  cpuRequestMilli: number
  cpuUsageMilli: number
  memRequestBytes: number
  memUsageBytes: number
  projects: ResourceProjectView[]
}

/**
 * 占比（%）：request 与 usage 都 > 0 才构成「超申请」信号。
 * - request <= 0：无申请量，无从谈占比；
 * - usage <= 0：无实际用量（刚部署/未采集到指标），是「没跑起来」而非「申请多用
 *   得少」的可测量浪费——把它当 0% 会与真实低占比空间混淆、淹没排序信号。
 * 两种情况对称返回 null（不参与占比排序/警示，展示为「—」）。
 */
const ratioPct = (request: number, usage: number): number | null =>
  request > 0 && usage > 0 ? (usage / request) * 100 : null

/** 超申请阈值：实际用量不足申请的 30% 即标「超申请」——定位低占比高申请空间的入口 */
const OVER_REQUEST_THRESHOLD = 30

/** 排序维度 → 占比取值函数；null 在升序中恒排末尾（无申请量不可能超申请） */
const DIM_RATIO = {
  cpu: (ns: ResourceNamespaceView) => ratioPct(ns.cpuRequestMilli, ns.cpuUsageMilli),
  mem: (ns: ResourceNamespaceView) => ratioPct(ns.memRequestBytes, ns.memUsageBytes),
} as const

type SortKey = keyof typeof DIM_RATIO

/** 占比状态：超申请（浪费）/ 正常 / 超用（超过申请量）/ 无数据 */
type BarTone = 'ok' | 'warn' | 'err' | 'mute'

/** 占比 → 状态 tone：<30% 超申请(warn) / 30-100% 正常(ok) / >100% 超用(err) / null 无数据(mute) */
const ratioTone = (ratio: number | null): BarTone =>
  ratio === null ? 'mute' : ratio < OVER_REQUEST_THRESHOLD ? 'warn' : ratio > 100 ? 'err' : 'ok'

/** tone → 占比数字文字色（与进度条填充色同源，一眼对上状态） */
const TONE_TEXT: Record<BarTone, string> = {
  ok: 'text-ok',
  warn: 'text-warn',
  err: 'text-err',
  mute: 'text-faint',
}

/** tone → 进度条填充色：绿=正常 / 橙=超申请 / 红=超用 / 灰=无数据 */
const TONE_FILL: Record<BarTone, string> = {
  ok: 'bg-ok',
  warn: 'bg-warn',
  err: 'bg-err',
  mute: 'bg-line',
}

/** 工作负载 DTO → 视图模型（字符串数值转 number） */
const toWorkloadView = (w: ResourceWorkloadDto): ResourceWorkloadView => ({
  kind: w.kind,
  name: w.name,
  podCount: w.podCount,
  cpuRequestMilli: Number(w.cpuRequestMilli),
  cpuUsageMilli: Number(w.cpuUsageMilli),
  memRequestBytes: Number(w.memRequestBytes),
  memUsageBytes: Number(w.memUsageBytes),
})

/** 项目 DTO → 视图模型（字符串数值转 number，工作负载明细同步转换） */
const toProjectView = (p: ResourceProjectDto): ResourceProjectView => ({
  name: p.name,
  podCount: p.podCount,
  cpuRequestMilli: Number(p.cpuRequestMilli),
  cpuUsageMilli: Number(p.cpuUsageMilli),
  memRequestBytes: Number(p.memRequestBytes),
  memUsageBytes: Number(p.memUsageBytes),
  workloads: (p.workloads ?? []).map(toWorkloadView),
})

/** 命名空间 DTO → 视图模型（字符串数值转 number，项目明细同步转换） */
const toNamespaceView = (ns: ResourceNamespaceDto): ResourceNamespaceView => ({
  name: ns.name,
  podCount: ns.podCount,
  cpuRequestMilli: Number(ns.cpuRequestMilli),
  cpuUsageMilli: Number(ns.cpuUsageMilli),
  memRequestBytes: Number(ns.memRequestBytes),
  memUsageBytes: Number(ns.memUsageBytes),
  projects: ns.projects.map(toProjectView),
})

/**
 * 空间资源管理（管理员后台）：
 * 卡片网格展示每个管理命名空间的 Pod requests/实际用量占比——找「申请了很多 requests
 * 却用不到多少资源」的空间。默认按 CPU 占比升序（超申请者置顶），可切内存占比；超申请
 * 卡片橙色描边 + 徽标列超出的量，进度条按状态着色（绿=正常/橙=超申请/红=超用）。点卡片
 * 展开项目明细（含工作负载细分）。诊断页不轮询（手动刷新即可，避免后台空转）。
 */
export function ResourceManagement() {
  const { t } = useTranslation()
  const [namespaces, setNamespaces] = useState<ResourceNamespaceView[]>([])
  const [loading, setLoading] = useState(true)
  // 刷新中：手动刷新按钮转圈 + 禁用，避免连点
  const [refreshing, setRefreshing] = useState(false)
  // 渐入版本号：每次取数成功 +1，RefreshFade 依 key 重挂载网格重播渐入
  const [version, setVersion] = useState(0)
  // 排序维度：默认 CPU 占比（超申请最严重的空间置顶）
  const [sortKey, setSortKey] = useState<SortKey>('cpu')
  // 展开卡片：一次只展开一个空间的项目明细
  const [expanded, setExpanded] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    setRefreshing(true)
    try {
      const { data, error } = await api.GET('/api/admin/cluster/resources')
      if (error) throw new Error(error.message ?? String(error))
      if (!data) return
      setNamespaces(data.namespaces.map(toNamespaceView))
      // 取数成功 → 版本号 +1，整网格重播一次渐入（keep-last-frame，不闪断）
      setVersion((v) => v + 1)
    } catch {
      // 拉取失败静默保留上一帧快照，避免闪断；首帧失败即落空态
    } finally {
      // loading 统一收敛到 finally：成功/失败/空响应（!data 早退）三条路径都清，
      // 避免 !data 早退漏清 loading 卡死骨架屏（对齐其他后台页 .finally 清 loading 的模式）
      setLoading(false)
      setRefreshing(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  // 按当前维度占比升序：占比最低（申请多、用得少）者置顶；无申请量（null）恒排末尾
  const sorted = [...namespaces].sort((a, b) => {
    const ra = DIM_RATIO[sortKey](a) ?? Infinity
    const rb = DIM_RATIO[sortKey](b) ?? Infinity
    return ra - rb
  })

  /** 分段控件样式：容器内嵌按钮，选中项浮起（bg-surface + 阴影），未选中 muted */
  const segCls = (active: boolean) =>
    `rounded-md px-2.5 py-1 text-[12px] transition-colors ${
      active ? 'bg-surface font-medium text-ink shadow-sm' : 'text-mute hover:text-ink'
    }`

  return (
    <div className="flex flex-col gap-4">
      {/* 页头：标题 + 手动刷新 */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-[16px] font-semibold text-ink">{t('resources.title')}</h2>
        <Button variant="outline" size="sm" onClick={refresh} disabled={refreshing}>
          {refreshing ? (
            <Icon name="loader" className="size-3.5 animate-spin" />
          ) : (
            <Icon name="refresh" className="size-3.5" />
          )}
          {t('common.refresh')}
        </Button>
      </div>

      {/* 排序维度切换：分段控件（CPU 占比 / 内存占比）——本地即时重排，不加 loading */}
      <div className="flex items-center gap-2">
        <div className="flex rounded-lg border border-line bg-raised p-0.5">
          <button type="button" onClick={() => setSortKey('cpu')} className={segCls(sortKey === 'cpu')}>
            {t('resources.sortCpu')}
          </button>
          <button type="button" onClick={() => setSortKey('mem')} className={segCls(sortKey === 'mem')}>
            {t('resources.sortMem')}
          </button>
        </div>
        <span className="ml-auto hidden text-[11px] text-faint sm:block">{t('resources.sortHint')}</span>
      </div>

      {/* 内容区：relative 包裹以便遮罩定位；手动刷新时降透明度禁交互 + 居中 spinner（排序切换为本地即时重排不加 loading） */}
      <div className="relative">
        {loading ? (
          <SkeletonGrid count={6} />
        ) : sorted.length === 0 ? (
          <div className="rounded-lg border border-line bg-surface">
            <Empty icon="namespace" text={t('resources.empty')} />
          </div>
        ) : (
          <RefreshFade
            version={version}
            className={`grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3 ${refreshing ? 'pointer-events-none opacity-40' : ''}`}
          >
            {sorted.map((ns) => (
              <NamespaceCard
                key={ns.name}
                ns={ns}
                sortKey={sortKey}
                open={expanded === ns.name}
                onToggle={() => setExpanded(expanded === ns.name ? null : ns.name)}
              />
            ))}
          </RefreshFade>
        )}
        {/* 手动刷新遮罩：已有数据时居中 spinner（首载骨架不遮罩，刷新按钮自带转圈防连点） */}
        {refreshing && namespaces.length > 0 && (
          <div className="absolute inset-0 z-10 flex items-center justify-center bg-surface/50">
            <Icon name="loader" className="size-5 animate-spin text-faint" />
          </div>
        )}
      </div>
    </div>
  )
}

/** 单命名空间卡片：名称/超申请徽标/Pod 数 + CPU·内存双维状态着色占比条，点卡头或底部展开项目明细 */
function NamespaceCard({ ns, sortKey, open, onToggle, className, style }: {
  ns: ResourceNamespaceView
  sortKey: SortKey
  open: boolean
  onToggle: () => void
  /** RefreshFade 经 cloneElement 注入的渐入 class/延迟——须转发到根元素才生效 */
  className?: string
  style?: CSSProperties
}) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const cpuRatio = ratioPct(ns.cpuRequestMilli, ns.cpuUsageMilli)
  const memRatio = ratioPct(ns.memRequestBytes, ns.memUsageBytes)
  // 超申请：任一维占比 < 阈值即整卡标出（列出该维申请了却没用的量）
  const overCpu = cpuRatio !== null && cpuRatio < OVER_REQUEST_THRESHOLD
  const overMem = memRatio !== null && memRatio < OVER_REQUEST_THRESHOLD
  const overRequest = overCpu || overMem
  // 浪费量：申请值 - 实际用量（仅展示 >0 的一维，避免「超 0」噪声）
  const cpuOver = ns.cpuRequestMilli - ns.cpuUsageMilli
  const memOver = ns.memRequestBytes - ns.memUsageBytes

  return (
    <article
      className={`flex min-w-0 flex-col rounded-lg border bg-surface transition-shadow hover:shadow-sm ${
        overRequest ? 'border-warn/50' : 'border-line'
      } ${className ?? ''}`}
      style={style}
    >
      {/* 卡头：名称 + Pod 数（点按展开项目明细） */}
      <button
        type="button"
        onClick={onToggle}
        aria-expanded={open}
        aria-label={t('resources.cardAria', { name: ns.name })}
        className="flex min-w-0 items-center gap-2 px-4 pt-3 text-left"
      >
        <span className="min-w-0 flex-1 truncate font-mono text-[13px] font-semibold text-ink">{ns.name}</span>
        <span className="shrink-0 font-mono text-[11px] tabular-nums text-faint">
          {t('resources.podCount', { count: ns.podCount })}
        </span>
      </button>

      {/* 超申请徽标：软底圆点标签（Tag warn 语义），列出各维申请了却没用的量 */}
      {overRequest && (
        <div className="px-4 pt-1.5">
          <Tag tone="warn">
            {t('resources.overRequest')}
            {overCpu && (
              <span className="font-mono">
                <span className="opacity-50">·</span>
                {t('resources.cpu')} {fmtCpuMilli(cpuOver)}
              </span>
            )}
            {overMem && (
              <span className="font-mono">
                <span className="opacity-50">·</span>
                {t('resources.mem')} {fmtMem(memOver)}
              </span>
            )}
          </Tag>
        </div>
      )}

      {/* 双维资源条：填充=使用占申请比，空槽=浪费的申请量；状态着色（超申请橙/正常绿/超用红） */}
      <div className="flex-1 px-4 pt-1">
        <MetricBar
          label={t('resources.cpu')}
          usage={fmtCpuMilli(ns.cpuUsageMilli)}
          request={fmtCpuMilli(ns.cpuRequestMilli)}
          ratio={cpuRatio}
          active={sortKey === 'cpu'}
        />
        <MetricBar
          label={t('resources.mem')}
          usage={fmtMem(ns.memUsageBytes)}
          request={fmtMem(ns.memRequestBytes)}
          ratio={memRatio}
          active={sortKey === 'mem'}
        />
      </div>

      {/* 底部操作条：展开项目明细 + 跳转命名空间管理（独立按钮，避免在 toggle 按钮内嵌按钮） */}
      <div className="flex items-center justify-between border-t border-line px-4 py-2">
        <button
          type="button"
          onClick={onToggle}
          className="flex items-center gap-1 text-[11px] text-faint transition-colors hover:text-primary"
        >
          <Icon name="chevron-right" className={`size-3 transition-transform ${open ? 'rotate-90' : ''}`} />
          {t('resources.projectsTitle')}
          <span className="font-mono">{ns.projects.length}</span>
        </button>
        <button
          type="button"
          onClick={() => navigate(`/admin/namespaces?ns=${encodeURIComponent(ns.name)}`)}
          title={t('resources.gotoManagement')}
          aria-label={t('resources.gotoManagement')}
          className="text-faint transition-colors hover:text-primary"
        >
          <Icon name="external" className="size-3.5" />
        </button>
      </div>

      {/* 项目明细：仅展开且有项目时渲染；垂直堆叠紧凑项目卡 */}
      {open && ns.projects.length > 0 && (
        <div className="flex flex-col gap-2 border-t border-line bg-bg/50 px-4 py-3">
          {ns.projects.map((p) => (
            <ProjectRow key={p.name} p={p} />
          ))}
        </div>
      )}
    </article>
  )
}

/** 单维资源占比条（卡片主视图）：维度名 + 占比%主数字 + 状态着色进度条 + 「使用/申请」对照行。
 * 占比%与填充色同源（TONE_TEXT/TONE_FILL），一眼看懂该维健康度；空槽即浪费的申请量。 */
function MetricBar({ label, usage, request, ratio, active }: {
  label: string
  usage: string
  request: string
  ratio: number | null
  active: boolean
}) {
  const { t } = useTranslation()
  const tone = ratioTone(ratio)
  const pct = ratio === null ? 0 : Math.min(ratio, 100)
  return (
    <div className="mt-3">
      <div className="flex items-center justify-between gap-2">
        <span className={`text-[11px] ${active ? 'font-medium text-mute' : 'text-faint'}`}>{label}</span>
        <span className={`font-mono text-[12px] font-semibold tabular-nums ${TONE_TEXT[tone]}`}>
          {ratio === null ? '—' : `${Math.round(ratio)}%`}
        </span>
      </div>
      <div className="mt-1 h-2 overflow-hidden rounded-full bg-raised">
        <div className={`h-full rounded-full transition-all ${TONE_FILL[tone]}`} style={{ width: `${pct}%` }} />
      </div>
      {/* 使用在前、申请在后：一眼先看「用了多少」，再看「申请了多少」——差距即浪费 */}
      <div className="mt-1 flex items-center justify-between font-mono text-[10px] tabular-nums text-faint">
        <span>{t('resources.used', { value: usage })}</span>
        <span>{t('resources.requested', { value: request })}</span>
      </div>
    </div>
  )
}

/** 单维紧凑占比条（项目明细卡内）：占比% + 细进度条 + 使用/申请一行，双维并排（grid-cols-2） */
function MiniBar({ label, usage, request, ratio }: {
  label: string
  usage: string
  request: string
  ratio: number | null
}) {
  const tone = ratioTone(ratio)
  const pct = ratio === null ? 0 : Math.min(ratio, 100)
  return (
    <div className="min-w-0">
      <div className="flex items-center justify-between gap-2">
        <span className="text-[10px] text-faint">{label}</span>
        <span className={`font-mono text-[11px] font-semibold tabular-nums ${TONE_TEXT[tone]}`}>
          {ratio === null ? '—' : `${Math.round(ratio)}%`}
        </span>
      </div>
      <div className="mt-1 h-1.5 overflow-hidden rounded-full bg-raised">
        <div className={`h-full rounded-full ${TONE_FILL[tone]}`} style={{ width: `${pct}%` }} />
      </div>
      <div className="mt-1 truncate font-mono text-[10px] tabular-nums text-faint">
        {usage} / {request}
      </div>
    </div>
  )
}

/** 单项目明细卡：名称 + Pod 数 + CPU·内存双维紧凑占比条 + 工作负载明细（按属主链拆到
 * Deployment/StatefulSet/DaemonSet，回答「项目下哪个 deploy 申请了多少/占比多少」）。 */
function ProjectRow({ p }: { p: ResourceProjectView }) {
  const { t } = useTranslation()
  return (
    <div className="rounded-md border border-line bg-surface px-3 py-2">
      <div className="flex min-w-0 items-center justify-between gap-2">
        <span className="min-w-0 flex-1 truncate font-mono text-[12px] font-medium text-ink">{p.name}</span>
        <span className="shrink-0 font-mono text-[10px] tabular-nums text-faint">
          {t('resources.podCount', { count: p.podCount })}
        </span>
      </div>
      <div className="mt-1.5 grid grid-cols-2 gap-3">
        <MiniBar
          label={t('resources.cpu')}
          usage={fmtCpuMilli(p.cpuUsageMilli)}
          request={fmtCpuMilli(p.cpuRequestMilli)}
          ratio={ratioPct(p.cpuRequestMilli, p.cpuUsageMilli)}
        />
        <MiniBar
          label={t('resources.mem')}
          usage={fmtMem(p.memUsageBytes)}
          request={fmtMem(p.memRequestBytes)}
          ratio={ratioPct(p.memRequestBytes, p.memUsageBytes)}
        />
      </div>
      {/* 工作负载明细：仅项目有 workload 拆分时渲染（裸 pod 计入项目总量但不单列） */}
      {p.workloads.length > 0 && (
        <div className="mt-1.5 border-t border-line pt-1.5">
          <div className="mb-1 text-[10px] font-medium text-mute">{t('resources.workloadsTitle')}</div>
          <div className="flex flex-col gap-0.5">
            {p.workloads.map((w) => (
              <div key={`${w.kind}/${w.name}`} className="flex min-w-0 items-center justify-between gap-2 font-mono text-[10px]">
                <span className="min-w-0 truncate text-mute">
                  {w.kind} · {w.name}
                </span>
                <span className="shrink-0 tabular-nums text-faint">
                  {t('resources.podCount', { count: w.podCount })}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
