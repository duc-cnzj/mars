import { memo, useCallback, useEffect, useRef, useState, type CSSProperties } from 'react'
import { useTranslation } from 'react-i18next'
import { useSearchParams } from 'react-router-dom'
import { toast } from '@/lib/toast'
import { SEARCH_DEBOUNCE_MS } from '@/lib/constants'
import { Icon } from '@/components/Icons'
import { SearchInput } from '@/components/SearchInput'
import { Empty, RefreshFade, SkeletonList, Tag, type Tone } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { Switch } from '@/components/ui/shadcn/switch'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/shadcn/popover'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/shadcn/tooltip'
import { formatDateTime } from '@/lib/format'
import { humanizeDateTime } from '@/lib/humanizeDateTime'
import { MemberInput } from '@/features/workbench/MemberInput'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import type { components } from '@/api/schema'
import type { TKey } from '@/i18n/keys'

type AdminItem = components['schemas']['namespace.AdminItem']

/** 每页条数（服务端分页，滚动触底追加下一页） */
const PAGE_SIZE = 15

/** 活跃度筛选：'all' = 不筛选（全部命名空间） */
type LivenessFilter = 'active' | 'dormant' | 'zombie' | 'all'

/** 活跃度分类（对齐服务端 LivenessKind 枚举字面量） */
type Liveness = 'active' | 'dormant' | 'zombie'

/** 活跃度 → 标签色：活跃 ok（常态）/ 低活跃 warn / 僵尸 err（建议治理） */
const LIVENESS_TONE: Record<Liveness, Tone> = {
  active: 'ok',
  dormant: 'warn',
  zombie: 'err',
}

/** 活跃度 → i18n 词条键 */
const LIVENESS_KEY: Record<Liveness, TKey> = {
  active: 'namespaces.livenessActive',
  dormant: 'namespaces.livenessDormant',
  zombie: 'namespaces.livenessZombie',
}

/** 活跃度筛选展示顺序：活跃在前，僵尸在后（与治理优先级呼应） */
const LIVENESS_ORDER: readonly Liveness[] = ['active', 'dormant', 'zombie']

/** 活跃度阈值（天）：与项目治理同阈值（服务端分类，前端仅 tooltip 展示用） */
const ACTIVE_DAYS = 30
const ZOMBIE_DAYS = 90

/** 成员展示列表：把所有者本人计入人数。member 表只存额外成员（Create 只写 creatorEmail，
 *  所有者不在其中），展示层补上 owner 行保证「人数算上所有者本人」；owner 已在 members 时原样返回。
 *  仅影响展示（计数 + 成员下钻列表），管理弹窗仍消费原始 ns.members（不含 owner，防误删本人）。 */
function memberDisplayList(
  members: components['schemas']['types.MemberModel'][],
  creatorEmail: string,
): components['schemas']['types.MemberModel'][] {
  return members.some((m) => m.email === creatorEmail) ? members : [{ id: 0, email: creatorEmail }, ...members]
}

/**
 * 单行命名空间（React.memo）：行数据与回调引用均未变则跳过整行重渲——
 * 搜索击键、筛选切换、触底翻页、弹窗开关等页面级 state 变化不再连带全部已加载行重渲，
 * 只重渲数据实际变化的行（无限下拉累积上百行时避免每次击键 O(n) 行重渲）。
 * 成员/项目下钻 Popover 与操作按钮均内联在本行，行 memo 后计算只在行数据变化时发生。
 */
const NamespaceRow = memo(function NamespaceRow({
  item,
  onManage,
  className,
  style,
}: {
  item: AdminItem
  onManage: (item: AdminItem) => void
  /** RefreshFade 经 cloneElement 注入的渐入 class/延迟——须转发到根元素才生效 */
  className?: string
  style?: CSSProperties
}) {
  const { t } = useTranslation()
  const { ns, lastActiveAt, livenessKind } = item
  // 成员展示列表（补上所有者本人）：属性计数与成员下钻共用，保证「人数算上所有者」
  const memberList = memberDisplayList(ns.members, ns.creatorEmail)
  // 后端 livenessKind 是自由 string（schema 未收成枚举字面量）：防御性归一化到三态，
  // 未知/空值一律按 active 展示——避免陌生字面量击穿 LIVENESS_TONE/LIVENESS_KEY 索引产生无 tone 标签
  const kind: Liveness =
    livenessKind === 'dormant' || livenessKind === 'zombie' ? livenessKind : 'active'
  return (
    <div
      className={`grid grid-cols-1 gap-2 border-b border-line px-4 py-2.5 last:border-b-0 sm:grid-cols-2 lg:grid-cols-[minmax(0,2fr)_minmax(0,1.5fr)_7rem_9rem_7rem_6rem] lg:items-center ${className ?? ''}`}
      style={style}
    >
      {/* 命名空间：名称 + 私有 tag + 描述 */}
      <div className="flex min-w-0 flex-col gap-0.5">
        <div className="flex items-center gap-1.5">
          <span className="truncate font-mono text-[13px] font-medium text-ink">{ns.name}</span>
          {ns.private && (
            <Tag tone="accent" dot={false}>
              {t('namespaces.privateTag')}
            </Tag>
          )}
        </div>
        <span className="truncate text-[11px] text-faint">{ns.description || '-'}</span>
      </div>

      {/* 创建者邮箱 */}
      <div className="flex min-w-0 items-center">
        <span className="truncate font-mono text-[13px] text-ink">{ns.creatorEmail}</span>
      </div>

      {/* 属性：点「成员 · 项目」下钻查看具体成员与项目 */}
      <Popover>
        <PopoverTrigger asChild>
          <button
            type="button"
            aria-label={t('namespaces.viewDetail')}
            className="w-fit cursor-pointer text-left text-[12px] text-mute underline decoration-dotted underline-offset-2 transition-colors hover:text-ink"
          >
            {t('namespaces.propsCount', {
              members: memberList.length,
              projects: ns.projects.length,
            })}
          </button>
        </PopoverTrigger>
        <PopoverContent align="start" className="w-80">
          <div className="grid gap-3">
            {/* 成员列表：邮箱展示，创建者带标记 */}
            <div>
              <div className="mb-1.5 text-[11px] font-medium text-faint">
                {t('namespaces.members')}（{memberList.length}）
              </div>
              <ul className="max-h-40 overflow-y-auto rounded-md border border-line">
                {memberList.map((m) => (
                  <li
                    key={m.id}
                    className="flex items-center justify-between gap-2 border-b border-line px-2 py-1 last:border-b-0"
                  >
                    <span className="truncate font-mono text-[12px] text-ink">{m.email}</span>
                    {m.email === ns.creatorEmail && (
                      <Tag tone="accent" dot={false} className="shrink-0">
                        {t('namespaces.owner')}
                      </Tag>
                    )}
                  </li>
                ))}
              </ul>
            </div>
            {/* 项目列表：chips 展示 */}
            <div>
              <div className="mb-1.5 text-[11px] font-medium text-faint">
                {t('namespaces.projects')}（{ns.projects.length}）
              </div>
              {ns.projects.length === 0 ? (
                <span className="text-[12px] text-faint">{t('namespaces.projectsEmpty')}</span>
              ) : (
                <div className="flex flex-wrap gap-1">
                  {ns.projects.map((p) => (
                    <span
                      key={p.id}
                      className="rounded-md bg-raised px-1.5 py-0.5 font-mono text-[11px] text-mute"
                    >
                      {p.name}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>
        </PopoverContent>
      </Popover>

      {/* 最近活跃时间：空间下所有项目 UpdatedAt 最大值；无项目（从未部署）显示「从未活跃」。
          非活跃空间前置活跃度标签（活跃是常态不标，对齐项目治理的行内语义） */}
      <span className="text-[11px] text-ink">
        {lastActiveAt ? (
          <span className="flex flex-wrap items-center gap-1.5">
            {kind !== 'active' && (
              <TooltipProvider delayDuration={100}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Tag tone={LIVENESS_TONE[kind]} dot={false}>
                      {t(LIVENESS_KEY[kind])}
                    </Tag>
                  </TooltipTrigger>
                  <TooltipContent>
                    {/* 行内徽标悬停解释状态含义：与筛选 chip 同文案精确插值（zombie 只 days，dormant 只 min/max） */}
                    {kind === 'zombie'
                      ? t('namespaces.livenessTipZombie', { days: ZOMBIE_DAYS })
                      : t('namespaces.livenessTipDormant', { min: ACTIVE_DAYS, max: ZOMBIE_DAYS })}
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            )}
            <time dateTime={lastActiveAt} title={formatDateTime(lastActiveAt)}>
              {humanizeDateTime(lastActiveAt)}
            </time>
          </span>
        ) : (
          <span className="text-faint">{t('namespaces.lastActiveEmpty')}</span>
        )}
      </span>

      {/* 创建时间：humanize 相对时间 + 精确时间 tooltip */}
      <span className="text-[11px] text-ink">
        <time dateTime={ns.createdAt} title={formatDateTime(ns.createdAt)}>
          {humanizeDateTime(ns.createdAt)}
        </time>
      </span>

      {/* 操作：管理（私有/成员/转让）——admin 对任意空间可操作 */}
      <div className="flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon-xs"
          onClick={() => onManage(item)}
          aria-label={t('namespaces.manage')}
          title={t('namespaces.manage')}
          className="text-faint hover:text-primary"
        >
          <Icon name="gear" className="size-4" />
        </Button>
      </div>
    </div>
  )
})

/**
 * 命名空间全局管理（管理员后台）
 *
 * 管理员视角查看「所有」命名空间（工作台只展示当前用户可访问的「我的空间」）：
 * - 按名称/创建者搜索 + 「只看私有」一键过滤 + 活跃度分类筛选（均走服务端 query，输入 300ms 防抖）
 * - 成员/项目下钻：点「N 人 · M 项目」弹 Popover 查看具体成员（邮箱）与项目
 * - 无限下拉（服务端分页，滚动触底追加）；每行提供「管理」操作——admin 对任意空间可编辑
 *   （后端 RequireNamespaceOwner 对 admin 绕过 owner 校验），管理弹窗私有/成员/转让一次提交
 *   update_config
 * - ?ns= 预选：从空间资源页跳转过来时把该空间作为初始搜索词（消费后清参，刷新不重复预选）
 * - 数据由 /api/admin/namespaces 提供
 */
export function NamespaceManager() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()
  // ?ns= 预选：从空间资源页「查看该空间」跳转带入搜索词，消费后即从 URL 清除
  const initNs = searchParams.get('ns') ?? ''
  const [items, setItems] = useState<AdminItem[]>([])
  const [count, setCount] = useState(0)
  const [keyword, setKeyword] = useState(initNs)
  // 防抖后的关键词：SearchInput 逐键触发 onChange，避免每次击键都打后端
  const [debouncedKeyword, setDebouncedKeyword] = useState(initNs)
  const [privateOnly, setPrivateOnly] = useState(false)
  // 活跃度分类筛选：'all' 不传 liveness（服务端返回全部）
  const [liveness, setLiveness] = useState<LivenessFilter>('all')
  const [page, setPage] = useState(1)
  const [initialLoading, setInitialLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  // 渐入版本号：整页取数成功 +1，RefreshFade 依 key 重挂载列表重播渐入（追加页不重播，避免闪断）
  const [version, setVersion] = useState(0)
  // 管理弹窗当前目标：null=关闭；非空=该空间配置弹窗打开（单例弹窗，复用 NamespaceCard 的管理交互）
  const [manageNs, setManageNs] = useState<AdminItem | null>(null)
  // 无限下拉哨兵：进入列表容器视口（提前 300px）即翻下一页
  const sentinelRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  // 请求忙碌锁：阻止触底加载在请求中/刷新时连环翻页
  const busyRef = useRef(false)
  const refreshingRef = useRef(false)
  // 请求去重基准：记录上次真正发起的 (过滤条件, 页码)，跳过「新条件×旧页码」过期请求与重复请求。
  // 初始值必须「不可能等于」首屏状态——null ≠ filterKey、0 ≠ 1 的 page，
  // 否则首屏挂载被去重判定吞掉，列表永远不请求（骨架屏常驻）
  const lastFilterKeyRef = useRef<string | null>(null)
  const lastPageRef = useRef(0)
  // 最新过滤条件快照（每次渲染同步）：校验在途响应落地时是否仍属当前过滤。
  // 拦截「旧筛选的追加页晚到、被 append 进新筛选结果」的跨条件污染——ProjectGovernance 用
  // ignore 取消在途，本页追加页不可取消，改落地校验丢弃过期响应
  const filterKeyRef = useRef('')
  const nsParamConsumedRef = useRef(false)
  const hasMore = items.length < count
  // 有旧数据时的重取 loading：首载 items 为空走骨架，不进遮罩；搜索/筛选/刷新重取则遮罩旧列表
  const refetching = initialLoading && items.length > 0

  // 消费 ?ns= 预选参数：写进搜索词后立即从 URL 清除（replace），刷新/分享链接不再重复预选
  useEffect(() => {
    if (nsParamConsumedRef.current) return
    nsParamConsumedRef.current = true
    if (!initNs) return
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        next.delete('ns')
        return next
      },
      { replace: true },
    )
  }, [initNs, setSearchParams])

  // 关键词防抖：输入停顿 300ms 后才把新关键词交给 fetch effect
  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedKeyword(keyword), SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [keyword])

  // 过滤条件指纹：搜索 / 只看私有 / 活跃度任一变化都视为新的过滤条件
  const filterKey = `${debouncedKeyword}|${privateOnly}|${liveness}`
  // 每渲染同步最新过滤条件（供 fetchList 落地校验，见 filterKeyRef 声明注释）
  filterKeyRef.current = filterKey

  // 过滤条件变化 → 回到顶部 + 重置第 1 页（由 page 变化驱动下面的 fetch）
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: 0 })
    setPage(1)
  }, [filterKey])

  /** 拉取第 p 页：append=true 追加到列表尾（无限滚动），否则整体替换（首屏/搜索/刷新）。 */
  const fetchList = useCallback(
    async (p: number, append: boolean) => {
      busyRef.current = true
      if (append) setLoadingMore(true)
      else setInitialLoading(true)
      try {
        const { data, error: err } = await api.GET(API.adminNamespaces, {
          params: {
            query: {
              page: p,
              pageSize: PAGE_SIZE,
              search: debouncedKeyword.trim() || undefined,
              privateOnly: privateOnly || undefined,
              liveness: liveness === 'all' ? undefined : liveness,
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
    [debouncedKeyword, privateOnly, liveness],
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

  // 触底加载：sentinel 进入列表容器视口（提前 300px）且非忙碌/非刷新时翻下一页
  useEffect(() => {
    const el = sentinelRef.current
    const root = scrollRef.current
    if (!el || !root) return
    const io = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !busyRef.current && !refreshing) {
          setPage((p) => p + 1)
        }
      },
      { root, rootMargin: '300px' },
    )
    io.observe(el)
    return () => io.disconnect()
  }, [hasMore, initialLoading, loadingMore, refreshing])

  /** 手动刷新：回到第 1 页拉最新一版命名空间（保留当前搜索/过滤）。
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

  // 行操作回调：引用稳定（useCallback），React.memo 行据此跳过不相关重渲
  const handleManage = useCallback((item: AdminItem) => setManageNs(item), [])

  const chipCls = (active: boolean) =>
    `cursor-pointer select-none rounded-full border px-3 py-1 text-[12px] transition-colors ${
      active
        ? 'border-primary bg-primary-soft font-medium text-primary'
        : 'border-line bg-surface text-mute hover:border-primary hover:text-primary'
    }`

  return (
    <div className="flex h-full flex-col gap-4">
      {/* 页头 + 工具栏 + 筛选（shrink-0，固定不随列表滚动） */}
      <div className="flex shrink-0 flex-col gap-3">
        {/* 页头：标题 + 搜索 + 刷新 */}
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2.5">
            <h2 className="text-[16px] font-semibold text-ink">{t('namespaces.title')}</h2>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {/* 搜索统一放标题行右上角（对齐 Events/治理/工作台） */}
            <SearchInput
              value={keyword}
              onChange={setKeyword}
              placeholder={t('namespaces.searchPlaceholder')}
              className="w-72"
            />
            <Button variant="outline" size="sm" onClick={refresh} disabled={refreshing}>
              {refreshing ? (
                <Icon name="loader" className="size-4 animate-spin" />
              ) : (
                <Icon name="refresh" className="size-4" />
              )}
              {t('common.refresh')}
            </Button>
          </div>
        </div>

        {/* 工具栏：只看私有 + 结果计数（SearchInput 内置 ⌘K 聚焦快捷键，已上移到标题行） */}
        <div className="flex flex-wrap items-center gap-3">
          <Button
            size="sm"
            variant={privateOnly ? 'default' : 'outline'}
            aria-pressed={privateOnly}
            onClick={() => setPrivateOnly((v) => !v)}
          >
            <Icon name="project" className="size-3.5" />
            {t('namespaces.filterPrivate')}
          </Button>
          <span className="text-[12px] text-faint">{t('namespaces.resultCount', { count })}</span>
        </div>

        {/* 活跃度筛选标签：与项目治理同阈值语义（服务端分类，过滤走 query）；
            chip hover 用 Tooltip 即时解释各状态含义（原生 title 有延迟且无样式，用户感知不到） */}
        <TooltipProvider delayDuration={100}>
          <div className="flex flex-wrap items-center gap-1.5">
            <span className="text-[12px] text-faint">{t('namespaces.filterLiveness')}</span>
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  type="button"
                  onClick={() => setLiveness('all')}
                  className={chipCls(liveness === 'all')}
                >
                  {t('namespaces.all')}
                </button>
              </TooltipTrigger>
              <TooltipContent>{t('namespaces.livenessTipAll')}</TooltipContent>
            </Tooltip>
            {LIVENESS_ORDER.map((l) => (
              <Tooltip key={l}>
                <TooltipTrigger asChild>
                  <button
                    type="button"
                    onClick={() => setLiveness(l)}
                    className={chipCls(liveness === l)}
                  >
                    {t(LIVENESS_KEY[l])}
                  </button>
                </TooltipTrigger>
                <TooltipContent>
                  {/* 按活跃度分类语义精确传插值参数（active/zombie 只 days，dormant 只 min/max） */}
                  {l === 'active'
                    ? t('namespaces.livenessTipActive', { days: ACTIVE_DAYS })
                    : l === 'dormant'
                      ? t('namespaces.livenessTipDormant', { min: ACTIVE_DAYS, max: ZOMBIE_DAYS })
                      : t('namespaces.livenessTipZombie', { days: ZOMBIE_DAYS })}
                </TooltipContent>
              </Tooltip>
            ))}
          </div>
        </TooltipProvider>
      </div>

      {/* 命名空间列表：表头固定 + 列表内部滚动（滚动容器作为哨兵 root） */}
      <section className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-line bg-surface">
        <div className="hidden grid-cols-[minmax(0,2fr)_minmax(0,1.5fr)_7rem_9rem_7rem_6rem] items-center gap-2 border-b border-line px-4 py-2 text-[11px] font-medium text-faint lg:grid">
          <span>{t('namespaces.namespace')}</span>
          <span>{t('namespaces.owner')}</span>
          <span>{t('namespaces.props')}</span>
          <span>{t('namespaces.lastActive')}</span>
          <span>{t('namespaces.createdAt')}</span>
          <span>{t('namespaces.action')}</span>
        </div>

        <div ref={scrollRef} className="relative min-h-0 flex-1 overflow-y-auto">
        {/* 内容区：重取遮罩时降透明度并禁止交互，保持旧帧不闪断（首载骨架不遮罩） */}
        <div className={refetching ? 'pointer-events-none opacity-40' : undefined}>
          {initialLoading && items.length === 0 ? (
            <SkeletonList count={8} bare />
          ) : error && items.length === 0 ? (
            <div className="flex min-h-0 items-center justify-center p-8">
              <Empty icon="namespace" text={error} />
            </div>
          ) : items.length === 0 ? (
            <div className="flex min-h-0 items-center justify-center p-8">
              <Empty
                icon="namespace"
                text={keyword ? t('namespaces.searchEmpty', { kw: keyword.trim() }) : t('common.empty')}
              />
            </div>
          ) : (
            <RefreshFade version={version}>
            {items.map((item) => (
              <NamespaceRow
                key={item.ns.id}
                item={item}
                onManage={handleManage}
              />
            ))}
            </RefreshFade>
          )}

          {/* 无限下拉哨兵：滚到底自动加载下一页；无更多数据时显示到底提示 */}
          <div ref={sentinelRef} className="flex h-10 items-center justify-center gap-2">
            {loadingMore ? (
              <span className="flex items-center gap-1.5 text-[12px] text-mute">
                <Icon name="loader" className="animate-spin text-[13px]" />
                {t('common.loadingMore')}
              </span>
            ) : items.length > 0 && !hasMore ? (
              <span className="text-[12px] text-faint">{t('common.noMore')}</span>
            ) : null}
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

      {/* 空间配置管理弹窗（单例）：私有/成员/转让，提交 update_config 后刷新列表 */}
      <ManageDialog
        item={manageNs}
        open={manageNs !== null}
        onOpenChange={(o) => !o && setManageNs(null)}
        onSaved={refresh}
      />
    </div>
  )
}

/**
 * 空间配置管理弹窗（管理员视角）：私有开关 + 成员标签输入 + 转让管理员，一次提交 update_config。
 * 表单字段仅在弹窗打开瞬间从当前 ns 快照（避免父级刷新 ns 时冲掉未保存的编辑），
 * 与工作台 NamespaceCard 的管理弹窗同一交互约定。
 */
function ManageDialog({
  item,
  open,
  onOpenChange,
  onSaved,
}: {
  item: AdminItem | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const [isPrivate, setIsPrivate] = useState(false)
  const [membersList, setMembersList] = useState<string[]>([])
  const [transferEmail, setTransferEmail] = useState('')
  const [saving, setSaving] = useState(false)

  // 弹窗打开瞬间从当前 ns 快照表单字段，打开期间父级刷新 ns 不清空未保存的编辑
  useEffect(() => {
    if (!open || !item) return
    setIsPrivate(item.ns.private)
    setMembersList(item.ns.members.map((m) => m.email))
    setTransferEmail('')
  }, [open, item])

  /** 一次提交全部空间配置（私有/成员/转让），后端 update_config 单事务原子落库 */
  const save = async () => {
    if (!item || saving) return
    setSaving(true)
    try {
      const email = transferEmail.trim()
      const { error } = await api.POST(API.namespacesUpdateConfig, {
        body: {
          id: item.ns.id,
          private: isPrivate,
          emails: membersList,
          ...(email ? { newAdminEmail: email } : {}),
        },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('namespaces.manageSaved'))
      onOpenChange(false)
      onSaved()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onOpenChange(false)}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-[15px]">
            <Icon name="gear" className="text-[14px]" />
            {t('namespaces.manage')}
            <span className="font-mono text-[12px] text-mute">· {item?.ns.name}</span>
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-5 py-1">
          {/* 私有 */}
          <div className="space-y-1.5">
            <div className="text-[12px] text-mute">{t('namespaces.privateLabel')}</div>
            <div className="flex items-center justify-between rounded-md border border-line px-3 py-2">
              <span className="text-[13px] text-ink">{t('namespaces.private')}</span>
              <Switch checked={isPrivate} onCheckedChange={setIsPrivate} />
            </div>
            <p className="text-[11px] text-faint">{t('namespaces.privateTip')}</p>
          </div>

          {/* 成员 */}
          <div className="space-y-1.5">
            <label className="text-[12px] text-mute">{t('namespaces.membersLabel')}</label>
            <MemberInput
              value={membersList}
              onChange={setMembersList}
              placeholder={t('namespaces.membersPlaceholder')}
            />
            <p className="text-[11px] text-faint">{t('namespaces.membersTip')}</p>
          </div>

          {/* 转让管理员 */}
          <div className="space-y-1.5">
            <label className="text-[12px] text-mute">{t('namespaces.transferLabel')}</label>
            <Input
              value={transferEmail}
              onChange={(e) => setTransferEmail(e.target.value)}
              placeholder={t('namespaces.transferPlaceholder')}
            />
            <p className="text-[11px] text-faint">
              {t('namespaces.transferTip', { email: item?.ns.creatorEmail ?? '' })}
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={saving}>
            {t('common.cancel')}
          </Button>
          <Button onClick={save} disabled={saving}>
            {saving && <Icon name="loader" className="size-4 animate-spin" />}
            {t('common.save')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
