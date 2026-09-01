import { memo, useCallback, useEffect, useRef, useState, type CSSProperties } from 'react'
import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { StatCard } from '@/components/StatCard'
import { Empty, RefreshFade, SkeletonList, Tag, type Tone } from '@/components/ui'
import { SearchInput } from '@/components/SearchInput'
import { Button } from '@/components/ui/shadcn/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/shadcn/tooltip'
import { formatDateTime } from '@/lib/format'
import { SEARCH_DEBOUNCE_MS } from '@/lib/constants'
import { humanizeDateTime } from '@/lib/humanizeDateTime'
import { toast } from '@/lib/toast'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import type { components } from '@/api/schema'
import type { TKey } from '@/i18n/keys'

/** 活跃度筛选：'all' = 不筛选（全部项目） */
type LivenessFilter = 'active' | 'dormant' | 'zombie' | 'all'

/** 最近部署状态（对齐后端 deployStatus 枚举语义，含服务端可能返回的未知态） */
type ProjectStatus = 'deployed' | 'progressing' | 'failed' | 'unknown'

/** 活跃度分类：活跃 / 低活跃 / 僵尸 */
type Liveness = 'active' | 'dormant' | 'zombie'

/** 服务端活跃度统计（基于搜索命中全量，不随分页/过滤裁剪） */
type LivenessStats = components['schemas']['project.LivenessStats']

/** 服务端活跃度清单单条项目 */
type LivenessItem = components['schemas']['project.LivenessItem']

/** 活跃阈值（天）：距最后更新时间不超过此值视为活跃 */
const ACTIVE_DAYS = 30

/** 僵尸阈值（天）：距最后更新时间超过此值视为僵尸（长期无人使用） */
const ZOMBIE_DAYS = 90

/** 每页条数（服务端分页，滚动触底追加下一页） */
const PAGE_SIZE = 15

/** 后端部署状态枚举 → 前端状态（服务端可能返回 StatusUnknown，故补 unknown 分支） */
const STATUS_MAP: Record<LivenessItem['deployStatus'], ProjectStatus> = {
  StatusUnknown: 'unknown',
  StatusDeploying: 'progressing',
  StatusDeployed: 'deployed',
  StatusFailed: 'failed',
}

/** 最近部署状态 → 标签色：部署成功 ok / 部署中 warn / 失败 err / 未知 mute */
const STATUS_TONE: Record<ProjectStatus, Tone> = {
  deployed: 'ok',
  progressing: 'warn',
  failed: 'err',
  unknown: 'mute',
}

/** 最近部署状态 → i18n 词条键（TKey 字面量类型，编译期校验词条不悬空） */
const PROJECT_STATUS_KEY: Record<ProjectStatus, TKey> = {
  deployed: 'governance.statusDeployed',
  progressing: 'governance.statusProgressing',
  failed: 'governance.statusFailed',
  unknown: 'governance.statusUnknown',
}

/** 活跃度 → 标签色：活跃 ok（常态）/ 低活跃 warn / 僵尸 err（建议治理） */
const LIVENESS_TONE: Record<Liveness, Tone> = {
  active: 'ok',
  dormant: 'warn',
  zombie: 'err',
}

/** 活跃度 → i18n 词条键 */
const LIVENESS_KEY: Record<Liveness, TKey> = {
  active: 'governance.livenessActive',
  dormant: 'governance.livenessDormant',
  zombie: 'governance.livenessZombie',
}

/** 活跃度筛选展示顺序：活跃在前，僵尸在后（与治理优先级呼应） */
const LIVENESS_ORDER: readonly Liveness[] = ['active', 'dormant', 'zombie']

/** 按距最后更新时间天数分类活跃度（阈值与服务端一致：<=30 活跃 / >=90 僵尸）。
 *  时间戳缺失/非法（new Date → NaN）时归为活跃：宁可不提示治理，
 *  也不把无时间戳的项目误标成「低活跃」治理对象（NaN 时两个阈值比较均 false 会误落 dormant）。 */
function classifyLiveness(updatedAt: string, now: Date): Liveness {
  const ts = new Date(updatedAt).getTime()
  if (Number.isNaN(ts)) return 'active'
  const days = Math.floor((now.getTime() - ts) / 86_400_000)
  if (days <= ACTIVE_DAYS) return 'active'
  if (days >= ZOMBIE_DAYS) return 'zombie'
  return 'dormant'
}

/** 项目治理行（React.memo）：status/liveness 由父级预计算为原始字符串 prop（非每次新建对象），
 *  配合 item 引用稳定，列表状态变化（关键词/活跃度筛选/排序切换/翻页追加）时行不重渲染。 */
const GovernanceRow = memo(function GovernanceRow({
  item,
  status,
  liveness,
  className,
  style,
}: {
  item: LivenessItem
  status: ProjectStatus
  liveness: Liveness
  /** RefreshFade 经 cloneElement 注入的渐入 class/延迟——须转发到根元素才生效 */
  className?: string
  style?: CSSProperties
}) {
  const { t } = useTranslation()
  // 最近提交 cell 溢出检测：scrollWidth > clientWidth 判定内容被省略，仅省略时挂 tooltip 展示全文
  const commitRef = useRef<HTMLDivElement>(null)
  const [commitOverflow, setCommitOverflow] = useState(false)
  // 完整提交信息字符串（省略时 tooltip 展示）：作者 · 标题 · 相对时间
  const commitFull = [
    item.gitCommitAuthor,
    item.gitCommitTitle,
    item.gitCommitDate ? humanizeDateTime(item.gitCommitDate) : '',
  ]
    .filter(Boolean)
    .join(' · ')

  // 监听提交 cell 尺寸/内容变化实时判定省略态（grid 列宽随窗口或布局重排时同样生效）
  useEffect(() => {
    const el = commitRef.current
    if (!el) return
    const check = () => setCommitOverflow(el.scrollWidth > el.clientWidth)
    check()
    const ro = new ResizeObserver(check)
    ro.observe(el)
    return () => ro.disconnect()
  }, [item.gitCommitAuthor, item.gitCommitTitle, item.gitCommitDate])

  return (
    <div
      className={`grid grid-cols-1 gap-1 border-b border-line px-4 py-2.5 last:border-b-0 sm:grid-cols-2 lg:grid-cols-[1fr_3fr_1fr_6rem_4.5rem_5.5rem_5.5rem] lg:items-center lg:gap-3 ${className ?? ''}`}
      style={style}
    >
      {/* 项目名 */}
      <div className="min-w-0 truncate text-[13px] font-medium text-ink" title={item.name}>
        {item.name}
      </div>

      {/* 最近提交：提交人 + 标题 + 时间（判断代码是否仍有人维护）；无提交信息时占位。
          超长省略（commitOverflow）时 tooltip 展示全文，未省略不高频弹窗。
          ref 挂在真正省略的内层 span 上：外层 div 若自身 overflow:hidden 会把子元素溢出裁掉，
          使外层 scrollWidth 恒等于 clientWidth，检测永远失败——必须在发生省略的元素上量尺寸 */}
      <div className="min-w-0 text-[11px] text-mute">
        {item.gitCommitAuthor ? (
          <TooltipProvider delayDuration={100}>
            <Tooltip>
              <TooltipTrigger asChild>
                <span ref={commitRef} className="block truncate">
                  <span className="font-medium text-ink">{item.gitCommitAuthor}</span>
                  {item.gitCommitTitle && <span className="text-faint"> · {item.gitCommitTitle}</span>}
                  {item.gitCommitDate && <span className="text-faint"> · {humanizeDateTime(item.gitCommitDate)}</span>}
                </span>
              </TooltipTrigger>
              {/* 仅省略时挂载 tooltip 内容；未省略时 Radix 无内容悬停不弹窗 */}
              {commitOverflow && <TooltipContent>{commitFull}</TooltipContent>}
            </Tooltip>
          </TooltipProvider>
        ) : (
          <span className="text-faint">—</span>
        )}
      </div>

      {/* 命名空间 */}
      <span className="truncate text-[12px] text-mute">{item.namespace}</span>

      {/* 最近部署状态 */}
      <Tag tone={STATUS_TONE[status]}>{t(PROJECT_STATUS_KEY[status])}</Tag>

      {/* 部署次数 */}
      <span className="font-mono text-[12px] text-ink">
        {item.deployCount}
        <span className="ml-0.5 text-[10px] text-faint">{t('governance.times')}</span>
      </span>

      {/* 最近更新人：人为更新（创建/成功部署）者，系统状态变更不刷新；历史项目可能为空 */}
      <span className="truncate text-[12px] text-mute" title={item.updatedBy}>
        {item.updatedBy || '—'}
      </span>

      {/* 最后更新时间：相对为主，精确放 tooltip；非活跃项目前置标签 */}
      <div className="flex min-w-0 flex-wrap items-center gap-1.5">
        {liveness !== 'active' && (
          <TooltipProvider delayDuration={100}>
            <Tooltip>
              <TooltipTrigger asChild>
                <Tag tone={LIVENESS_TONE[liveness]} dot={false}>
                  {t(LIVENESS_KEY[liveness])}
                </Tag>
              </TooltipTrigger>
              <TooltipContent>
                {/* 行内徽标悬停解释状态含义：与筛选 chip 同文案精确插值（zombie 只 days，dormant 只 min/max） */}
                {liveness === 'zombie'
                  ? t('governance.livenessTipZombie', { days: ZOMBIE_DAYS })
                  : t('governance.livenessTipDormant', { min: ACTIVE_DAYS, max: ZOMBIE_DAYS })}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )}
        <time dateTime={item.updatedAt} title={formatDateTime(item.updatedAt)} className="text-[12px] text-ink">
          {humanizeDateTime(item.updatedAt)}
        </time>
      </div>
    </div>
  )
})

/**
 * 项目治理（管理员后台）
 *
 * 跨命名空间聚合的项目活跃度清单，用于识别「哪些项目已经没有人使用了」：
 * - 顶部三卡统计：项目总数 / 活跃（30 天内更新）/ 僵尸（90 天未更新），数据来自服务端 stats
 * - 关键词搜索 + 活跃度筛选 + 最后更新时间排序均走服务端 query（搜索输入 300ms 防抖），
 *   更新时间列点击切换升/降序（服务端排序）
 * - 无限下拉（服务端分页，滚动触底追加下一页）；行内活跃度标签按同阈值前端推导
 *   （服务端不回传分类），行同时展示仓库/最近提交（判断代码是否仍有人维护）
 * - 数据由 /api/admin/projects/liveness 提供
 */
export function ProjectGovernance() {
  const { t } = useTranslation()
  const [items, setItems] = useState<LivenessItem[]>([])
  const [count, setCount] = useState(0)
  const [stats, setStats] = useState<LivenessStats>({ total: 0, active: 0, dormant: 0, zombie: 0 })
  const [keyword, setKeyword] = useState('')
  // 防抖后的关键词：SearchInput 逐键触发 onChange，避免每次击键都打后端
  const [debouncedKeyword, setDebouncedKeyword] = useState('')
  const [livenessFilter, setLivenessFilter] = useState<LivenessFilter>('all')
  const [page, setPage] = useState(1)
  const [initialLoading, setInitialLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  // 渐入版本号：整页取数成功 +1，RefreshFade 依 key 重挂载列表重播渐入（追加页不重播，避免闪断）
  const [version, setVersion] = useState(0)
  // 最后更新时间排序方向（默认降序 = 最近更新在前），作为 sort 参数交给服务端
  const [timeSort, setTimeSort] = useState<'asc' | 'desc'>('desc')
  // 列表滚动容器（IntersectionObserver 的 root）与底部哨兵
  const scrollRef = useRef<HTMLDivElement>(null)
  const sentinelRef = useRef<HTMLDivElement>(null)
  // 请求忙碌锁：阻止触底加载在请求中/刷新时连环翻页
  const busyRef = useRef(false)
  const refreshingRef = useRef(false)
  // 请求去重基准：记录上次真正发起的 (过滤条件, 页码)，跳过「新条件×旧页码」过期请求与重复请求。
  // 初始值必须「不可能等于」首屏状态——null ≠ filterKey、0 ≠ 1 的 page，
  // 否则首屏挂载被去重判定吞掉，列表永远不请求（骨架屏常驻）
  const lastFilterKeyRef = useRef<string | null>(null)
  const lastPageRef = useRef(0)
  // 最新过滤条件快照（每次渲染同步）：校验在途响应落地时是否仍属当前过滤。
  // 拦截「旧筛选的追加页晚到、被 append 进新筛选结果」的跨条件污染——追加页不可取消，
  // 改落地校验丢弃过期响应
  const filterKeyRef = useRef('')
  const hasMore = items.length < count
  // 有旧数据时的重取 loading：首载 items 为空走骨架，不进遮罩；搜索/筛选/刷新重取则遮罩旧列表
  const refetching = initialLoading && items.length > 0

  // 关键词防抖：输入停顿 300ms 后才把新关键词交给 fetch effect
  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedKeyword(keyword), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [keyword])

  // 过滤条件指纹：搜索 / 活跃度筛选 / 排序方向任一变化都视为新的过滤条件
  //（排序切换走服务端重新拉取，故也纳入指纹）
  const filterKey = `${debouncedKeyword}|${livenessFilter}|${timeSort}`
  // 每渲染同步最新过滤条件（供 fetchList 落地校验，见 filterKeyRef 声明注释）
  filterKeyRef.current = filterKey

  // 过滤条件变化 → 回到顶部 + 重置第 1 页（由 page 变化驱动下面的 fetch）
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: 0 })
    setPage(1)
  }, [filterKey])

  /** 拉取第 p 页活跃度清单：append=true 追加到列表尾（无限滚动），否则整体替换（首屏/搜索/刷新）。
   *  search/liveness/sort 交给服务端，stats 为搜索命中全量统计。 */
  const fetchList = useCallback(
    async (p: number, append: boolean) => {
      busyRef.current = true
      if (append) setLoadingMore(true)
      else setInitialLoading(true)
      try {
        const { data, error: err } = await api.GET(API.adminProjectsLiveness, {
          params: {
            query: {
              page: p,
              pageSize: PAGE_SIZE,
              search: debouncedKeyword.trim() || undefined,
              liveness: livenessFilter === 'all' ? undefined : livenessFilter,
              sort: timeSort,
            },
          },
        })
        if (err) throw new Error(err.message ?? String(err))
        if (!data) return
        // 在途响应落地时若过滤条件已切换（旧筛选的追加页晚到）→ 丢弃，不 append 进新筛选结果
        if (filterKeyRef.current !== filterKey) return
        setError('')
        setItems((prev) => (append ? [...prev, ...data.items] : data.items))
        setCount(data.count)
        setStats(data.stats)
        // 仅整页取数（非追加）重播渐入：追加页复用已有动画，避免整表重挂闪断
        if (!append) setVersion((v) => v + 1)
      } catch (e) {
        // 追加页失败不清空已加载列表，toast 提示即可；首屏/整页失败落 error 空态
        if (append) toast.error(e instanceof Error ? e.message : String(e))
        else setError(e instanceof Error ? e.message : String(e))
      } finally {
        busyRef.current = false
        setInitialLoading(false)
        setLoadingMore(false)
      }
    },
    [debouncedKeyword, livenessFilter, timeSort],
  )

  useEffect(() => {
    // 刷新已直接拉取第 1 页，跳过 setPage(1) 触发的重复请求
    if (page === 1 && refreshingRef.current) return
    // 过滤条件刚变、page 尚未重置到 1（filterKey effect 已 setPage(1)，本 commit 里拿到的还是旧 page）：
    // 跳过这次「新条件×旧页码」的过期请求，真正的第 1 页由 setPage(1) 触发的下一次 effect 承担
    if (lastFilterKeyRef.current !== filterKey && page !== 1) return
    // 同 (filterKey, page) 去重：刷新完成后的 effect 重跑不会重复拉第 1 页
    if (lastFilterKeyRef.current === filterKey && lastPageRef.current === page) return
    lastFilterKeyRef.current = filterKey
    lastPageRef.current = page
    void fetchList(page, page > 1)
  }, [page, filterKey, fetchList])

  /** 手动刷新：回到第 1 页拉最新一版活跃度清单（保留当前搜索/筛选/排序）。
   *  重置期间同步锁 busyRef，阻止触底加载在刷新时连环翻页。 */
  const refresh = async () => {
    if (refreshingRef.current) return
    refreshingRef.current = true
    setRefreshing(true)
    busyRef.current = true
    // 回到顶部：让哨兵离开视口，刷新后不会自动加载后续页
    if (scrollRef.current) scrollRef.current.scrollTop = 0
    setPage(1)
    // 预标记本次刷新要拉的 (filterKey, 1)，刷新完成后的 effect 重跑据此去重
    lastFilterKeyRef.current = filterKey
    lastPageRef.current = 1
    await fetchList(1, false)
    busyRef.current = false
    refreshingRef.current = false
    setRefreshing(false)
  }

  // 底部哨兵进入列表容器视口（含 300px 预加载区）且非忙碌/非刷新时 → 翻下一页（服务端分页）
  useEffect(() => {
    const sentinel = sentinelRef.current
    const root = scrollRef.current
    if (!sentinel || !root) return
    const io = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !busyRef.current && !refreshing) {
          setPage((p) => p + 1)
        }
      },
      { root, rootMargin: '300px' },
    )
    io.observe(sentinel)
    return () => io.disconnect()
  }, [hasMore, initialLoading, loadingMore, refreshing])

  // 活跃度分类基准时间：每轮渲染取一次，行标签与更新时间刻度一致
  const now = new Date()

  const chipCls = (active: boolean) =>
    `cursor-pointer select-none rounded-full border px-3 py-1 text-[12px] transition-colors ${
      active
        ? 'border-primary bg-primary-soft font-medium text-primary'
        : 'border-line bg-surface text-mute hover:border-primary hover:text-primary'
    }`

  return (
    <div className="flex h-full flex-col gap-4">
      {/* 页头：标题 + 搜索 + 刷新 */}
      <div className="flex shrink-0 flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <h2 className="text-[16px] font-semibold text-ink">{t('governance.title')}</h2>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <SearchInput
            value={keyword}
            onChange={setKeyword}
            placeholder={t('governance.searchPlaceholder')}
            className="w-56"
          />
          <span className="text-[12px] text-faint">{t('governance.count', { count })}</span>
          <Button size="sm" variant="outline" disabled={refreshing} onClick={refresh}>
            {refreshing ? (
              <Icon name="loader" className="size-4 animate-spin" />
            ) : (
              <Icon name="refresh" className="size-4" />
            )}
            {t('common.refresh')}
          </Button>
        </div>
      </div>

      {/* 顶部三卡统计（服务端基于搜索命中全量统计） */}
      <div className="grid shrink-0 grid-cols-1 gap-3 sm:grid-cols-3">
        <StatCard label={t('governance.total')} value={stats.total} icon="project" tone="mute" />
        <StatCard label={t('governance.active')} value={stats.active} icon="rocket" tone="ok" />
        <StatCard label={t('governance.zombie')} value={stats.zombie} icon="alert" tone="accent" />
      </div>

      {/* 活跃度筛选标签；chip hover 用 Tooltip 即时解释各状态含义（原生 title 有延迟且无样式，用户感知不到） */}
      <TooltipProvider delayDuration={100}>
        <div className="flex shrink-0 flex-wrap items-center gap-1.5">
          <span className="text-[12px] text-faint">{t('governance.filterLiveness')}</span>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                onClick={() => setLivenessFilter('all')}
                className={chipCls(livenessFilter === 'all')}
              >
                {t('governance.all')}
              </button>
            </TooltipTrigger>
            <TooltipContent>{t('governance.livenessTipAll')}</TooltipContent>
          </Tooltip>
          {LIVENESS_ORDER.map((l) => (
            <Tooltip key={l}>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  onClick={() => setLivenessFilter(l)}
                  className={chipCls(livenessFilter === l)}
                >
                  {t(LIVENESS_KEY[l])}
                </button>
              </TooltipTrigger>
              <TooltipContent>
                {/* 按活跃度分类语义精确传插值参数（active/zombie 只 days，dormant 只 min/max） */}
                {l === 'active'
                  ? t('governance.livenessTipActive', { days: ACTIVE_DAYS })
                  : l === 'dormant'
                    ? t('governance.livenessTipDormant', { min: ACTIVE_DAYS, max: ZOMBIE_DAYS })
                    : t('governance.livenessTipZombie', { days: ZOMBIE_DAYS })}
              </TooltipContent>
            </Tooltip>
          ))}
        </div>
      </TooltipProvider>

      {/* 项目列表：固定表头 + 内部滚动容器（无限下拉 root） */}
      <section className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-line bg-surface">
        <div className="hidden grid-cols-[1fr_3fr_1fr_6rem_4.5rem_5.5rem_5.5rem] items-center gap-3 border-b border-line px-4 py-2 text-[11px] font-medium text-faint lg:grid">
          <span>{t('governance.project')}</span>
          <span>{t('governance.lastCommit')}</span>
          <span>{t('governance.namespace')}</span>
          <span>{t('governance.status')}</span>
          <span>{t('governance.deployCount')}</span>
          <span>{t('governance.updatedBy')}</span>
          <button
            type="button"
            onClick={() => setTimeSort((v) => (v === 'desc' ? 'asc' : 'desc'))}
            className="flex w-fit items-center gap-0.5 hover:text-ink"
            title={timeSort === 'desc' ? t('governance.sortDesc') : t('governance.sortAsc')}
          >
            {t('governance.lastUpdate')}
            <Icon name={timeSort === 'desc' ? 'chevron-down' : 'chevron-up'} className="size-3 opacity-70" />
          </button>
        </div>

        <div ref={scrollRef} className="relative min-h-0 flex-1 overflow-y-auto">
        {/* 内容区：重取遮罩时降透明度并禁止交互，保持旧帧不闪断（首载骨架不遮罩） */}
        <div className={refetching ? 'pointer-events-none opacity-40' : undefined}>
        {initialLoading && items.length === 0 ? (
          <SkeletonList count={8} bare />
        ) : error && items.length === 0 ? (
          <Empty icon="alert" text={error} />
        ) : items.length === 0 ? (
          <Empty
            icon="project"
            text={keyword ? t('governance.searchEmpty', { kw: keyword.trim() }) : t('common.empty')}
          />
        ) : (
          <RefreshFade version={version}>
          {items.map((p) => (
            <GovernanceRow
              key={p.id}
              item={p}
              status={STATUS_MAP[p.deployStatus]}
              liveness={classifyLiveness(p.updatedAt, now)}
            />
          ))}
          </RefreshFade>
        )}
        {/* 无限下拉哨兵：进入视口翻下一页；无更多数据时显示到底提示 */}
        <div ref={sentinelRef} className="flex h-10 items-center justify-center gap-2">
          {loadingMore ? (
            <span className="flex items-center gap-1.5 text-[12px] text-mute">
              <Icon name="loader" className="animate-spin text-[13px]" />
              {t('common.loadingMore')}
            </span>
          ) : items.length > 0 && !hasMore ? (
            <span className="text-[11px] text-faint">{t('common.noMore')}</span>
          ) : null}
        </div>
        </div>
        {/* 重取遮罩：有旧数据时的 loading（搜索/活跃度筛选/排序/刷新），居中 spinner */}
        {refetching && (
          <div className="absolute inset-0 z-10 flex items-center justify-center bg-surface/50">
            <Icon name="loader" className="size-5 animate-spin text-faint" />
          </div>
        )}
        </div>
      </section>
    </div>
  )
}
