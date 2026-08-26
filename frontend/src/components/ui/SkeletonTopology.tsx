import { Skeleton } from './shadcn/skeleton'

/**
 * 资源拓扑骨架：占位结构与 TopologyTab 真实图对齐（分层树：Application 根 → Service/工作负载
 * → Pod 叶）。节点盒按真实 TopologyNode 的 52px 高 + 「图标列 / 名称 / 状态位」三段式布局，
 * 层级用宽窄收窄 + 细竖线连接示意「图在加载」。加载完成切真图时盒位不跳（reserve space）。
 * 容器自身 h-full 铺满画布；内部用 m-auto 包裹树：画布够高时垂直水平居中，画布过矮时内容
 * 顶对齐且父容器可滚动——避免 justify-center 溢出把 Application 根裁掉（顶部缺失）。
 */
export function SkeletonTopology() {
  return (
    <div className="flex h-full min-h-0 flex-col overflow-auto rounded-lg border border-line bg-surface">
      <div className="m-auto flex shrink-0 flex-col items-center px-6 py-8">
        {/* 根：Application 聚合节点（最宽） */}
        <SkeletonNode className="w-56" />
        <div className="h-6 w-px shrink-0 bg-line" />
        {/* 中间层：2 个 Service/工作负载分支，各挂 2 个 Pod 叶 */}
        <div className="flex items-start gap-10">
          <Branch />
          <Branch />
        </div>
      </div>
    </div>
  )
}

/** 单个分支：Service 节点 + 下方 2 个 Pod 叶（宽递减，暗示层级越深节点越密） */
function Branch() {
  return (
    <div className="flex flex-col items-center gap-4">
      <SkeletonNode className="w-40" />
      <div className="flex items-start gap-3">
        <SkeletonNode className="w-28" />
        <SkeletonNode className="w-28" />
      </div>
    </div>
  )
}

/** 拓扑节点盒：52px 高（对齐真实 NODE_H），左图标列 + 中名称/kind + 右状态位 */
function SkeletonNode({ className }: { className?: string }) {
  return (
    <div
      className={`flex h-[52px] shrink-0 items-center gap-3 rounded-md border border-line bg-bg px-3 ${className ?? ''}`}
    >
      <Skeleton className="size-5 shrink-0 rounded-md" />
      <div className="flex min-w-0 flex-1 flex-col gap-2">
        <Skeleton className="h-2.5 w-3/4" />
        <Skeleton className="h-2 w-1/3" />
      </div>
      <Skeleton className="size-2.5 shrink-0 rounded-full" />
    </div>
  )
}
