import type { IconName } from '@/components/Icons'
import type { Tone } from '@/components/ui/Tag'
import type { NodeKind, NodeStatus, TopoEdge, TopoGraph, TopoNode } from './topologyTypes'

/** 资源类型 → 现有线性图标（沿用 Icon 图集，零新增图标） */
export const KIND_ICON: Record<NodeKind, IconName> = {
  Application: 'grid', // 应用根（Argo 资源树根节点）
  Deployment: 'rocket', // 部署/发布
  ReplicaSet: 'copy', // 副本
  Pod: 'boxes', // 容器
  Service: 'network', // 网络端点
  Ingress: 'external', // 外部入口
  ConfigMap: 'gear', // 配置
  Secret: 'key', // 密钥
  HPA: 'gauge', // 指标扩缩
  StatefulSet: 'database', // 有状态工作负载（redis 等持久化实例）
  DaemonSet: 'cpu', // 节点级守护（logger 等每节点一个）
}

/** 节点健康状态 → 语义 tone（与 Tag/StatusDot 联动，随主题换肤） */
export const STATUS_TONE: Record<NodeStatus, Tone> = {
  healthy: 'ok',
  degraded: 'err',
  // 进行中：琥珀橙（用户要求避开默认 info 蓝；与 Pod 生命周期五色均不冲突）
  progressing: 'warn',
  unknown: 'mute', // 未知（资源树中无工作负载/状态缺失）
}

/** 默认隐藏的资源类型：ReplicaSet 是 Deployment→Pod 的中间实现细节，简化展示时压缩掉 */
export const HIDDEN_KINDS: NodeKind[] = ['ReplicaSet']

/**
 * 应用级健康状态：由全部 Pod 聚合（最重优先 degraded > progressing > healthy；无 Pod → unknown）。
 * 用于 Application 根节点的状态点配色（根自身不报状态，健康度看它承载的 pod）。
 */
export function deriveAppStatus(nodes: TopoNode[]): NodeStatus {
  const pods = nodes.filter((n) => n.kind === 'Pod')
  if (pods.length === 0) return 'unknown'
  if (pods.some((p) => p.status === 'degraded')) return 'degraded'
  if (pods.some((p) => p.status === 'progressing')) return 'progressing'
  return 'healthy'
}

/** 把 Application 根节点状态替换为 pod 聚合结果（直播资源树 Tab：根不报状态，健康看 pod） */
export function withDerivedAppStatus(graph: TopoGraph): TopoGraph {
  const appStatus = deriveAppStatus(graph.nodes)
  return {
    ...graph,
    nodes: graph.nodes.map((n) => (n.kind === 'Application' && n.status !== appStatus ? { ...n, status: appStatus } : n)),
  }
}

/**
 * 从图中剔除指定 kind 的节点并重连边（跨过被删节点，把它的 keep 祖先与 keep 后代直连）。
 * 用于隐藏中间层（如 ReplicaSet）把 Deployment 直挂 Pod；被删节点成链时逐层穿透。
 * 保留被删节点出边的 type（owner 链仍是 owner）。重复的父→子组合去重（id 取 "父->子"）。
 */
export function dropKinds(graph: TopoGraph, kinds: NodeKind[]): TopoGraph {
  const drop = new Set<NodeKind>(kinds)
  const nodes = graph.nodes.filter((n) => !drop.has(n.kind))
  const nodeKind = new Map(graph.nodes.map((n) => [n.id, n.kind]))

  // 邻接（按方向遍历用）：跨过被删节点找 keep 端点
  const parents = new Map<string, string[]>()
  const children = new Map<string, string[]>()
  for (const edge of graph.edges) {
    if (!children.has(edge.source)) children.set(edge.source, [])
    children.get(edge.source)!.push(edge.target)
    if (!parents.has(edge.target)) parents.set(edge.target, [])
    parents.get(edge.target)!.push(edge.source)
  }

  /** 从 start 沿 dir 邻接表走，跨过被删节点，返回所有 keep 端点（start 本身 keep 则含它） */
  const reach = (start: string, dir: Map<string, string[]>): string[] => {
    const out: string[] = []
    const stack = [start]
    const seenNode = new Set<string>()
    while (stack.length) {
      const cur = stack.pop()!
      if (seenNode.has(cur)) continue
      seenNode.add(cur)
      if (!drop.has(nodeKind.get(cur) ?? ('Application' as NodeKind))) {
        out.push(cur)
        continue
      }
      for (const next of dir.get(cur) ?? []) stack.push(next)
    }
    return out
  }

  const edges: TopoEdge[] = []
  const seen = new Set<string>()
  for (const edge of graph.edges) {
    const sDropped = drop.has(nodeKind.get(edge.source) ?? ('Application' as NodeKind))
    const tDropped = drop.has(nodeKind.get(edge.target) ?? ('Application' as NodeKind))
    if (!sDropped && !tDropped) {
      edges.push(edge)
      continue
    }
    const sources = sDropped ? reach(edge.source, parents) : [edge.source]
    const targets = tDropped ? reach(edge.target, children) : [edge.target]
    for (const s of sources) {
      for (const t of targets) {
        if (s === t) continue
        const key = `${s}->${t}`
        if (seen.has(key)) continue
        seen.add(key)
        edges.push({ id: key, type: edge.type, source: s, target: t })
      }
    }
  }
  return { nodes, edges }
}
