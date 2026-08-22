import { useEffect, useId, useRef, useState, type MouseEvent as ReactMouseEvent } from 'react'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/shadcn/tooltip'
import { nextZIndex } from '@/lib/zIndex'
import { Icon } from '@/components/Icons'

/**
 * 零依赖 SVG 迷你面积图：标题 + 当前值 + 平滑面积折线。
 * 输入为数值序列，自动 min/max 归一化；颜色用 CSS 变量（如 var(--primary)）随主题联动。
 * - 新数据进入：800ms easeInOutCubic 曲线 morph——环形缓冲满员时旧曲线向左错位滑动一 slot、
 *   新点从右侧滑入（flow 观感而非原地形变），新旧各自独立归一化只插值 y；
 *   渲染阶段即对齐 morph 起点，无闪帧
 * - hover：十字线对齐最近采样点，显示该点数值（labels 传并行人类可读值）；
 *   气泡水平夹紧在容器内，接近左右缘不越界被裁切
 * 用作 CPU/内存等时序指标的轻量图表，避免引入重型图表库。
 */
export function AreaSpark({
  label,
  value,
  points,
  labels,
  color,
  height = 24,
  hint,
}: {
  label: string
  value: string
  points: number[]
  /** 并行的人类可读采样值（如 "400m"），hover 时展示；缺省回退数值本身 */
  labels?: string[]
  color: string
  height?: number
  /** 标签旁的小问号提示（如 CPU 的 1 CPU = 1000m 单位说明），hover 展示 */
  hint?: string
}) {
  const id = useId()
  // morph 起点曲线：新数据进入时保持旧曲线，随 morphT 平滑变形到当前 points
  const morphFromRef = useRef<number[]>([])
  const prevPointsRef = useRef(points)
  const [morphT, setMorphT] = useState(1)
  // 数据真变化时由 derived-state 置位，effect 依据它启动播放——
  // 不能读 morphT 判断：渲染期 setMorphT(0) 的 retry 提交前，effect 拿到的是陈旧 morphT=1，
  // 若用 morphT>=1 提前返回会把 morphFromRef 清成新数据，导致 morph 根本不播（曲线瞬跳终态）
  const needMorphRef = useRef(false)
  // hover 对齐到的采样点下标（数据点，非渲染点）
  const [hover, setHover] = useState<number | null>(null)
  // tooltip 水平偏移：接近容器左右缘时把气泡拉回容器内，避免被裁切或盖住
  const [tipShift, setTipShift] = useState(0)
  const tipRef = useRef<HTMLDivElement>(null)
  // 单位提示 tooltip：portal 挂在 body，须盖过可拖拽宿主弹窗的动态 z-index（z-51+），
  // 每次打开时通过共享计数器拿一个新 z（复用强杀确认框同款机制）
  const [hintOpen, setHintOpen] = useState(false)
  const [hintZ, setHintZ] = useState(() => nextZIndex())

  const W = 100
  const H = 40
  const PAD = 2
  /** 固定屏幕空间采样密度：跨长度 morph 的关键——在固定 x 上对新旧曲线各自插值再混合 */
  const RENDER_POINTS = 60
  const n = points.length

  // 数据变化：渲染阶段立即把 morphT 置 0（呈现旧曲线形态），避免先闪一帧新曲线再弹回旧的。
  // 官方 derived-state 模式——用 ref 记录上次 points 引用，变化时同步 morph 起点。
  if (prevPointsRef.current !== points) {
    const from = prevPointsRef.current
    const same = from.length === points.length && from.every((v, i) => v === points[i])
    if (!same) {
      morphFromRef.current = from
      setMorphT(0)
      needMorphRef.current = true
    }
    prevPointsRef.current = points
  }

  // morph 播放：morphT 0→1（easeInOutCubic），800ms——数据向左流动、新点从右侧滑入。
  // 起点到终点先慢后快再慢，避免生硬的匀加速；deps 只挂 points，needMorphRef 兜底保证必播。
  useEffect(() => {
    if (!needMorphRef.current) return
    needMorphRef.current = false
    const start = performance.now()
    const dur = 800
    let raf = 0
    const step = (now: number) => {
      const p = Math.min(1, (now - start) / dur)
      // easeInOutCubic
      setMorphT(p < 0.5 ? 4 * p * p * p : 1 - Math.pow(-2 * p + 2, 3) / 2)
      if (p >= 1) morphFromRef.current = points
      if (p < 1) raf = requestAnimationFrame(step)
    }
    raf = requestAnimationFrame(step)
    return () => cancelAnimationFrame(raf)
  }, [points])

  // tooltip 水平夹紧：越界时计算偏移量，让气泡始终落在图表容器内（右侧缘不再跑出卡片）
  useEffect(() => {
    const tip = tipRef.current
    const host = tip?.parentElement
    if (!tip || !host || hover === null) return
    const hostW = host.getBoundingClientRect().width
    const tipW = tip.getBoundingClientRect().width
    const pct = n <= 1 ? 50 : (hover / (n - 1)) * 100
    const leftPx = (pct / 100) * hostW - tipW / 2
    const clamped = Math.min(Math.max(leftPx, 2), Math.max(hostW - tipW - 2, 2))
    setTipShift(clamped - leftPx)
  }, [hover, n, labels])

  /** 在数值序列上按小数下标线性插值（新旧曲线长度不同也能按比例对齐取值） */
  const evalY = (arr: number[], idx: number): number | null => {
    if (arr.length === 0) return null
    if (arr.length === 1) return arr[0]
    const i = Math.min(arr.length - 1, Math.max(0, idx))
    const i0 = Math.floor(i)
    const i1 = Math.min(arr.length - 1, i0 + 1)
    return arr[i0] + (arr[i1] - arr[i0]) * (i - i0)
  }

  let line = ''
  let area = ''
  if (n > 0) {
    const to = points
    const from = morphFromRef.current.length ? morphFromRef.current : to
    // 渲染期兜底：数据刚变（needMorphRef 待消费）但 morphT 还是旧值 1 的窗口
    // （render-phase retry 或首个 rAF 生效前的那一帧），强制按起点曲线渲染，杜绝新曲线闪现闪烁
    const t = needMorphRef.current && morphT >= 1 ? 0 : morphT
    // 新旧各自独立归一化（只插值 y）：中途 min/max 变化不拉扯整条线的垂直尺度
    const maxF = Math.max(...from)
    const minF = Math.min(...from)
    const maxT = Math.max(...to)
    const minT = Math.min(...to)
    const rangeF = maxF - minF || 1
    const rangeT = maxT - minT || 1
    const xs = (i: number) => (RENDER_POINTS <= 1 ? W / 2 : (i / (RENDER_POINTS - 1)) * W)
    // 环形缓冲满员时（长度恒定）每采样一点整体左移一格 → 旧曲线错位左移一 slot、
    // 新点从右侧滑入；同位对齐会让整条线原地"喘气"，错位滑动才是"流动"观感。
    // 长度变化（首次填充）时 slot=0 退化为同位插值。
    const slot = from.length === to.length && from.length > 1 ? 1 / (from.length - 1) : 0
    const ys: number[] = []
    for (let i = 0; i < RENDER_POINTS; i++) {
      const frac = (i / (RENDER_POINTS - 1)) * (to.length - 1)
      // t=0 从旧曲线原位出发（无瞬跳），随 t 向左滑动一 slot → 新点从右侧滑入
      const fracF = frac + t * slot
      const yF = H - PAD - ((evalY(from, fracF)! - minF) / rangeF) * (H - PAD * 2)
      const yT = H - PAD - ((evalY(to, frac)! - minT) / rangeT) * (H - PAD * 2)
      ys.push(yF + (yT - yF) * t)
    }
    // Catmull-Rom → 三次贝塞尔：采样点之间以平滑曲线连接（recharts monotone 同款思路），
    // 避免直线折线显得生硬。
    const pts = ys.map((sy, i) => ({ x: xs(i), y: sy }))
    let d = `M${pts[0].x.toFixed(2)},${pts[0].y.toFixed(2)}`
    for (let i = 0; i < pts.length - 1; i++) {
      const p0 = pts[i - 1] ?? pts[i]
      const p1 = pts[i]
      const p2 = pts[i + 1]
      const p3 = pts[i + 2] ?? p2
      const c1x = p1.x + (p2.x - p0.x) / 6
      const c1y = p1.y + (p2.y - p0.y) / 6
      const c2x = p2.x - (p3.x - p1.x) / 6
      const c2y = p2.y - (p3.y - p1.y) / 6
      d += ` C${c1x.toFixed(2)},${c1y.toFixed(2)} ${c2x.toFixed(2)},${c2y.toFixed(2)} ${p2.x.toFixed(2)},${p2.y.toFixed(2)}`
    }
    line = d
    area = `${line} L${xs(RENDER_POINTS - 1).toFixed(2)},${H} L${xs(0).toFixed(2)},${H} Z`
  }

  /** hover 对齐：DOM 横向位置 → 数据下标（与渲染采样无关） */
  const pctX = (i: number) => (n <= 1 ? 50 : (i / (n - 1)) * 100)
  const onMove = (e: ReactMouseEvent<HTMLDivElement>) => {
    if (n === 0) return
    const rect = e.currentTarget.getBoundingClientRect()
    const ratio = (e.clientX - rect.left) / (rect.width || 1)
    setHover(Math.min(n - 1, Math.max(0, Math.round(ratio * (n - 1)))))
  }

  return (
    <div className="flex h-6 min-w-0 items-center px-1">
      <div className="flex w-full flex-col">
        <div className="flex items-baseline justify-between gap-1 leading-none">
          <span className="flex min-w-0 items-center gap-1 text-[10px] leading-none text-faint">
            <span className="truncate">{label}</span>
            {hint && (
              <TooltipProvider>
                <Tooltip
                  open={hintOpen}
                  onOpenChange={(o) => {
                    if (o) setHintZ(nextZIndex())
                    setHintOpen(o)
                  }}
                >
                  <TooltipTrigger asChild>
                    <span className="inline-flex cursor-help items-center text-faint/70 hover:text-faint" tabIndex={-1}>
                      <Icon name="info" className="size-3" aria-hidden />
                    </span>
                  </TooltipTrigger>
                  <TooltipContent sideOffset={4} className="max-w-[220px] text-center" style={{ zIndex: hintZ }}>
                    {hint}
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            )}
          </span>
          <span className="font-mono text-[11px] font-semibold leading-none tabular-nums" style={{ color }}>
            {value}
          </span>
        </div>
        <div
          className="relative mt-0.5 cursor-crosshair"
          style={{ height }}
          onMouseMove={onMove}
          onMouseLeave={() => setHover(null)}
        >
          <svg viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none" className="h-full w-full" aria-hidden>
            <defs>
              <linearGradient id={id} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={color} stopOpacity="0.3" />
                <stop offset="100%" stopColor={color} stopOpacity="0" />
              </linearGradient>
            </defs>
            {area && <path d={area} fill={`url(#${id})`} />}
            {line && (
              <path
                d={line}
                fill="none"
                stroke={color}
                strokeWidth={1.5}
                strokeLinejoin="round"
                strokeLinecap="round"
                vectorEffect="non-scaling-stroke"
              />
            )}
          </svg>
          {/* hover 十字线 + 对应采样点数值 */}
          {hover !== null && n > 0 && (
            <>
              <div
                className="pointer-events-none absolute inset-y-0 -translate-x-1/2"
                style={{ left: `${pctX(hover)}%`, width: 1, backgroundColor: color, opacity: 0.4 }}
              />
              <div
                ref={tipRef}
                className="pointer-events-none absolute z-50 whitespace-nowrap rounded bg-black/85 px-1.5 py-0.5 font-mono text-[10px] text-white"
                style={{
                  left: `${pctX(hover)}%`,
                  top: -2,
                  transform: `translateX(calc(-50% + ${tipShift}px)) translateY(-100%)`,
                }}
              >
                {labels?.[hover] ?? String(points[hover])}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
