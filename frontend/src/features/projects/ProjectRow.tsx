import { useEffect, useRef, useState } from 'react'
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
import { Tag } from '@/components/ui'
import { Button, buttonVariants } from '@/components/ui/shadcn/button'
import { Switch } from '@/components/ui/shadcn/switch'
import { DeployStatusIcon } from './DeployStatusIcon'
import { hibernateWindowLabel, replicaFieldLabels } from './hibernate'
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
  onChanged,
}: {
  project: ProjectModel
  onClick: () => void
  /** 休眠开关切换成功后的回调：通知工作台原地刷新该空间卡片（不传则仅更新行内弹层本地态） */
  onChanged?: () => void
}) {
  // 行级 i18n：休眠中 zzz 图标的取值（子组件各自 useTranslation，行内用例走这里）
  const { t } = useTranslation()
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
        <TooltipProvider delayDuration={100}>
          <Tooltip open={truncated && nameTipHover} onOpenChange={() => {}}>
            <TooltipTrigger asChild>
              <span
                ref={nameRef}
                onMouseEnter={() => setNameTipHover(true)}
                onMouseLeave={() => setNameTipHover(false)}
                className="min-w-0 truncate text-[13px] font-medium text-ink"
              >
                {project.name}
              </span>
            </TooltipTrigger>
            <TooltipContent side="top" className="max-w-[min(320px,80vw)] break-words">
              {project.name}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
        {/* 休眠中：名称紧贴 zzz 图标，一眼看出该项目已入睡（副本字段已缩容为 0） */}
        {sleeping && (
          <Icon
            name="zzz"
            className="size-4 shrink-0 text-hibernate"
            aria-label={t('project.hibernateStateSleeping')}
          />
        )}
      </span>
      {/* 工具簇：CPU/内存 + 访问地址 + 定时休眠三个 24px 命中区图标，用独立容器 gap-0.5（2px）收紧，
          与行级 gap-2 解耦——只收紧三图标间距，不动「部署状态↔名称」的间距。
          shrink-0：簇整体不被 flex 压缩；items-center 对齐图标垂直中线。 */}
      <span className="flex shrink-0 items-center gap-0.5">
        {/* 对齐旧版 ProjectDetail：仅部署成功时显示 CPU/内存 + 访问地址（URL 仅在存在时显示） */}
        {project.deployStatus === 'StatusDeployed' && (
          <>
            <ProjectCpuMemory projectId={project.id} />
            <ProjectEndpoints projectId={project.id} />
          </>
        )}
        {/* 定时休眠入口：图标恒显示（对齐访问地址图标的常驻策略），数据点击时才懒拉。
            enabled 取自行数据（列表项已带），让图标即时区分开启/未开启，无需等懒拉详情 */}
        <ProjectHibernate
          projectId={project.id}
          enabled={project.hibernateEnabled}
          onChanged={onChanged}
        />
      </span>
    </div>
  )
}

/** 弹层触发图标基类（几何/命中区/焦点环）：24px 命中区、可聚焦、点击不冒泡到父 button（避免打开详情弹窗）。
    颜色态由调用方拼：未激活走 iconTriggerCls，激活态（如休眠开启）走 hibernateIconActive */
const iconTriggerBase =
  'flex size-6 shrink-0 cursor-pointer items-center justify-center rounded-md transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50'

/** 未激活态：淡色图标，hover 转主色 */
const iconTriggerCls = cn(iconTriggerBase, 'text-faint hover:bg-raised hover:text-primary')

/** 休眠开启态：不着底色块，只把月亮本体填实成主色（图标切 moon-fill），与未激活的淡色描边月亮一眼可辨；
    hover 仅垫一层 raised 底提示可点，不改变月亮的实心主色 */
const hibernateIconActive = cn(iconTriggerBase, 'text-primary hover:bg-raised')

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
 * 项目定时休眠：图标恒显示（开启休眠时常驻实心主色底、未开为淡色，一眼区分），
 * 数据点击弹层时才懒拉取：/api/projects/{id}（该项目详情的 repo.marsConfig.hibernate
 * 决定仓库是否有休眠能力 + hibernateEnabled/hibernateState 回显开关）+ 全局时间窗列表
 * （把 window_name 映射成「窗名 (时段)」，走缓存模块去重）。
 * 与 CPU/端点图标同策略：切页/渲染零请求，避免每行一个请求的风暴。
 * 无仓库级能力时给出提示；有能力时提供「开关 + 只读时间窗 / 副本字段」，
 * 切换仅改项目级开关、不触发重部署，成功后通知宿主刷新。
 */
function ProjectHibernate({
  projectId,
  enabled,
  onChanged,
}: {
  projectId: number
  /** 行数据（列表项）已带项目级休眠开关，供图标即时反映开启态，无需等弹层懒拉详情 */
  enabled: boolean
  onChanged?: () => void
}) {
  const { t } = useTranslation()
  const [detail, setDetail] = useState<ProjectModel | null>(null)
  const [windows, setWindows] = useState<HibernateWindowModel[]>([])
  const [saving, setSaving] = useState(false)

  /** 打开弹层时按需加载：项目详情（取休眠能力/开关/状态）+ 全局时间窗（把 window_name 映射成时段展示）。
   *  两者都只拉一次（详情靠 detail 守卫，时间窗靠缓存模块去重），切页/渲染零请求。 */
  const load = () => {
    if (!detail) {
      api
        .GET(API.projectsDetail, { params: { path: { id: projectId } } })
        .then(({ data }) => {
          if (data?.item) setDetail(data.item)
        })
        .catch(() => {})
    }
    void getHibernateWindows().then(setWindows)
  }

  /** 切换项目级休眠开关：乐观更新，失败回滚并提示 */
  const toggle = async (next: boolean) => {
    if (!detail) return
    setDetail({ ...detail, hibernateEnabled: next })
    setSaving(true)
    try {
      const { error } = await api.POST(API.projectsHibernate, {
        params: { path: { projectId } },
        body: { projectId, enabled: next },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(
        next ? t('project.hibernateToggleOn') : t('project.hibernateToggleOff'),
      )
      onChanged?.()
    } catch (e) {
      setDetail((d) => (d ? { ...d, hibernateEnabled: !next } : d))
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setSaving(false)
    }
  }

  const hb = detail?.repo?.marsConfig?.hibernate
  // 图标开启态：优先用弹层已拉的详情（切换后即时更新），否则回退行数据（未开弹层时也能区分）
  const hibOn = detail?.hibernateEnabled ?? enabled
  return (
    // 冒泡拦截（仅此一件事）：trigger 的 stopPropagation 防止点图标冒到行按钮误开详情；
    // 内容 onClick stopPropagation 拦断 portal 内容经 React 树冒泡到行按钮（点弹层内也误开详情）。
    <Popover onOpenChange={(open) => open && load()}>
      <PopoverTrigger asChild>
        <span
          role="button"
          tabIndex={0}
          onClick={(e) => e.stopPropagation()}
          // 开启休眠：月亮本体填实成主色（moon-fill），不着底色块；未开启为淡色描边月亮
          className={hibOn ? hibernateIconActive : iconTriggerCls}
          aria-label={t('project.hibernate')}
          title={t('project.hibernate')}
        >
          <Icon name={hibOn ? 'moon-fill' : 'moon'} className="size-4" />
        </span>
      </PopoverTrigger>
      <PopoverContent
        side="top"
        onClick={(e) => e.stopPropagation()}
        className="w-[min(260px,80vw)] p-2"
      >
        <div className="mb-1 px-1 text-[12px] font-medium">{t('project.hibernate')}</div>
        {!detail ? (
          <div className="flex items-center gap-1.5 px-1 py-1 text-[12px] text-faint">
            <Icon name="loader" className="size-3 animate-spin" />
            {t('common.loading')}
          </div>
        ) : !hb?.enabled ? (
          <div className="px-1 py-1 text-[12px] leading-snug text-faint">
            {t('project.hibernateNoCapability')}
            <div className="mt-1 text-[11px]">{t('project.hibernateNoCapabilityTip')}</div>
          </div>
        ) : (
          <div className="flex flex-col gap-1.5 px-1">
            <div className="flex items-center justify-between gap-2">
              <span className="flex items-center gap-1.5 text-[12px] text-ink">
                {detail.hibernateEnabled && (
                  <Tag
                    tone={detail.hibernateState === 'sleeping' ? 'warn' : 'ok'}
                    dot={false}
                    className="shrink-0"
                  >
                    {detail.hibernateState === 'sleeping'
                      ? t('project.hibernateStateSleeping')
                      : t('project.hibernateStateAwake')}
                  </Tag>
                )}
              </span>
              <Switch
                checked={detail.hibernateEnabled}
                disabled={saving}
                onCheckedChange={toggle}
              />
            </div>
            <div className="text-[11px] text-mute">
              {t('project.hibernateWindow')}:{' '}
              <span className="font-mono">{hibernateWindowLabel(hb.windowName, windows) || '-'}</span>
            </div>
            <div className="text-[11px] text-mute">
              {t('project.hibernateReplicaFields')}:{' '}
              <span className="font-mono">
                {replicaFieldLabels(hb.replicaPaths, detail.repo?.marsConfig?.elements).join(', ') ||
                  '-'}
              </span>
            </div>
            {/* 一句话说明：休眠时间段内副本字段被置 0，避免用户看到数量变化时误判 */}
            <div className="text-[11px] leading-snug text-faint">
              {t('project.hibernateReplicaZeroNote')}
            </div>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}
