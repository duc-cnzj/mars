import { memo, useEffect, useRef, useState, type CSSProperties } from 'react'
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
import { humanizeDateTime } from '@/lib/humanizeDateTime'
import { api } from '@/api/client'
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

/** 单次拉取上限：项目量级百级，全量拉回后本地排序（对齐服务端内存分页语义） */
const FETCH_LIMIT = 100

/** 无限下拉滚动揭示块大小：一次拉全量后本地排序，滚动到底逐块揭示 */
const CHUNK = 20

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
 *  配合 item 引用稳定，列表状态变化（关键词/活跃度筛选/排序切换/滚动揭示）时行不重渲染。 */
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
  return (
    <div
      className={`grid grid-cols-1 gap-1 border-b border-line px-4 py-2.5 last:border-b-0 sm:grid-cols-2 lg:grid-cols-[1fr_1fr_6rem_5rem_1fr] lg:items-center lg:gap-3 ${className ?? ''}`}
      style={style}
    >
      {/* 项目：名称 + 仓库·分支@commit + 最近提交人（判断代码是否仍有人维护） */}
      <div className="min-w-0">
        <div className="truncate text-[13px] font-medium text-ink">{item.name}</div>
        <div className="truncate font-mono text-[11px] text-faint">
          {item.repo}
          {item.gitBranch && item.gitCommit ? ` · ${item.gitBranch}@${item.gitCommit}` : ''}
        </div>
        {item.gitCommitAuthor && (
          <div className="truncate text-[11px] text-mute" title={item.gitCommitTitle}>
            <span className="font-medium text-ink">{item.gitCommitAuthor}</span>
            {item.gitCommitTitle && <span className="text-faint"> · {item.gitCommitTitle}</span>}
            {item.gitCommitDate && <span className="text-faint"> · {humanizeDateTime(item.gitCommitDate)}</span>}
          </div>
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
 * - 关键词搜索 + 活跃度筛选均走服务端 query（输入 300ms 防抖），更新时间列可点击升降序切换
 * - 数据由 /api/admin/projects/liveness 提供；行内活跃度标签按同阈值前端推导（服务端不回传分类），
 *   行同时展示仓库/最近提交（判断代码是否仍有人维护）
 */
export function ProjectGovernance() {
  const { t } = useTranslation()
  const [keyword, setKeyword] = useState('')
  // 防抖后的关键词：SearchInput 逐键触发 onChange，避免每次击键都打后端
  const [debouncedKeyword, setDebouncedKeyword] = useState('')
  const [livenessFilter, setLivenessFilter] = useState<LivenessFilter>('all')
  const [items, setItems] = useState<LivenessItem[]>([])
  const [stats, setStats] = useState<LivenessStats>({ total: 0, active: 0, dormant: 0, zombie: 0 })
  const [count, setCount] = useState(0)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  // 手动刷新计数：递增触发 fetch effect 重跑
  const [reloadKey, setReloadKey] = useState(0)
  // 渐入版本号：取数成功 +1，RefreshFade 依 key 重挂载列表重播渐入
  const [version, setVersion] = useState(0)
  // 更新时间排序方向（默认降序 = 最近更新在前）
  const [timeSort, setTimeSort] = useState<'asc' | 'desc'>('desc')
  // 滚动揭示量：只渲染 sorted.slice(0, visible)，滚动到底 +CHUNK（客户端无限下拉）
  const [visible, setVisible] = useState(CHUNK)
  // 列表滚动容器（IntersectionObserver 的 root）与底部哨兵
  const scrollRef = useRef<HTMLDivElement>(null)
  const sentinelRef = useRef<HTMLDivElement>(null)

  // 关键词防抖：输入停顿 300ms 后才把新关键词交给 fetch effect
  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedKeyword(keyword), 300)
    return () => window.clearTimeout(timer)
  }, [keyword])

  // 拉取活跃度清单：搜索/活跃度过滤交给服务端，stats 为搜索命中全量统计
  useEffect(() => {
    let ignore = false
    setLoading(true)
    void api
      .GET('/api/admin/projects/liveness', {
        params: {
          query: {
            page: 1,
            pageSize: FETCH_LIMIT,
            search: debouncedKeyword.trim() || undefined,
            liveness: livenessFilter === 'all' ? undefined : livenessFilter,
          },
        },
      })
      .then(({ data, error: err }) => {
        if (ignore) return
        if (err) {
          setError(err.message ?? String(err))
          setItems([])
          return
        }
        setError('')
        if (!data) return
        setItems(data.items)
        setCount(data.count)
        setStats(data.stats)
        // 取数成功 → 版本号 +1，整表重播一次渐入（keep-last-frame，不闪断）
        setVersion((v) => v + 1)
      })
      .finally(() => {
        if (!ignore) {
          setLoading(false)
          setRefreshing(false)
        }
      })
    return () => {
      ignore = true
    }
  }, [debouncedKeyword, livenessFilter, reloadKey])

  /** 手动刷新：拉取最新一版活跃度清单 */
  const refresh = () => {
    setRefreshing(true)
    setReloadKey((k) => k + 1)
  }

  // 搜索 / 活跃度筛选 / 手动刷新 / 排序方向变化 → 回顶部并重置揭示量
  //（排序切换走本地重排，无需重新拉取，故不进入 fetch effect 依赖）
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: 0 })
    setVisible(CHUNK)
  }, [debouncedKeyword, livenessFilter, reloadKey, timeSort])

  const chipCls = (active: boolean) =>
    `cursor-pointer select-none rounded-full border px-3 py-1 text-[12px] transition-colors ${
      active
        ? 'border-primary bg-primary-soft font-medium text-primary'
        : 'border-line bg-surface text-mute hover:border-primary hover:text-primary'
    }`

  /** 全量按最后更新时间排序（默认最近更新在前，僵尸自然沉底） */
  const sorted = [...items].sort((a, b) => {
    const diff = new Date(a.updatedAt).getTime() - new Date(b.updatedAt).getTime()
    return timeSort === 'desc' ? -diff : diff
  })

  // 滚动揭示切片 + 是否还有未揭示行（揭示完毕哨兵显示「没有更多了」）
  const revealed = sorted.slice(0, visible)
  const hasMore = visible < sorted.length
  // 服务端返回条数少于总条数 → 列表被后端截断（如搜索命中超过服务端上限）：
  // 哨兵不再显示「没有更多了」误导，改提示仅展示前 N 条
  const truncated = count > items.length
  // 有旧数据时的重取 loading：首载 items 为空走骨架，不进遮罩；搜索/筛选/刷新重取则遮罩旧列表
  const refetching = loading && items.length > 0

  // 底部哨兵进入列表视口（含 300px 预加载区）→ 揭示下一块（客户端无限下拉）
  useEffect(() => {
    const sentinel = sentinelRef.current
    const root = scrollRef.current
    if (!sentinel || !root) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) setVisible((v) => Math.min(v + CHUNK, sorted.length))
      },
      { root, rootMargin: '300px' },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [sorted.length])

  // 活跃度分类基准时间：每轮渲染取一次，行标签与更新时间刻度一致
  const now = new Date()

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
        <div className="hidden grid-cols-[1fr_1fr_6rem_5rem_1fr] items-center gap-3 border-b border-line px-4 py-2 text-[11px] font-medium text-faint lg:grid">
          <span>{t('governance.project')}</span>
          <span>{t('governance.namespace')}</span>
          <span>{t('governance.status')}</span>
          <span>{t('governance.deployCount')}</span>
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
        {loading && items.length === 0 ? (
          <SkeletonList count={8} bare />
        ) : error ? (
          <Empty icon="alert" text={error} />
        ) : sorted.length === 0 ? (
          <Empty
            icon="project"
            text={keyword ? t('governance.searchEmpty', { kw: keyword.trim() }) : t('common.empty')}
          />
        ) : (
          <RefreshFade version={version}>
          {revealed.map((p) => (
            <GovernanceRow
              key={p.id}
              item={p}
              status={STATUS_MAP[p.deployStatus]}
              liveness={classifyLiveness(p.updatedAt, now)}
            />
          ))}
          </RefreshFade>
        )}
        {/* 无限下拉哨兵：进入视口揭示下一块；揭示完毕显示「没有更多了」 */}
        <div ref={sentinelRef} className="flex h-10 items-center justify-center gap-2">
          {hasMore ? (
            <Icon name="loader" className="size-3.5 animate-spin text-faint" />
          ) : truncated ? (
            // 后端截断：不谎报「没有更多」，如实说明仅展示前 N 条（warn 提示治理可见范围）
            <span className="text-[11px] text-warn">
              {t('governance.truncated', { shown: items.length, total: count })}
            </span>
          ) : (
            <span className="text-[11px] text-faint">{t('common.noMore')}</span>
          )}
        </div>
        </div>
        {/* 重取遮罩：有旧数据时的 loading（搜索/活跃度筛选/刷新），居中 spinner */}
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
