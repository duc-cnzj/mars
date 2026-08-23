import type { EdgeType, TopoEdge, TopoGraph, TopoNode, NodeKind } from './topologyTypes'

/** 二维向量（逻辑视口坐标） */
export interface Vec {
  x: number
  y: number
}

/** 节点 id → 左上角坐标 的映射（渲染权威位置） */
export type PositionMap = Record<string, Vec>

/** 逻辑世界尺寸（fitView 会按需缩放适配画布，此值只决定初始坐标系范围）
 *  世界宽 2700 ≈ demo 树 6 列内容宽度 2412 + 边距，保证平移钳制不会盖住最右列 */
export const WORLD_W = 2700
export const WORLD_H = 900
export const WORLD_CENTER: Vec = { x: WORLD_W / 2, y: WORLD_H / 2 }
/** 世界软边界：拖拽/平移允许越界到该余量内 */
export const WORLD_MARGIN = 250

/** 参与层级（rank）计算的边类型集合：owner + route + selector 都算层级 ——
 *  Ingress→Service（route）、Service→Deployment（selector）构成多父树，父子决定列次，
 *  把 app→ingress→svc→deploy→pod 排成逐列链。 */
export const HIERARCHY_TREE: ReadonlySet<EdgeType> = new Set<EdgeType>(['owner', 'selector', 'route'])

/** 节点盒尺寸（对齐 Argo 资源树的 282×52） */
export const NODE_W = 282
export const NODE_H = 52
/** 相邻列（rank）间距：上一列节点右缘 → 下一列节点左缘 */
export const RANK_SEP = 160
/** 同列相邻行间距 */
export const NODE_SEP = 26
/** 首列左缘偏移（世界坐标） */
export const RANK_OFFSET_X = 120

/** 布局结果：节点坐标 + Application 根节点 id */
export interface TopoLayout {
  positions: PositionMap
  rootId: string
}

/** 一条边的正交路由结果（折线顶点序列，首尾为源/目标盒边缘落点） */
export interface EdgeRoute {
  id: string
  type: EdgeType
  source: string
  target: string
  points: Vec[]
}

/** k8s kind 显示顺序：同列按此排序（对齐 Argo compareNodes 的按名排列习惯） */
const KIND_ORDER: Record<NodeKind, number> = {
  ConfigMap: 0,
  Deployment: 1,
  HPA: 2,
  Ingress: 3,
  ReplicaSet: 4,
  Secret: 5,
  Service: 6,
  Pod: 7,
  StatefulSet: 8,
  DaemonSet: 9,
  Application: 10,
}

/** 同列内比较：kind 序为主，其次名称自然序 */
function compareRankNodes(a: TopoNode, b: TopoNode): number {
  const ka = KIND_ORDER[a.kind] ?? 99
  const kb = KIND_ORDER[b.kind] ?? 99
  return ka - kb || a.name.localeCompare(b.name, undefined, { numeric: true })
}

/** kind → 列位次序（同值归同一列，数字小者靠左）。对齐资源链
 *  App → Ingress → Service → [Deployment|StatefulSet|DaemonSet|ConfigMap|Secret|HPA] → [ReplicaSet] → Pod；
 *  ReplicaSet 恒被 HIDDEN_KINDS 剔除，其列位仅占位不参与压缩。 */
const KIND_COLUMN_ORDER: Record<NodeKind, number> = {
  Application: 0,
  Ingress: 1,
  Service: 2,
  Deployment: 3,
  StatefulSet: 3,
  DaemonSet: 3,
  ConfigMap: 3,
  Secret: 3,
  HPA: 3,
  ReplicaSet: 4,
  Pod: 5,
}

/** 图实际占用的列位次序（升序去重）：列位只按「出现的类型」分配，压缩掉空列 */
function columnSlotsOf(graph: TopoGraph): number[] {
  return [...new Set(graph.nodes.map((n) => KIND_COLUMN_ORDER[n.kind] ?? 99))].sort((a, b) => a - b)
}

/**
 * 按 kind 定列位：同 kind 恒同列（列位 = kind 在资源链中的次序，未出现的列位压缩）。
 * 替代按深度分层——深度会把不同分支的同 kind 节点拆到不同列（如无 svc 覆盖的
 * Deployment 在深 1、有 svc 覆盖的在深 3），导致同类型散到多列、数量多时显乱。
 */
function computeKindRanks(graph: TopoGraph): Map<string, number> {
  const slots = columnSlotsOf(graph)
  const colOfSlot = new Map(slots.map((s, i) => [s, i]))
  const rank = new Map<string, number>()
  for (const n of graph.nodes) rank.set(n.id, colOfSlot.get(KIND_COLUMN_ORDER[n.kind] ?? 99)!)
  return rank
}

/** kind 列位签名（逗号串）：图当前占用哪些列位。live Tab 用它判断 kind 集是否变化——
 *  变化即列位重映射，需整体重排而非钉住追加（否则新旧列位错位重叠）。 */
export function kindColumnSignature(graph: TopoGraph): string {
  return columnSlotsOf(graph).join(',')
}

/**
 * 按 rank 分组后做层内排序：初始按 kind+name；再以 barycenter 启发式
 * （子节点贴近其「层级父节点」在上一列的行号）迭代若干次减少交叉。
 */
function orderRanks(
  graph: TopoGraph,
  rank: Map<string, number>,
  nodeById: Map<string, TopoNode>,
  hierarchy: ReadonlySet<EdgeType>,
): string[][] {
  const maxRank = Math.max(...graph.nodes.map((n) => rank.get(n.id)!))
  const ranks: string[][] = []
  for (let r = 0; r <= maxRank; r++) {
    ranks.push(
      graph.nodes
        .filter((n) => rank.get(n.id) === r)
        .map((n) => n.id)
        .sort((a, b) => compareRankNodes(nodeById.get(a)!, nodeById.get(b)!)),
    )
  }
  // barycenter：每个节点取层级父节点在上一列的行号均值，贴近父节点排
  for (let pass = 0; pass < 3; pass++) {
    for (let r = 1; r < ranks.length; r++) {
      const prevIdx = new Map(ranks[r - 1].map((id, i) => [id, i]))
      const bary = new Map<string, { sum: number; n: number }>()
      for (const edge of graph.edges) {
        if (!hierarchy.has(edge.type)) continue
        if (prevIdx.has(edge.source) && ranks[r].includes(edge.target)) {
          const cur = bary.get(edge.target) ?? { sum: 0, n: 0 }
          cur.sum += prevIdx.get(edge.source)!
          cur.n += 1
          bary.set(edge.target, cur)
        }
      }
      ranks[r].sort((a, b) => {
        const ba = bary.get(a)
        const bb = bary.get(b)
        const avgA = ba ? ba.sum / ba.n : 0
        const avgB = bb ? bb.sum / bb.n : 0
        return avgA - avgB || compareRankNodes(nodeById.get(a)!, nodeById.get(b)!)
      })
    }
  }
  return ranks
}

/**
 * 分层布局（Argo 资源树风格）：LR 树形，每列一个 rank，同列节点垂直居中排布。
 * 返回节点左上角坐标（top-left，非中心），供渲染与边路由共用。
 * 列位按 kind 定（computeKindRanks）：同 kind 恒同列，App→Ingress→Service→Workload→Pod
 * 排成逐列链；hierarchy 仅参与列内 barycenter 排序（子贴近上一列的父），不决定列位。
 */
export function layoutGraph(graph: TopoGraph, hierarchy: ReadonlySet<EdgeType> = HIERARCHY_TREE): TopoLayout {
  const root = graph.nodes.find((n) => n.kind === 'Application')
  if (!root) throw new Error('layoutGraph: 图中缺少 Application 根节点')
  const rank = computeKindRanks(graph)
  const nodeById = new Map(graph.nodes.map((n) => [n.id, n]))
  const ranks = orderRanks(graph, rank, nodeById, hierarchy)
  const positions: PositionMap = {}
  for (let r = 0; r < ranks.length; r++) {
    const list = ranks[r]
    const colH = list.length * (NODE_H + NODE_SEP) - NODE_SEP
    const top = WORLD_CENTER.y - colH / 2
    const x = RANK_OFFSET_X + r * (NODE_W + RANK_SEP)
    list.forEach((id, i) => {
      positions[id] = { x, y: top + i * (NODE_H + NODE_SEP) }
    })
  }
  return { positions, rootId: root.id }
}

/**
 * 在既有布局之上追加新节点（钉住布局）：模拟部署期间把新生成的 ReplicaSet/Pod
 * 追加到所属列底部，而**不移动任何既有节点**——保证模拟开始/结束切回基础布局时
 * 基础节点一像素不动，只有新节点淡入/淡出。rank 与列内排序沿用 layoutGraph 的规则。
 */
export function extendLayout(
  union: TopoGraph,
  base: PositionMap,
  hierarchy: ReadonlySet<EdgeType> = HIERARCHY_TREE,
): PositionMap {
  const rank = computeKindRanks(union)
  const nodeById = new Map(union.nodes.map((n) => [n.id, n]))
  const ranks = orderRanks(union, rank, nodeById, hierarchy)
  const pos: PositionMap = {}
  for (let r = 0; r < ranks.length; r++) {
    const x = RANK_OFFSET_X + r * (NODE_W + RANK_SEP)
    // 既有节点保持原始坐标，同时记录该列底部，供下方追加新节点
    let bottom = -Infinity
    for (const id of ranks[r]) {
      const b = base[id]
      if (b) {
        pos[id] = b
        bottom = Math.max(bottom, b.y + NODE_H)
      }
    }
    if (bottom === -Infinity) bottom = WORLD_CENTER.y - NODE_H / 2
    // 新节点（不在 base 中）从列底下方依次排开
    let y = bottom + NODE_SEP
    for (const id of ranks[r]) {
      if (base[id]) continue
      pos[id] = { x, y }
      y += NODE_H + NODE_SEP
    }
  }
  return pos
}

/**
 * 边路由：把每条边转成避开节点盒的正交折线（对齐 Argo 的 L 形/缩进折线）。
 * - 跨列边：出源右缘 → 折点 → 进目标左缘（箭头向右）
 * - 同列边（如 hpa→deployment、configmap→deployment）：出源右缘经右间隙绕行，
 *   进目标右缘（箭头向左），避免与跨列边挤在左间隙里
 * - 多边汇入同一目标 / 同一源发散：按 id 稳定分槽，入口/出口 y 均摊到盒高内
 */
export function routeEdges(edges: TopoEdge[], pos: PositionMap): EdgeRoute[] {
  // 汇入同一目标的边：按 id 排序后分槽（决定目标侧入口 y）
  const incomingByTarget = new Map<string, string[]>()
  for (const e of edges) {
    const arr = incomingByTarget.get(e.target) ?? []
    arr.push(e.id)
    incomingByTarget.set(e.target, arr)
  }
  const targetSlot = new Map<string, number>()
  for (const ids of incomingByTarget.values()) {
    ids.sort()
    ids.forEach((id, i) => targetSlot.set(id, i))
  }
  // 同一源发出的边：分槽（决定源侧出口 y）
  const outgoingBySource = new Map<string, string[]>()
  for (const e of edges) {
    const arr = outgoingBySource.get(e.source) ?? []
    arr.push(e.id)
    outgoingBySource.set(e.source, arr)
  }
  const sourceSlot = new Map<string, number>()
  for (const ids of outgoingBySource.values()) {
    ids.sort()
    ids.forEach((id, i) => sourceSlot.set(id, i))
  }

  return edges.map((e) => {
    const s = pos[e.source]
    const t = pos[e.target]
    if (!s || !t) return { id: e.id, type: e.type, source: e.source, target: e.target, points: [] }
    const outN = outgoingBySource.get(e.source)?.length ?? 1
    const outSlot = sourceSlot.get(e.id) ?? 0
    const inN = incomingByTarget.get(e.target)?.length ?? 1
    const inSlot = targetSlot.get(e.id) ?? 0
    const sy = s.y + (NODE_H * (outSlot + 0.5)) / outN
    const ty = t.y + (NODE_H * (inSlot + 0.5)) / inN

    const points: Vec[] = []
    if (Math.abs(s.x - t.x) < 1) {
      // 同列：右缘出 → 右间隙绕行 → 目标右缘入（箭头向左）
      const sx = s.x + NODE_W
      const tx = t.x + NODE_W
      const bendX = sx + RANK_SEP / 2
      points.push({ x: sx, y: sy }, { x: bendX, y: sy })
      if (Math.abs(sy - ty) > 0.5) points.push({ x: bendX, y: ty })
      points.push({ x: tx, y: ty })
    } else {
      // 跨列：右缘出 → 折点（按汇入槽位错开）→ 目标左缘入（箭头向右）
      const sx = s.x + NODE_W
      const tx = t.x
      const bendX = sx + ((tx - sx) * (inSlot + 1)) / (inN + 1)
      points.push({ x: sx, y: sy })
      if (Math.abs(sy - ty) > 0.5) {
        points.push({ x: bendX, y: sy })
        points.push({ x: bendX, y: ty })
      }
      points.push({ x: tx, y: ty })
    }
    return { id: e.id, type: e.type, source: e.source, target: e.target, points }
  })
}
