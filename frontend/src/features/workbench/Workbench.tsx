import { lazy, memo, Suspense, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronLeftIcon, ChevronRightIcon, GripVertical, Loader2, RefreshCw } from 'lucide-react'
import {
  DndContext,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import type { DragEndEvent } from '@dnd-kit/core'
import {
  SortableContext,
  useSortable,
  arrayMove,
  rectSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import type { components } from '../../api/schema'
import { api } from '../../api/client'
import { toast } from '@/lib/toast'
import { Icon } from '../../components/icons'
import { Empty, SkeletonGrid } from '../../components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Pagination, PaginationContent, PaginationEllipsis, PaginationItem, PaginationLink } from '@/components/ui/shadcn/pagination'
import { cn } from '@/lib/utils'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/shadcn/tabs'
import { SearchInput } from '../../components/SearchInput'
import { AddNamespaceModal } from './AddNamespaceModal'
import { NamespaceCard } from './NamespaceCard'
import { useWebsocket } from '../../realtime/useWebsocket'
// 项目详情弹窗静态依赖 TabEdit→CodeEditor(CodeMirror 620KB)/DiffViewer(react-diff-viewer)/prism，
// 懒加载后这些重依赖延迟到真正点开项目卡片时才拉取，不进工作台首屏。
const ProjectDetailModal = lazy(() =>
  import('../projects/ProjectDetailModal').then((m) => ({ default: m.ProjectDetailModal })),
)

type NamespaceModel = components['schemas']['types.NamespaceModel']
type ProjectModel = components['schemas']['types.ProjectModel']
type TabKey = 'all' | 'favorite'
/** 已打开的项目详情弹窗条目：完整 ProjectModel（渲染/初始 tab 用）+ 所在空间名 */
type OpenEntry = {
  project: ProjectModel
  namespaceName: string
}

const PAGE_SIZE = 12

// 拖拽性能：NamespaceCard 是重组件（弹窗/工具提示/项目行），memo 后拖拽过程的
// 反复渲染只落在外层 dnd-kit 节点上，卡片本体引用不变则跳过重渲染。
const MemoizedNamespaceCard = memo(NamespaceCard)

// 平台判断：macOS 显示 ⌘，Windows/Linux 显示 Ctrl（快捷键两者都支持，与 SearchInput 一致）
const isMac = /Mac|iPod|iPhone|iPad/i.test(navigator.platform || navigator.userAgent)

/** 工作台 Tab 选择持久化 key（与旧版 AppContent 的 'active-tabs' 一致） */
const TABS_KEY = 'active-tabs'

type PageItem = number | '...'

/** 经典省略号分页：首尾固定两页 + 当前页±1，中间用省略号折叠（如 1 2 … 5 6 7 … 99 100）。
 * 总页数 ≤ 5 时全量展示——省略号折叠不省空间，直接列全。 */
function buildPages(page: number, total: number): PageItem[] {
  if (total <= 5) return Array.from({ length: total }, (_, i) => i + 1)
  const items: PageItem[] = []
  const start = Math.max(3, page - 1)
  const end = Math.min(total - 2, page + 1)
  items.push(1, 2)
  if (start > 3) items.push('...')
  for (let i = start; i <= end; i += 1) items.push(i)
  if (end < total - 2) items.push('...')
  items.push(total - 1, total)
  return items
}

/**
 * 工作台：命名空间列表页。
 * 「全部」Tab 分页加载（页码 URL 化：刷新停留在原页）；「关注」Tab 无限下拉（滚动到底自动追加下一页）。
 * 支持按名称搜索（防抖）、新建与删除，数据全部走真实后端。
 */
export function Workbench() {
  const { t } = useTranslation()

  // 初始化从 localStorage 恢复选中 Tab（兼容旧版存的 '1'/'2'），刷新后保持
  const [tab, setTab] = useState<TabKey>(() => {
    const saved = localStorage.getItem(TABS_KEY)
    if (saved === 'favorite' || saved === '2') return 'favorite'
    return 'all'
  })
  const [items, setItems] = useState<NamespaceModel[]>([])
  const [count, setCount] = useState(0)
  // 页码初始值从 ?page= 恢复（仅「全部」Tab 有分页），刷新后停留在原页
  const [page, setPage] = useState(() => {
    const raw = new URLSearchParams(window.location.search).get('page')
    const n = raw ? Number(raw) : NaN
    return Number.isInteger(n) && n > 0 ? n : 1
  })
  const [hasMore, setHasMore] = useState(false)
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [debouncedKw, setDebouncedKw] = useState('')
  const [addOpen, setAddOpen] = useState(false)
  // 已打开的项目详情弹窗（多开）：从 NamespaceCard 提升到工作台层单一真源，
  // 集合同步到 URL ?open=1,2 持久化——刷新/翻页/切 Tab 都不关弹窗
  const [openProjects, setOpenProjects] = useState<OpenEntry[]>([])
  // 重复点击已打开的卡片 → 对应弹窗置顶（frontAt 递增），原在 NamespaceCard
  const [frontBumps, setFrontBumps] = useState<Record<number, number>>({})
  // URL 恢复扫描完成前不写 open 参数：挂载帧（openProjects=[]）不能把 ?open= 清掉，
  // 否则刷新窗口期里 URL 短暂失去参数，极快再次刷新会丢弹窗
  const [hydrated, setHydrated] = useState(false)
  const restoredRef = useRef(false)
  // 其他用户部署/删除/变更触发的空间刷新（WS ReloadProjects）：命中的空间卡显示 loading，
  // 对齐旧版 setNamespaceReload(true, nsID) → 仅该空间卡 Spin；刷新完成后清除。
  const [reloadingNsId, setReloadingNsId] = useState<number | null>(null)
  const handledReloadRevRef = useRef(0)
  const { reloadProjectsRev, reloadNsId } = useWebsocket()
  // 手动刷新按钮的自身 loading：刷新列表但保持网格挂载（与后台刷新同语义，不动卡片覆盖层）
  const [refreshing, setRefreshing] = useState(false)
  // 无限滚动的哨兵节点：进入视口即加载下一页
  const sentinelRef = useRef<HTMLDivElement>(null)
  // 总页数（派生自 count，供分页渲染与键盘翻页 guard 共用）
  const totalPages = Math.max(1, Math.ceil(count / PAGE_SIZE))

  // 首屏是否已成功加载过数据：仅当首拉失败（多为 URL 页码越界被后端拒绝）才回退第 1 页；
  // 浏览中翻页失败只 toast 不改页码，避免网络抖动把用户踢回首页。
  const loadedRef = useRef(false)

  // 翻页辅助：pageRef 实时页 + 请求序列号。
  // pageRef 由 goPage 同步写、渲染期回写 page，键盘 handler 在 React 重渲前即可读到最新页；
  // fetchSeqRef 每次 fetchPage 递增，响应回来与最新序列比对——过期请求整包丢弃，
  // 杜绝「快速连按翻页 → 旧响应 setPage 回写 → 页码来回跳」的乱序落地。
  const pageRef = useRef(page)
  pageRef.current = page // 渲染期同步最新页（幂等，StrictMode 双渲无害）
  const fetchSeqRef = useRef(0)

  // 首屏空闲预热：详情弹窗/配置编辑器已懒加载（见上方 lazy 注释），这里在首帧渲染并
  // 等数据请求发出后（1.2s）后台拉取对应 chunk，既保住首屏体积又让「点开项目卡片/
  // 切配置 Tab」零等待。TabShell(xterm) 使用频率低，不预热，保持按需。
  useEffect(() => {
    const timer = window.setTimeout(() => {
      void import('../projects/ProjectDetailModal')
      void import('../projects/CreateProjectModal')
      void import('../projects/TabEdit') // 静态依赖 CodeMirror(~590KB)，随配置 Tab 一起预热
    }, 1200)
    return () => window.clearTimeout(timer)
  }, [])

  // goPage：同步写 pageRef（让键盘在重渲前读到实时页）再 setPage。所有翻页入口统一走它。
  const goPage = useCallback((p: number) => {
    pageRef.current = p
    setPage(p)
  }, [])

  /**
   * 拉取第 p 页：append=true 追加到列表尾（无限滚动），否则整体替换（分页/首屏/搜索）。
   * 同步维护 page / hasMore，供分页与哨兵判断。
   * 越界自纠正：目标页超出总页数时保持骨架屏直接切到最后一页（成功路径），
   * 或首屏失败回退第 1 页（错误路径），由 [page] effect 重拉并同步 URL。
   */
  const fetchPage = useCallback(
    async (p: number, append: boolean) => {
      const seq = ++fetchSeqRef.current
      let corrected = false
      if (append) setLoadingMore(true)
      else setLoading(true)
      try {
        const { data, error } = await api.GET('/api/namespaces', {
          params: {
            query: {
              page: p,
              pageSize: PAGE_SIZE,
              favorite: tab === 'favorite' ? true : undefined,
              name: debouncedKw.trim() || undefined,
            },
          },
        })
        // 过期响应（翻页后又发起更新请求）：整包丢弃，不 setItems/setPage/不 toast，
        // finally 里也不复位 loading——加载态交棒给最新一次请求，避免页码被旧响应回写。
        if (seq !== fetchSeqRef.current) return
        if (error) throw new Error(error.message ?? String(error))
        if (!data) return
        const total = Math.max(1, Math.ceil(data.count / PAGE_SIZE))
        // 页码越界：URL/翻页给了超出总页数的页码 → 不落地越界空页，
        // 加载态保持，goPage(total) 触发 [page] effect 重拉最后一页并重写 ?page=
        if (!append && p > total) {
          corrected = true
          goPage(total)
          setLoading(true)
          return
        }
        setItems((prev) => (append ? [...prev, ...data.items] : data.items))
        setCount(data.count)
        // 追加不改 page：page 只代表「当前替换页」，由 Tab/搜索/翻页驱动
        if (!append) goPage(p)
        setHasMore(p * PAGE_SIZE < data.count)
        loadedRef.current = true
      } catch (e) {
        // 首屏且非首页拉取失败（URL 页码越界/过期被后端拒绝）：回退第 1 页重试
        if (!append && p > 1 && !loadedRef.current) {
          corrected = true
          goPage(1)
          setLoading(true)
          return
        }
        toast.error(e instanceof Error ? e.message : String(e))
      } finally {
        // corrected 或已过期时加载态已交棒给下一轮拉取，不再置 loading=false，避免闪空态
        if (!corrected && seq === fetchSeqRef.current) {
          setLoading(false)
          setLoadingMore(false)
        }
      }
    },
    [tab, debouncedKw, goPage],
  )

  // 关注 Tab 无搜索时进入「全量 + 可排序」模式：一次性拉完该用户全部关注（跨页顺序对拖拽重排
  // 是必要前提——后端 FavoriteSort 按整份有序 id 列表回写 sort_order），拖拽后按此全量回传。
  const loadAllFavorites = useCallback(async () => {
    const seq = ++fetchSeqRef.current
    setLoading(true)
    const all: NamespaceModel[] = []
    try {
      for (let p = 1; ; p += 1) {
        const { data, error } = await api.GET('/api/namespaces', {
          params: { query: { page: p, pageSize: PAGE_SIZE, favorite: true } },
        })
        if (seq !== fetchSeqRef.current) return
        if (error) throw new Error(error.message ?? String(error))
        if (!data || data.items.length === 0) break
        all.push(...data.items)
        if (all.length >= data.count) break
      }
      setItems(all)
      setCount(all.length)
      setHasMore(false)
      loadedRef.current = true
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      if (seq === fetchSeqRef.current) setLoading(false)
    }
  }, [])

  // 统一刷新入口：关注 Tab 无搜索时走全量加载（保序），其余路径走分页刷新。
  const refresh = useCallback(() => {
    if (tab === 'favorite' && !debouncedKw) return loadAllFavorites()
    return fetchPage(page, false)
  }, [tab, debouncedKw, page, loadAllFavorites, fetchPage])

  // 其他用户部署/删除项目 → WS ReloadProjects（WS 层已 debounce 500ms）→ 刷新列表，
  // 命中空间卡显示 loading、刷新后状态图标随之更新（对齐旧版 AppContent + ItemCard）。
  // reloadProjectsRev 与 reloadNsId 同批更新；handledReloadRevRef 防止依赖 refresh 身份变化时
  // effect 重复触发（翻页/切 Tab 会让 refresh 换引用，但 rev 未变不应再刷一次）。
  useEffect(() => {
    if (reloadProjectsRev === 0 || reloadProjectsRev === handledReloadRevRef.current) return
    handledReloadRevRef.current = reloadProjectsRev
    const nsId = reloadNsId
    setReloadingNsId(nsId)
    void refresh().finally(() => setReloadingNsId(null))
  }, [reloadProjectsRev, reloadNsId, refresh])

  // 手动刷新：整页拉取当前条件（分页/关注全量），loading 由按钮 spinner 表达
  const handleRefresh = useCallback(() => {
    if (refreshing) return
    setRefreshing(true)
    void refresh().finally(() => setRefreshing(false))
  }, [refreshing, refresh])

  const handleDeleted = useCallback(
    (nsId: number) => {
      setOpenProjects((prev) => prev.filter((e) => e.project.namespaceId !== nsId))
      void refresh()
    },
    [refresh],
  )

  // 关注 Tab 拖拽重排：乐观重排列表 + 提交被移动与落点两个空间 id，失败回滚。
  const sortable = tab === 'favorite' && !debouncedKw
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
  )
  const handleDragEnd = useCallback(
    ({ active, over }: DragEndEvent) => {
      if (!over || active.id === over.id) return
      const from = items.findIndex((ns) => String(ns.id) === active.id)
      const to = items.findIndex((ns) => String(ns.id) === over.id)
      if (from === -1 || to === -1) return
      const prev = items
      const next = arrayMove(items, from, to)
      setItems(next)
      setCount(next.length)
      // 后端 FavoriteSort 契约：firstId=被拖拽空间，secondId=落点位置原空间（移动前位置）。
      // prev 是移动前数组：prev[from] 即被拖拽项，prev[to] 即落点参照项。
      void api
        .PUT('/api/namespaces/favorite/sort', {
          body: { firstId: prev[from].id, secondId: prev[to].id },
        })
        .then(({ error }) => {
          if (error) {
            setItems(prev)
            toast.error(error.message ?? String(error))
          } else {
            toast.success(t('workbench.sortSaved'))
          }
        })
    },
    [items, t],
  )

  // 搜索输入防抖 400ms：触发时回到第 1 页并携带新关键词查询。
  // 挂载时 keyword===debouncedKw（同为空）直接跳过——否则 400ms 后的 setPage(1)
  // 会覆盖从 URL 恢复的初始分页（如 ?page=3 刷新后被拉回第 1 页），
  // 并与 fetchPage 的 setPage(p) 形成 [3,1,3,1] 乒乓死循环。
  useEffect(() => {
    const timer = setTimeout(() => {
      if (keyword === debouncedKw) return
      goPage(1)
      setDebouncedKw(keyword)
    }, 400)
    return () => clearTimeout(timer)
  }, [keyword, debouncedKw])

  // 页码/弹窗集合同步到 URL：翻页/切 Tab/搜索回第 1 页/开合弹窗用 replaceState 重写。
  // 替换式无弹栈副作用，刷新后停留在当前页；「关注」Tab 无分页不写页码。
  // open 参数只在水合完成后写：挂载帧（openProjects=[]）不把 ?open= 清掉，
  // 扫描解析出的 id 在 hydrated 置位后由本 effect 一次性写回（顺带清掉已删除/越界的 id）。
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    if (tab === 'all' && page > 1) params.set('page', String(page))
    else params.delete('page')
    if (hydrated) {
      const ids = openProjects.map((e) => e.project.id)
      if (ids.length) params.set('open', ids.join(','))
      else params.delete('open')
    }
    const qs = params.toString()
    window.history.replaceState(
      {},
      '',
      window.location.pathname + (qs ? `?${qs}` : '') + window.location.hash,
    )
  }, [page, tab, openProjects, hydrated])

  // Tab 切换：回到第 1 页并持久化选择。
  // setLoading(true) 同步置骨架屏：否则切换帧会渲染出「新 tab + 旧空 items + loading=false」
  // 的错空态（如关注空态切全部时，在提示文字位置闪出「新建空间」按钮）。
  const changeTab = (key: TabKey) => {
    setTab(key)
    localStorage.setItem(TABS_KEY, key)
    goPage(1)
    // 清空旧 tab 的 items：骨架屏判定是 loading && items.length===0，
    // 否则切换帧会把旧 tab 的卡片列表露在新 tab 头下（stale-while-revalidate 只该用于
    // 同数据集的后台刷新——部署成功/删除/改配置，那些场合网格必须保持挂载不换骨架）
    setItems([])
    setLoading(true)
  }

  // 查询条件（Tab/关键词）或用户翻页 → 整体替换对应页。
  // 统一由 (tab, page, debouncedKw) 驱动：Tab 切换、搜索、翻页（含回到第 1 页）都会重新拉取；
  // 「关注」Tab 无搜索时进入全量模式（loadAllFavorites，拖拽排序的整份顺序来源），
  // 追加走哨兵 fetchPage(append)，不改 page，因此不会误触发本 effect。
  useEffect(() => {
    if (tab === 'favorite' && !debouncedKw) {
      void loadAllFavorites()
      return
    }
    void fetchPage(page, false)
  }, [tab, page, debouncedKw, fetchPage, loadAllFavorites])

  // 「关注」Tab 无限滚动：哨兵进入视口（提前 240px 预载）且未在加载中 → 追加下一页。
  // 追加页码由已加载条数推导（items.length / PAGE_SIZE + 1），不依赖 page 状态，
  // 避免追加回写 page 干扰上方的替换式拉取 effect。
  useEffect(() => {
    if (tab !== 'favorite') return
    const el = sentinelRef.current
    if (!el) return
    const obs = new IntersectionObserver(
      (entries) => {
        if (!entries[0].isIntersecting) return
        if (loading || loadingMore || !hasMore) return
        void fetchPage(Math.floor(items.length / PAGE_SIZE) + 1, true)
      },
      { rootMargin: '240px 0px' },
    )
    obs.observe(el)
    return () => obs.disconnect()
  }, [tab, items.length, hasMore, loading, loadingMore, fetchPage])

  // 收藏切换乐观更新：关注 Tab 上取消关注 → 直接从列表移除（成功即保持移除；
  // 失败时 NamespaceCard 用原 ns 回调回滚 → 列表已无该卡，追加回尾部恢复）。
  // 全部 Tab 仅原地更新 favorite 字段，不涉及增删。
  const optimisticToggle = (ns: NamespaceModel) => {
    const removing = tab === 'favorite' && !ns.favorite
    if (removing) {
      if (items.some((it) => it.id === ns.id)) {
        setItems(items.filter((it) => it.id !== ns.id))
        setCount((c) => Math.max(0, c - 1))
      }
      return
    }
    if (!items.some((it) => it.id === ns.id)) {
      setItems([...items, ns]) // 回滚恢复被移除的卡片
      setCount((c) => c + 1)
      return
    }
    setItems(items.map((it) => (it.id === ns.id ? ns : it)))
  }
  // ctrl/cmd+k 聚焦搜索由 SearchInput 内置（与 events/repos 一致）

  /** 打开项目详情弹窗（多开）：已开同 id 只置顶（bringToFront），否则追加。卡片行点击上报 */
  const openProject = (p: ProjectModel, namespaceName: string) => {
    if (openProjects.some((e) => e.project.id === p.id)) {
      setFrontBumps((b) => ({ ...b, [p.id]: (b[p.id] ?? 0) + 1 }))
      return
    }
    setOpenProjects((prev) => [...prev, { project: p, namespaceName }])
  }
  const closeProject = (id: number) =>
    setOpenProjects((prev) => prev.filter((e) => e.project.id !== id))

  /** URL 恢复：无过滤全量翻页解析 ?open= 里的项目 id（单趟扫描，兼容旧 ?pid=）。
   *  找不到（已删除/越界）的 id 自然丢弃；解析结果合并进已开集合，完成后置水合位。 */
  const restoreOpenFromUrl = useCallback(async (ids: number[]) => {
    const missing = new Set(ids)
    const found: OpenEntry[] = []
    let page = 1
    for (;;) {
      const { data, error } = await api.GET('/api/namespaces', {
        params: { query: { page, pageSize: PAGE_SIZE } },
      })
      if (error || !data) break
      for (const ns of data.items) {
        for (const proj of ns.projects) {
          if (missing.delete(proj.id))
            found.push({ project: proj, namespaceName: ns.name })
        }
      }
      if (missing.size === 0) break
      if (page * PAGE_SIZE >= data.count) break
      page += 1
    }
    setOpenProjects((prev) => {
      const seen = new Set(prev.map((e) => e.project.id))
      return [...prev, ...found.filter((e) => !seen.has(e.project.id))]
    })
    setHydrated(true)
  }, [])

  // 挂载时解析 ?open=1,2（兼容旧 ?pid= 单值别名）→ 全量扫描重开。restoredRef 防 StrictMode 双跑
  useEffect(() => {
    if (restoredRef.current) return
    restoredRef.current = true
    const params = new URLSearchParams(window.location.search)
    const ids = (params.get('open') ?? '')
      .split(',')
      .map((s) => Number(s))
      .filter((n) => Number.isInteger(n) && n > 0)
    const legacy = Number(params.get('pid'))
    if (Number.isInteger(legacy) && legacy > 0) ids.push(legacy)
    if (ids.length === 0) {
      setHydrated(true)
      return
    }
    void restoreOpenFromUrl([...new Set(ids)])
  }, [restoreOpenFromUrl])

  // 键盘翻页：cmd/ctrl + ←/→ 切上一页/下一页（裸方向键不触发，保护浏览器快捷键与输入光标）。
  // 仅「全部」Tab 多页时生效（关注 Tab 无分页）。
  // 放行场景：任何弹窗/弹层打开（role=dialog，含 Radix Dialog/Popover，避免翻页藏在弹层后）、
  // 焦点在输入类控件（保护搜索框光标左右移动）。
  // 尾缘防抖 + 累计目标页：连按不逐击发请求，而是把目标页累进 pendingPageRef，
  // 静默 250ms 后只发一笔请求到最终页——快速 →→→→→ = 1 请求翻到第 5 页，中间意图不丢；
  // 同时忽略 e.repeat 长按 auto-repeat，避免按住不放连续翻页。
  // （此前是前缘节流：250ms 窗口内连按直接丢弃中间几次，慢速连按仍每次按键发一笔请求。）
  useEffect(() => {
    if (tab !== 'all' || count <= PAGE_SIZE) return
    let pendingPage: number | null = null
    let debounceTimer: ReturnType<typeof setTimeout> | null = null
    const onKeyDown = (e: KeyboardEvent) => {
      if (document.querySelector('[role="dialog"]')) return
      const el = document.activeElement as HTMLElement | null
      if (
        el &&
        (el.tagName === 'INPUT' ||
          el.tagName === 'TEXTAREA' ||
          el.tagName === 'SELECT' ||
          el.isContentEditable)
      )
        return
      if (!(e.metaKey || e.ctrlKey)) return // 仅 cmd/ctrl + ←/→ 触发，裸方向键不翻页
      if (e.repeat) return // 长按 auto-repeat 不连续翻页
      const current = pendingPage ?? pageRef.current
      let target: number | null = null
      if (e.key === 'ArrowLeft' && current > 1) target = current - 1
      else if (e.key === 'ArrowRight' && current < totalPages) target = current + 1
      if (target === null) return
      e.preventDefault()
      pendingPage = target
      if (debounceTimer !== null) clearTimeout(debounceTimer)
      debounceTimer = setTimeout(() => {
        debounceTimer = null
        const final = pendingPage
        pendingPage = null
        if (final !== null && final !== pageRef.current) goPage(final)
      }, 250)
    }
    window.addEventListener('keydown', onKeyDown)
    return () => {
      window.removeEventListener('keydown', onKeyDown)
      if (debounceTimer !== null) clearTimeout(debounceTimer)
    }
  }, [tab, count, totalPages, goPage])

  const pageItems = buildPages(page, totalPages)

  return (
    <div className="flex flex-col gap-4">
      {/* 工具栏：Tab 左 + 搜索/新建 右（旧版排布，品牌由 Header 承载） */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="border-b border-line">
          <Tabs value={tab} onValueChange={(v) => changeTab(v as TabKey)}>
            <TabsList variant="line">
              <TabsTrigger value="all" className="group">
                <Icon
                  name="grid"
                  className="size-4 group-data-[state=active]:fill-current group-data-[state=active]:text-primary"
                />
                {t('workbench.allProjects')}
              </TabsTrigger>
              <TabsTrigger value="favorite" className="group">
                <Icon
                  name="star"
                  className="size-4 group-data-[state=active]:fill-current group-data-[state=active]:text-primary"
                />
                {t('workbench.favorites')}
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>
        <div className="flex min-w-0 flex-1 items-center justify-end gap-3">
          <SearchInput
            value={keyword}
            onChange={setKeyword}
            placeholder={t('workbench.searchPlaceholder')}
            className="w-full max-w-xs"
          />
          <Button
            variant="outline"
            size="icon"
            onClick={handleRefresh}
            disabled={refreshing}
            title={t('workbench.refresh')}
            aria-label={t('workbench.refresh')}
          >
            {refreshing ? <Loader2 className="size-4 animate-spin" /> : <RefreshCw className="size-4" />}
          </Button>
          <Button variant="default" onClick={() => setAddOpen(true)}>
            <Icon name="plus" className="size-4" />
            {t('workbench.addNamespaceShort')}
          </Button>
        </div>
      </div>

      {/* 列表：骨架屏只在无数据时换出。已有 items 的后台刷新（部署成功 onChanged/删除/改配置）
          保持网格挂载——否则 loading 硬切换会把 NamespaceCard 整个卸载，
          其 openProjects（打开的详情弹窗）随之销毁，部署成功弹窗就被误关了 */}
      {loading && items.length === 0 ? (
        <div role="status" aria-busy="true">
          <SkeletonGrid count={9} />
        </div>
      ) : items.length === 0 ? (
        <Empty
          icon="namespace"
          text={
            debouncedKw
              ? t('workbench.searchEmpty', { kw: debouncedKw })
              : tab === 'favorite'
                ? t('workbench.favoritesEmpty')
                : t('workbench.emptyAll')
          }
          action={
            debouncedKw ? (
              <Button variant="outline" size="sm" onClick={() => setKeyword('')}>
                <Icon name="close" className="size-3.5" />
                {t('common.clearSearch')}
              </Button>
            ) : tab === 'favorite' ? (
              <span className="text-[12px] text-faint">{t('workbench.favoritesEmptyHint')}</span>
            ) : (
              <Button variant="default" size="sm" onClick={() => setAddOpen(true)}>
                <Icon name="plus" className="size-4" />
                {t('workbench.addNamespaceShort')}
              </Button>
            )
          }
        />
      ) : sortable ? (
        // 关注 Tab 无搜索：全量列表包进 DndContext 支持拖拽重排（后端按整份有序 id 回写）
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
        >
          <SortableContext
            items={items.map((ns) => String(ns.id))}
            strategy={rectSortingStrategy}
          >
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
              {items.map((ns, index) => (
                <SortableNamespaceCard
                  key={ns.id}
                  id={String(ns.id)}
                  index={index}
                  ns={ns}
                  loading={reloadingNsId !== null && ns.id === reloadingNsId}
                  onToggleFavorite={optimisticToggle}
                  onOpenProject={(p) => openProject(p, ns.name)}
                  onDeleted={handleDeleted}
                  onChanged={() => void refresh()}
                />
              ))}
            </div>
          </SortableContext>
        </DndContext>
      ) : (
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          {items.map((ns, index) => (
            <div
              key={ns.id}
              className="animate-list-in"
              style={{ animationDelay: `${Math.min(index, 10) * 30}ms` }}
            >
              <MemoizedNamespaceCard
                ns={ns}
                loading={reloadingNsId !== null && ns.id === reloadingNsId}
                onToggleFavorite={optimisticToggle}
                onOpenProject={(p) => openProject(p, ns.name)}
                // 删除命名空间 → 关闭该空间下已打开的详情弹窗（原卡片卸载自带的行为，提升后显式接管）
                onDeleted={handleDeleted}
                onChanged={() => void refresh()}
              />
            </div>
          ))}
        </div>
      )}

      {/* 关注 Tab 无限滚动哨兵：滚到底部即追加下一页（未加载完时才渲染） */}
      {tab === 'favorite' && items.length > 0 && hasMore && (
        <div ref={sentinelRef} className="flex h-10 items-center justify-center">
          {loadingMore && <Loader2 className="size-4 animate-spin text-faint" />}
        </div>
      )}

      {/* 分页 / 结果数：有数据时恒显数量反馈；「全部」多页出分页控件，「关注」只显数量。
          关注 Tab 可排序时把底栏固定在视口底部（sticky）：滚动长列表，「拖拽提示 + 共 N 个空间」
          始终可见。负边距撑满内容区宽度（抵消 main 的水平 padding），玻璃底避免卡片透穿。 */}
      {!loading && items.length > 0 && (
        <div
          className={cn(
            'flex items-center justify-end gap-2',
            sortable &&
              'sticky bottom-0 z-20 -mx-4 border-t border-line/60 bg-bg/85 px-4 py-2.5 backdrop-blur-md sm:-mx-6 sm:px-6 lg:-mx-10 lg:px-10',
          )}
        >
          {tab === 'all' && count > PAGE_SIZE && (
            <>
              {/* 键盘翻页提示：⌘/Ctrl + ←/→ 仅在全部项目页生效（mac 显 ⌘，win/linux 显 Ctrl；移动端无实体键盘隐藏） */}
              <span className="hidden items-center gap-1.5 text-[11px] text-faint sm:flex">
                <kbd className="flex h-5 items-center gap-0.5 whitespace-nowrap rounded border border-line bg-raised px-1 font-mono text-[10px] leading-none text-mute">
                  {isMac ? <span className="font-sans text-[11px]">⌘</span> : <span>Ctrl</span>}←
                </kbd>
                <kbd className="flex h-5 items-center gap-0.5 whitespace-nowrap rounded border border-line bg-raised px-1 font-mono text-[10px] leading-none text-mute">
                  {isMac ? <span className="font-sans text-[11px]">⌘</span> : <span>Ctrl</span>}→
                </kbd>
                {t('workbench.pageTurnHint')}
              </span>
              {/* m-0 覆盖 Pagination 基类 mx-auto：否则在 justify-end 行里 nav 会居中、
                  把左侧键盘提示顶到页面最左，无法紧贴分页 */}
              <Pagination className="w-fit m-0">
              <PaginationContent>
                <PaginationItem>
                  {/* 用 PaginationLink 而非 PaginationPrevious/Next：shadcn 那两个硬编码
                      英文 "Previous/Next" 文案，双语应用需自行本地化（aria-label + href 语义） */}
                  <PaginationLink
                    href={page > 1 ? `?page=${page - 1}` : undefined}
                    aria-label={t('common.previousPage')}
                    aria-disabled={page <= 1 || undefined}
                    className={cn(page <= 1 && 'pointer-events-none opacity-50')}
                    onClick={(e) => {
                      e.preventDefault()
                      if (page > 1) goPage(page - 1)
                    }}
                  >
                    <ChevronLeftIcon className="size-4" />
                  </PaginationLink>
                </PaginationItem>
                {pageItems.map((item, i) =>
                  item === '...' ? (
                    <PaginationEllipsis key={`ellipsis-${i}`} />
                  ) : (
                    <PaginationItem key={item}>
                      <PaginationLink
                        href={`?page=${item}`}
                        isActive={item === page}
                        onClick={(e) => {
                          e.preventDefault()
                          goPage(item)
                        }}
                      >
                        {item}
                      </PaginationLink>
                    </PaginationItem>
                  ),
                )}
                <PaginationItem>
                  <PaginationLink
                    href={page < totalPages ? `?page=${page + 1}` : undefined}
                    aria-label={t('common.nextPage')}
                    aria-disabled={page >= totalPages || undefined}
                    className={cn(page >= totalPages && 'pointer-events-none opacity-50')}
                    onClick={(e) => {
                      e.preventDefault()
                      if (page < totalPages) goPage(page + 1)
                    }}
                  >
                    <ChevronRightIcon className="size-4" />
                  </PaginationLink>
                </PaginationItem>
              </PaginationContent>
              </Pagination>
            </>
          )}
          {/* 关注 Tab 无搜索时展示拖拽排序提示（有可排序卡片才提示） */}
          {sortable && items.length > 0 && (
            <span className="hidden items-center gap-1 font-mono text-[11px] text-faint sm:flex">
              <GripVertical className="size-3" />
              {t('workbench.dragSortHint')}
            </span>
          )}
          <span className="font-mono text-[11px] text-faint">
            {tab === 'all' && count > PAGE_SIZE
              ? t('workbench.pagination', { count, totalPages })
              : t('workbench.searchResultCount', { count })}
          </span>
        </div>
      )}

      <AddNamespaceModal
        open={addOpen}
        onClose={() => setAddOpen(false)}
        onCreated={() => void fetchPage(1, false)}
      />

      {/* 已打开的项目详情弹窗（多开，URL ?open= 持久化）：刷新/翻页/切 Tab 保持打开。
          删除/配置变更 = 关弹窗 + 刷新列表（语义与原 NamespaceCard 渲染时一致）。
          Suspense fallback=null：弹窗懒加载完成即弹出，不占首屏。 */}
      <Suspense fallback={null}>
        {openProjects.map(({ project, namespaceName }) => (
          <ProjectDetailModal
            key={project.id}
            project={project}
            namespaceName={namespaceName}
            open
            frontAt={frontBumps[project.id] ?? 0}
            onClose={() => closeProject(project.id)}
            onDeleted={() => {
              closeProject(project.id)
              void refresh()
            }}
            onChanged={() => void refresh()}
          />
        ))}
      </Suspense>
    </div>
  )
}

/**
 * 可拖拽排序的关注卡片（仅关注 Tab 无搜索时启用）：useSortable 包一层 NamespaceCard，
 * 顶部图标簇左端注入拖拽手柄。拖拽中整卡轻微抬升（z + 半透明），其余样式透传。
 */
function SortableNamespaceCard({
  id,
  index,
  ns,
  loading,
  onToggleFavorite,
  onOpenProject,
  onDeleted,
  onChanged,
}: {
  id: string
  index: number
  ns: NamespaceModel
  loading: boolean
  onToggleFavorite: (ns: NamespaceModel) => void
  onOpenProject: (p: ProjectModel) => void
  onDeleted: (nsId: number) => void
  onChanged: () => void
}) {
  const { t } = useTranslation()
  // 入场动画播完即摘掉 animate-list-in：fill-mode:both 会让动画"永久存活"，
  // 浏览器每帧都重算这 6 个元素的样式/布局（拖拽时的 StyleAndLayout 风暴）。
  const [entered, setEntered] = useState(false)
  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id,
    // 让位动画 200ms→150ms：拖拽过程中列表跟手更紧，降低“橡皮筋”滞后感
    transition: { duration: 150, easing: 'ease' },
  })

  // 手柄引用稳定（dnd-kit 的 useNodeRef 用 useCallback 包了），useMemo 后拖拽过程
  // 该节点引用不变 → 下游 memo(NamespaceCard) 能跳过重渲染，拖拽只动外层廉价节点。
  const dragHandle = useMemo(
    () => (
      <button
        type="button"
        ref={setActivatorNodeRef}
        {...attributes}
        {...listeners}
        aria-label={t('workbench.dragSort')}
        title={t('workbench.dragSort')}
        className="cursor-grab rounded-md p-1 text-faint transition-colors hover:bg-raised hover:text-primary active:cursor-grabbing focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40"
      >
        <GripVertical className="size-[15px]" />
      </button>
    ),
    [t, attributes, listeners, setActivatorNodeRef],
  )

  return (
    <div
      ref={setNodeRef}
      style={{
        transform: CSS.Transform.toString(transform),
        transition,
      }}
      className={cn('will-change-transform', isDragging && 'relative z-10')}
    >
      <div
        className={cn(
          'h-full',
          !entered && 'animate-list-in',
          // 拖拽高亮用细 ring（2px 无模糊，GPU 合成便宜）；不要大 blur 阴影——每帧重绘是真机卡顿源
          isDragging && 'rounded-xl ring-2 ring-primary/50',
        )}
        style={{
          animationDelay: `${Math.min(index, 10) * 30}ms`,
          // 拖拽抬起：scale 是独立 CSS 属性，不与 dnd-kit 的 inline transform 冲突
          scale: isDragging ? '1.03' : undefined,
        }}
        onAnimationEnd={() => setEntered(true)}
      >
        <MemoizedNamespaceCard
        ns={ns}
        loading={loading}
        onToggleFavorite={onToggleFavorite}
        onOpenProject={onOpenProject}
        onDeleted={onDeleted}
        onChanged={onChanged}
        dragHandle={dragHandle}
        />
      </div>
    </div>
  )
}
