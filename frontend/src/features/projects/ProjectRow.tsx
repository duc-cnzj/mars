import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import { Gauge, Link, Loader2 } from 'lucide-react'
import type { components } from '../../api/schema'
import { api } from '../../api/client'
import { getEndpoints } from '../../api/endpointsCache'
import { copyText } from '../../utils/copy'
import { Icon } from '../../components/icons'
import { DeployStatusIcon } from './DeployStatusIcon'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/shadcn/popover'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/shadcn/tooltip'

type ProjectModel = components['schemas']['types.ProjectModel']
type ServiceEndpointModel = components['schemas']['types.ServiceEndpoint']

/**
 * 项目行：命名空间卡片内联的项目入口（旧版虚线按钮样式）。
 * 展示部署状态 + 项目名；部署成功时右侧追加访问地址与 CPU/内存（对齐旧版 ProjectDetail）。
 * 项目名被省略号截断时，hover 名称文字本体弹完整名 tooltip；未截断不弹。
 * 名称 span 点击冒泡到行按钮，正常打开详情弹窗；仅 CPU/URL 图标为独立交互区（onClick stopPropagation）。
 * URL 与资源均用点击 Popover（而非仅 hover）——ui-ux-pro-max「Hover vs Tap」规则；
 * 图标为 24px 命中区且可聚焦，满足 WCAG 2.2 目标尺寸。点击行（含项目名）打开 ProjectDetailModal。
 */
export function ProjectRow({
  project,
  onClick,
}: {
  project: ProjectModel
  onClick: () => void
}) {
  // 全名 tooltip：受控 open = 名称截断 && 悬停在名称 span 上。
  // hover 挂在名称 span（文字本体），行其他区域悬停不弹；点击名称冒泡到行按钮，打开详情弹窗。
  const nameRef = useRef<HTMLSpanElement>(null)
  const [truncated, setTruncated] = useState(false)
  const [nameTipHover, setNameTipHover] = useState(false)

  // 截断检测：scrollWidth > clientWidth（+1px 缓冲防亚像素抖动）；ResizeObserver 在容器/文字变化时重算
  useEffect(() => {
    const el = nameRef.current
    if (!el) return
    const measure = () => setTruncated(el.scrollWidth > el.clientWidth + 1)
    measure()
    const ro = new ResizeObserver(measure)
    ro.observe(el)
    return () => ro.disconnect()
  }, [project.name])

  return (
    <button
      type="button"
      onClick={onClick}
      className="flex w-full items-center gap-2 rounded-lg border border-dashed border-line bg-surface px-3 py-2 text-left transition-colors hover:border-primary/50 hover:bg-raised"
    >
      <DeployStatusIcon status={project.deployStatus} />
      <TooltipProvider delayDuration={100}>
        <Tooltip open={truncated && nameTipHover} onOpenChange={() => {}}>
          <TooltipTrigger asChild>
            <span
              ref={nameRef}
              onMouseEnter={() => setNameTipHover(true)}
              onMouseLeave={() => setNameTipHover(false)}
              className="min-w-0 flex-1 truncate text-[13px] font-medium text-ink"
            >
              {project.name}
            </span>
          </TooltipTrigger>
          <TooltipContent side="top" className="max-w-[min(320px,80vw)] break-words">
            {project.name}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
      {/* 对齐旧版 ProjectDetail：仅部署成功时显示 CPU/内存 + 访问地址（URL 仅在存在时显示） */}
      {project.deployStatus === 'StatusDeployed' && (
        <>
          <ProjectCpuMemory projectId={project.id} />
          <ProjectEndpoints projectId={project.id} />
        </>
      )}
    </button>
  )
}

/** 弹层触发图标基类：24px 命中区、可聚焦、点击不冒泡到父 button（避免打开详情弹窗） */
const iconTriggerCls =
  'flex size-6 shrink-0 cursor-pointer items-center justify-center rounded-md text-faint transition-colors hover:bg-raised hover:text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50'

/**
 * 项目资源用量：点击 Popover 懒拉取 /api/metrics/projects/{projectId}/cpu_memory（对齐旧版 CpuMemory）。
 * 旧版为 hover Tooltip，触屏/键盘不可达；改为点击弹层满足「Hover vs Tap」。
 */
function ProjectCpuMemory({ projectId }: { projectId: number }) {
  const { t } = useTranslation()
  const [usage, setUsage] = useState<{ cpu: string; memory: string } | null>(null)

  const fetchUsage = () => {
    if (usage) return
    api
      .GET('/api/metrics/projects/{projectId}/cpu_memory', {
        params: { path: { projectId } },
      })
      .then(({ data }) => {
        if (data) setUsage({ cpu: data.cpu, memory: data.memory })
      })
      .catch(() => setUsage({ cpu: '-', memory: '-' }))
  }

  return (
    // 冒泡拦截（仅此一件事）：trigger 的 stopPropagation 防止点图标冒到行按钮误开详情；
    // 内容 onClick stopPropagation 拦断 portal 内容经 React 树冒泡到行按钮（点弹窗内也误开详情）。
    <Popover onOpenChange={(open) => open && fetchUsage()}>
      <PopoverTrigger asChild>
        <span
          role="button"
          tabIndex={0}
          onClick={(e) => e.stopPropagation()}
          className={iconTriggerCls}
          aria-label={t('project.cpuMemory')}
          title={t('project.cpuMemory')}
        >
          <Gauge size={16} />
        </span>
      </PopoverTrigger>
      <PopoverContent
        side="top"
        onClick={(e) => e.stopPropagation()}
        className="w-[min(240px,80vw)] p-2"
      >
        <div className="mb-1 px-1 text-[12px] font-medium">{t('project.cpuMemory')}</div>
        {usage ? (
          <div className="flex flex-col gap-0.5 px-1 pb-1 font-mono text-[12px]">
            <span>cpu: {usage.cpu || '-'}</span>
            <span>memory: {usage.memory || '-'}</span>
          </div>
        ) : (
          <div className="flex items-center gap-1.5 px-1 py-1 text-[12px] text-faint">
            <Loader2 className="size-3 animate-spin" />
            {t('common.loading')}
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}

/**
 * 项目访问地址：图标恒显示（无论有无端点），但数据懒加载——仅点击打开弹层时才经 endpointsCache
 * 拉取 /api/endpoints/projects/{projectId}（TTL 60s + inflight 去重）。切页/渲染零端点请求，
 * 彻底消除原请求风暴；弹层内做加载 / 空 / 列表三态（对齐 NamespaceCard.NamespaceEndpoints）。
 * 点击 Popover 展示链接与复制（对齐旧版 ServiceEndpoint）。
 */
function ProjectEndpoints({ projectId }: { projectId: number }) {
  const { t } = useTranslation()
  const [eps, setEps] = useState<ServiceEndpointModel[] | null>(null)

  const fetchEndpoints = () => {
    if (eps) return // 已加载（含空）直接复用；缓存命中时 getEndpoints 同步返回
    void getEndpoints(projectId).then(setEps)
  }

  const copyUrl = async (url: string) => {
    const ok = await copyText(url)
    if (ok) toast.success(t('common.copied'))
    else toast.error(t('common.copyFailed'))
  }

  return (
    // 冒泡拦截（仅此一件事）：trigger 的 stopPropagation 防止点图标冒到行按钮误开详情；
    // 内容 onClick stopPropagation 拦断 portal 内容经 React 树冒泡到行按钮（点弹窗内也误开详情）。
    <Popover onOpenChange={(open) => open && fetchEndpoints()}>
      <PopoverTrigger asChild>
        <span
          role="button"
          tabIndex={0}
          onClick={(e) => e.stopPropagation()}
          className={iconTriggerCls}
          aria-label={t('project.endpoints')}
          title={t('project.endpoints')}
        >
          <Link size={16} />
        </span>
      </PopoverTrigger>
      <PopoverContent
        side="top"
        onClick={(e) => e.stopPropagation()}
        className="w-[max-content] max-w-[min(480px,90vw)] p-2"
      >
        <div className="mb-1 px-1 text-[12px] font-medium">
          {t('project.endpoints')}
        </div>
        {eps === null ? (
          <div className="flex items-center gap-1.5 px-1 py-1 text-[12px] text-faint">
            <Loader2 className="size-3 animate-spin" />
            {t('common.loading')}
          </div>
        ) : eps.length === 0 ? (
          <div className="px-1 py-1 text-[12px] text-faint">{t('common.empty')}</div>
        ) : (
          <div className="flex max-h-48 flex-col gap-0.5 overflow-auto">
            {eps.map((ep, i) => (
              <div
                key={i}
                className="flex items-center gap-1.5 rounded-md px-1 py-1 text-[12px] hover:bg-raised"
              >
                <span className="shrink-0 text-faint">
                  {ep.name}
                  {ep.portName ? `(${ep.portName})` : ''}:
                </span>
                {ep.url.startsWith('http') ? (
                  <a
                    href={ep.url}
                    target="_blank"
                    rel="noreferrer"
                    className="min-w-0 flex-1 truncate text-primary hover:underline"
                  >
                    {ep.url}
                  </a>
                ) : (
                  <span className="min-w-0 flex-1 truncate text-mute">{ep.url}</span>
                )}
                <button
                  type="button"
                  onClick={() => copyUrl(ep.url)}
                  title={t('common.copied')}
                  className="flex size-6 shrink-0 items-center justify-center rounded-md text-faint transition-colors hover:text-primary"
                >
                  <Icon name="copy" className="text-[11px]" />
                </button>
              </div>
            ))}
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
