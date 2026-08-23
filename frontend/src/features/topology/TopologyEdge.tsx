import { memo } from 'react'
import type { EdgeRoute } from './layout'
import type { EdgeType } from './topologyTypes'

/** 边样式表：按关系类型区分线型（实线 owner / 短划点线 route / 虚线 selector / 点线 config） */
const EDGE_STYLE: Record<EdgeType, { dash: string | undefined; color: string }> = {
  owner: { dash: undefined, color: 'stroke-line-strong' },
  route: { dash: '10 4 2 4', color: 'stroke-line-strong' },
  selector: { dash: '7 5', color: 'stroke-line' },
  config: { dash: '2 5', color: 'stroke-line' },
}

export interface TopologyEdgeProps {
  route: EdgeRoute
  /** 该边是否与当前高亮节点直接相连（active → primary 高亮 + 箭头变主色） */
  active: boolean
  /** 存在高亮节点但本边不相关 → 置灰 */
  dimmed: boolean
}

/**
 * 拓扑边：按布局路由出的正交折线画 <path>（对齐 Argo 资源树的折线 + 箭头）。
 * 箭头用 svg marker（orient=auto 自动对齐末段方向，进目标即指向左/右）。
 * vector-effect=non-scaling-stroke：缩放时线宽恒定。opacity 分三档：active 1 / 常态 0.6 / 置灰 0.15。
 */
export const TopologyEdge = memo(function TopologyEdge({ route, active, dimmed }: TopologyEdgeProps) {
  const style = EDGE_STYLE[route.type]
  const d = route.points.map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x} ${p.y}`).join(' ')
  const opacity = active ? 1 : dimmed ? 0.15 : 0.6
  return (
    <g opacity={opacity} className="transition-opacity duration-150">
      <path
        d={d}
        fill="none"
        vectorEffect="non-scaling-stroke"
        strokeDasharray={style.dash}
        strokeWidth={active ? 2 : 1.4}
        markerEnd={active ? 'url(#topo-arrow-primary)' : 'url(#topo-arrow-line)'}
        // selector 虚线边加「行军蚁」流动，暗示服务→Pod 的流量在走
        className={[active ? 'stroke-primary' : style.color, route.type === 'selector' ? 'topo-dash-flow' : ''].join(' ')}
      />
    </g>
  )
})
