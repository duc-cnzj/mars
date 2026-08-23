/** 资源拓扑图 demo 页：领域类型定义（纯前端 mock 数据模型） */

/** 节点健康状态：healthy 健康 / degraded 异常 / progressing 进行中 / unknown 未知 */
export type NodeStatus = 'healthy' | 'degraded' | 'progressing' | 'unknown'

/** Pod 生命周期细粒度状态（对齐 TabLog 的 StateContainer：就绪 / 未就绪 / 启动中 / 停止中 / 即将停止）。
 *  资源树接口不携带此粒度，仅直播 Tab 由 /containers 聚合后附加到 Pod 节点；缺失时状态点走健康色兜底 */
export type PodLifecycle = 'isOld' | 'terminating' | 'pending' | 'notReady' | 'ready'

/** 支持的 k8s 资源类型（demo 演示集；Application 为拓扑根节点，对齐 Argo 资源树） */
export type NodeKind =
  | 'Application'
  | 'Deployment'
  | 'ReplicaSet'
  | 'Pod'
  | 'Service'
  | 'Ingress'
  | 'ConfigMap'
  | 'Secret'
  | 'HPA'
  | 'StatefulSet'
  | 'DaemonSet'

/** 边关系类型：owner 属主引用 / selector 标签选择器 / config 引用挂载 / route 路由转发 */
export type EdgeType = 'owner' | 'selector' | 'config' | 'route'

/** 节点事件（详情面板展示，mock） */
export interface TopoEvent {
  type: 'normal' | 'warning'
  /** 相对时间文案（mock，如 "3m"） */
  time: string
  message: string
}

/** 拓扑图节点：一个 k8s 资源 */
export interface TopoNode {
  id: string
  kind: NodeKind
  name: string
  namespace: string
  status: NodeStatus
  labels: Record<string, string>
  events: TopoEvent[]
  /** Pod 生命周期（仅直播 Tab 由 /containers 聚合附加；demo 节点无此字段，状态点走健康色兜底） */
  lifecycle?: PodLifecycle
}

/** 项目访问地址（对齐后端 types.ServiceEndpoint：name / url / portName）。
 *  非节点字段：全部平铺在 Application 根节点详情卡片（直播 Tab 由 /api/endpoints 传入） */
export interface TopoEndpoint {
  name: string
  url: string
  portName?: string
}

/** 拓扑图边：两个节点间的关系 */
export interface TopoEdge {
  id: string
  type: EdgeType
  source: string
  target: string
}

/** 整个拓扑图（demo 静态数据） */
export interface TopoGraph {
  nodes: TopoNode[]
  edges: TopoEdge[]
}
