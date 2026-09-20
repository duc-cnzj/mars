import { useEffect, useRef, useState, type ReactNode } from 'react'
import { Skeleton } from './shadcn/skeleton'

/**
 * 资源拓扑骨架：占位结构与 TopologyTab 真实图对齐（从左到右的 dagre 分层列布局）。
 * 节点盒按真实 TopologyNode 统一 282px 宽（NODE_W） + 「图标列 / 名称+kind / 状态位」
 * 三段式布局，每列用 ▶ 连接示意流向：Application → Service/工作负载 → Pod。
 * 中间列密度高（4 节点模拟多分支）展示中等规模拓扑的骨架轮廓。
 * 自动测量容器尺寸并 scale 适配，确保骨架在任意视口下完整可见、不溢出。
 */
export function SkeletonTopology() {
  const outerRef = useRef<HTMLDivElement>(null)
  const innerRef = useRef<HTMLDivElement>(null)
  const [scale, setScale] = useState(1)

  useEffect(() => {
    const outer = outerRef.current
    const inner = innerRef.current
    if (!outer || !inner) return
    const fit = () => {
      const ow = outer.clientWidth - 2
      const oh = outer.clientHeight - 2
      const iw = inner.scrollWidth
      const ih = inner.scrollHeight
      if (ow > 0 && oh > 0 && (iw > ow || ih > oh)) {
        setScale(Math.max(0.25, Math.min(ow / iw, oh / ih)))
      } else {
        setScale(1)
      }
    }
    fit()
    const ro = new ResizeObserver(fit)
    ro.observe(outer)
    return () => ro.disconnect()
  }, [])

  return (
    <div
      ref={outerRef}
      className="flex h-full min-h-0 items-center justify-center overflow-hidden rounded-lg border border-line bg-surface"
    >
      <div
        ref={innerRef}
        className="flex shrink-0 items-center gap-8 px-6 py-8"
        style={{
          transform: `scale(${scale})`,
          transformOrigin: 'center center',
        }}
      >
        {/* Rank 0: Application（根节点） */}
        <Rank>
          <SkeletonNode />
        </Rank>

        <Connector />

        {/* Rank 1: Ingress / Service / Deployment */}
        <Rank>
          <SkeletonNode />
          <SkeletonNode />
          <SkeletonNode />
        </Rank>

        <Connector />

        {/* Rank 2: 工作负载层 */}
        <Rank>
          <SkeletonNode />
          <SkeletonNode />
          <SkeletonNode />
          <SkeletonNode />
          <SkeletonNode />
        </Rank>

        <Connector />

        {/* Rank 3: Pod */}
        <Rank>
          <SkeletonNode />
          <SkeletonNode />
          <SkeletonNode />
          <SkeletonNode />
          <SkeletonNode />
          <SkeletonNode />
          <SkeletonNode />
        </Rank>
      </div>
    </div>
  )
}

/** 单列（Rank）：节点垂直堆叠，间隔 NODE_SEP=26px */
function Rank({ children }: { children: ReactNode }) {
  return <div className="flex shrink-0 flex-col gap-[26px]">{children}</div>
}

/** 列间箭头连接器 */
function Connector() {
  return (
    <span className="shrink-0 text-[10px] leading-none text-line/60">▶</span>
  )
}

/** 拓扑节点盒：52px 高（对齐真实 NODE_H=52），左图标列 + 中名称/kind + 右状态位 */
function SkeletonNode() {
  return (
    <div className="flex h-[52px] w-[282px] shrink-0 items-center gap-3 rounded-md border border-line bg-bg px-3">
      <Skeleton className="size-5 shrink-0 rounded-md" />
      <div className="flex min-w-0 flex-1 flex-col gap-2">
        <Skeleton className="h-2.5 w-3/4" />
        <Skeleton className="h-2 w-1/3" />
      </div>
      <Skeleton className="size-2.5 shrink-0 rounded-full" />
    </div>
  )
}
