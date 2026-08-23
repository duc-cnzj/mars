import { memo, type PointerEvent as ReactPointerEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { IconPaths } from '@/components/Icons'
import type { Tone } from '@/components/ui/Tag'
import { NODE_H, NODE_W } from './layout'
import { KIND_ICON, STATUS_TONE } from './mockTopology'
import type { PodLifecycle, TopoNode } from './topologyTypes'

export interface TopologyNodeProps {
  node: TopoNode
  /** 世界坐标（节点 <g> 的平移量，左上角） */
  x: number
  y: number
  /** 高亮类型：hover 悬浮 / selected 选中 / chain 链路成员（决定描边样式） */
  highlight: 'hover' | 'selected' | 'chain' | null
  /** 非邻居节点置灰 */
  dimmed: boolean
  /** 搜索命中（外圈虚线 ring） */
  matched: boolean
  /** 拖拽中（改变 cursor 样式） */
  isDragging: boolean
  /** Application 根节点：加大图标 + 强调色（对齐 Argo 根节点视觉） */
  isRoot: boolean
  /** 节点 pointerdown：启动拖拽手势（graph 层负责坐标换算与 DOM 直改） */
  onPointerDown: (e: ReactPointerEvent<SVGGElement>, id: string) => void
  /** 悬浮进入/离开：传 node.id / null */
  onHover: (id: string | null) => void
  /** <g> 元素注册（graph 层拖拽期间直接 setAttribute 改 transform） */
  refCallback: (el: SVGGElement | null) => void
}

/** tone → 状态点填充类（随主题语义色联动） */
const TONE_FILL: Record<Tone, string> = {
  ok: 'fill-ok',
  warn: 'fill-warn',
  err: 'fill-err',
  info: 'fill-info',
  accent: 'fill-primary',
  mute: 'fill-faint',
}

/** Pod 生命周期 → 状态标签底色（与 TabLog PodStateTag 完全一致，刻意不随主题，对齐状态色硬编码约定） */
const POD_LIFECYCLE_FILL: Record<PodLifecycle, string> = {
  ready: '#a78bfa', // 就绪
  notReady: '#93c5fd', // 未就绪
  pending: '#67e8f9', // 启动中
  terminating: '#fca5a5', // 停止中
  isOld: '#fde047', // 即将停止
}

/** 生命周期胶囊文字/转圈色：近黑深字 + 加粗（浅色 pastel 底配深字对比足够；不要白晕/描边，用户定稿）。
 *  刻意硬编码不随主题 —— 胶囊底是硬编码 pastel，亮暗主题同色 */
const LIFECYCLE_TEXT = '#0f172a'

/** Pod 生命周期 → 节点标签文案（复用图例短文案；isOld 用紧凑版 podOld，避免 en 的 "About to stop" 撑爆 pill） */
const LIFECYCLE_LABEL_KEY = {
  ready: 'topology.legendReady',
  notReady: 'topology.legendNotReady',
  pending: 'topology.legendStarting',
  terminating: 'topology.legendStopping',
  isOld: 'topology.podOld',
} as const

/** 节点状态标签字号 */
const LABEL_FONT = 11

/** 估算 SVG 文本宽度（CJK 整宽 / ASCII 0.62×字号），用于 pill 自适应宽度 */
function textWidthPx(s: string, fs: number): number {
  let w = 0
  for (const ch of s) {
    const code = ch.codePointAt(0) ?? 0
    w += code > 0x2e80 ? fs : fs * 0.62
  }
  return w
}

/** 节点名可用宽度估算：CJK 整宽 / ASCII 0.5×字号。0.62 是状态胶囊的保守估（胶囊内留白友好）；
 *  名称用 0.5 更贴 13px 系统字体实际字宽，能吃掉更多可用宽度再截断（用户：省略的文字再少一点） */
function nameWidthPx(s: string, fs: number): number {
  let w = 0
  for (const ch of s) {
    const code = ch.codePointAt(0) ?? 0
    w += code > 0x2e80 ? fs : fs * 0.5
  }
  return w
}

/** 名称按像素宽度截断：maxW 内放尽可能多的字符，保留头 + 省略号 + 末 4 位 */
function fitName(name: string, maxW: number, fs: number): string {
  if (nameWidthPx(name, fs) <= maxW) return name
  const tail = name.slice(-4)
  const ell = '…'
  const tailW = nameWidthPx(ell + tail, fs)
  let head = name.length - 4
  while (head > 0 && tailW + nameWidthPx(name.slice(0, head), fs) > maxW) head--
  return head > 0 ? `${name.slice(0, head)}${ell}${tail}` : tail
}

/** 节点健康状态 → 状态标签文案（复用详情面板 statusKey 同一组词条） */
const STATUS_LABEL_KEY = {
  healthy: 'topology.statusHealthy',
  degraded: 'topology.statusDegraded',
  progressing: 'topology.statusProgressing',
  unknown: 'topology.statusUnknown',
} as const

/** 语义 tone → 状态标签文字色（对齐 Tag 的 text-* 组合） */
const TONE_TEXT_CLASS: Record<Tone, string> = {
  ok: 'text-ok',
  warn: 'text-warn',
  err: 'text-err',
  info: 'text-info',
  accent: 'text-primary',
  mute: 'text-mute',
}

/** 语义 tone → 状态标签底（对齐 Tag 的 *-soft 组合；mute 用 raised 免浅底白字） */
const TONE_PILL_FILL: Record<Tone, string> = {
  ok: 'fill-ok-soft',
  warn: 'fill-warn-soft',
  err: 'fill-err-soft',
  info: 'fill-info-soft',
  accent: 'fill-primary-soft',
  mute: 'fill-raised',
}

/**
 * 拓扑节点（Argo 资源树盒样式）：282×52 圆角盒，左列 kind 图标 + 下方 kind 标签，
 * 右侧名称 + 健康状态点。memo 组件：拖拽期间 props 不变（位置由画布层直接改 DOM
 * transform），React 完全跳过重渲。文本/图标/状态点 pointer-events-none，命中只有矩形。
 * 悬浮置灰（内层 <g>，150ms 快速过渡）与拖拽 cursor 由外层 <g> 承载。
 */
export const TopologyNode = memo(function TopologyNode({
  node,
  x,
  y,
  highlight,
  dimmed,
  matched,
  isDragging,
  isRoot,
  onPointerDown,
  onHover,
  refCallback,
}: TopologyNodeProps) {
  const tone = STATUS_TONE[node.status]
  const { t } = useTranslation()
  // Pod 生命周期 → 节点状态标签（替代状态点；仅直播 Tab 的 Pod 带 lifecycle，其余走原健康色点）
  const lifecycle = node.kind === 'Pod' ? node.lifecycle : undefined
  const lifecycleLabel = lifecycle ? t(LIFECYCLE_LABEL_KEY[lifecycle]) : undefined
  const lifecycleHex = lifecycle ? POD_LIFECYCLE_FILL[lifecycle] : undefined
  const lifecycleTransitioning = lifecycle === 'pending' || lifecycle === 'terminating' || lifecycle === 'isOld'
  const lifecyclePillW = lifecycleLabel
    ? textWidthPx(lifecycleLabel, LABEL_FONT) + (lifecycleTransitioning ? 12 : 0) + 14
    : 0
  // 节点健康状态 → 状态标签（Application 根：聚合健康度，主题 tone 色软底胶囊，对齐 Tag 的 text-* + *-soft）
  const appLabel = node.kind === 'Application' ? t(STATUS_LABEL_KEY[node.status]) : undefined
  const appTransitioning = node.status === 'progressing'
  const appPillW = appLabel
    ? textWidthPx(appLabel, LABEL_FONT) + (appTransitioning ? 12 : 0) + 14
    : 0
  const strokeClass =
    highlight === 'selected' || highlight === 'hover'
      ? 'stroke-primary'
      : highlight === 'chain'
        ? 'stroke-primary/60'
        : isRoot
          ? 'stroke-primary/40'
          : 'stroke-line'
  const strokeWidth = highlight === 'selected' ? 2 : highlight === 'hover' ? 1.6 : 1.2
  const fillClass = highlight === 'selected' ? 'fill-raised' : 'fill-surface'
  const cursorClass = isDragging ? 'cursor-grabbing' : 'cursor-grab'
  // 名称可用宽度 = 右侧状态指示区左缘 - 名称起点 - 间距；按像素截断（比固定 22/26 字符更贴实际：
  // 短状态胶囊/纯状态点时名称吃掉更多宽度，省略更少）。nameTruncated 驱动全名 hover 提示条
  const nameStartX = 62
  const indicatorLeft = lifecycle
    ? NODE_W - 16 - lifecyclePillW
    : node.kind === 'Application'
      ? NODE_W - 16 - appPillW
      : node.kind === 'Pod'
        ? NODE_W - 30 // 无 lifecycle 的 Pod：右缘健康状态点（圆心 NODE_W-18、r=5，留 7px 边距）
        : NODE_W - 16 // 其余 kind（Service/Ingress/Deployment…）无右侧指示区，名称可用满右缘留白
  const displayName = fitName(node.name, indicatorLeft - nameStartX - 6, 13)
  const nameTruncated = displayName !== node.name
  // 内层置灰：悬浮高亮用短过渡（跟手）
  const dimmedClass = dimmed ? 'opacity-20' : 'opacity-100'

  return (
    <g
      ref={refCallback}
      transform={`translate(${x} ${y})`}
      className={cursorClass}
      style={{ touchAction: 'none' }}
      onPointerDown={(e) => onPointerDown(e, node.id)}
      onMouseEnter={() => onHover(node.id)}
      onMouseLeave={() => onHover(null)}
    >
      {/* 内容层：悬浮置灰（150ms 快速过渡） */}
      <g className={`transition-opacity duration-150 ${dimmedClass}`}>
        {/* 搜索命中 ring：外圈虚线，标记高亮 */}
        {matched && (
          <rect
            x={-5}
            y={-5}
            width={NODE_W + 10}
            height={NODE_H + 10}
            rx={10}
            fill="none"
            strokeDasharray="5 4"
            strokeWidth={1.5}
            className="stroke-primary"
            pointerEvents="none"
          />
        )}
        {/* 主体框：表面填充 + 按高亮/根节点切换描边；是唯一的命中区（其余子元素 pointer-events-none） */}
        <rect
          width={NODE_W}
          height={NODE_H}
          rx={6}
          className={`${fillClass} ${strokeClass} shadow-sm`}
          strokeWidth={strokeWidth}
        />
        {/* kind 图标：左列（对齐 Argo 60px 图标区），根节点放大并强调色 */}
        <g
          transform={`translate(30 ${isRoot ? 10 : 13}) scale(${isRoot ? 1.05 : 0.8})`}
          className={isRoot ? 'text-primary' : 'text-faint'}
          pointerEvents="none"
        >
          <IconPaths name={KIND_ICON[node.kind]} />
        </g>
        {/* kind 标签：图标下方（对齐 Argo __node-kind） */}
        <text
          x={30}
          y={46}
          fontSize={9}
          letterSpacing="0.4"
          textAnchor="middle"
          fill="currentColor"
          className="text-faint uppercase"
          pointerEvents="none"
        >
          {node.kind}
        </text>
        {/* 名称：截断（保留头尾…末4位），完整名在详情面板展示；被省略时 hover 节点在名称上方
            弹 SVG 提示条显示全名（复用已有 hover 高亮信号，不依赖文字自身命中；随 pan/zoom 走） */}
        <text
          x={62}
          y={24}
          fontSize={13}
          fontWeight={500}
          fill="currentColor"
          className={isRoot ? 'text-primary font-semibold' : 'text-ink'}
          pointerEvents="none"
        >
          {displayName}
        </text>
        {/* 全名提示条：节点名被省略且 hover 节点时展示（SVG 内绘制，主题 surface 底 + 边框，
            盖在节点上方；native title 在该 SVG 场景下不可靠，用户反馈无效，改用此方案） */}
        {nameTruncated && highlight === 'hover' && (
          <g transform={`translate(62 ${-34})`} pointerEvents="none">
            <rect
              x={0}
              y={-10}
              width={textWidthPx(node.name, 12) + 14}
              height={22}
              rx={6}
              fill="var(--overlay)"
              stroke="var(--border)"
              strokeWidth={1}
            />
            <text x={7} y={4} fontSize={12} fill="var(--text)">
              {node.name}
            </text>
          </g>
        )}
        {/* 状态指示：只在 Pod（叶子运行单元）和 Application 根（状态由 pod 聚合）上显示。
            直播 Tab 的 Pod 带 lifecycle → 生命周期胶囊（PodStateTag 底色 + 白字，过渡态前导转圈）；
            Application 根 → 聚合健康度胶囊（主题 tone 软底 + tone 文字，进行中前导转圈）；
            无 lifecycle 的 Pod（demo / 容器接口缺该 pod）走健康色点（进行中脉冲、其余呼吸） */}
        {(node.kind === 'Pod' || node.kind === 'Application') &&
          (lifecycle ? (
            // 生命周期胶囊：右缘贴 NODE_W-16、垂直居中于原状态点位置；宽度按文本估算自适应。
            // 声呐椭圆先画（垫在胶囊底下）；CSS transform 会覆盖 SVG transform 属性，定位 translate
            // 必须留在外层 g，动画类放椭圆自身（transform-box: fill-box 绕自身中心缩放）
            <g transform={`translate(${NODE_W - 16} ${NODE_H - 16})`} pointerEvents="none">
              {/* 声呐环：与胶囊同几何的圆角矩形（rx=10，非椭圆），外扩淡出；全图同周期同相位一起扫 */}
              <rect
                x={-lifecyclePillW}
                y={-10}
                width={lifecyclePillW}
                height={20}
                rx={10}
                fill="none"
                strokeWidth={2.5}
                stroke={lifecycleHex}
                className="topo-pill-ping"
              />
              <rect
                x={-lifecyclePillW}
                y={-10}
                width={lifecyclePillW}
                height={20}
                rx={10}
                fill={lifecycleHex}
              />
              {lifecycleTransitioning && (
                // 过渡态：pill 内前导小转圈（启动中青 / 停止中红 / 即将停止黄），绕圆心旋转
                <g className="animate-spin" style={{ transformOrigin: `${-lifecyclePillW + 12}px 0px` }}>
                  <circle
                    cx={-lifecyclePillW + 12}
                    cy={0}
                    r={4}
                    fill="none"
                    stroke={LIFECYCLE_TEXT}
                    strokeWidth={2}
                    strokeLinecap="round"
                    strokeDasharray="7 6"
                  />
                </g>
              )}
              <text
                x={-lifecyclePillW + 7 + (lifecycleTransitioning ? 12 : 0)}
                y={4}
                fontSize={LABEL_FONT}
                fontWeight={600}
                fill={LIFECYCLE_TEXT}
              >
                {lifecycleLabel}
              </text>
            </g>
          ) : node.kind === 'Application' ? (
            // 应用状态胶囊：tone 文字 + *-soft 软底（对齐 Tag 组合），currentColor 让文字/转圈/声呐同色；
            // 进行中（progressing）前导转圈，其余稳态纯文本
            <g
              transform={`translate(${NODE_W - 16} ${NODE_H - 16})`}
              className={TONE_TEXT_CLASS[tone]}
              pointerEvents="none"
            >
              {/* 声呐环：currentColor 跟随 tone 文字色；圆角矩形与胶囊同几何，全图同节拍一起外扩 */}
              <rect
                x={-appPillW}
                y={-10}
                width={appPillW}
                height={20}
                rx={10}
                fill="none"
                strokeWidth={2.5}
                stroke="currentColor"
                className="topo-pill-ping"
              />
              <rect
                className={TONE_PILL_FILL[tone]}
                x={-appPillW}
                y={-10}
                width={appPillW}
                height={20}
                rx={10}
              />
              {appTransitioning && (
                <g className="animate-spin" style={{ transformOrigin: `${-appPillW + 12}px 0px` }}>
                  <circle
                    cx={-appPillW + 12}
                    cy={0}
                    r={4}
                    fill="none"
                    stroke="currentColor"
                    strokeWidth={2}
                    strokeLinecap="round"
                    strokeDasharray="7 6"
                  />
                </g>
              )}
              <text
                x={-appPillW + 7 + (appTransitioning ? 12 : 0)}
                y={4}
                fontSize={LABEL_FONT}
                fill="currentColor"
              >
                {appLabel}
              </text>
            </g>
          ) : (
            <circle
              cx={NODE_W - 18}
              cy={NODE_H - 16}
              r={5}
              className={`${TONE_FILL[tone]} ${node.status === 'progressing' ? 'topo-dot-pulse' : 'topo-dot-breathe'}`}
              pointerEvents="none"
            />
          ))}
      </g>
    </g>
  )
})
