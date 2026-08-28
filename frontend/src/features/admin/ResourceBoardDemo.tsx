import { Badge } from '@/components/ui/shadcn/badge'
import { Icon, type IconName } from '@/components/Icons'
import { cn } from '@/lib/utils'
import type { components } from '@/api/schema'

type ResourceNamespace = components['schemas']['cluster.ResourceNamespace']
type ResourceProject = components['schemas']['cluster.ResourceProject']
type ResourceWorkload = components['schemas']['cluster.ResourceProjectWorkload']

/** 工作负载类型 → 展示名 + 语义色（token 换肤） */
const KIND_TONE: Record<string, { label: string; badge: string }> = {
  Deployment: { label: 'Deployment', badge: 'bg-primary-soft text-primary' },
  StatefulSet: { label: 'StatefulSet', badge: 'bg-info-soft text-info' },
  DaemonSet: { label: 'DaemonSet', badge: 'bg-ok-soft text-ok' },
}
const KIND_DEFAULT = { label: 'Workload', badge: 'bg-mute-soft text-mute' }

/** CPU 毫核 → 人类可读：≥1 core 转核，否则保留 m。 */
function formatCpu(milli: number): string {
  if (milli >= 1000) return `${(milli / 1000).toFixed(2)} C`
  return `${Math.round(milli)}m`
}

/** 字节 → 人类可读（KiB/MiB/GiB/TiB，1024 进制）。 */
function formatMem(bytes: number): string {
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB']
  let v = bytes
  let i = 0
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i += 1
  }
  return `${v >= 100 ? Math.round(v) : v.toFixed(1)} ${units[i]}`
}

/** 用量 / 申请 占比（%）；申请为 0 时按 0 处理（无申请即无占比）。 */
function ratio(usage: number, request: number): number {
  if (request <= 0) return 0
  return (usage / request) * 100
}

/** 占比语义：>100% 超申请（红）/ <30% 低效申请（琥珀）/ 其余正常（绿）。 */
function ratioTone(pct: number): string {
  if (pct > 100) return 'text-err'
  if (pct < 30) return 'text-warn'
  return 'text-ok'
}
function ratioBar(pct: number): string {
  if (pct > 100) return 'bg-err'
  if (pct < 30) return 'bg-warn'
  return 'bg-ok'
}

/** 单条资源占比行：图标 + 申请值 + 用量值 + 填充条 + 百分比。 */
function UsageRow({ icon, label, usage, request }: { icon: IconName; label: string; usage: number; request: number }) {
  const pct = ratio(usage, request)
  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center justify-between text-[11px]">
        <span className="flex items-center gap-1 text-mute">
          <Icon name={icon} className="text-[12px]" />
          {label}
        </span>
        <span className="font-medium text-ink">
          用量 {formatCpu(usage)} / 申请 {formatCpu(request)}
        </span>
        <span className={cn('font-semibold tabular-nums', ratioTone(pct))}>{pct.toFixed(0)}%</span>
      </div>
      <div className="h-1.5 w-full overflow-hidden rounded-full bg-primary/15">
        <div className={cn('h-full rounded-full transition-all', ratioBar(pct))} style={{ width: `${Math.min(pct, 100)}%` }} />
      </div>
    </div>
  )
}

/** 工作负载卡片：kind 徽标 + 名称 + pod 数 + CPU/内存占比。 */
function WorkloadCard({ wl }: { wl: ResourceWorkload }) {
  const tone = KIND_TONE[wl.kind] ?? KIND_DEFAULT
  return (
    <div className="flex flex-col gap-2.5 rounded-lg border border-line bg-surface p-3">
      <div className="flex items-center justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2">
          <Badge className={cn('px-1.5 py-0.5 text-[10px] font-semibold', tone.badge)}>{tone.label}</Badge>
          <span className="truncate text-[12.5px] font-semibold text-ink">{wl.name}</span>
        </div>
        <span className="shrink-0 text-[11px] text-mute">{wl.podCount} pods</span>
      </div>
      <UsageRow icon="cpu" label="CPU" usage={Number(wl.cpuUsageMilli)} request={Number(wl.cpuRequestMilli)} />
      <UsageRow icon="memory" label="内存" usage={Number(wl.memUsageBytes)} request={Number(wl.memRequestBytes)} />
    </div>
  )
}

/** 项目卡片：项目名 + 资源占比 + 工作负载卡片网格。 */
function ProjectCard({ proj }: { proj: ResourceProject }) {
  const cpuPct = ratio(Number(proj.cpuUsageMilli), Number(proj.cpuRequestMilli))
  const memPct = ratio(Number(proj.memUsageBytes), Number(proj.memRequestBytes))
  return (
    <div className="flex flex-col gap-3 rounded-xl border border-line bg-raised p-4">
      <div className="flex items-center justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2">
          <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary-soft text-primary">
            <Icon name="project" className="text-[14px]" />
          </div>
          <span className="truncate text-[14px] font-bold text-ink">{proj.name}</span>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <span className="text-[11px] text-mute">{proj.podCount} pods</span>
          <span className={cn('rounded-full px-2 py-0.5 text-[11px] font-semibold tabular-nums', ratioTone(cpuPct))}>
            CPU {cpuPct.toFixed(0)}%
          </span>
          <span className={cn('rounded-full px-2 py-0.5 text-[11px] font-semibold tabular-nums', ratioTone(memPct))}>
            内存 {memPct.toFixed(0)}%
          </span>
        </div>
      </div>
      <UsageRow icon="cpu" label="CPU" usage={Number(proj.cpuUsageMilli)} request={Number(proj.cpuRequestMilli)} />
      <UsageRow icon="memory" label="内存" usage={Number(proj.memUsageBytes)} request={Number(proj.memRequestBytes)} />
      {proj.workloads.length > 0 && (
        <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2 2xl:grid-cols-3">
          {proj.workloads.map((wl) => (
            <WorkloadCard key={`${wl.kind}/${wl.name}`} wl={wl} />
          ))}
        </div>
      )}
    </div>
  )
}

/** 命名空间卡片：空间名 + 总占比 + 项目网格。 */
function NamespaceCard({ ns }: { ns: ResourceNamespace }) {
  const cpuPct = ratio(Number(ns.cpuUsageMilli), Number(ns.cpuRequestMilli))
  const memPct = ratio(Number(ns.memUsageBytes), Number(ns.memRequestBytes))
  return (
    <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-5">
      <div className="flex items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-2.5">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary">
            <Icon name="namespace" className="text-[16px]" />
          </div>
          <div className="min-w-0">
            <div className="truncate text-[15px] font-bold text-ink">{ns.name}</div>
            <div className="text-[11px] text-mute">
              {ns.projects.length} 个项目 · {ns.podCount} pods · CPU {formatCpu(Number(ns.cpuUsageMilli))} /{' '}
              {formatCpu(Number(ns.cpuRequestMilli))} · 内存 {formatMem(Number(ns.memUsageBytes))} /{' '}
              {formatMem(Number(ns.memRequestBytes))}
            </div>
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-2 text-[12px]">
          <span className={cn('rounded-full px-2.5 py-1 font-semibold tabular-nums', ratioTone(cpuPct))}>CPU {cpuPct.toFixed(0)}%</span>
          <span className={cn('rounded-full px-2.5 py-1 font-semibold tabular-nums', ratioTone(memPct))}>内存 {memPct.toFixed(0)}%</span>
        </div>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        {ns.projects.map((proj) => (
          <ProjectCard key={proj.name} proj={proj} />
        ))}
      </div>
    </section>
  )
}

/** demo 数据：模拟「一个空间多个项目、项目内多个 sts/deploy/daemon」的资源快照。 */
const demoData: components['schemas']['cluster.ResourceBoardResponse'] = {
  namespaces: [
    {
      name: 'devops-frontend',
      podCount: 14,
      cpuRequestMilli: '2900',
      cpuUsageMilli: '1220',
      memRequestBytes: '6442450944', // 6 GiB
      memUsageBytes: '3221225472', // 3 GiB
      projects: [
        {
          name: 'web-portal',
          podCount: 4,
          cpuRequestMilli: '1000',
          cpuUsageMilli: '430',
          memRequestBytes: '2147483648',
          memUsageBytes: '1073741824',
          workloads: [
            { kind: 'Deployment', name: 'web-portal', podCount: 3, cpuRequestMilli: '750', cpuUsageMilli: '380', memRequestBytes: '1610612736', memUsageBytes: '805306368' },
            { kind: 'DaemonSet', name: 'web-portal-cache', podCount: 1, cpuRequestMilli: '250', cpuUsageMilli: '50', memRequestBytes: '536870912', memUsageBytes: '268435456' },
          ],
        },
        {
          name: 'gateway',
          podCount: 2,
          cpuRequestMilli: '600',
          cpuUsageMilli: '540',
          memRequestBytes: '1073741824',
          memUsageBytes: '805306368',
          workloads: [
            { kind: 'Deployment', name: 'gateway', podCount: 2, cpuRequestMilli: '600', cpuUsageMilli: '540', memRequestBytes: '1073741824', memUsageBytes: '805306368' },
          ],
        },
        {
          name: 'user-service',
          podCount: 6,
          cpuRequestMilli: '900',
          cpuUsageMilli: '70',
          memRequestBytes: '1610612736',
          memUsageBytes: '201326592',
          workloads: [
            { kind: 'Deployment', name: 'user-service', podCount: 3, cpuRequestMilli: '600', cpuUsageMilli: '55', memRequestBytes: '1073741824', memUsageBytes: '134217728' },
            { kind: 'StatefulSet', name: 'user-db', podCount: 2, cpuRequestMilli: '200', cpuUsageMilli: '10', memRequestBytes: '536870912', memUsageBytes: '67108864' },
            { kind: 'DaemonSet', name: 'user-agent', podCount: 1, cpuRequestMilli: '100', cpuUsageMilli: '5', memRequestBytes: '268435456', memUsageBytes: '0' },
          ],
        },
        {
          name: 'order-service',
          podCount: 3,
          cpuRequestMilli: '400',
          cpuUsageMilli: '180',
          memRequestBytes: '1610612736',
          memUsageBytes: '1127428915',
          workloads: [
            { kind: 'Deployment', name: 'order-service', podCount: 2, cpuRequestMilli: '400', cpuUsageMilli: '180', memRequestBytes: '1610612736', memUsageBytes: '1127428915' },
          ],
        },
      ],
    },
    {
      name: 'devops-backend',
      podCount: 8,
      cpuRequestMilli: '2400',
      cpuUsageMilli: '1820',
      memRequestBytes: '4294967296', // 4 GiB
      memUsageBytes: '3758096384', // 3.5 GiB
      projects: [
        {
          name: 'billing-core',
          podCount: 3,
          cpuRequestMilli: '1200',
          cpuUsageMilli: '1500', // 超申请（>100%）示例
          memRequestBytes: '2147483648',
          memUsageBytes: '2684354560',
          workloads: [
            { kind: 'Deployment', name: 'billing-core', podCount: 2, cpuRequestMilli: '1200', cpuUsageMilli: '1500', memRequestBytes: '2147483648', memUsageBytes: '2684354560' },
          ],
        },
        {
          name: 'data-pipeline',
          podCount: 5,
          cpuRequestMilli: '1200',
          cpuUsageMilli: '320',
          memRequestBytes: '2147483648',
          memUsageBytes: '1073741824',
          workloads: [
            { kind: 'Deployment', name: 'data-worker', podCount: 2, cpuRequestMilli: '800', cpuUsageMilli: '240', memRequestBytes: '1610612736', memUsageBytes: '805306368' },
            { kind: 'StatefulSet', name: 'data-queue', podCount: 2, cpuRequestMilli: '300', cpuUsageMilli: '60', memRequestBytes: '536870912', memUsageBytes: '268435456' },
            { kind: 'DaemonSet', name: 'data-collector', podCount: 1, cpuRequestMilli: '100', cpuUsageMilli: '20', memRequestBytes: '268435456', memUsageBytes: '0' },
          ],
        },
      ],
    },
  ],
}

/** 空间资源（管理员后台）demo：展示「空间 → 项目 → 工作负载（sts/deploy/daemon）」三级卡片。 */
export function ResourceBoardDemo() {
  const legend = Object.entries(KIND_TONE)
  return (
    <div className="mx-auto flex max-w-7xl flex-col gap-5 p-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-lg font-bold text-ink">空间资源</h1>
        <p className="text-[13px] text-mute">
          每个命名空间的资源申请 / 实际用量占比，按项目与工作负载（Deployment / StatefulSet / DaemonSet）拆分，定位「申请了很多却用不到多少」的空间。
        </p>
      </div>
      <div className="flex items-center gap-4">
        <span className="text-[12px] text-mute">工作负载类型：</span>
        {legend.map(([kind, { label, badge }]) => (
          <Badge key={kind} className={cn('gap-1.5 px-2 py-0.5 text-[11px] font-medium', badge)}>
            {label}
          </Badge>
        ))}
        <div className="ml-auto flex items-center gap-3 text-[11px] text-mute">
          <span className="flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-err" /> 超申请
          </span>
          <span className="flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-warn" /> 低效（&lt;30%）
          </span>
          <span className="flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-ok" /> 正常
          </span>
        </div>
      </div>
      <div className="flex flex-col gap-5">
        {demoData.namespaces.map((ns) => (
          <NamespaceCard key={ns.name} ns={ns} />
        ))}
      </div>
    </div>
  )
}
