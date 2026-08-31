import { useCallback, useEffect, useRef, useState } from 'react'
import { api } from '@/api/client'
import {
  buildOverview,
  buildPodCount,
  type ClusterOverview,
  type ClusterStatus,
  type NamespaceMetric,
  type NodeMetric,
  type PodMetric,
} from './board'

/** 趋势环形缓冲上限（对齐后端 150s/5s 的 30 点窗口，这里 16 点） */
const TREND_POINTS = 16

/** 自动刷新间隔：轮询 /api/admin/cluster/board 获取最新快照 */
export const REFRESH_INTERVAL_MS = 3000

/**
 * 集群资源看板数据源（轮询真实后端）：
 * - 每 REFRESH_INTERVAL_MS 拉取 /api/admin/cluster/board 快照，转数值后从节点聚合出
 *   集群总览并推进 CPU/内存使用率趋势（环形缓冲 16 点）；refresh() 手动立即拉新一版
 * - 轮询失败静默保留上一帧快照（看板不闪断），下次成功自动恢复；返回值全部为最新快照
 * - 渐入语义：轮询（fetchSnapshot）静默不重播，手动刷新（refresh）成功才 +version，
 *   避免每 3s 轮询重播渐入导致的闪屏（对齐 Repos 的「重新获取数据时内容渐入」体验）
 */
export function useResourceBoard() {
  const [nodes, setNodes] = useState<NodeMetric[]>([])
  const [namespaces, setNamespaces] = useState<NamespaceMetric[]>([])
  const [pods, setPods] = useState<PodMetric[]>([])
  // Top Pod 排行维度：cpu（默认）/ mem——切换后随查询重新拉取（后端按维度取 TopN，前端不重排旧列表）
  const [topSort, setTopSort] = useState<'cpu' | 'mem'>('cpu')
  // 最新 topSort 快照（每渲染同步）：校验在途响应落地时是否仍属当前维度——
  // 维度快速连切时旧维度请求晚到会被丢弃，避免「UI 高亮 mem 但列表是 CPU TopN」的旧响应覆盖新状态
  const topSortRef = useRef(topSort)
  topSortRef.current = topSort
  const [cpuTrend, setCpuTrend] = useState<number[]>([])
  const [memTrend, setMemTrend] = useState<number[]>([])
  // 集群健康状态：单一事实来源 = 后端 overview.status（看板与顶栏同口径，不再前端按阈值重算）
  const [status, setStatus] = useState<ClusterStatus>('health')
  const [lastUpdate, setLastUpdate] = useState<Date>(() => new Date())
  // 手动刷新中：刷新按钮转圈 + 禁用，避免连点
  const [refreshing, setRefreshing] = useState(false)
  // 渐入版本号：仅手动刷新成功 +1，RefreshFade 依 key 重挂载区块重播渐入（轮询不 bump，防闪屏）
  const [version, setVersion] = useState(0)

  /** 拉取快照（轮询 + 手动刷新共用取数逻辑）：返回是否成功应用到状态，供手动刷新判定是否重播渐入 */
  const fetchSnapshot = useCallback(async (): Promise<boolean> => {
    try {
      const { data, error } = await api.GET('/api/admin/cluster/board', {
        params: { query: { topSort } },
      })
      if (error) throw new Error(error.message ?? String(error))
      if (!data) return false
      // 维度切换后在途旧响应作废：仅最新 topSort 的响应落地，旧维度 TopN 不得覆盖新维度状态
      if (topSortRef.current !== topSort) return false
      // 后端节点/命名空间/Pod 的资源字段为字符串，统一转数值喂给聚合纯函数
      const nextNodes: NodeMetric[] = data.nodes.map((n) => ({
        name: n.name,
        role: n.role as NodeMetric['role'],
        status: n.status as NodeMetric['status'],
        cpuCapacity: Number(n.cpuCapacity),
        cpuUsage: Number(n.cpuUsage),
        cpuRequest: Number(n.cpuRequest),
        memCapacity: Number(n.memCapacity),
        memUsage: Number(n.memUsage),
        memRequest: Number(n.memRequest),
      }))
      const nextNamespaces: NamespaceMetric[] = data.namespaces.map((n) => ({
        namespace: n.name,
        cpuMilli: Number(n.cpuMilli),
        memoryBytes: Number(n.memoryBytes),
        podCount: n.podCount,
      }))
      const nextPods: PodMetric[] = data.pods.map((p) => ({
        namespace: p.namespace,
        pod: p.pod,
        cpuMilli: Number(p.cpuMilli),
        memoryBytes: Number(p.memoryBytes),
      }))
      setNodes(nextNodes)
      setNamespaces(nextNamespaces)
      setPods(nextPods)
      // 后端 status 未知/缺失时回退 health，绝不本地重算阈值
      setStatus((data.overview?.status as ClusterStatus) ?? 'health')
      const overview = buildOverview(nextNodes, (data.overview?.status as ClusterStatus) ?? 'health')
      setCpuTrend((prev) => [...prev, overview.usageCpuRate].slice(-TREND_POINTS))
      setMemTrend((prev) => [...prev, overview.usageMemRate].slice(-TREND_POINTS))
      setLastUpdate(new Date())
      return true
    } catch {
      // 轮询失败静默：看板保持上一帧快照，避免闪断刷屏；下次成功自动恢复
      return false
    }
  }, [topSort])

  /** 手动刷新：拉最新快照并重播一次区块渐入；失败保留上一帧不重播（避免误导） */
  const refresh = useCallback(async () => {
    setRefreshing(true)
    try {
      const ok = await fetchSnapshot()
      if (ok) setVersion((v) => v + 1)
    } finally {
      setRefreshing(false)
    }
  }, [fetchSnapshot])

  // 首次立即拉取 + 周期自动刷新（StrictMode 双挂载由 cleanup 兜底：clearInterval 保证只跑一份）；
  // topSort 变化 → fetchSnapshot 重建 → 本 effect 重跑（立即拉新维度 + 重启轮询）
  useEffect(() => {
    void fetchSnapshot()
    const timer = window.setInterval(fetchSnapshot, REFRESH_INTERVAL_MS)
    return () => window.clearInterval(timer)
  }, [fetchSnapshot])

  const overview: ClusterOverview = buildOverview(nodes, status)
  return {
    overview,
    podCount: buildPodCount(namespaces),
    nodes,
    namespaces,
    pods,
    topSort,
    setTopSort,
    cpuTrend,
    memTrend,
    lastUpdate,
    version,
    refreshing,
    refresh,
  }
}
