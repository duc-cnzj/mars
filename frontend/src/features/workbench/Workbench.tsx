import { lazy, memo, Suspense, useCallback, useEffect, useMemo, useRef, useState, type CSSProperties, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
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
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { toast } from '@/lib/toast'
import { isMac } from '@/lib/platform'
import { buildPages } from '@/lib/pagination'
import { Icon } from '@/components/Icons'
import { Empty, SkeletonGrid } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Pagination, PaginationContent, PaginationEllipsis, PaginationItem, PaginationLink } from '@/components/ui/shadcn/pagination'
import { cn } from '@/lib/utils'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/shadcn/tabs'
import { SearchInput } from '@/components/SearchInput'
import { AddNamespaceModal } from './AddNamespaceModal'
import { NamespaceCard } from './NamespaceCard'
import { useWebsocket } from '@/hooks/useWebsocket'
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

/** 工作台 Tab 选择持久化 key（与旧版 AppContent 的 'active-tabs' 一致） */
const TABS_KEY = 'active-tabs'

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
  // 对齐旧版 setNamespaceReload(true, nsID) → 仅这批受影响空间卡 Spin；刷新完成后清除。
  const [reloadingNsIds, setReloadingNsIds] = useState<number[] | null>(null)
  const handledReloadRevRef = useRef(0)
  const { reloadProjectsRev, reloadNsIds, ready } = useWebsocket()
  // WS 是否已建立过连接：首连不触发重同步（首屏数据已由 fetchList 拉取），
  // 后续每次 ready=false→true（断线重连）都整页重拉一次兜底。
  const wsConnectedRef = useRef(false)
  // 上一拍 ready 值：依赖 refresh 换引用（翻页/切 Tab）重跑 effect 时 ready 未翻转，不得误判为重连
  const wsPrevReadyRef = useRef(false)
  // 本地操作刚刷新过的空间标记（nsId → 时间戳）：TTL 内到达的 WS 批次刷新直接跳过，
  // 避免「本地 onChanged 直刷一次 + 后端广播又刷一次」的重复请求与二次 loading 闪（④）。
  const recentlyRefreshedRef = useRef(new Map<number, number>())
  // 手动刷新按钮的自身 loading：刷新列表但保持网格挂载（与后台刷新同语义，不动卡片覆盖层）
  const [refreshing, setRefreshing] = useState(false)
  // 手动刷新完成的递增记号：ListEnter.replayKey 用它重播列表入场动画（刷新不重挂卡片，
  // 靠摘类回加触发动画重启，NamespaceCard 内部展开等状态原样保留）
  const [listRev, setListRev] = useState(0)
  // 无限滚动的哨兵节点：进入视口即加载下一页
  const sentinelRef = useRef<HTMLDivElement>(null)
  // 总页数（派生自 count，供分页渲染与键盘翻页 guard 共用）
  const totalPages = Math.max(1, Math.ceil(count / PAGE_SIZE))

  // 首屏是否已成功加载过数据：仅当首拉失败（多为 URL 页码越界被后端拒绝）才回退第 1 页；
  // 浏览中翻页失败只 toast 不改页码，避免网络抖动把用户踢回首页。
  const loadedRef = useRef(false)

  // 翻页辅助：pageRef 实时页 + 请求序列号。
  // pageRef 由 goPage 同步写、渲染期回写 page，键盘 handler 在 React 重渲前即可读到最新页；
  // fetchSeqRef 每次 fetchList 递增，响应回来与最新序列比对——过期请求整包丢弃，
  // 杜绝「快速连按翻页 → 旧响应 setPage 回写 → 页码来回跳」的乱序落地。
  const pageRef = useRef(page)
  pageRef.current = page // 渲染期同步最新页（幂等，StrictMode 双渲无害）
  const fetchSeqRef = useRef(0)
  // 实时 items 镜像：异步回调（refreshNamespace 404 移除前判存在、避免向未加载页下刀）需同步读当前列表，
  // 不依赖会过期的闭包 items；渲染期同步，幂等。
  const itemsRef = useRef(items)
  itemsRef.current = items
  // 实时 openProjects 镜像：稳定 openProject 回调（memo 生效前提）里读最新已开集合做去重。
  const openProjectsRef = useRef(openProjects)
  openProjectsRef.current = openProjects

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
  const fetchList = useCallback(
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

  // 统一刷新入口：整页拉取当前条件（分页/关注分页）。
  // 关注 Tab 拖拽排序是相对移动契约（firstId=被拖拽空间，secondId=落点参照项，见 handleDragEnd），
  // 无需全量加载——部分加载下已加载卡片间的拖拽同样正确，未加载卡片保持后端相对顺序。
  const refresh = useCallback(() => {
    return fetchList(page, false)
  }, [page, fetchList])

  /**
   * 只刷新单个空间的详情并原地替换列表项（不整页重拉）。
   * 适用：单空间内的项目/配置变更（本地操作 onChanged + WS ReloadProjects）——尤其关注 Tab
   * 无限滚动下整页重拉会折叠回第 1 页并丢滚动位置；原地替换保持卡片挂载、内部展开状态不丢。
   * 空间已被删除（show 返回 404/grpc code 5）→ 从列表移除并同步计数；其余错误仅 toast。
   * 目标空间未加载（关注 Tab 的未翻到页 / 他端新建）→ 原地无匹配则静默，由后续整页刷新兜底。
   */
  const refreshNamespace = useCallback(
    async (nsId: number, silent = false): Promise<boolean> => {
      try {
        const { data, error } = await api.GET('/api/namespaces/{id}', {
          params: { path: { id: String(nsId) } },
        })
        if (error) {
          if (error.code === 5) {
            if (itemsRef.current.some((it) => it.id === nsId)) {
              setItems((prev) => prev.filter((it) => it.id !== nsId))
              setCount((c) => Math.max(0, c - 1))
            }
            // 空间已被删除：其名下项目弹窗一并关掉，否则残留孤儿弹窗
            //（弹窗内 TabLog/TabShell 请求持续 404；handleDeleted 卡片流程会关，WS 404 路径漏了）
            setOpenProjects((prev) => prev.filter((e) => e.project.namespaceId !== nsId))
            return true // 404 = 正常收敛，不算刷新失败
          }
          if (!silent) toast.error(error.message ?? String(error))
          return false
        }
        if (!data?.item) return true
        // Show 响应不含 favorite（后端 transformer 不计算，仅 List 路径回填）：
        // 原地替换时保留旧值，否则已收藏空间刷新后星标变空心，关注 Tab 点星会被
        // `removing = tab==='favorite' && !ns.favorite` 误判成"取消关注"（数据级误操作）
        setItems((prev) =>
          prev.map((it) => (it.id === nsId ? { ...it, ...data.item, favorite: it.favorite } : it)),
        )
        return true
      } catch (e) {
        // 网络层失败（fetch reject）：不向外抛——onChanged/WS 批量直接传本函数时不会产生未处理拒绝
        if (!silent) toast.error(e instanceof Error ? e.message : String(e))
        return false
      }
    },
    [],
  )

  // 本地操作触发的空间刷新：先记录 nsId+时间戳（④ 去重依据）再执行真实刷新。
  // 部署/创建/删除项目后端都会广播 ReloadProjects，本地直刷后 WS 批次（500ms debounce）又刷一次
  // = 两次请求 + 两次 loading 闪。后端时序（runner.OnFinally：UpdateDeployStatus → ToAll → SendDeployedResult）
  // 保证本地 onChanged 刷新时最终状态已落库，跳过 WS 批次安全；TTL 窗口覆盖 debounce + 余量。
  const localRefresh = useCallback(
    (nsId: number) => {
      recentlyRefreshedRef.current.set(nsId, Date.now())
      return refreshNamespace(nsId)
    },
    [refreshNamespace],
  )

  // 其他用户部署/删除/创建项目 → WS ReloadProjects（WS 层已 debounce 500ms，窗口内累积受影响空间集合）
  // → 逐个 refreshNamespace 只刷新对应空间，命中空间卡显示 loading、刷新后状态图标随之更新。
  // 按空间详情原地替换，不整页重拉——关注 Tab 无限滚动下整页重拉会折叠回第 1 页、丢滚动位置。
  // reloadNsIds 为 null（窗口内有坏帧未解出）时退化为整页刷新。
  // reloadProjectsRev 与 reloadNsIds 同批更新；handledReloadRevRef 防止依赖身份变化时
  // effect 重复触发（翻页/切 Tab 会让回调换引用，但 rev 未变不应再刷一次）。
  useEffect(() => {
    if (reloadProjectsRev === 0 || reloadProjectsRev === handledReloadRevRef.current) return
    handledReloadRevRef.current = reloadProjectsRev
    const ids = reloadNsIds
    if (ids === null) {
      // 坏帧/未知空间：整页刷新（不标单卡 loading）
      setReloadingNsIds(null)
      void refresh().finally(() => setReloadingNsIds(null))
      return
    }
    if (ids.length === 0) return // 防御：空集合无刷新目标
    // ④ 去重：过滤掉 TTL 内本地刚刷过的空间（部署成功/创建项目 onChanged 直刷），
    //    只刷真正来自其他端的变更——同空间避免二次请求 + 二次 loading 闪。
    const now = Date.now()
    const TTL = 2000
    let hasExpired = false
    const freshIds = ids.filter((id) => {
      const ts = recentlyRefreshedRef.current.get(id)
      if (ts === undefined) return true
      if (now - ts < TTL) return false
      recentlyRefreshedRef.current.delete(id)
      hasExpired = true
      return true
    })
    // 顺带清掉过期的标记，避免 Map 无限增长
    if (hasExpired) {
      for (const [id, ts] of recentlyRefreshedRef.current) {
        if (now - ts > TTL) recentlyRefreshedRef.current.delete(id)
      }
    }
    if (freshIds.length === 0) return
    setReloadingNsIds(freshIds)
    // silent=true：单卡失败不逐卡 toast；allSettled 汇总失败数合并成一条（⑤）。
    // allSettled 同时保证单卡 show 失败不影响其余空间刷新、不产生未处理拒绝。
    void Promise.allSettled(freshIds.map((nsId) => refreshNamespace(nsId, true))).then((results) => {
      setReloadingNsIds(null)
      const failed = results.filter((r) => r.status === 'fulfilled' && r.value === false).length
      if (failed > 0) toast.error(t('workbench.refreshNsFailed', { count: failed }))
    })
  }, [reloadProjectsRev, reloadNsIds, refresh, refreshNamespace, t])

  // WS 断线重连 → 重同步：ReloadProjects 是实时推送，断线窗口内其他用户的部署/删除/创建
  // 事件不会补发，本地列表可能陈旧。重连成功（ready false→true 沿）后整页重拉当前条件兜底。
  // 用 refresh 而非 handleRefresh：静默原地替换，不重播入场动画（用户没主动刷新）、
  // 不整页重挂（卡片展开状态/已开弹窗保留）；首连不刷，首屏数据已由 fetchList 拉取。
  // ready 沿检测用 wsPrevReadyRef：refresh 依赖 page/tab/kw，其引用变化会重跑本 effect，
  // 但 ready 未翻转（仍是 true）时不得误触发——`!ready || prev` 直接短路。
  useEffect(() => {
    const prev = wsPrevReadyRef.current
    wsPrevReadyRef.current = ready
    if (!ready || prev) return
    if (wsConnectedRef.current) {
      if (tab === 'favorite' && itemsRef.current.length > 0) {
        // 关注 Tab 无限滚动下 page 恒为 1（哨兵走 append 不动 page），整页 refresh 会把
        // 已加载列表折叠回第 1 页丢滚动位置——与 WS 批次同一套语义，改对已加载空间逐个
        // refreshNamespace 原地补刷断线窗口内他端的部署/删除/变更；未加载到的空间（他端新建）
        // 由后续翻页拉取天然兜底。reconnect 罕见，逐个请求量级可接受。
        void Promise.allSettled(itemsRef.current.map((it) => refreshNamespace(it.id, true)))
      } else {
        void refresh()
      }
    } else {
      wsConnectedRef.current = true
    }
  }, [ready, refresh, tab, refreshNamespace])

  // 手动刷新：整页拉取当前条件（分页/关注分页），loading 由按钮 spinner 表达
  const handleRefresh = useCallback(() => {
    if (refreshing) return
    setRefreshing(true)
    void refresh().finally(() => {
      setRefreshing(false)
      // 数据就位后再重播入场动画（刷新按钮 = 重新进入列表的语义）
      setListRev((r) => r + 1)
    })
  }, [refreshing, refresh])

  const handleDeleted = useCallback(
    (nsId: number) => {
      setOpenProjects((prev) => prev.filter((e) => e.project.namespaceId !== nsId))
      // 与其他 onChanged/onDeleted 路径一致：TTL 标记 + 按空间详情原地刷新。
      // 整页 refresh 会让关注 Tab 折叠回第 1 页丢滚动位置；refreshNamespace 的 404 分支
      // 已负责「从列表移除 + 同步计数 + 关闭该空间名下弹窗」，直接复用即可。
      void localRefresh(nsId)
    },
    [localRefresh],
  )

  // 关注 Tab 拖拽重排（无搜索时）：乐观重排已加载卡片 + 提交相对移动（firstId=被拖拽空间，
  // secondId=落点参照项，后端在整份顺序里重定位，未加载卡片相对顺序不受影响——部分加载下拖拽同样正确），失败回滚。
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
      const dragged = items[from]
      const next = arrayMove(items, from, to)
      setItems(next)
      // 不 setCount：count 是后端返回的关注总数（fetchList 的 data.count），拖拽重排不改变
      // 集合大小。next.length 只是已加载条数，部分加载下覆盖会把"共 N 个空间"错算成已加载数
      //（loadAllFavorites 全量时代 items.length===count 才成立，无限滚动后已失配）。
      // 后端 FavoriteSort 契约：firstId=被拖拽空间，secondId=落点位置原空间（移动前位置）。
      // items 是移动前数组：items[from] 即被拖拽项，items[to] 即落点参照项。
      void api
        .PUT('/api/namespaces/favorite/sort', {
          body: { firstId: dragged.id, secondId: items[to].id },
        })
        .then(({ error }) => {
          if (error) {
            // 失败回滚：只把被拖拽项移回原位，不做整表快照替换——乐观移动与失败之间的窗口里
            // WS ReloadProjects 的 refreshNamespace 会 in-place 合并该空间（部署状态等）、触底
            // 追加也可能落地，整表回滚会把它们一并抹掉（对齐 optimisticToggle 只合 favorite 字段、
            // 不整对象替换的既有约定）。若该空间已被并发删除，原位无可恢复项则保持现状。
            setItems((cur) => {
              const idx = cur.findIndex((it) => it.id === dragged.id)
              if (idx === -1) return cur
              const restored = [...cur]
              const [moved] = restored.splice(idx, 1)
              restored.splice(Math.min(from, restored.length), 0, moved)
              return restored
            })
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
  // 并与 fetchList 的 setPage(p) 形成 [3,1,3,1] 乒乓死循环。
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
  // 追加走哨兵 fetchList(append)，不改 page，因此不会误触发本 effect。
  useEffect(() => {
    void fetchList(page, false)
  }, [tab, page, debouncedKw, fetchList])

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
        void fetchList(Math.floor(items.length / PAGE_SIZE) + 1, true)
      },
      { rootMargin: '240px 0px' },
    )
    obs.observe(el)
    return () => obs.disconnect()
  }, [tab, items.length, hasMore, loading, loadingMore, fetchList])

  // 收藏切换乐观更新：关注 Tab 上取消关注 → 直接从列表移除（成功即保持移除；
  // 失败时 NamespaceCard 用原 ns 回调回滚 → 列表已无该卡，追加回尾部恢复）。
  // 全部 Tab 仅原地更新 favorite 字段，不涉及增删。
  // useCallback 稳定引用（依赖 tab 而非 items，列表用 itemsRef 现读）：让 memo(NamespaceCard) 生效，
  // items 变化（单卡刷新/翻页）不应让全部卡片重渲染；setItems 改用函数式，避免闭包过期。
  const optimisticToggle = useCallback(
    (ns: NamespaceModel) => {
      const removing = tab === 'favorite' && !ns.favorite
      if (removing) {
        if (itemsRef.current.some((it) => it.id === ns.id)) {
          setItems((prev) => prev.filter((it) => it.id !== ns.id))
          setCount((c) => Math.max(0, c - 1))
        }
        return
      }
      if (!itemsRef.current.some((it) => it.id === ns.id)) {
        setItems((prev) => [...prev, ns]) // 回滚恢复被移除的卡片
        setCount((c) => c + 1)
        return
      }
      // 只合并 favorite 字段，不整体替换 ns：请求期间 WS 批刷新可能已更新该空间
      //（部署状态等），整体替换会把那些新字段一并打回旧值
      setItems((prev) =>
        prev.map((it) => (it.id === ns.id ? { ...it, favorite: ns.favorite } : it)),
      )
    },
    [tab],
  )
  // ctrl/cmd+k 聚焦搜索由 SearchInput 内置（与 events/repos 一致）

  /** 打开项目详情弹窗（多开）：已开同 id 只置顶（bringToFront），否则追加。卡片行点击上报。
   *  稳定引用（useCallback []）：卡片 onOpenProject prop 需要稳定才能让 memo(NamespaceCard) 生效——
   *  namespace 名从 itemsRef 现查、去重从 openProjectsRef 现读，不依赖会过期的闭包。 */
  const openProject = useCallback((p: ProjectModel) => {
    const namespaceName = itemsRef.current.find((it) => it.id === p.namespaceId)?.name ?? ''
    if (openProjectsRef.current.some((e) => e.project.id === p.id)) {
      setFrontBumps((b) => ({ ...b, [p.id]: (b[p.id] ?? 0) + 1 }))
      return
    }
    setOpenProjects((prev) => [...prev, { project: p, namespaceName }])
  }, [])
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
                {t('workbench.allSpaces')}
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
            {refreshing ? <Icon name="loader" className="size-4 animate-spin" /> : <Icon name="refresh-cw" className="size-4" />}
          </Button>
          <Button variant="default" onClick={() => setAddOpen(true)}>
            <Icon name="plus" className="size-4" />
            {t('workbench.addNamespaceShort')}
          </Button>
        </div>
      </div>

      {/* 列表：骨架屏只在无数据时换出。已有 items 的变更（部署成功 onChanged/删除项目/改配置/WS）
          走 refreshNamespace 原地替换该空间——保持网格挂载，不整页重拉（关注 Tab 折叠回第 1 页、
          且 loading 硬切换会把 NamespaceCard 整个卸载，其打开的详情弹窗随之销毁） */}
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
              <span className="text-[12px] text-faint">{t('workbench.favoritesEmptyTip')}</span>
            ) : (
              <Button variant="default" size="sm" onClick={() => setAddOpen(true)}>
                <Icon name="plus" className="size-4" />
                {t('workbench.addNamespaceShort')}
              </Button>
            )
          }
        />
      ) : sortable ? (
        // 关注 Tab 无搜索：分页列表包进 DndContext 支持拖拽重排（相对移动契约，见 handleDragEnd）
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
                  replayKey={listRev}
                  ns={ns}
                  loading={reloadingNsIds?.includes(ns.id) ?? false}
                  onToggleFavorite={optimisticToggle}
                  onOpenProject={openProject}
                  onDeleted={handleDeleted}
                  onChanged={localRefresh}
                />
              ))}
            </div>
          </SortableContext>
        </DndContext>
      ) : (
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          {items.map((ns, index) => (
            <ListEnter
              key={ns.id}
              delay={(index % PAGE_SIZE) * LIST_ENTER_STAGGER_MS}
              replayKey={listRev}
            >
              <MemoizedNamespaceCard
                ns={ns}
                loading={reloadingNsIds?.includes(ns.id) ?? false}
                onToggleFavorite={optimisticToggle}
                onOpenProject={openProject}
                // 删除命名空间 → 关闭该空间下已打开的详情弹窗（原卡片卸载自带的行为，提升后显式接管）
                onDeleted={handleDeleted}
                onChanged={localRefresh}
              />
            </ListEnter>
          ))}
        </div>
      )}

      {/* 关注 Tab 无限滚动哨兵：滚到底部即追加下一页（未加载完时才渲染） */}
      {tab === 'favorite' && items.length > 0 && hasMore && (
        <div ref={sentinelRef} className="flex h-10 items-center justify-center">
          {loadingMore && <Icon name="loader" className="size-4 animate-spin text-faint" />}
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
              {/* 键盘翻页提示：⌘/Ctrl + ←/→ 仅在全部空间页生效（mac 显 ⌘，win/linux 显 Ctrl；移动端无实体键盘隐藏） */}
              <span className="hidden items-center gap-1.5 text-[11px] text-faint sm:flex">
                <kbd className="flex h-5 items-center gap-0.5 whitespace-nowrap rounded border border-line bg-raised px-1 font-mono text-[10px] leading-none text-mute">
                  {isMac ? <span className="font-sans text-[11px]">⌘</span> : <span>Ctrl</span>}←
                </kbd>
                <kbd className="flex h-5 items-center gap-0.5 whitespace-nowrap rounded border border-line bg-raised px-1 font-mono text-[10px] leading-none text-mute">
                  {isMac ? <span className="font-sans text-[11px]">⌘</span> : <span>Ctrl</span>}→
                </kbd>
                {t('workbench.pageTurnTip')}
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
                    <Icon name="chevron-left" className="size-4" />
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
                    <Icon name="chevron-right" className="size-4" />
                  </PaginationLink>
                </PaginationItem>
              </PaginationContent>
              </Pagination>
            </>
          )}
          {/* 关注 Tab 无搜索时展示拖拽排序提示（有可排序卡片才提示） */}
          {sortable && items.length > 0 && (
            <span className="hidden items-center gap-1 font-mono text-[11px] text-faint sm:flex">
              <Icon name="grip-vertical" className="size-3" />
              {t('workbench.dragSortTip')}
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
        onCreated={() => void fetchList(1, false)}
      />

      {/* 已打开的项目详情弹窗（多开，URL ?open= 持久化）：刷新/翻页/切 Tab 保持打开。
          删除/配置变更 = 关弹窗 + 按空间详情原地刷新列表（只更新该空间卡，不整页重拉）。
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
              // localRefresh（记录标记 + 刷新）：删除项目也会广播 ReloadProjects，
              // 直刷 + WS 批次双重刷新会闪两次 loading，与 ④ 去重保持一致
              void localRefresh(project.namespaceId)
            }}
            onChanged={() => void localRefresh(project.namespaceId)}
          />
        ))}
      </Suspense>
    </div>
  )
}

/** 列表项入场错峰步长：每张卡片在上张开始后 120ms 再进入（0.45s 动画 → 肉眼可分辨的逐个进入级联）。
 *  延迟按「批内相对位置」(index % PAGE_SIZE) 计算：无限下拉追加的每批新卡片从 0 重新级联，
 *  否则第 2 页起全部卡片同取 1200ms 延迟 → fill-mode:both 下前 1.2s 全不可见，级联丢失（2026-08-22 修）。 */
const LIST_ENTER_STAGGER_MS = 120

/**
 * 列表项进入动画容器：挂载时播 animate-list-in（淡入上浮 + delay 错峰），播完即摘掉类——
 * fill-mode:both 会让动画"永久存活"，浏览器每帧都重算这 6 个元素的样式/布局（拖拽时的
 * StyleAndLayout 风暴）。replayKey 变化（主页手动刷新完成）时把已摘的类回加，浏览器重启
 * 动画；不重挂卡片子树，NamespaceCard 内部展开等状态原样保留。
 */
function ListEnter({
  delay,
  replayKey,
  className,
  style,
  children,
}: {
  delay: number
  replayKey: number
  className?: string
  style?: CSSProperties
  children: ReactNode
}) {
  const [entered, setEntered] = useState(false)
  // 仅当 replayKey 变化且动画已播完（entered=true、类已摘）时回加类触发重启。
  // 省略 entered 依赖是有意的：放进依赖会在 onAnimationEnd 置 true 后立刻回加类 → 无限重播。
  useEffect(() => {
    if (entered) setEntered(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [replayKey])
  return (
    <div
      className={cn(className, !entered && 'animate-list-in')}
      style={{ animationDelay: `${delay}ms`, ...style }}
      onAnimationEnd={() => setEntered(true)}
    >
      {children}
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
  replayKey,
  ns,
  loading,
  onToggleFavorite,
  onOpenProject,
  onDeleted,
  onChanged,
}: {
  id: string
  index: number
  replayKey: number
  ns: NamespaceModel
  loading: boolean
  onToggleFavorite: (ns: NamespaceModel) => void
  onOpenProject: (p: ProjectModel) => void
  onDeleted: (nsId: number) => void
  onChanged: (nsId: number) => void
}) {
  const { t } = useTranslation()
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
      <Button
        variant="ghost"
        size="icon-xs"
        ref={setActivatorNodeRef}
        {...attributes}
        {...listeners}
        aria-label={t('workbench.dragSort')}
        title={t('workbench.dragSort')}
        className="cursor-grab text-faint hover:text-primary active:cursor-grabbing"
      >
        <Icon name="grip-vertical" className="size-4" />
      </Button>
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
      <ListEnter
        delay={(index % PAGE_SIZE) * LIST_ENTER_STAGGER_MS}
        replayKey={replayKey}
        className={cn(
          'h-full',
          // 拖拽高亮用细 ring（2px 无模糊，GPU 合成便宜）；不要大 blur 阴影——每帧重绘是真机卡顿源
          isDragging && 'rounded-xl ring-2 ring-primary/50',
        )}
        style={{
          // 拖拽抬起：scale 是独立 CSS 属性，不与 dnd-kit 的 inline transform 冲突
          scale: isDragging ? '1.03' : undefined,
        }}
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
      </ListEnter>
    </div>
  )
}
