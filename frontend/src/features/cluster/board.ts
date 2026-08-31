/**
 * 集群资源看板 · 视图契约与纯格式化工具
 *
 * 仅保留视图层依赖的类型、格式化与聚合纯函数（不含任何假数据/抖动模拟）：
 * - 数据形态对齐后端 cluster.BoardResponse 聚合契约（节点/命名空间/Pod 三视图 + 集群总览）
 * - 集群总览由节点聚合派生（buildOverview 是单一事实来源：总览/趋势都从这里派生）
 * - 真实数据由 useResourceBoard 轮询 /api/admin/cluster/board 提供
 */

/** 集群健康状态（对齐 ClusterStatus 的三态文案/色系） */
export type ClusterStatus = 'health' | 'not good' | 'bad'

/** 节点维度资源：CPU 用毫核（m），内存用字节（B） */
export interface NodeMetric {
  name: string
  /** master / worker */
  role: 'master' | 'worker'
  /** 调度状态：Ready / NotReady / SchedulingDisabled */
  status: 'Ready' | 'NotReady' | 'SchedulingDisabled'
  cpuCapacity: number
  cpuUsage: number
  cpuRequest: number
  memCapacity: number
  memUsage: number
  memRequest: number
}

/** 命名空间维度聚合：CPU/内存用量 + Pod 数 */
export interface NamespaceMetric {
  namespace: string
  cpuMilli: number
  memoryBytes: number
  podCount: number
}

/** Pod 维度采样（服务端取 Top N，按 CPU 降序排好） */
export interface PodMetric {
  namespace: string
  pod: string
  cpuMilli: number
  memoryBytes: number
}

/** 集群总览（由节点聚合派生，与后端 ClusterInfo 字段语义对齐） */
export interface ClusterOverview {
  status: ClusterStatus
  totalCpu: string
  usedCpu: string
  freeCpu: string
  /** 百分比数值 0-100，供趋势图与进度条使用 */
  usageCpuRate: number
  requestCpuRate: number
  totalMemory: string
  usedMemory: string
  freeMemory: string
  usageMemRate: number
  requestMemRate: number
  nodeTotal: number
  nodeReady: number
}

/** CPU 毫核 → "N core"（节点/集群维度展示） */
export const fmtCpuCore = (milli: number): string => `${(milli / 1000).toFixed(2)} core`

/** CPU 毫核 → "N m"（Pod/命名空间维度展示，对齐 PodMetrics 的毫核约定） */
export const fmtCpuMilli = (milli: number): string => `${milli} m`

/** 字节 → 1024 进制人类可读（"GiB"） */
export const fmtMem = (bytes: number): string => {
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB']
  let v = bytes
  let u = units[0]
  for (let i = 1; v >= 1024 && i < units.length; i += 1) {
    v /= 1024
    u = units[i]
  }
  return `${v >= 100 ? Math.round(v) : v.toFixed(1)} ${u}`
}

/** 使用率百分比（用量/容量；cap<=0 时返回 0，避免除零/Infinity）——导出供 NodeTable 复用，口径统一 */
export const usageRate = (used: number, cap: number): number => (cap > 0 ? (used / cap) * 100 : 0)

/**
 * 由节点集合聚合出集群总览（单一事实来源：总览/趋势都从这里派生）。
 * 健康状态由调用方注入后端 status（单一事实来源，不再前端按阈值重算——
 * 旧 deriveStatus 阈值与后端 getStatus 不一致，导致看板「异常」/顶栏「亚健康」矛盾）。
 */
export const buildOverview = (nodes: NodeMetric[], status: ClusterStatus): ClusterOverview => {
  const total = nodes.reduce(
    (acc, n) => {
      acc.cpuCap += n.cpuCapacity
      acc.cpuUse += n.cpuUsage
      acc.cpuReq += n.cpuRequest
      acc.memCap += n.memCapacity
      acc.memUse += n.memUsage
      acc.memReq += n.memRequest
      return acc
    },
    { cpuCap: 0, cpuUse: 0, cpuReq: 0, memCap: 0, memUse: 0, memReq: 0 },
  )
  const usageCpuRate = usageRate(total.cpuUse, total.cpuCap)
  const usageMemRate = usageRate(total.memUse, total.memCap)
  const requestCpuRate = usageRate(total.cpuReq, total.cpuCap)
  const requestMemRate = usageRate(total.memReq, total.memCap)
  return {
    status,
    totalCpu: fmtCpuCore(total.cpuCap),
    usedCpu: fmtCpuCore(total.cpuUse),
    freeCpu: fmtCpuCore(total.cpuCap - total.cpuUse),
    usageCpuRate,
    requestCpuRate,
    totalMemory: fmtMem(total.memCap),
    usedMemory: fmtMem(total.memUse),
    freeMemory: fmtMem(total.memCap - total.memUse),
    usageMemRate,
    requestMemRate,
    nodeTotal: nodes.length,
    nodeReady: nodes.filter((n) => n.status === 'Ready').length,
  }
}

/** 命名空间聚合 → Pod 总数 */
export const buildPodCount = (namespaces: NamespaceMetric[]): number =>
  namespaces.reduce((acc, n) => acc + n.podCount, 0)
