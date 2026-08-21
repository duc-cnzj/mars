import type { ReactNode, SVGProps } from 'react'

/** 1.6px 描边线性图标基底，随 currentColor 着色 */
function I({ children, ...rest }: SVGProps<SVGSVGElement> & { children: ReactNode }) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="1em"
      height="1em"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.6}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
      {...rest}
    >
      {children}
    </svg>
  )
}

export type IconName =
  | 'grid'
  | 'cluster'
  | 'namespace'
  | 'project'
  | 'rocket'
  | 'boxes'
  | 'terminal'
  | 'pulse'
  | 'repo'
  | 'key'
  | 'link'
  | 'gear'
  | 'search'
  | 'bell'
  | 'chevron-down'
  | 'chevron-right'
  | 'chevron-left'
  | 'plus'
  | 'filter'
  | 'refresh'
  | 'logs'
  | 'cpu'
  | 'memory'
  | 'network'
  | 'database'
  | 'check'
  | 'moon'
  | 'sun'
  | 'collapse'
  | 'expand'
  | 'more'
  | 'play'
  | 'pause'
  | 'cube'
  | 'close'
  | 'crown'
  | 'copy'
  | 'power'
  | 'star'
  | 'loader'
  | 'clock'
  | 'alert'
  | 'external'
  | 'check-circle-fill'
  | 'close-circle-fill'
  | 'warning-fill'
  | 'user'

const paths: Record<IconName, ReactNode> = {
  grid: (
    <>
      <rect x="3" y="3" width="7" height="7" rx="1.5" />
      <rect x="14" y="3" width="7" height="7" rx="1.5" />
      <rect x="3" y="14" width="7" height="7" rx="1.5" />
      <rect x="14" y="14" width="7" height="7" rx="1.5" />
    </>
  ),
  cluster: (
    <>
      <rect x="3" y="4" width="18" height="6" rx="1.5" />
      <rect x="3" y="14" width="18" height="6" rx="1.5" />
      <circle cx="7" cy="7" r="0.9" fill="currentColor" stroke="none" />
      <circle cx="7" cy="17" r="0.9" fill="currentColor" stroke="none" />
      <path d="M12 10v4" />
    </>
  ),
  namespace: (
    <>
      <path d="m12 3 8 4.5v9L12 21l-8-4.5v-9L12 3Z" />
      <path d="M12 12 4 7.5M12 12l8-4.5M12 12v9" />
    </>
  ),
  project: (
    <>
      <path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7Z" />
      <path d="M3 12h18" />
    </>
  ),
  rocket: (
    <>
      <path d="M12 3c3 1.5 5 4 5 8l-2 4-3 2-3-2-2-4c0-4 2-6.5 5-8Z" />
      <path d="M9 14 4.5 18l2 2L10 15.5" />
      <path d="M15 14l4.5 4-2 2L14 15.5" />
      <circle cx="12" cy="9" r="1.6" />
    </>
  ),
  boxes: (
    <>
      <path d="m3 8 4-2 4 2v5l-4 2-4-2V8Z" />
      <path d="m11 8 4-2 4 2v5l-4 2-4-2V8Z" />
      <path d="M7 14v4l4 2 4-2v-4" />
    </>
  ),
  terminal: (
    <>
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="m7 9 3 3-3 3M12.5 15H17" />
    </>
  ),
  pulse: (
    <>
      <path d="M3 12h4l2-6 4 12 2-6h6" />
    </>
  ),
  repo: (
    <>
      <circle cx="6" cy="6" r="2" />
      <circle cx="6" cy="18" r="2" />
      <circle cx="18" cy="9" r="2" />
      <path d="M6 8v8M18 11l-8.5 5.5M17.9 7.1 8.8 5" />
    </>
  ),
  key: (
    <>
      <circle cx="8" cy="15" r="4.5" />
      <path d="m11.2 11.8 8.3-8.3M16.5 6.5 19 9M13.5 9.5 15.5 11.5" />
    </>
  ),
  // lucide Link 同款几何：与外层 ProjectRow URL 图标对齐（旧版 TabInfo 地址 section 用 LinkOutlined）
  link: (
    <>
      <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
      <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
    </>
  ),
  gear: (
    <>
      <circle cx="12" cy="12" r="3" />
      <path d="M19.4 15a1.7 1.7 0 0 0 .34 1.87l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.7 1.7 0 0 0-1.87-.34 1.7 1.7 0 0 0-1 1.55V21a2 2 0 1 1-4 0v-.09a1.7 1.7 0 0 0-1-1.55 1.7 1.7 0 0 0-1.87.34l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.7 1.7 0 0 0 .34-1.87 1.7 1.7 0 0 0-1.55-1H3a2 2 0 1 1 0-4h.09a1.7 1.7 0 0 0 1.55-1 1.7 1.7 0 0 0-.34-1.87l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.7 1.7 0 0 0 1.87.34h0a1.7 1.7 0 0 0 1-1.55V3a2 2 0 1 1 4 0v.09a1.7 1.7 0 0 0 1 1.55h0a1.7 1.7 0 0 0 1.87-.34l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.7 1.7 0 0 0-.34 1.87v0a1.7 1.7 0 0 0 1.55 1H21a2 2 0 1 1 0 4h-.09a1.7 1.7 0 0 0-1.55 1Z" />
    </>
  ),
  search: (
    <>
      <circle cx="11" cy="11" r="6.5" />
      <path d="m20 20-3.8-3.8" />
    </>
  ),
  bell: (
    <>
      <path d="M6 9a6 6 0 1 1 12 0c0 4 1.5 5.5 1.5 5.5H4.5S6 13 6 9Z" />
      <path d="M10 18a2 2 0 0 0 4 0" />
    </>
  ),
  'chevron-down': <path d="m6 9 6 6 6-6" />,
  'chevron-right': <path d="m9 6 6 6-6 6" />,
  'chevron-left': <path d="m15 6-6 6 6 6" />,
  plus: <path d="M12 5v14M5 12h14" />,
  filter: <path d="M4 5h16l-6 7v5.5L10 20v-8L4 5Z" />,
  refresh: (
    <>
      <path d="M20 11a8 8 0 0 0-14.9-3M4 5v4h4" />
      <path d="M4 13a8 8 0 0 0 14.9 3M20 19v-4h-4" />
    </>
  ),
  logs: (
    <>
      <path d="M4 6h16M4 12h16M4 18h10" />
    </>
  ),
  cpu: (
    <>
      <rect x="6" y="6" width="12" height="12" rx="2" />
      <rect x="10" y="10" width="4" height="4" />
      <path d="M9 2v4M15 2v4M9 18v4M15 18v4M2 9h4M2 15h4M18 9h4M18 15h4" />
    </>
  ),
  memory: (
    <>
      <rect x="3" y="6" width="18" height="12" rx="2" />
      <path d="M7 10v4M12 10v4M17 10v4M3 10h0M21 10h0" />
    </>
  ),
  network: (
    <>
      <circle cx="12" cy="12" r="2.5" />
      <circle cx="12" cy="5" r="1.5" />
      <circle cx="12" cy="19" r="1.5" />
      <circle cx="19" cy="12" r="1.5" />
      <circle cx="5" cy="12" r="1.5" />
      <path d="m12 14.5-1.8 3M12 14.5l1.8 3M13.9 13.7l5.2-1.2M10.1 13.7l-5.2-1.2" />
    </>
  ),
  database: (
    <>
      <ellipse cx="12" cy="5.5" rx="8" ry="3" />
      <path d="M4 5.5v13c0 1.66 3.58 3 8 3s8-1.34 8-3v-13" />
      <path d="M4 12c0 1.66 3.58 3 8 3s8-1.34 8-3" />
    </>
  ),
  check: <path d="m5 12 4.5 4.5L19 7" />,
  moon: <path d="M21 12.8A9 9 0 1 1 11.2 3 7 7 0 0 0 21 12.8Z" />,
  sun: (
    <>
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
    </>
  ),
  collapse: <path d="m9 4 6 6-6 6" />,
  expand: <path d="m15 4-6 6 6 6" />,
  more: (
    <>
      <circle cx="5" cy="12" r="0.9" fill="currentColor" stroke="none" />
      <circle cx="12" cy="12" r="0.9" fill="currentColor" stroke="none" />
      <circle cx="19" cy="12" r="0.9" fill="currentColor" stroke="none" />
    </>
  ),
  play: <path d="M8 5.5v13l11-6.5-11-6.5Z" fill="currentColor" stroke="none" />,
  pause: (
    <>
      <rect x="6" y="5" width="4" height="14" rx="1" fill="currentColor" stroke="none" />
      <rect x="14" y="5" width="4" height="14" rx="1" fill="currentColor" stroke="none" />
    </>
  ),
  cube: (
    <>
      <path d="m12 2 8 4.5v11L12 22l-8-4.5v-11L12 2Z" />
      <path d="M12 12 4 7.5M12 12l8-4.5M12 12v10" />
    </>
  ),
  close: (
    <>
      <path d="M6 6l12 12M18 6 6 18" />
    </>
  ),
  crown: (
    <>
      <path d="m2.5 6 4.2 3.8L12 3.5l5.3 6.3L21.5 6 19.3 16H4.7L2.5 6Z" />
      <path d="M5.5 19.5h13" />
    </>
  ),
  copy: (
    <>
      <rect x="9" y="9" width="11" height="11" rx="2" />
      <path d="M5 15H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h10a1 1 0 0 1 1 1v1" />
    </>
  ),
  alert: (
    <>
      <path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
      <path d="M12 9v4" />
      <path d="M12 17h.01" />
    </>
  ),
  loader: (
    <>
      <path d="M12 3a9 9 0 1 0 9 9" />
    </>
  ),
  power: (
    <>
      <path d="M12 3v9" />
      <path d="M6.4 5.5a8 8 0 1 0 11.2 0" />
    </>
  ),
  star: (
    <>
      <path d="m12 3.5 2.6 5.3 5.9.9-4.3 4.1 1 5.8-5.2-2.7-5.2 2.7 1-5.8L3.5 9.7l5.9-.9L12 3.5Z" />
    </>
  ),
  clock: (
    <>
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    </>
  ),
  // lucide external-link 同款：箭头穿出方框，用于「查看 pipeline 详情」外链（对齐旧版外链 SVG）
  external: (
    <>
      <path d="M15 3h6v6" />
      <path d="M10 14 21 3" />
      <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
    </>
  ),
  // antd CheckCircleFilled 同款：实心圆 + 白色对勾，用作 pipeline 成功状态图标（沿用旧版）
  'check-circle-fill': (
    <>
      <circle cx="12" cy="12" r="10" fill="currentColor" stroke="none" />
      <path
        d="m7.5 12.5 3 3 6-6.5"
        fill="none"
        stroke="#fff"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </>
  ),
  // antd CloseCircleFilled 同款：实心圆 + 白色叉，用作 pipeline 失败状态图标（沿用旧版）
  'close-circle-fill': (
    <>
      <circle cx="12" cy="12" r="10" fill="currentColor" stroke="none" />
      <path d="m8.5 8.5 7 7M15.5 8.5l-7 7" fill="none" stroke="#fff" strokeWidth="2" strokeLinecap="round" />
    </>
  ),
  // antd WarningFilled 同款：实心三角形 + 白色叹号，用作 pipeline 执行中状态图标（沿用旧版）
  'warning-fill': (
    <>
      <path
        d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0Z"
        fill="currentColor"
        stroke="none"
      />
      <path d="M12 9v4" fill="none" stroke="#fff" strokeWidth="2" strokeLinecap="round" />
      <path d="M12 17h.01" fill="none" stroke="#fff" strokeWidth="2" strokeLinecap="round" />
    </>
  ),
  user: (
    <>
      <circle cx="12" cy="8" r="3.5" />
      <path d="M4.5 20c0-3.5 3.4-5.5 7.5-5.5s7.5 2 7.5 5.5" />
    </>
  ),
}

/** 线性图标：<Icon name="cluster" className="text-xl" /> */
export function Icon({
  name,
  className,
  ...rest
}: { name: IconName; className?: string } & Omit<SVGProps<SVGSVGElement>, 'name' | 'children'>) {
  return (
    <I className={className} {...rest}>
      {paths[name]}
    </I>
  )
}
