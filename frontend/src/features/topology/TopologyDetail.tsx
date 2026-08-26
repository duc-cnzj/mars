import { useEffect, useRef, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { Button } from '@/components/ui/shadcn/button'
import { Tag } from '@/components/ui/Tag'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/shadcn/tooltip'
import { nextZIndex } from '@/lib/zIndex'
import { KIND_ICON, STATUS_TONE } from './mockTopology'
import type { NodeStatus, TopoEndpoint, TopoNode } from './topologyTypes'

/** 健康状态 → i18n 词条键（字面量联合，保证 t() 类型收窄通过） */
function statusKey(
  status: NodeStatus,
): 'topology.statusHealthy' | 'topology.statusDegraded' | 'topology.statusProgressing' | 'topology.statusUnknown' {
  if (status === 'healthy') return 'topology.statusHealthy'
  if (status === 'degraded') return 'topology.statusDegraded'
  if (status === 'progressing') return 'topology.statusProgressing'
  return 'topology.statusUnknown'
}

/**
 * 节点名：宽幅超长时截断（省略号），hover 出全名 tooltip。
 * 截断检测：scrollWidth > clientWidth（+1px 缓冲防亚像素抖动），ResizeObserver 重算。
 * tooltip 走 portal 挂 body，须盖过可拖拽宿主弹窗的动态 z（z-51+），hover 打开时取 nextZIndex
 * （对齐 Elements FieldLabel / 强杀确认框机制）。仅截断时才真正打开才 bump 共享 zCounter。
 */
function NodeName({ name }: { name: string }) {
  const ref = useRef<HTMLDivElement>(null)
  const [truncated, setTruncated] = useState(false)
  const [tipHover, setTipHover] = useState(false)
  const [tipZ, setTipZ] = useState(50)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const measure = () => setTruncated(el.scrollWidth > el.clientWidth + 1)
    measure()
    const ro = new ResizeObserver(measure)
    ro.observe(el)
    return () => ro.disconnect()
  }, [name])

  return (
    <TooltipProvider delayDuration={100}>
      <Tooltip open={truncated && tipHover} onOpenChange={() => {}}>
        <TooltipTrigger asChild>
          <div
            ref={ref}
            onMouseEnter={() => {
              if (truncated) setTipZ(nextZIndex())
              setTipHover(true)
            }}
            onMouseLeave={() => setTipHover(false)}
            className="truncate text-[13px] font-medium text-ink"
          >
            {name}
          </div>
        </TooltipTrigger>
        <TooltipContent side="top" style={{ zIndex: tipZ }}>
          {name}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}

interface TopologyDetailProps {
  node: TopoNode
  onClose: () => void
  /** 面板主体自定义操作区（直播 Tab 传 Pod 日志/强杀） */
  actions?: ReactNode
  /** 项目访问地址列表（仅 Application 根节点展示全部；直播 Tab 由 /api/endpoints 传入） */
  endpoints?: TopoEndpoint[]
}

/**
 * 节点详情面板：画布容器内的绝对定位覆盖层（非 portal，避免与 pan/zoom 打架）。
 * 展示状态/命名空间；仅 Pod 节点展示事件（运行单元层面信号），Application 根节点额外
 * 平铺展示全部项目访问地址；操作区（Pod 日志/强杀）由父组件注入。
 */
export function TopologyDetail({ node, onClose, actions, endpoints }: TopologyDetailProps) {
  const { t } = useTranslation()
  return (
    <aside className="animate-in fade-in slide-in-from-right-2 absolute right-3 top-3 bottom-3 z-10 flex w-80 flex-col overflow-hidden rounded-lg border border-line bg-overlay shadow-lg duration-150">
      {/* 头部：kind + 名称 + 关闭 */}
      <div className="flex items-start gap-2 border-b border-line px-4 py-3">
        <span className="mt-0.5 shrink-0 text-faint">
          <Icon name={KIND_ICON[node.kind]} className="size-4" />
        </span>
        <div className="min-w-0 flex-1">
          <div className="text-[10px] font-medium tracking-[0.08em] text-faint uppercase">{node.kind}</div>
          {/* 长名截断时 hover 弹全名 tooltip（超宽检测见组件） */}
          <NodeName name={node.name} />
        </div>
        <Button
          variant="ghost"
          size="icon-xs"
          onClick={onClose}
          aria-label={t('common.close')}
          className="text-mute hover:text-ink"
        >
          <Icon name="close" className="size-4" />
        </Button>
      </div>

      {/* 主体：滚动区域 */}
      <div className="min-h-0 flex-1 space-y-3 overflow-y-auto px-4 py-3">
        {/* 状态 / 命名空间 */}
        <div className="flex items-center justify-between py-1">
          <span className="text-[12px] text-mute">{t('topology.status')}</span>
          <Tag tone={STATUS_TONE[node.status]}>{t(statusKey(node.status))}</Tag>
        </div>
        <div className="flex items-center justify-between py-1">
          <span className="text-[12px] text-mute">{t('topology.namespace')}</span>
          <span className="font-mono text-[12px] text-ink">{node.namespace}</span>
        </div>
        {/* 项目访问地址（仅 Application 根节点）：全部端点平铺展示，http(s) 链接可点开、
            hostname/IP 纯文本（对齐 ProjectRow 端点展示策略）。数据由父组件注入：
            直播 Tab 传 /api/endpoints 全量；空列表不渲染该区块 */}
        {node.kind === 'Application' && endpoints && endpoints.length > 0 && (
          <div className="py-1">
            <div className="mb-1.5 text-[12px] text-mute">{t('topology.accessAddresses')}</div>
            <div className="flex flex-col gap-1">
              {endpoints.map((ep, i) => (
                <div key={i} className="flex items-center gap-1.5 py-0.5">
                  <span className="shrink-0 font-mono text-[11px] text-faint">
                    {ep.name}
                    {ep.portName ? `(${ep.portName})` : ''}:
                  </span>
                  {ep.url.startsWith('http') ? (
                    <a
                      href={ep.url}
                      target="_blank"
                      rel="noreferrer"
                      translate="no"
                      className="min-w-0 flex-1 truncate text-[12px] text-primary hover:underline"
                    >
                      {ep.url}
                    </a>
                  ) : (
                    <span className="min-w-0 flex-1 truncate text-[12px] text-ink" translate="no">
                      {ep.url}
                    </span>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* 直播 Tab 的节点操作区（Pod 日志/强杀），由父组件注入 */}
        {actions}

        {/* 事件列表（仅 Pod 节点展示：事件是运行单元层面的信号，其余 kind 无业务意义） */}
        {node.kind === 'Pod' && (
          <div className="py-1">
            <div className="mb-1.5 text-[12px] text-mute">{t('topology.events')}</div>
            <div className="flex flex-col gap-1.5">
              {node.events.map((ev, i) => (
                <div key={i} className="flex gap-2 rounded-md border border-line bg-raised px-2.5 py-2">
                  <Icon
                    name={ev.type === 'warning' ? 'alert' : 'info'}
                    className={`mt-0.5 size-3.5 shrink-0 ${ev.type === 'warning' ? 'text-err' : 'text-faint'}`}
                  />
                  <div className="min-w-0">
                    <div
                      className={`text-[12px] leading-snug ${ev.type === 'warning' ? 'text-err' : 'text-ink'}`}
                    >
                      {ev.message}
                    </div>
                    <div className="mt-0.5 font-mono text-[11px] text-faint">{ev.time}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </aside>
  )
}
