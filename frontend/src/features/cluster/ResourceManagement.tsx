import { useCallback, useEffect, useState, type CSSProperties } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { Icon } from '@/components/Icons'
import { RefreshFade, Tag } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { SkeletonList } from '@/components/ui/SkeletonList'
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
/** 占比计算导出供治理页复用：request 与 usage 都 > 0 才构成占比（其余为 null → 「—」）。 */
export const ratioPct = (request: number, usage: number): number | null =>
  request > 0 && usage > 0 ? (usage / request) * 100 : null

/** 超申请阈值：实际用量不足申请的 30% 即标「超申请」——定位低占比高申请空间的入口 */
const OVER_REQUEST_THRESHOLD = 30

/** 排序维度 → 占比取值函数；null 在升序中恒排末尾（无申请量不可能超申请） */
const DIM_RATIO = {
  cpu: (ns: ResourceNamespaceView) => ratioPct(ns.cpuRequestMilli, ns.cpuUsageMilli),
  mem: (ns: ResourceNamespaceView) => ratioPct(ns.memRequestBytes, ns.memUsageBytes),
} as const

type SortKey = keyof typeof DIM_RATIO

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
 * 每个管理命名空间的 Pod requests/实际用量占比列表——找出「申请了很多 requests 却用不到
 * 多少资源」的空间。默认按 CPU 占比升序（超申请者置顶），可切内存占比；行点击展开项目明细。
 * 诊断页不轮询（手动刷新即可，避免后台空转）。
 */
export function ResourceManagement() {
  const { t } = useTranslation()
  const [namespaces, setNamespaces] = useState<ResourceNamespaceView[]>([])
  const [loading, setLoading] = useState(true)
  // 刷新中：手动刷新按钮转圈 + 禁用，避免连点
  const [refreshing, setRefreshing] = useState(false)
  // 渐入版本号：每次取数成功 +1，RefreshFade 依 key 重挂载列表重播渐入
  const [version, setVersion] = useState(0)
  // 排序维度：默认 CPU 占比（超申请最严重的空间置顶）
  const [sortKey, setSortKey] = useState<SortKey>('cpu')
  // 展开行：一次只展开一个空间的项目明细
  const [expanded, setExpanded] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    setRefreshing(true)
    try {
      const { data, error } = await api.GET('/api/admin/cluster/resources')
      if (error) throw new Error(error.message ?? String(error))
      if (!data) return
      setNamespaces(data.namespaces.map(toNamespaceView))
      // 取数成功 → 版本号 +1，整表重播一次渐入（keep-last-frame，不闪断）
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

  const chipCls = (active: boolean) =>
    `rounded-md border px-1.5 py-0.5 text-[11px] transition-colors ${
      active
        ? 'border-primary bg-primary-soft font-medium text-primary'
        : 'border-line text-mute hover:border-primary hover:text-primary'
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

      {/* 排序维度切换：CPU 占比 / 内存占比 */}
      <div className="flex items-center gap-1">
        <button type="button" onClick={() => setSortKey('cpu')} className={chipCls(sortKey === 'cpu')}>
          {t('resources.sortCpu')}
        </button>
        <button type="button" onClick={() => setSortKey('mem')} className={chipCls(sortKey === 'mem')}>
          {t('resources.sortMem')}
        </button>
        <span className="ml-auto hidden text-[11px] text-faint sm:block">{t('resources.sortHint')}</span>
      </div>

      {loading ? (
        <SkeletonList count={5} />
      ) : sorted.length === 0 ? (
        <div className="rounded-lg border border-line bg-surface px-4 py-10 text-center text-[12px] text-faint">
          {t('resources.empty')}
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border border-line bg-surface">
          <RefreshFade version={version}>
            {sorted.map((ns) => (
              <NamespaceRow
                key={ns.name}
                ns={ns}
                sortKey={sortKey}
                open={expanded === ns.name}
                onToggle={() => setExpanded(expanded === ns.name ? null : ns.name)}
              />
            ))}
          </RefreshFade>
        </div>
      )}
    </div>
  )
}

/** 单命名空间行：名称/超申请警示/Pod 数 + CPU·内存双维占比条，点击展开项目明细 */
function NamespaceRow({ ns, sortKey, open, onToggle, className, style }: {
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
  // 超申请警示取当前排序维度占比：低占比且有申请量才标（无申请量的空间占比为 null）
  const dimRatio = DIM_RATIO[sortKey](ns)
  const overRequest = dimRatio !== null && dimRatio < OVER_REQUEST_THRESHOLD
  // 超申请量：申请量 - 实际用量。两维各自独立计算，仅展示 >0 的一维（避免「超 0」噪声）
  const cpuOver = ns.cpuRequestMilli - ns.cpuUsageMilli
  const memOver = ns.memRequestBytes - ns.memUsageBytes

  return (
    <div className={`border-b border-line last:border-b-0 ${className ?? ''}`} style={style}>
      <div className="flex items-stretch">
        <button
          type="button"
          onClick={onToggle}
          className="flex min-w-0 flex-1 flex-col gap-1.5 px-4 py-3 text-left transition-colors hover:bg-bg"
          aria-label={t('resources.rowAria', { name: ns.name })}
        >
          {/* 首行：名称 + 警示 tag + 展开箭头 + Pod 数 */}
          <div className="flex items-center gap-2">
            <span className="truncate font-mono text-[13px] font-medium text-ink">{ns.name}</span>
            {overRequest && (
              <span className="flex shrink-0 items-center gap-1 rounded-md border border-warn/60 bg-warn-soft px-1.5 py-0.5 text-[11px] text-warn">
                <span>{t('resources.overRequest')}</span>
                {cpuOver > 0 && (
                  <>
                    <span className="text-warn/40">·</span>
                    <span className="font-mono">{t('resources.cpu')} {fmtCpuMilli(cpuOver)}</span>
                  </>
                )}
                {memOver > 0 && (
                  <>
                    <span className="text-warn/40">·</span>
                    <span className="font-mono">{t('resources.mem')} {fmtMem(memOver)}</span>
                  </>
                )}
              </span>
            )}
            {ns.projects.length > 0 && (
              <Icon
                name="chevron-right"
                className={`size-3.5 shrink-0 text-faint transition-transform ${open ? 'rotate-90' : ''}`}
              />
            )}
            <span className="ml-auto shrink-0 font-mono text-[11px] tabular-nums text-faint">
              {t('resources.podCount', { count: ns.podCount })}
            </span>
          </div>
          {/* 双维占比条：当前排序维度高亮，另一维度弱化；占比统一 Tag（低占比橙标超申请） */}
          <ResourceBar
            label={t('resources.cpu')}
            request={fmtCpuMilli(ns.cpuRequestMilli)}
            usage={fmtCpuMilli(ns.cpuUsageMilli)}
            ratio={ratioPct(ns.cpuRequestMilli, ns.cpuUsageMilli)}
            active={sortKey === 'cpu'}
            percentTag
          />
          <ResourceBar
            label={t('resources.mem')}
            request={fmtMem(ns.memRequestBytes)}
            usage={fmtMem(ns.memUsageBytes)}
            ratio={ratioPct(ns.memRequestBytes, ns.memUsageBytes)}
            active={sortKey === 'mem'}
            percentTag
          />
        </button>
        {/* 跳到命名空间管理并预选该空间：独立入口（避免在 toggle 按钮内嵌按钮的无效 HTML） */}
        <button
          type="button"
          onClick={() => navigate(`/admin/namespaces?ns=${encodeURIComponent(ns.name)}`)}
          title={t('resources.gotoManagement')}
          aria-label={t('resources.gotoManagement')}
          className="flex shrink-0 items-center justify-center border-l border-line px-3 text-faint transition-colors hover:bg-bg hover:text-primary"
        >
          <Icon name="external" className="size-4" />
        </button>
      </div>
      {/* 项目明细：仅展开且有项目时渲染；项目卡片响应式网格（宽屏三列，窄屏逐级降列） */}
      {open && ns.projects.length > 0 && (
        <div className="border-t border-line bg-bg/50 px-4 py-3">
          <div className="mb-2 text-[12px] font-medium text-mute">{t('resources.projectsTitle')}</div>
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {ns.projects.map((p) => (
              <ProjectRow key={p.name} p={p} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

/** 单维资源占比条：标签 + 用量占比填充条 + 「当前值 / 申请值」数值 + 占比%。
 * percentTag=true 时占比渲染为圆角 Tag（低占比 warn 橙标超申请、正常 mute 灰），
 * 供治理页行内扫读；默认内联文本供项目卡片使用。导出双消费方，两页占比口径一致。 */
export function ResourceBar({ label, request, usage, ratio, active, percentTag = false }: {
  label: string
  request: string
  usage: string
  ratio: number | null
  active: boolean
  percentTag?: boolean
}) {
  const pct = ratio === null ? 0 : Math.min(ratio, 100)
  return (
    <div className="flex items-center gap-2">
      <span className={`w-7 shrink-0 text-[11px] ${active ? 'font-medium text-mute' : 'text-faint'}`}>{label}</span>
      <div className="h-1.5 min-w-0 flex-1 overflow-hidden rounded-full bg-raised">
        <div className={`h-full rounded-full ${active ? 'bg-primary/80' : 'bg-line'}`} style={{ width: `${pct}%` }} />
      </div>
      {/* 当前值(实际用量)在前、申请值在后——一眼先看「用了多少」，再看「申请了多少」 */}
      <span className="shrink-0 font-mono text-[11px] tabular-nums text-mute">
        {usage} / {request}
      </span>
      {ratio === null ? (
        <span className="w-10 shrink-0 text-right text-[11px] text-faint">—</span>
      ) : percentTag ? (
        // 占比 Tag：低占比(<30% 超申请)warn 橙标，正常 mute 灰——行内扫读「谁在超申请」
        <Tag tone={ratio < OVER_REQUEST_THRESHOLD ? 'warn' : 'mute'} dot={false} className="shrink-0">
          {Math.round(ratio)}%
        </Tag>
      ) : (
        <span
          className={`w-10 shrink-0 text-right font-mono text-[11px] tabular-nums ${
            ratio < OVER_REQUEST_THRESHOLD ? 'text-warn' : 'text-faint'
          }`}
        >
          {Math.round(ratio)}%
        </span>
      )}
    </div>
  )
}

/** 单项目明细卡片：名称 + Pod 数 + CPU·内存双维占比条 + 工作负载明细（按属主链拆到
 * Deployment/StatefulSet/DaemonSet，回答「项目下哪个 deploy 申请了多少/占比多少」）。 */
function ProjectRow({ p }: { p: ResourceProjectView }) {
  const { t } = useTranslation()
  return (
    <div className="flex min-w-0 flex-col gap-1.5 rounded-lg border border-line bg-surface px-3 py-2.5">
      {/* 卡头：项目名 + Pod 数 */}
      <div className="flex items-center gap-2">
        <span className="min-w-0 flex-1 truncate font-mono text-[12px] font-medium text-ink">{p.name}</span>
        <span className="shrink-0 font-mono text-[11px] tabular-nums text-faint">
          {t('resources.podCount', { count: p.podCount })}
        </span>
      </div>
      {/* 双维占比条：全部高亮（卡片内无弱化维度，超申请低占比自动 warn 标色） */}
      <ResourceBar
        label={t('resources.cpu')}
        request={fmtCpuMilli(p.cpuRequestMilli)}
        usage={fmtCpuMilli(p.cpuUsageMilli)}
        ratio={ratioPct(p.cpuRequestMilli, p.cpuUsageMilli)}
        active
        percentTag
      />
      <ResourceBar
        label={t('resources.mem')}
        request={fmtMem(p.memRequestBytes)}
        usage={fmtMem(p.memUsageBytes)}
        ratio={ratioPct(p.memRequestBytes, p.memUsageBytes)}
        active
        percentTag
      />
      {/* 工作负载明细：仅项目有 workload 拆分时渲染（裸 pod 计入项目总量但不单列） */}
      {p.workloads.length > 0 && (
        <div className="mt-1 border-t border-line pt-2">
          <div className="mb-1.5 text-[11px] font-medium text-mute">{t('resources.workloadsTitle')}</div>
          <div className="flex flex-col gap-1.5">
            {p.workloads.map((w) => (
              <WorkloadRow key={`${w.kind}/${w.name}`} w={w} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

/** 单工作负载明细行：kind · name + Pod 数 + CPU·内存双维紧凑占比条（复用 ResourceBar） */
function WorkloadRow({ w }: { w: ResourceWorkloadView }) {
  const { t } = useTranslation()
  return (
    <div className="flex min-w-0 flex-col gap-1 rounded-md bg-bg/60 px-2 py-1.5">
      {/* 行头：工作负载类型 · 名 + Pod 数 */}
      <div className="flex items-center gap-1.5">
        <span className="min-w-0 flex-1 truncate font-mono text-[11px] font-medium text-ink">
          {w.kind} · {w.name}
        </span>
        <span className="shrink-0 font-mono text-[10px] tabular-nums text-faint">
          {t('resources.podCount', { count: w.podCount })}
        </span>
      </div>
      <ResourceBar
        label={t('resources.cpu')}
        request={fmtCpuMilli(w.cpuRequestMilli)}
        usage={fmtCpuMilli(w.cpuUsageMilli)}
        ratio={ratioPct(w.cpuRequestMilli, w.cpuUsageMilli)}
        active
        percentTag
      />
      <ResourceBar
        label={t('resources.mem')}
        request={fmtMem(w.memRequestBytes)}
        usage={fmtMem(w.memUsageBytes)}
        ratio={ratioPct(w.memRequestBytes, w.memUsageBytes)}
        active
        percentTag
      />
    </div>
  )
}
