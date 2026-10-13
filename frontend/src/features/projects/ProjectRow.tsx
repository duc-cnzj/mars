import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import { getEndpoints } from '@/api/endpointsCache'
import { getHibernateWindows } from '@/api/hibernateWindowsCache'
import { copyText } from '@/lib/copy'
import { cn } from '@/lib/utils'
import { Icon } from '@/components/Icons'
import { EndpointUrl } from '@/components/EndpointUrl'
import { TruncatedText } from '@/components/TruncatedText'
import { Button, buttonVariants } from '@/components/ui/shadcn/button'
import { DeployStatusIcon } from './DeployStatusIcon'
import { hibernateWindowLabel, replicaFieldLabels } from './hibernate'
import { iconTriggerCls } from './iconTrigger'
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
type HibernateWindowModel = components['schemas']['hibernate.HibernateWindow']

/**
 * 项目行：命名空间卡片内联的项目入口（对齐标准 dashed variant：同「次级入口」虚线语言，hover 点亮 primary）。
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
  // 全名 tooltip：名称超长截断时 hover 弹完整名——统一走 TruncatedText（测量 / hover / 受控 open 封装，
  // 与 SystemSettings / ProjectGovernance 同源）；点击名称仍冒泡到行按钮打开详情弹窗。

  // 休眠中：项目级开关已开且已进入时间窗（副本字段已缩容为 0）——名称右侧挂 zzz 图标。
  // 区别于「已开启但还在运行」（awake）：那只是工具簇月亮实心，不挂 zzz，避免把「已武装」误当「已睡着」。
  const sleeping = project.hibernateEnabled && project.hibernateState === 'sleeping'

  return (
    /* 对齐标准 dashed variant（buttonVariants），但外层用 <div role="button"> 而非 <button>：
       行内嵌 CPU/内存、端点等独立交互区（PopoverTrigger span），若外层用 <button> 会形成
       button-in-button 语义冲突（HTML 无效 + 无障碍树嵌套）。div role=button 容器可合法容纳
       交互子元素，视觉与按钮完全一致（Enter/Space 同样触发打开详情）。 */
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          onClick()
        }
      }}
      className={cn(
        buttonVariants({ variant: 'dashed' }),
        // 行 gap 保持 buttonVariants 基类 gap-2（8px）：此间距作用于「部署状态↔名称↔工具簇」，
        // 视觉上只体现在部署状态与名称之间。工具簇内部三图标另用独立容器收紧（见下）
        'h-auto w-full cursor-pointer justify-start rounded-lg px-3 py-2 text-left',
      )}
    >
      <DeployStatusIcon status={project.deployStatus} />
      {/* 名称与 zzz 图标同处一个 flex-1 容器：容器占满中段把工具簇顶到最右，而 zzz 紧贴名称——
          名称短时 zzz 就挨着名字，名称过长被截断时 zzz 仍紧随其后，不会被推到行尾 */}
      <span className="flex min-w-0 flex-1 items-center gap-1">
        <TruncatedText text={project.name} className="min-w-0 text-[13px] font-medium text-ink" />
        {/* 休眠中：名称紧贴 zzz 图标，一眼看出该项目已入睡（副本字段已缩容为 0）；
            hover 展开只读明细（生效时间窗 / 副本字段 / 置零说明），明细按需懒拉 */}
        {sleeping && (
          <HibernateSleepTip projectId={project.id} windowNames={project.hibernateWindowNames} />
        )}
      </span>
      {/* 工具簇：CPU/内存 + 访问地址两个 24px 命中区图标，用独立容器 gap-0.5（2px）收紧，
          与行级 gap-2 解耦——只收紧图标间距，不动「部署状态↔名称」的间距。
          shrink-0：簇整体不被 flex 压缩；items-center 对齐图标垂直中线。
          定时休眠编辑入口已迁至详情弹窗标题栏，列表行只留 zzz 被动信号。 */}
      <span className="flex shrink-0 items-center gap-0.5">
        {/* 对齐旧版 ProjectDetail：仅部署成功时显示 CPU/内存 + 访问地址（URL 仅在存在时显示） */}
        {project.deployStatus === 'StatusDeployed' && (
          <>
            <ProjectCpuMemory projectId={project.id} />
            <ProjectEndpoints projectId={project.id} />
          </>
        )}
      </span>
    </div>
  )
}

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
      .GET(API.metricsProjectCpuMemory, {
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
          <Icon name="gauge" className="size-4" />
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
            <Icon name="loader" className="size-3 animate-spin" />
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
          <Icon name="link" className="size-4" />
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
            <Icon name="loader" className="size-3 animate-spin" />
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
                {(ep.svcName || ep.ingressName) && (
                  <span className="shrink-0 text-faint">{ep.svcName || ep.ingressName}:</span>
                )}
                <EndpointUrl url={ep.url} isLink={ep.url.startsWith('http')} />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-xs"
                  onClick={() => copyUrl(ep.url)}
                  title={t('common.copied')}
                  className="shrink-0 text-faint hover:text-primary"
                >
                  <Icon name="copy" className="text-[11px]" />
                </Button>
              </div>
            ))}
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}

/**
 * 休眠中 zzz 图标的悬停提示（只读）：hover 时才懒拉一次项目详情（取仓库声明的窗列表 + 副本字段）
 * 与全局时间窗（缓存去重），把「生效时间窗 + 副本字段 + 置零说明」铺开——列表行只放信号（zzz），
 * 明细按需展开。与 CPU/端点图标同策略：切页/渲染零请求，仅悬停该行（睡眠态稀少）才有 1 次详情请求。
 * 详情未就绪时先给加载态；拉取失败静默保留图标（不弹提示），不阻断列表。
 */
function HibernateSleepTip({
  projectId,
  windowNames,
}: {
  projectId: number
  /** 列表项已带项目选定子集（空 = 未选），先于详情返回即可用于兜底展示 */
  windowNames?: string[]
}) {
  const { t } = useTranslation()
  const [detail, setDetail] = useState<ProjectModel | null>(null)
  const [windows, setWindows] = useState<HibernateWindowModel[]>([])

  /** 悬停时按需加载一次：项目详情（取仓库窗列表 + 副本字段 + 元素描述）+ 全局时间窗时段。
   *  detail 守卫保证每次挂载只拉一次，切页/渲染零请求。 */
  const load = () => {
    if (detail) return
    api
      .GET(API.projectsDetail, { params: { path: { id: projectId } } })
      .then(({ data }) => {
        if (data?.item) setDetail(data.item)
      })
      .catch(() => {})
    void getHibernateWindows().then(setWindows)
  }

  const hb = detail?.repo?.marsConfig?.hibernate
  // 生效窗 = 项目选定子集本身（空 = 未选，不继承仓库全量）；时段缺失时回退窗名本身
  const winText = hibernateWindowLabel(windowNames ?? [], windows)
  const fields = replicaFieldLabels(hb?.replicaPaths ?? [], detail?.repo?.marsConfig?.elements)
  return (
    <TooltipProvider delayDuration={100}>
      <Tooltip>
        <TooltipTrigger asChild>
          <span
            className="inline-flex shrink-0 items-center"
            onMouseEnter={load}
          >
            <Icon
              name="zzz"
              className="size-4 shrink-0 text-hibernate"
              aria-label={t('project.hibernateStateSleeping')}
            />
          </span>
        </TooltipTrigger>
        <TooltipContent side="top" className="max-w-[min(320px,85vw)]">
          {!detail ? (
            <div className="flex items-center gap-1.5 text-[11px]">
              <Icon name="loader" className="size-3 animate-spin" />
              {t('common.loading')}
            </div>
          ) : (
            <div className="flex flex-col gap-1.5 text-[11px] leading-snug">
              <div className="flex flex-col gap-0.5">
                <span className="text-background/60">{t('project.hibernateWindow')}:</span>
                <span className="font-mono">{winText || '-'}</span>
              </div>
              <div className="flex flex-col gap-0.5">
                <span className="text-background/60">{t('project.hibernateReplicaFields')}:</span>
                <span className="font-mono">{fields.join(', ') || '-'}</span>
              </div>
              <div className="text-background/70">{t('project.hibernateReplicaZeroNote')}</div>
            </div>
          )}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
