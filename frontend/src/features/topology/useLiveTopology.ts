import { useCallback, useRef, useState, type MutableRefObject } from 'react'
import {
  extendLayout,
  HIERARCHY_TREE,
  kindColumnSignature,
  layoutGraph,
  type PositionMap,
} from './layout'
import type { TopoGraph } from './topologyTypes'

export interface UseLiveTopology {
  /** 当前资源树图（首次数据到达后非空；刷新为全新引用驱动重渲） */
  graph: TopoGraph | null
  /** 渲染位置快照：随 commitGraph / 拖拽结束更新 */
  positions: PositionMap
  /** 权威位置（可变）：拖拽期间由画布层直接改 DOM，不触发 re-render */
  posRef: MutableRefObject<PositionMap>
  /** Application 根节点 id（画布层做根节点强调样式；空图时 ''） */
  rootId: string
  /** 当前被拖拽的节点 id */
  draggingId: string | null
  beginDrag: (id: string) => void
  endDrag: () => void
  /** 提交一张新资源树：首载走确定性分层布局，刷新钉住存活节点 + 新节点追加列底 */
  commitGraph: (g: TopoGraph) => void
  /** 重排一次：对当前图重跑全量确定性分层布局（丢弃手动拖拽位置），供「适应视图」按钮联动 */
  relayout: (g: TopoGraph | null) => void
}

/**
 * 实时拓扑布局 hook（资源树 Tab 专用）：持有不断刷新的图与一套自管理的布局位置。
 * 图是「活的」（pod 事件驱动重拉），布局须在刷新间保持——
 * 首次 commitGraph 用 layoutGraph 出确定性 LR 树；之后每次用 extendLayout 把旧布局
 * 当 base：存活节点坐标一像素不动（含用户拖拽结果）、新出现节点追加列底、消失节点
 * 自动剔除（extendLayout 只遍历新图节点）。幂等：节点集不变 → 位置不变，只重渲状态色。
 * 分层用 HIERARCHY_TREE（owner+route+selector 全成层级）：把
 * app→ingress→svc→deploy→pod 排成逐列链，而非只按 owner 塌缩成稀疏列。
 */
export function useLiveTopology(): UseLiveTopology {
  const [graph, setGraph] = useState<TopoGraph | null>(null)
  const [positions, setPositions] = useState<PositionMap>({})
  const posRef = useRef<PositionMap>({})
  // 上次已提交布局（作 extendLayout 的 base）：endDrag 时同步拖拽结果，刷新后保留
  const layoutRef = useRef<PositionMap | null>(null)
  // 上次提交图的 kind 列位签名：kind 集变化 → 列位重映射 → 需整体重排
  const kindSigRef = useRef('')
  const [draggingId, setDraggingId] = useState<string | null>(null)

  /** 提交新图：首载分层布局，刷新 extendLayout（钉存活 + 追加新节点 + 剔除旧节点）；
   *  kind 集变化（如新出现 Ingress）时列位映射变了，整体重排而非钉住追加，避免错位重叠 */
  const commitGraph = useCallback((g: TopoGraph) => {
    setGraph(g)
    const sig = kindColumnSignature(g)
    const mappingChanged = sig !== kindSigRef.current
    kindSigRef.current = sig
    const base = layoutRef.current
    const next = !base || mappingChanged
      ? layoutGraph(g, HIERARCHY_TREE).positions
      : extendLayout(g, base, HIERARCHY_TREE)
    layoutRef.current = next
    posRef.current = next
    setPositions(next)
  }, [])

  /** 拖拽开始：标记被拖节点（画布层暂停悬浮高亮） */
  const beginDrag = useCallback((id: string) => {
    setDraggingId(id)
  }, [])

  /** 拖拽结束：清标记、提交位置快照，并把拖拽结果同步进布局 base（刷新后保留） */
  const endDrag = useCallback(() => {
    setDraggingId(null)
    layoutRef.current = { ...posRef.current }
    setPositions({ ...posRef.current })
  }, [])

  const rootId = graph?.nodes.find((n) => n.kind === 'Application')?.id ?? ''

  /** 重排：layoutGraph 全量确定性重排，同步 layoutRef（后续刷新钉住这份新布局）/posRef（fitView 读它）/positions（重渲）。
   *  不更新 kindSigRef：kind 集不变时后续 commitGraph 走 extendLayout 钉住新布局，符合「重排一次」语义 */
  const relayout = useCallback((g: TopoGraph | null) => {
    if (!g) return
    const fresh = layoutGraph(g, HIERARCHY_TREE).positions
    layoutRef.current = fresh
    posRef.current = fresh
    setPositions(fresh)
  }, [])

  return { graph, positions, posRef, rootId, draggingId, beginDrag, endDrag, commitGraph, relayout }
}
