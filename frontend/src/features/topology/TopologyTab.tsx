import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import { nextZIndex } from '@/lib/zIndex'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { useWebsocket } from '@/hooks/useWebsocket'
import { Icon } from '@/components/Icons'
import { SearchInput } from '@/components/SearchInput'
import { StatusDot } from '@/components/ui/StatusDot'
import { Empty, SkeletonTopology } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/shadcn/popover'
import { ContainerLogModal, type ContainerLogTarget } from '../projects/ContainerLogModal'
import { dropKinds, HIDDEN_KINDS, withDerivedAppStatus } from './mockTopology'
import { TopologyDetail } from './TopologyDetail'
import { TopologyGraph, type TopologyGraphHandle } from './TopologyGraph'
import { useLiveTopology } from './useLiveTopology'
import type {
  EdgeType,
  NodeKind,
  NodeStatus,
  PodLifecycle,
  TopoEdge,
  TopoEndpoint,
  TopoGraph,
  TopoNode,
} from './topologyTypes'

type ProjectModel = components['schemas']['types.ProjectModel']
type ResourceTreeNode = components['schemas']['project.ResourceTreeNode']
type ResourceTreeEdge = components['schemas']['project.ResourceTreeEdge']
type ResourceTreeResponse = components['schemas']['project.ResourceTreeResponse']
type StateContainer = components['schemas']['types.StateContainer']

/** pod 事件 → 重拉资源树的去抖间隔（与 TabLog/TabShell 的 POD_DEBOUNCE_MS=500 对齐，避免频繁重拉） */
const POD_DEBOUNCE_MS = 500
/** 兜底轮询间隔：pod 事件丢失（断流/订阅漏）时的安全网 */
const POLL_MS = 30000

/** 后端节点状态 → 前端 NodeStatus（后端已是同一套字符串，直映，未知兜底 unknown） */
const STATUS: Record<string, NodeStatus> = {
  healthy: 'healthy',
  degraded: 'degraded',
  progressing: 'progressing',
  unknown: 'unknown',
}

/** 后端 kind → 前端 NodeKind（对齐 demo 渲染集：Application/Deployment/ReplicaSet/Pod/
 *  Service/Ingress/ConfigMap/Secret/HPA/StatefulSet/DaemonSet；真正未知的 kind 才降级
 *  Pod，防 KIND_ICON 查表崩溃——如 sts/ds 漏配会被误标成 Pod） */
const KIND: Record<string, NodeKind> = {
  Application: 'Application',
  Deployment: 'Deployment',
  ReplicaSet: 'ReplicaSet',
  Pod: 'Pod',
  Service: 'Service',
  Ingress: 'Ingress',
  ConfigMap: 'ConfigMap',
  Secret: 'Secret',
  HPA: 'HPA',
  StatefulSet: 'StatefulSet',
  DaemonSet: 'DaemonSet',
}

/** 后端边 type → 前端 EdgeType（后端 proto 明确 owner|selector|route 三种；
 *  真正未知的 type 才降级 owner，防 EDGE_STYLE 查表崩溃——route 漏配会被误标成
 *  owner 实线，且失去 route 的线型/层级语义） */
const EDGE_TYPE: Record<string, EdgeType> = {
  owner: 'owner',
  selector: 'selector',
  route: 'route',
}

/** StateContainer → Pod 生命周期，优先级：isOld > terminating > pending > 任一容器未就绪 > 全就绪。
 *  isOld/terminating/pending 是同 pod 各容器共有的（pod 级修订/删除时间/相位），ready 才逐容器，
 *  取最重即对（如 rolling 中旧 pod 的 isOld 不会被同 pod 容器 ready 覆盖）。 */
const LIFECYCLE_RANK: Record<PodLifecycle, number> = {
  ready: 0,
  notReady: 1,
  pending: 2,
  terminating: 3,
  isOld: 4,
}
function buildPodStates(items: StateContainer[]): Record<string, PodLifecycle> {
  const map: Record<string, PodLifecycle> = {}
  for (const c of items) {
    const s: PodLifecycle = c.isOld
      ? 'isOld'
      : c.terminating
        ? 'terminating'
        : c.pending
          ? 'pending'
          : c.ready
            ? 'ready'
            : 'notReady'
    if (!map[c.pod] || LIFECYCLE_RANK[s] > LIFECYCLE_RANK[map[c.pod]]) map[c.pod] = s
  }
  return map
}

/** proto ResourceTree → 前端 TopoGraph：old 标记进 labels，events 置空（v1 不拉事件），
 *  容器接口聚合的 Pod 生命周期附到 Pod 节点（无记录留空 → 状态点走健康色兜底） */
function toTopoGraph(tree: ResourceTreeResponse, podStates: Record<string, PodLifecycle>): TopoGraph {
  const nodes: TopoNode[] = (tree.nodes ?? []).map((n: ResourceTreeNode) => ({
    id: n.id,
    kind: KIND[n.kind] ?? 'Pod',
    name: n.name,
    namespace: n.namespace,
    status: STATUS[n.status] ?? 'unknown',
    // 旧版本副本（滚动升级中的旧 RS/Pod）：打 old 标签供详情面板/后续样式弱化识别
    labels: n.old ? { ...(n.labels ?? {}), old: 'true' } : (n.labels ?? {}),
    events: [],
    // 仅后端确认是 Pod 的节点才挂生命周期（未知 kind 降级 Pod 的除外）；容器接口过滤 Failed 死透 pod
    lifecycle: n.kind === 'Pod' && podStates[n.name] ? podStates[n.name] : undefined,
  }))
  const edges: TopoEdge[] = (tree.edges ?? []).map((e: ResourceTreeEdge) => ({
    id: e.id,
    type: EDGE_TYPE[e.type] ?? 'owner',
    source: e.source,
    target: e.target,
  }))
  // 隐藏 ReplicaSet 中间层：Deployment 直挂 Pod；根状态按 pod 聚合（状态点配色），与 demo 简化展示一致
  return withDerivedAppStatus(dropKinds({ nodes, edges }, HIDDEN_KINDS))
}

/**
 * 拓扑 Tab（项目详情弹窗第 5 Tab）：渲染项目真实资源树（后端 resource_tree 接口），
 * pod 事件驱动下实时刷新。
 * - 数据：挂载拉一次 + pod 事件 debounce 重拉 + 30s 兜底轮询
 * - 布局：useLiveTopology 保持刷新间节点不跳位（存活钉住、新节点追加列底）
 * - 交互：点击选中看详情、拖拽、适应视图、手动刷新
 */
export function TopologyTab({
  project,
  resizeAt,
}: {
  project: ProjectModel
  /** 宿主弹窗尺寸变化信号（缩放/最大化/还原时 bump，见 useDraggableDialog onResize）→ 自动适应视图 */
  resizeAt?: number
}) {
  const { t } = useTranslation()
  const { subscribeProjectPodEvent } = useWebsocket()
  const live = useLiveTopology()
  const graphRef = useRef<TopologyGraphHandle>(null)
  // 根容器 ref：hidden 态（display:none）下判可见性，避免 0 尺寸容器触发 fitView 算错 view
  const rootRef = useRef<HTMLDivElement>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [hoveredId, setHoveredId] = useState<string | null>(null)
  const [query, setQuery] = useState('')
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)
  // Pod 操作：查看日志（ContainerLogModal）/ 强杀（force_delete，Popover 确认）
  const [logTarget, setLogTarget] = useState<ContainerLogTarget | null>(null)
  const [killOpen, setKillOpen] = useState(false)
  const [confirmZ, setConfirmZ] = useState(0)
  const [killing, setKilling] = useState(false)
  const [podContainers, setPodContainers] = useState<StateContainer[]>([])
  // 项目访问地址（/api/endpoints 全量，Application 根节点详情卡片平铺展示；随每次树刷新一起拉）
  const [endpoints, setEndpoints] = useState<TopoEndpoint[]>([])
  // 手动刷新中：刷新按钮转圈禁用（兜底轮询/pod 事件不置位，不打扰用户）
  const [refreshing, setRefreshing] = useState(false)

  /** 拉取资源树 + 容器列表并提交给布局（幂等：节点集不变则位置不变，只重渲状态色/lifecycle）。
   *  容器接口失败不影响树刷新（该次无生命周期点，状态点回退健康色）。 */
  const fetchTree = useCallback(async () => {
    try {
      const [treeRes, containerRes, endpointRes] = await Promise.all([
        api.GET(API.projectsResourceTree, { params: { path: { id: project.id } } }),
        api.GET(API.projectsContainers, { params: { path: { id: project.id } } }).catch(() => null),
        // 项目访问地址（Application 根节点详情卡片全部展示）：失败不回退树刷新，仅当次地址列表为空
        api
          .GET(API.endpointsProject, { params: { path: { projectId: project.id } } })
          .catch(() => null),
      ])
      if (treeRes.error) throw new Error(treeRes.error.message ?? String(treeRes.error))
      if (!treeRes.data) return
      const items = containerRes?.data?.items ?? []
      setPodContainers(items)
      setEndpoints(endpointRes?.data?.items ?? [])
      live.commitGraph(toTopoGraph(treeRes.data, buildPodStates(items)))
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    }
  }, [project.id, live.commitGraph])

  // 挂载即拉一次（部署成功后首次切入 Tab，全量落地终态）
  useEffect(() => {
    void fetchTree()
  }, [fetchTree])

  // pod 事件 → debounce 重拉（新 pod 出现/旧 pod 消失/状态翻转都走这条信号）+ 兜底轮询
  useEffect(() => {
    const unsub = subscribeProjectPodEvent(project.id, () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
      debounceRef.current = setTimeout(() => {
        debounceRef.current = null
        void fetchTree()
      }, POD_DEBOUNCE_MS)
    })
    pollRef.current = setInterval(() => void fetchTree(), POLL_MS)
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
      if (pollRef.current) clearInterval(pollRef.current)
      unsub()
    }
  }, [subscribeProjectPodEvent, project.id, fetchTree])

  const graph = live.graph

  // 搜索命中集：按名称或类型前缀匹配；空查询 → null（关闭高亮，对齐 demo 页逻辑）
  const queryActive = query.trim().length > 0
  const matchedIds = useMemo(() => {
    if (!queryActive) return null
    const q = query.trim().toLowerCase()
    const set = new Set<string>()
    for (const node of graph?.nodes ?? []) {
      if (node.name.toLowerCase().includes(q) || node.kind.toLowerCase().includes(q)) set.add(node.id)
    }
    return set
  }, [query, queryActive, graph])

  // 选中节点在刷新中被移除（pod 重建/缩容）→ 清空选中，避免详情面板指向幽灵节点
  useEffect(() => {
    if (selectedId && graph && !graph.nodes.some((n) => n.id === selectedId)) {
      setSelectedId(null)
    }
  }, [graph, selectedId])

  const selectedNode = graph?.nodes.find((n) => n.id === selectedId) ?? null

  // 应用状态从「进行中」翻转「健康」（部署完成落地）时，重跑一次确定性分层布局并重新适配：
  // 部署期间新 pod 逐次 extendLayout 追加列底，完成时一次重排让终态树归位对齐（含丢弃拖拽位）。
  // 只在此翻转边界触发一次（ref 记上次状态），其余刷新走 extendLayout 钉住不动。
  const prevAppStatusRef = useRef<NodeStatus | null>(null)
  useEffect(() => {
    const appNode = graph?.nodes.find((n) => n.id === live.rootId)
    const cur = appNode?.status ?? null
    const prev = prevAppStatusRef.current
    prevAppStatusRef.current = cur
    if (prev === 'progressing' && cur === 'healthy') {
      live.relayout(graph)
      graphRef.current?.fitView()
    }
  }, [graph, live.rootId])

  // 容器列表（查看日志按钮 / 生命周期点用）已由 fetchTree 随树同频拉取（挂载、pod 事件、30s 轮询），
  // 不再按选中单独拉；容器接口过滤 Failed 阶段 pod（buildStateContainers）：死透 pod 无日志按钮，强杀只认 namespace+pod 不受影响。

  /** 适应视图：仅按当前布局自适应居中（不重排，手动拖拽位置保留） */
  const handleFit = () => graphRef.current?.fitView()

  /** 重置布局：丢弃手动拖拽，重跑确定性分层布局再自适应居中 */
  const handleReset = () => {
    live.relayout(live.graph)
    graphRef.current?.fitView()
  }

  /** 手动刷新：按钮转圈 loading（fetchTree 内部 catch 并 toast，永不 throw，await 即完成时） */
  const handleRefresh = async () => {
    setRefreshing(true)
    try {
      await fetchTree()
    } finally {
      setRefreshing(false)
    }
  }

  // 宿主弹窗放大/还原/缩放完成后自动适应视图（resizeAt 只在宽/高真正变化时 bump，对齐 TabShell 的 refit 机制）。
  // 仅当前 Tab 可见时跟随：拓扑 Tab 切走后 wrapper 是 hidden（display:none），容器 0 尺寸，
  // fitView 会按 0 视口把 view 算到错误位置，跳过；切回可见由 TopologyGraph 挂载时的初始 fit 兜底。
  useEffect(() => {
    if (rootRef.current && rootRef.current.getBoundingClientRect().width > 0) {
      graphRef.current?.fitView()
    }
  }, [resizeAt])

  /** 强杀选中 pod：对齐 TabShell 的 force_delete（gracePeriodSeconds=0 立即删），成功后立即拉一次树（pod 事件也会触发） */
  const killPod = async () => {
    if (!selectedNode) return
    setKilling(true)
    try {
      const { error } = await api.POST(API.containerForceDelete, {
        params: { path: { namespace: selectedNode.namespace, pod: selectedNode.name } },
        body: { namespace: selectedNode.namespace, pod: selectedNode.name, gracePeriodSeconds: '0' },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('project.forceDeleteSuccess', { pod: selectedNode.name }))
      void fetchTree()
    } catch (e) {
      toast.error(
        t('project.forceDeleteFailed', { pod: selectedNode.name }) +
          (e instanceof Error ? `: ${e.message}` : ''),
      )
    } finally {
      setKilling(false)
      setKillOpen(false)
    }
  }

  /** 选中 Pod 的容器名（按 pod 名过滤项目容器列表） */
  const podContainerNames = podContainers
    .filter((c) => c.pod === selectedNode?.name)
    .map((c) => c.container)

  /** Pod 节点操作区：容器名列表 + 底部操作行（左「查看日志」逐容器、右「强杀」pod 级；非 Pod 节点不渲染） */
  const podActions = selectedNode?.kind === 'Pod' ? (
    <div className="py-1">
      <div className="mb-1.5 text-[12px] text-mute">{t('topology.containers')}</div>
      {podContainerNames.length === 0 ? (
        <div className="py-0.5 text-[12px] text-faint">{t('project.noContainers')}</div>
      ) : (
        <div className="flex flex-col gap-1">
          {podContainerNames.map((c) => (
            <span key={c} className="truncate font-mono text-[12px] text-ink" title={c}>
              {c}
            </span>
          ))}
        </div>
      )}
      {/* 操作行：左「查看日志」（逐容器，单容器最常见；多容器时同排各一个，title 标容器名）+
          右「强杀」（pod 级，只出现一次）。对齐 TabShell 的 Popover 小确认框（nextZIndex 压过叠加弹窗） */}
      <div className="mt-2 flex items-center justify-between gap-2">
        <div className="flex min-w-0 flex-wrap gap-1">
          {podContainerNames.map((c) => (
            <button
              key={c}
              type="button"
              onClick={() =>
                setLogTarget({
                  namespace: selectedNode.namespace,
                  pod: selectedNode.name,
                  container: c,
                })
              }
              title={c}
              className="shrink-0 rounded border border-primary/50 px-1.5 py-px text-[11px] text-primary transition-colors hover:bg-primary/20"
            >
              {t('project.viewLog')}
            </button>
          ))}
        </div>
        <Popover
          open={killOpen}
          onOpenChange={(o) => {
            if (o) setConfirmZ(nextZIndex())
            setKillOpen(o)
          }}
        >
          <PopoverTrigger asChild>
            <button
              type="button"
              className="inline-flex shrink-0 items-center gap-1 rounded-md border border-err/40 px-2 py-0.5 text-[11px] text-err transition-colors hover:bg-err-soft"
            >
              <Icon name="power" className="text-[11px]" />
              {t('project.forceDeletePod')}
            </button>
          </PopoverTrigger>
          <PopoverContent align="start" sideOffset={6} className="w-72 p-3" style={{ zIndex: confirmZ }}>
            <div className="flex flex-col gap-2">
              <div className="text-[12px] font-medium text-ink">{t('project.forceDeleteTitle')}</div>
              <div className="text-[11px] leading-snug text-mute">
                {t('project.forceDeleteDesc', { pod: selectedNode.name })}
              </div>
              <div className="flex justify-end gap-1.5">
                <Button size="xs" variant="outline" onClick={() => setKillOpen(false)}>
                  {t('common.cancel')}
                </Button>
                <Button
                  size="xs"
                  variant="destructive"
                  disabled={killing}
                  onClick={() => void killPod()}
                >
                  {killing && <Icon name="loader" className="size-3 animate-spin" />}
                  {t('project.forceDeletePod')}
                </Button>
              </div>
            </div>
          </PopoverContent>
        </Popover>
      </div>
    </div>
  ) : null

  return (
    <>
    <div ref={rootRef} className="flex h-full min-h-0 flex-col gap-3">
      {/* 小工具条：标题 + 实时徽标 + 缩放 / 适应视图 / 重置布局 / 刷新 */}
      {/* pt-1 给 focus ring（3px 外扩）留出顶部空间，否则被外层 overflow-auto 容器裁掉 */}
      <div className="flex shrink-0 flex-wrap items-center justify-between gap-2 pt-1">
        <div className="flex items-center gap-2">
          <h2 className="text-[14px] font-semibold text-ink">{t('topology.title')}</h2>
          <span className="flex items-center gap-1.5 text-[11px] text-primary">
            <span className="size-1.5 animate-pulse rounded-full bg-primary" />
            {t('topology.liveUpdated')}
          </span>
        </div>
        <div className="flex flex-wrap items-center gap-1.5">
          <SearchInput
            value={query}
            onChange={setQuery}
            placeholder={t('topology.searchPlaceholder')}
            size="sm"
            className="w-56"
          />
          {queryActive && (
            <span className="text-[12px] text-mute">
              {t('topology.searchMatchCount', { count: matchedIds?.size ?? 0 })}
            </span>
          )}
          <Button size="sm" variant="outline" className="h-7 px-2" onClick={() => graphRef.current?.zoomOut()} title={t('topology.zoomOut')}>
            <Icon name="minus" className="size-3.5" />
          </Button>
          <Button size="sm" variant="outline" className="h-7 px-2" onClick={() => graphRef.current?.zoomIn()} title={t('topology.zoomIn')}>
            <Icon name="plus" className="size-3.5" />
          </Button>
          <Button size="sm" variant="outline" onClick={handleFit} className="h-7" title={t('topology.fitView')}>
            <Icon name="expand" className="size-3.5" />
            {t('topology.fitView')}
          </Button>
          <Button size="sm" variant="outline" onClick={handleReset} className="h-7" title={t('topology.resetLayout')}>
            <Icon name="refresh-cw" className="size-3.5" />
            {t('topology.resetLayout')}
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={() => void handleRefresh()}
            className="h-7"
            disabled={refreshing}
            title={t('topology.liveRefresh')}
          >
            <Icon
              name={refreshing ? 'loader' : 'refresh'}
              className={`size-3.5 ${refreshing ? 'animate-spin' : ''}`}
            />
            {t('topology.liveRefresh')}
          </Button>
        </div>
      </div>

      {/* 画布区域：图 + 骨架/空态 + 图例 + 详情面板。
          graph 为 null = 首次加载在途 → 骨架屏（占位结构与真实分层树对齐）；
          已加载但节点为空 = 真正无资源 → Empty 提示；有节点 → 渲染真图 */}
      <div className="relative min-h-0 flex-1">
        {graph === null ? (
          <SkeletonTopology />
        ) : graph.nodes.length === 0 ? (
          <div className="flex h-full items-center justify-center rounded-lg border border-line bg-surface">
            <Empty text={t('topology.liveEmpty')} icon="network" />
          </div>
        ) : (
          <TopologyGraph
            ref={graphRef}
            graph={graph}
            rootId={live.rootId}
            positions={live.positions}
            posRef={live.posRef}
            draggingId={live.draggingId}
            selectedId={selectedId}
            hoveredId={hoveredId}
            matchedIds={matchedIds}
            onSelect={setSelectedId}
            onHover={setHoveredId}
            beginDrag={live.beginDrag}
            endDrag={live.endDrag}
          />
        )}

        {/* 空态提示：无选中时引导点击节点（仅真图非空时显示，空图/骨架不出现） */}
        {graph && graph.nodes.length > 0 && !selectedId && (
          <div className="pointer-events-none absolute top-3 left-3 z-10 flex items-center gap-1.5 rounded-md bg-overlay/85 px-2.5 py-1.5 text-[12px] text-faint backdrop-blur-sm">
            <Icon name="info" className="size-3.5" />
            {t('topology.noSelection')}
          </div>
        )}

        {/* 图例：状态色 + 关系线型（仅真图非空时显示，空图/骨架不出现） */}
        {graph && graph.nodes.length > 0 && (
          <div className="absolute bottom-3 left-3 z-10 rounded-lg border border-line bg-overlay px-3 py-2 shadow-sm">
            <div className="mb-1.5 text-[11px] font-medium text-faint">{t('topology.legend')}</div>
            <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-[11px] text-mute">
              <span className="text-faint">{t('topology.legendAppStatus')}</span>
              <span className="flex items-center gap-1.5">
                <StatusDot tone="ok" />
                {t('topology.legendHealthy')}
              </span>
              <span className="flex items-center gap-1.5">
                <StatusDot tone="err" />
                {t('topology.legendDegraded')}
              </span>
              <span className="flex items-center gap-1.5">
                <StatusDot tone="warn" />
                {t('topology.legendProgressing')}
              </span>
              <span className="mx-1 hidden h-3 w-px bg-line sm:block" />
              <span className="flex items-center gap-1.5">
                <span className="h-0 w-5 border-t-2 border-line-strong" />
                {t('topology.legendOwner')}
              </span>
              <span className="flex items-center gap-1.5">
                <span className="h-0 w-5 border-t-2 border-dashed border-line" />
                {t('topology.legendSelector')}
              </span>
            </div>
            {/* Pod 生命周期状态点（直播 Tab Pod 节点换色；其余节点/Application 根仍用上方健康色） */}
            <div className="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-line pt-1.5 text-[11px] text-mute">
              <span className="text-faint">{t('topology.legendPodStates')}</span>
              <span className="flex items-center gap-1.5">
                <span className="size-2 rounded-full" style={{ background: '#a78bfa' }} />
                {t('topology.legendReady')}
              </span>
              <span className="flex items-center gap-1.5">
                <span className="size-2 rounded-full" style={{ background: '#93c5fd' }} />
                {t('topology.legendNotReady')}
              </span>
              <span className="flex items-center gap-1.5">
                <span className="size-2 rounded-full" style={{ background: '#67e8f9' }} />
                {t('topology.legendStarting')}
              </span>
              <span className="flex items-center gap-1.5">
                <span className="size-2 rounded-full" style={{ background: '#fca5a5' }} />
                {t('topology.legendStopping')}
              </span>
              <span className="flex items-center gap-1.5">
                <span className="size-2 rounded-full" style={{ background: '#fde047' }} />
                {t('topology.legendAboutToStop')}
              </span>
            </div>
          </div>
        )}

        {/* 节点详情面板 */}
        {selectedNode && (
          <TopologyDetail
            node={selectedNode}
            onClose={() => setSelectedId(null)}
            actions={podActions}
            endpoints={endpoints}
          />
        )}
      </div>

      {/* Pod 容器实时日志弹窗（复用 DeployLog 的 ContainerLogModal：SSE 流 + ANSI 着色 + 自动滚底） */}
      {logTarget && (
        <ContainerLogModal
          open
          onClose={() => setLogTarget(null)}
          title={`${logTarget.pod}/${logTarget.container}`}
          container={logTarget}
        />
      )}
    </div>
    </>
  )
}
