import {
  forwardRef,
  useCallback,
  useId,
  useImperativeHandle,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type MutableRefObject,
  type PointerEvent as ReactPointerEvent,
} from 'react'
import { cn } from '@/lib/utils'
import { NODE_H, NODE_W, routeEdges, WORLD_H, WORLD_MARGIN, WORLD_W, type PositionMap, type Vec } from './layout'
import { TopologyNode } from './TopologyNode'
import { TopologyEdge } from './TopologyEdge'
import type { TopoGraph } from './topologyTypes'

/** 视图变换（world → 屏幕）：屏幕 = view.x + world * view.scale */
interface View {
  x: number
  y: number
  scale: number
}

export interface TopologyGraphHandle {
  /** 按当前节点位置把图适配到画布（重新居中 + 缩放） */
  fitView: () => void
  /** 放大一级（绕画布中心缩放，钳制与滚轮一致 [0.4, 2.5]） */
  zoomIn: () => void
  /** 缩小一级 */
  zoomOut: () => void
}

interface TopologyGraphProps {
  graph: TopoGraph
  /** Application 根节点 id（节点层做根强调样式） */
  rootId: string
  /** 渲染位置快照 */
  positions: PositionMap
  /** 权威位置（拖拽期间直接改，画布层同步 DOM） */
  posRef: MutableRefObject<PositionMap>
  draggingId: string | null
  selectedId: string | null
  hoveredId: string | null
  /** 搜索命中的节点 id 集合；null 表示未搜索 */
  matchedIds: Set<string> | null
  onSelect: (id: string | null) => void
  onHover: (id: string | null) => void
  beginDrag: (id: string) => void
  endDrag: () => void
}

const clamp = (v: number, lo: number, hi: number) => Math.min(Math.max(v, lo), hi)

/** 平移/缩放后的软钳制：保证世界图（含边距）至少有一部分留在画布内 */
function clampView(v: View, cw: number, ch: number): View {
  const minX = cw - (WORLD_W + WORLD_MARGIN) * v.scale
  const maxX = WORLD_MARGIN * v.scale
  const minY = ch - (WORLD_H + WORLD_MARGIN) * v.scale
  const maxY = WORLD_MARGIN * v.scale
  return {
    scale: v.scale,
    x: clamp(v.x, Math.min(minX, maxX), Math.max(minX, maxX)),
    y: clamp(v.y, Math.min(minY, maxY), Math.max(minY, maxY)),
  }
}

/** 拖拽手势内部态 */
interface DragState {
  id: string
  /** 按下点的客户端坐标（判定 click vs drag 阈值用） */
  startClient: Vec
  /** 抓取偏移：节点位置 = 光标世界点 - 偏移，保证拖起不跳位 */
  grabOffset: Vec
  moved: boolean
}

/** 平移手势内部态 */
interface PanState {
  startClient: Vec
  startView: View
  moved: boolean
}

/**
 * 拓扑图画布：SVG 手写引擎（零依赖）。
 * - 视图：view 变换应用在世界组 <g>，滚轮朝光标缩放（非 passive wheel）+ 空白拖拽平移
 * - 节点：位置来自 positions 快照，拖拽期间画布层直接改节点 <g> 的 transform（无 re-render）
 * - 高亮：hover 算直接邻居（轻量预览）；选中算完整链路（根 → 该对象 → 子孙，含旁支兄弟）；
 *   非高亮节点置灰；搜索命中打外圈虚线 ring
 */
export const TopologyGraph = forwardRef<TopologyGraphHandle, TopologyGraphProps>(
  function TopologyGraph(
    { graph, rootId, positions, posRef, draggingId, selectedId, hoveredId, matchedIds, onSelect, onHover, beginDrag, endDrag },
    ref,
  ) {
    const containerRef = useRef<HTMLDivElement>(null)
    const svgRef = useRef<SVGSVGElement | null>(null)
    const worldRef = useRef<SVGGElement>(null)
    const [view, setView] = useState<View>({ x: 0, y: 0, scale: 1 })
    // 空白处平移手势进行中（背景 cursor 切换 grab ↔ grabbing）
    const [panning, setPanning] = useState(false)
    const gridId = useId()

    // 节点 <g> 元素注册表：拖拽期间按 id 直接 setAttribute('transform')，绕开 React
    const nodeEls = useRef(new Map<string, SVGGElement>())
    const registerNodeEl = useCallback((id: string, el: SVGGElement | null) => {
      if (el) nodeEls.current.set(id, el)
      else nodeEls.current.delete(id)
    }, [])

    // 拖拽/平移手势态（ref，不参与渲染）
    const dragRef = useRef<DragState | null>(null)
    const panRef = useRef<PanState | null>(null)

    /** 客户端坐标 → 世界坐标：世界组 CTM 逆变换（含 translate+scale） */
    const toWorld = useCallback((clientX: number, clientY: number): Vec => {
      const svg = svgRef.current
      const world = worldRef.current
      if (!svg || !world) return { x: 0, y: 0 }
      const ctm = world.getScreenCTM()
      if (!ctm) return { x: 0, y: 0 }
      const pt = svg.createSVGPoint()
      pt.x = clientX
      pt.y = clientY
      const wp = pt.matrixTransform(ctm.inverse())
      return { x: wp.x, y: wp.y }
    }, [])

    /** 把某节点的位置直写到它的 <g> transform（拖拽高频路径，不触发 React） */
    const applyNodeTransform = useCallback(
      (id: string) => {
        const p = posRef.current[id]
        nodeEls.current.get(id)?.setAttribute('transform', `translate(${p.x} ${p.y})`)
      },
      [posRef],
    )

    /** 适应视图：按当前节点 bbox 重算 scale + 居中平移 */
    const fitView = useCallback(() => {
      const container = containerRef.current
      if (!container) return
      const rect = container.getBoundingClientRect()
      const cw = rect.width
      const ch = rect.height
      const pos = posRef.current
      // 节点盒真实范围是 [p.x, p.x+NODE_W] × [p.y, p.y+NODE_H]（p 为左上角，
      // 见 layout.ts 列起始坐标 / TopologyNode rect 从 0,0 起）。此前按 p 为盒中心
      // 加减 NODE_W/2，bbox 相对整体左移半格，居中后内容整体偏右 NODE_W/2×scale。
      let minX = Infinity
      let minY = Infinity
      let maxX = -Infinity
      let maxY = -Infinity
      for (const id of Object.keys(pos)) {
        const p = pos[id]
        minX = Math.min(minX, p.x)
        maxX = Math.max(maxX, p.x + NODE_W)
        minY = Math.min(minY, p.y)
        maxY = Math.max(maxY, p.y + NODE_H)
      }
      if (!Number.isFinite(minX)) return
      const bw = maxX - minX
      const bh = maxY - minY
      // 留白因子：0.95 让内容占视口 95%（左右各 ~2.5%），比默认 0.85 的左右空隙明显更小；
      // 用户要求「画布左右空隙小一些」——调高此因子即可（0.85 → 0.95，均匀收四边）
      let scale = Math.min(cw / bw, ch / bh) * 0.95
      scale = clamp(scale, 0.4, 2.5)
      const x = (cw - bw * scale) / 2 - minX * scale
      const y = (ch - bh * scale) / 2 - minY * scale
      setView(clampView({ x, y, scale }, cw, ch))
    }, [posRef])

    /** 绕画布中心缩放 factor 倍（钳制 [0.4, 2.5]，与滚轮一致） */
    const zoomBy = useCallback((factor: number) => {
      const container = containerRef.current
      if (!container) return
      const r = container.getBoundingClientRect()
      const cx = r.width / 2
      const cy = r.height / 2
      setView((v) => {
        const scale = clamp(v.scale * factor, 0.4, 2.5)
        const wx = (cx - v.x) / v.scale
        const wy = (cy - v.y) / v.scale
        const next = { scale, x: v.x + wx * (v.scale - scale), y: v.y + wy * (v.scale - scale) }
        return clampView(next, r.width, r.height)
      })
    }, [])

    const zoomIn = useCallback(() => zoomBy(1.2), [zoomBy])
    const zoomOut = useCallback(() => zoomBy(1 / 1.2), [zoomBy])

    // 暴露 fitView / zoomIn / zoomOut 给父页工具栏
    useImperativeHandle(ref, () => ({ fitView, zoomIn, zoomOut }))

    // 挂载即适配（useLayoutEffect 在绘制前执行，避免首帧偏移闪烁）
    useLayoutEffect(() => {
      fitView()
    }, [fitView])

    /** 滚轮朝光标缩放：factor = e^-deltaY，钳制 [0.4, 2.5]，光标所在世界点保持屏幕位置不变 */
    const onWheel = useCallback(
      (e: WheelEvent) => {
        e.preventDefault()
        setView((v) => {
          const factor = Math.exp(-e.deltaY * 0.0012)
          const scale = clamp(v.scale * factor, 0.4, 2.5)
          const wx = (e.clientX - v.x) / v.scale
          const wy = (e.clientY - v.y) / v.scale
          const next = { scale, x: v.x + wx * (v.scale - scale), y: v.y + wy * (v.scale - scale) }
          const container = containerRef.current
          if (!container) return next
          const r = container.getBoundingClientRect()
          return clampView(next, r.width, r.height)
        })
      },
      [],
    )

    /** svg ref 回调：保存元素 + 挂非 passive wheel（React onWheel 无法 preventDefault） */
    const svgRefCallback = useCallback(
      (el: SVGSVGElement | null) => {
        if (svgRef.current) svgRef.current.removeEventListener('wheel', onWheel)
        svgRef.current = el
        if (el) el.addEventListener('wheel', onWheel, { passive: false })
      },
      [onWheel],
    )

    /** 拖拽移动：改 posRef + 直写节点 transform（全程无 React re-render） */
    const handleWindowMove = useCallback(
      (e: PointerEvent) => {
        const d = dragRef.current
        if (!d) return
        const dx = Math.abs(e.clientX - d.startClient.x)
        const dy = Math.abs(e.clientY - d.startClient.y)
        if (!d.moved && dx + dy < 3) return // 3px 阈值内视为点击，不移动
        d.moved = true
        const wp = toWorld(e.clientX, e.clientY)
        posRef.current[d.id] = { x: wp.x - d.grabOffset.x, y: wp.y - d.grabOffset.y }
        applyNodeTransform(d.id)
      },
      [toWorld, posRef, applyNodeTransform],
    )

    /** 拖拽结束：移 listener；<3px 视为点击 → 选中节点；无论何种都恢复模拟 */
    const handleWindowUp = useCallback(() => {
      const d = dragRef.current
      dragRef.current = null
      window.removeEventListener('pointermove', handleWindowMove)
      window.removeEventListener('pointerup', handleWindowUp)
      if (d && !d.moved) onSelect(d.id)
      endDrag()
    }, [handleWindowMove, onSelect, endDrag])

    /** 节点 pointerdown：启动拖拽手势（暂停模拟），并区分 click 与 drag */
    const handleNodePointerDown = useCallback(
      (e: ReactPointerEvent<SVGGElement>, id: string) => {
        e.stopPropagation()
        e.preventDefault()
        const wp = toWorld(e.clientX, e.clientY)
        const cur = posRef.current[id]
        dragRef.current = {
          id,
          startClient: { x: e.clientX, y: e.clientY },
          grabOffset: { x: wp.x - cur.x, y: wp.y - cur.y },
          moved: false,
        }
        beginDrag(id)
        onHover(null)
        window.addEventListener('pointermove', handleWindowMove)
        window.addEventListener('pointerup', handleWindowUp)
      },
      [toWorld, posRef, beginDrag, onHover, handleWindowMove, handleWindowUp],
    )

    /** 平移移动：按手势起始视图增量更新（闭包捕获 startView，无陈旧问题） */
    const handlePanMove = useCallback((e: PointerEvent) => {
      const p = panRef.current
      if (!p) return
      if (!p.moved) {
        const dx = Math.abs(e.clientX - p.startClient.x)
        const dy = Math.abs(e.clientY - p.startClient.y)
        if (dx + dy < 3) return
        p.moved = true
      }
      setView((v) => {
        const next = {
          ...v,
          x: p.startView.x + (e.clientX - p.startClient.x),
          y: p.startView.y + (e.clientY - p.startClient.y),
        }
        const container = containerRef.current
        if (!container) return next
        const r = container.getBoundingClientRect()
        return clampView(next, r.width, r.height)
      })
    }, [])

    /** 平移结束：<3px 视为背景点击 → 清空选中与悬浮 */
    const handlePanUp = useCallback(() => {
      const p = panRef.current
      panRef.current = null
      setPanning(false)
      window.removeEventListener('pointermove', handlePanMove)
      window.removeEventListener('pointerup', handlePanUp)
      if (p && !p.moved) {
        onSelect(null)
        onHover(null)
      }
    }, [handlePanMove, onSelect, onHover])

    /** 空白处 pointerdown：启动平移手势（<3px 视为背景点击 → 取消选中） */
    const handleBackgroundPointerDown = useCallback(
      (e: ReactPointerEvent<SVGSVGElement>) => {
        if (e.button !== 0) return
        setPanning(true)
        panRef.current = {
          startClient: { x: e.clientX, y: e.clientY },
          startView: view,
          moved: false,
        }
        window.addEventListener('pointermove', handlePanMove)
        window.addEventListener('pointerup', handlePanUp)
      },
      [view, handlePanMove, handlePanUp],
    )

    // 悬浮回调：拖拽期间忽略（避免拖动经过其它节点触发高亮闪烁）
    const handleHover = useCallback(
      (id: string | null) => {
        if (draggingId) return
        onHover(id)
      },
      [draggingId, onHover],
    )

    // 高亮集（节点不置灰的集合）：完整链路 = 根 + 所悬浮/选中节点所属顶层分支的整棵子树
    // （含旁支兄弟及其子孙）。例：hover deployment-web → App + Ingress(web) + Service(web) +
    // 两个 Deployment + 全部 pod，而 admin/worker 分支置灰。「顶层分支」= 祖先链中直接挂根的那棵子树。
    const activeId = hoveredId ?? selectedId
    const chainSet = useMemo(() => {
      if (!activeId) return null
      // 有向边（source 为父/上层）：收集祖先与子孙
      const parents = new Map<string, string[]>()
      const children = new Map<string, string[]>()
      for (const edge of graph.edges) {
        if (!children.has(edge.source)) children.set(edge.source, [])
        children.get(edge.source)!.push(edge.target)
        if (!parents.has(edge.target)) parents.set(edge.target, [])
        parents.get(edge.target)!.push(edge.source)
      }
      // 祖先链：从当前节点沿 parent 一路向上到根
      const ancestors = new Set<string>()
      const upStack = [activeId]
      while (upStack.length) {
        const cur = upStack.pop()!
        if (ancestors.has(cur)) continue
        ancestors.add(cur)
        for (const p of parents.get(cur) ?? []) upStack.push(p)
      }
      // 是根 → 整棵树；否则取祖先中直接挂根的那些（顶层分支根），展开它整棵子树 + 根
      if (activeId === rootId) return ancestors
      const branchRoots = [...ancestors].filter((a) => a !== rootId && (parents.get(a)?.includes(rootId) ?? false))
      const set = new Set<string>([rootId])
      const downStack = branchRoots.length ? branchRoots : [activeId]
      while (downStack.length) {
        const cur = downStack.pop()!
        if (set.has(cur)) continue
        set.add(cur)
        for (const c of children.get(cur) ?? []) downStack.push(c)
      }
      return set
    }, [activeId, graph.edges, rootId])

    // 生效高亮集：hover 与选中共用同一套完整链路
    const activeSet = chainSet

    const queryActive = matchedIds !== null && matchedIds.size > 0

    // 边路由：由当前节点位置实时算正交折线（拖拽提交后 positions 变化 → 边跟着重排）
    const routes = useMemo(() => routeEdges(graph.edges, positions), [graph.edges, positions])

    return (
      <div
        ref={containerRef}
        className="relative h-full w-full overflow-hidden rounded-lg border border-line bg-surface"
      >
        <svg
          ref={svgRefCallback}
          className={cn('h-full w-full select-none', panning ? 'cursor-grabbing' : 'cursor-grab')}
          onPointerDown={handleBackgroundPointerDown}
        >
          <defs>
            {/* 画布点阵：浅色网格点缀，营造「拓扑画布」质感（随主题 line 色） */}
            <pattern id={gridId} width={26} height={26} patternUnits="userSpaceOnUse">
              <circle cx={1.5} cy={1.5} r={1.2} fill="var(--line)" opacity={0.55} />
            </pattern>
            {/* 边箭头：orient=auto 自动对齐末段方向（进目标即指向盒内）；常态线色 / 高亮主色两套 */}
            <marker id="topo-arrow-line" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6.5" markerHeight="6.5" orient="auto-start-reverse">
              <path d="M0 0 L10 5 L0 10 z" fill="var(--line)" />
            </marker>
            <marker id="topo-arrow-primary" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6.5" markerHeight="6.5" orient="auto-start-reverse">
              <path d="M0 0 L10 5 L0 10 z" fill="var(--primary)" />
            </marker>
          </defs>
          {/* 点阵层：屏幕坐标（不受世界变换影响） */}
          <rect width="100%" height="100%" fill={`url(#${gridId})`} pointerEvents="none" />
          {/* 世界组：统一应用 view 平移缩放 */}
          <g ref={worldRef} transform={`translate(${view.x} ${view.y}) scale(${view.scale})`}>
            {routes.map((route) => {
              // 链路内两端都在高亮集 → 点亮；其余置灰
              const onChain = activeSet ? activeSet.has(route.source) && activeSet.has(route.target) : false
              const edgeActive = onChain
              const edgeDimmed =
                (activeSet ? !edgeActive : false) ||
                (queryActive && !matchedIds.has(route.source) && !matchedIds.has(route.target))
              return (
                <TopologyEdge key={route.id} route={route} active={edgeActive} dimmed={edgeDimmed} />
              )
            })}
            {graph.nodes.map((node) => {
              const p = positions[node.id]
              const nodeDimmed =
                (activeSet ? !activeSet.has(node.id) : false) ||
                (queryActive && !matchedIds.has(node.id))
              return (
                <TopologyNode
                  key={node.id}
                  node={node}
                  x={p.x}
                  y={p.y}
                  highlight={
                    selectedId === node.id
                      ? 'selected'
                      : hoveredId === node.id
                        ? 'hover'
                        : chainSet?.has(node.id)
                          ? 'chain'
                          : null
                  }
                  dimmed={nodeDimmed}
                  matched={queryActive && matchedIds.has(node.id)}
                  isDragging={draggingId === node.id}
                  isRoot={node.id === rootId}
                  onPointerDown={handleNodePointerDown}
                  onHover={handleHover}
                  refCallback={(el) => registerNodeEl(node.id, el)}
                />
              )
            })}
          </g>
        </svg>
      </div>
    )
  },
)
