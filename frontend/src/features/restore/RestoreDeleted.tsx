import {
  memo,
  useCallback,
  useEffect,
  useRef,
  useState,
  type CSSProperties,
  type KeyboardEvent as ReactKeyboardEvent,
  type ReactElement,
} from 'react'
import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { SearchInput } from '@/components/SearchInput'
import { Empty, RefreshFade, SkeletonList, Tag } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/shadcn/tabs'
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
} from '@/components/ui/shadcn/pagination'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/shadcn/alert-dialog'
import { buildPages } from '@/lib/pagination'
import { SEARCH_DEBOUNCE_MS } from '@/lib/constants'
import { formatDateTime } from '@/lib/format'
import { humanizeDateTime } from '@/lib/humanizeDateTime'
import { toast } from '@/lib/toast'
import { cn } from '@/lib/utils'
import { api } from '@/api/client'
import { API } from '@/api/endpoints'
import type { components } from '@/api/schema'

type NamespaceModel = components['schemas']['types.NamespaceModel']
type ProjectModel = components['schemas']['types.ProjectModel']

/** 实体 Tab 键（对齐 i18n 的 restore.tab*） */
type TabKey = 'namespaces' | 'projects'

/** 每页条数（服务端分页，与后端 InitByDefault 的默认 15 对齐） */
const PAGE_SIZE = 15

/** 已删除列表的响应形状（两个实体的分页响应同构）：Count 是未分页总数，分页器靠它算总页数 */
type DeletedPage<T> = { count: number; items: T[] }

/** 可注入 className/style 的行元素（RefreshFade 靠 cloneElement 注入渐入动画，行须转发到根元素） */
type RowElement = ReactElement<{ className?: string; style?: CSSProperties }>

/**
 * 已删除列表取数状态机（空间/项目两个 Tab 共用）：关键词防抖 + 显式分页 + 过期响应丢弃。
 *
 * 两个 Tab 的取数语义完全一致（只有端点与实体不同），故收成一个 hook。用显式页码而非无限下拉：
 * 恢复是「找到那一条并还原」的排障动作，来回翻页比一路下拉更顺手，也少一层触底/追加的竞态。
 *
 * @param fetch     取数函数（由调用方注入 typed 的 api.GET 调用，避免在 hook 内对路径字符串做类型断言）
 * @param reloadTick 外部重取信号（恢复成功后 +1）：与页码解耦，避免「重取当前页」需要伪造页码变化
 */
function useDeletedList<T>(fetch: (page: number, search: string) => Promise<DeletedPage<T>>, reloadTick: number) {
  const [items, setItems] = useState<T[]>([])
  const [count, setCount] = useState(0)
  const [keyword, setKeyword] = useState('')
  const [debounced, setDebounced] = useState('')
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  // 成功取数版本号：驱动 RefreshFade 重播渐入（失败态不播）
  const [version, setVersion] = useState(0)
  // fetch 每次渲染都是新闭包，存 ref 当「最新实现」用，不写进 effect 依赖（否则每次渲染都重取）
  const fetchRef = useRef(fetch)
  fetchRef.current = fetch
  // 请求序号：连打搜索/翻页时只允许最后发出的请求落地，旧响应晚到一律丢弃（否则会被旧数据覆盖）
  const seqRef = useRef(0)

  // 关键词防抖：停顿 300ms 后才把新词交给 effect；页码重置写在同一批更新里，
  // effect 只会看到「新词 × 第 1 页」，不会先发一次「新词 × 旧页码」的废请求
  useEffect(() => {
    const timer = window.setTimeout(() => {
      setDebounced(keyword)
      setPage(1)
    }, SEARCH_DEBOUNCE_MS)
    return () => window.clearTimeout(timer)
  }, [keyword])

  useEffect(() => {
    const seq = ++seqRef.current
    setLoading(true)
    fetchRef.current(page, debounced)
      .then((data) => {
        if (seq !== seqRef.current) return
        // 恢复掉末页最后一条后回读会拿到空页（Count 仍 > 0）：回退一页，不停在空页上
        if (data.items.length === 0 && data.count > 0 && page > 1) {
          setPage(page - 1)
          return
        }
        setItems(data.items)
        setCount(data.count)
        setError('')
        setVersion((v) => v + 1)
      })
      .catch((e: unknown) => {
        if (seq !== seqRef.current) return
        setItems([])
        setCount(0)
        setError(e instanceof Error ? e.message : String(e))
      })
      .finally(() => {
        if (seq === seqRef.current) setLoading(false)
      })
  }, [page, debounced, reloadTick])

  return { items, count, keyword, setKeyword, debounced, page, setPage, loading, error, version }
}

/**
 * 页码的可键盘激活属性。
 *
 * PaginationLink 渲染的是 <a>，本页分页又是页内 state（不进 URL，故不能像空间治理/工作台那样
 * 用 href 提供原生可激活能力）——无 href 的 <a> 只给 onClick 会变成「Tab 能聚焦、Enter/Space
 * 按不动」的那种假交互。这里补 role=button（AT 才按按钮宣告，而非无 href 的链接）+ 显式按键
 * 处理；Space 在无 href 的 <a> 上会滚动页面，故必须 preventDefault。
 */
function pageActivation(activate: () => void) {
  return {
    role: 'button' as const,
    onClick: activate,
    onKeyDown: (e: ReactKeyboardEvent) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault()
        activate()
      }
    },
  }
}

/**
 * 已删除列表外壳（两个 Tab 共用）：搜索 / 表头 / 骨架-空态-错误态 / 分页 / 行渐入的统一装配。
 * 实体差异全部由 props 注入——列定义（表头文案 + 网格模板）与行渲染函数；本组件只管列表骨架与分页交互。
 */
function DeletedList<T>({
  hint,
  searchPlaceholder,
  emptyText,
  searchEmpty,
  header,
  headerGrid,
  fetch,
  renderRow,
  onRestore,
  reloadTick,
}: {
  hint: string
  searchPlaceholder: string
  emptyText: string
  /** 搜索无结果文案（按当前关键词插值） */
  searchEmpty: (kw: string) => string
  header: string[]
  /** 表头网格模板（lg 起生效，窄屏隐藏表头），列数/宽度须与行组件的网格模板严格一致 */
  headerGrid: string
  fetch: (page: number, search: string) => Promise<DeletedPage<T>>
  renderRow: (item: T, onRestore: (item: T) => void) => RowElement
  onRestore: (item: T) => void
  /** 外部重取信号（恢复成功后 +1） */
  reloadTick: number
}) {
  const { t } = useTranslation()
  const { items, count, keyword, setKeyword, debounced, page, setPage, loading, error, version } =
    useDeletedList(fetch, reloadTick)
  const totalPages = Math.max(1, Math.ceil(count / PAGE_SIZE))
  const pageItems = buildPages(page, totalPages)

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      {/* 提示条：说明本 Tab 的恢复范围（空间=连带项目 / 项目=只列单独删除的） */}
      <p className="flex items-start gap-1.5 text-[12px] leading-relaxed text-mute">
        <Icon name="info" className="mt-0.5 size-3.5 shrink-0 text-faint" />
        {hint}
      </p>

      {/* 计数 + 搜索（搜索放右上角，对齐空间治理/项目治理的工具栏排布） */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <span className="text-[12px] text-faint">{t('restore.count', { count })}</span>
        <SearchInput
          value={keyword}
          onChange={setKeyword}
          placeholder={searchPlaceholder}
          size="sm"
          className="w-72"
        />
      </div>

      <section className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-line bg-surface">
        <div
          className={cn(
            'hidden items-center gap-2 border-b border-line px-4 py-2 text-[11px] font-medium text-faint lg:grid',
            headerGrid,
          )}
        >
          {header.map((h) => (
            <span key={h} className="truncate">
              {h}
            </span>
          ))}
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading ? (
            <SkeletonList count={6} bare />
          ) : error ? (
            <div className="flex min-h-0 items-center justify-center p-8">
              <Empty icon="restore" text={error} />
            </div>
          ) : items.length === 0 ? (
            <div className="flex min-h-0 items-center justify-center p-8">
              <Empty
                icon="restore"
                text={debounced.trim() ? searchEmpty(debounced.trim()) : emptyText}
              />
            </div>
          ) : (
            <RefreshFade version={version}>{items.map((it) => renderRow(it, onRestore))}</RefreshFade>
          )}
        </div>

        {/* 翻页条：仅多页时出现。分页为页内 state（不进 URL），故用 PaginationLink + onClick
            （无 href）+ pageActivation 补键盘激活——保留全站分页器外观，只把跳转语义换成 state 更新 */}
        {totalPages > 1 && (
          <div className="flex justify-end border-t border-line px-2 py-1.5">
            <Pagination className="m-0 w-fit">
              <PaginationContent>
                <PaginationItem>
                  <PaginationLink
                    aria-label={t('common.previousPage')}
                    aria-disabled={page <= 1 || undefined}
                    tabIndex={page <= 1 ? -1 : 0}
                    className={cn(page <= 1 && 'pointer-events-none opacity-50')}
                    {...pageActivation(() => page > 1 && setPage(page - 1))}
                  >
                    <Icon name="chevron-left" className="size-4" />
                  </PaginationLink>
                </PaginationItem>
                {pageItems.map((it, i) =>
                  it === '...' ? (
                    <PaginationEllipsis key={`ellipsis-${i}`} />
                  ) : (
                    <PaginationItem key={it}>
                      <PaginationLink
                        isActive={it === page}
                        tabIndex={0}
                        {...pageActivation(() => setPage(it))}
                      >
                        {it}
                      </PaginationLink>
                    </PaginationItem>
                  ),
                )}
                <PaginationItem>
                  <PaginationLink
                    aria-label={t('common.nextPage')}
                    aria-disabled={page >= totalPages || undefined}
                    tabIndex={page >= totalPages ? -1 : 0}
                    className={cn(page >= totalPages && 'pointer-events-none opacity-50')}
                    {...pageActivation(() => page < totalPages && setPage(page + 1))}
                  >
                    <Icon name="chevron-right" className="size-4" />
                  </PaginationLink>
                </PaginationItem>
              </PaginationContent>
            </Pagination>
          </div>
        )}
      </section>
    </div>
  )
}

/** 删除时间：humanize 相对时间 + 精确时间 tooltip（全站时间展示约定） */
function DeletedAt({ value }: { value: string }) {
  if (!value) return <span className="text-[12px] text-faint">-</span>
  return (
    <time dateTime={value} title={formatDateTime(value)} className="text-[12px] text-ink">
      {humanizeDateTime(value)}
    </time>
  )
}

/** 空间行网格：空间 / 创建者 / 连带恢复的项目 / 删除时间 / 操作（窄屏堆叠，lg 起 5 列）。
 *  项目列给到 2.4fr：该列是唯一的内容型胖列（项目名 chip 要换行铺开），其余列是定长标识 */
const NS_ROW_GRID =
  'grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-[minmax(0,1.6fr)_minmax(0,1.2fr)_minmax(0,2.4fr)_9rem_5rem] lg:items-center'
/** 空间表头网格：列宽须与 NS_ROW_GRID 严格一致 */
const NS_HEADER_GRID = 'lg:grid-cols-[minmax(0,1.6fr)_minmax(0,1.2fr)_minmax(0,2.4fr)_9rem_5rem]'

/**
 * 已删除空间行（React.memo）：行数据与回调引用未变则跳过重渲——翻页/搜索击键等页面级
 * state 变化不再连带已加载行重渲（onRestore 是稳定的 setState，行 memo 才能命中）。
 */
const NamespaceDeletedRow = memo(function NamespaceDeletedRow({
  item,
  onRestore,
  className,
  style,
}: {
  item: NamespaceModel
  onRestore: (item: NamespaceModel) => void
  /** RefreshFade 经 cloneElement 注入的渐入 class/延迟——须转发到根元素才生效 */
  className?: string
  style?: CSSProperties
}) {
  const { t } = useTranslation()
  return (
    <div className={cn(NS_ROW_GRID, 'border-b border-line px-4 py-2.5 last:border-b-0', className)} style={style}>
      <div className="flex min-w-0 items-center gap-1.5">
        <span className="truncate font-mono text-[13px] font-medium text-ink">{item.name}</span>
        {item.private && (
          <Tag tone="accent" dot={false}>
            {t('namespaces.privateTag')}
          </Tag>
        )}
      </div>
      <span className="truncate font-mono text-[12px] text-ink">{item.creatorEmail}</span>
      {/* 随空间连带恢复的项目**逐个列名**：恢复前操作者要判断「这个空间是不是我要找的那个」，
          用户数据里没有任何可核对的线索。计数给不出甄别信息，所以名字必须直接可见。
          换行铺开不设折叠上限——「到底哪几个会一起回来」正是这一步要确认的全部内容 */}
      {item.projects.length === 0 ? (
        // 空态仍用计数文案（0 个项目）而非「-」：明确告诉操作者「此空间不会带出任何项目」
        <span className="text-[12px] text-faint">{t('restore.itemCount', { count: 0 })}</span>
      ) : (
        <div
          className="flex min-w-0 flex-wrap items-center gap-1"
          title={t('restore.itemCount', { count: item.projects.length })}
        >
          {item.projects.map((p) => (
            // 只压宽度、不截断：Badge 是 inline-flex，truncate 的 nowrap 会把超长项目名直接裁掉——
            // 而「看清楚是哪几个项目」正是这一列存在的理由。超宽时让它在线内换行，宁可高一行不丢字。
            <Tag key={p.id} tone="info" dot={false} className="max-w-full">
              {p.name}
            </Tag>
          ))}
        </div>
      )}
      <DeletedAt value={item.deletedAt} />
      <div className="flex items-center">
        <Button variant="outline" size="sm" onClick={() => onRestore(item)}>
          <Icon name="restore" className="size-3.5" />
          {t('restore.restore')}
        </Button>
      </div>
    </div>
  )
})

/** 项目行网格：项目 / 所属空间 / 最后操作人 / 删除时间 / 操作（窄屏堆叠，lg 起 5 列） */
const PROJECT_ROW_GRID =
  'grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-[minmax(0,2fr)_minmax(0,1.5fr)_minmax(0,1.5fr)_9rem_5rem] lg:items-center'
/** 项目表头网格：列宽须与 PROJECT_ROW_GRID 严格一致 */
const PROJECT_HEADER_GRID = 'lg:grid-cols-[minmax(0,2fr)_minmax(0,1.5fr)_minmax(0,1.5fr)_9rem_5rem]'

/**
 * 已删除项目行（React.memo）：所属空间列是恢复请求的定位键（后端按「空间名 + 项目名」还原），
 * 故必须与项目名一起展示，缺了这一列操作者无法确认恢复的是哪个空间下的同名项目。
 */
const ProjectDeletedRow = memo(function ProjectDeletedRow({
  item,
  onRestore,
  className,
  style,
}: {
  item: ProjectModel
  onRestore: (item: ProjectModel) => void
  /** RefreshFade 经 cloneElement 注入的渐入 class/延迟——须转发到根元素才生效 */
  className?: string
  style?: CSSProperties
}) {
  const { t } = useTranslation()
  return (
    <div
      className={cn(PROJECT_ROW_GRID, 'border-b border-line px-4 py-2.5 last:border-b-0', className)}
      style={style}
    >
      <span className="truncate font-mono text-[13px] font-medium text-ink">{item.name}</span>
      {/* 空间边由列表查询带出（恢复请求据此定位），空值仅作兜底展示 */}
      <span className="truncate font-mono text-[12px] text-ink">{item.namespace?.name || '-'}</span>
      <span className="truncate font-mono text-[12px] text-mute">{item.updatedBy || '-'}</span>
      <DeletedAt value={item.deletedAt} />
      <div className="flex items-center">
        <Button variant="outline" size="sm" onClick={() => onRestore(item)}>
          <Icon name="restore" className="size-3.5" />
          {t('restore.restore')}
        </Button>
      </div>
    </div>
  )
})

/**
 * 恢复二次确认弹窗（两个 Tab 共用）：恢复会改动集群侧资源与 DB 状态，先确认再落请求。
 * 提交中/失败都保持打开（失败原因就地展示），确认前只读回显目标名，避免点错行恢复错对象。
 */
function RestoreConfirmDialog({
  open,
  onOpenChange,
  title,
  desc,
  target,
  submitting,
  error,
  onConfirm,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  desc: string
  /** 只读回显的恢复目标（空间名 / 空间名 + 项目名） */
  target: string
  submitting: boolean
  error: string
  onConfirm: () => void
}) {
  const { t } = useTranslation()
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent className="max-w-md">
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{desc}</AlertDialogDescription>
        </AlertDialogHeader>
        <p className="rounded-md border border-line bg-bg px-3 py-2 font-mono text-[12px] break-all text-ink">
          {target}
        </p>
        {error && (
          <p className="flex items-start gap-1.5 text-[12px] leading-relaxed text-err">
            <Icon name="info" className="mt-0.5 size-3.5 shrink-0" />
            {error}
          </p>
        )}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={submitting}>{t('common.cancel')}</AlertDialogCancel>
          <AlertDialogAction
            onClick={(e) => {
              // 阻止 Radix 默认关闭：提交中/失败都要保持弹窗打开（失败原因就地展示）
              e.preventDefault()
              onConfirm()
            }}
            disabled={submitting}
          >
            {submitting && <Icon name="loader" className="size-4 animate-spin" />}
            {t('restore.confirmAction')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

/**
 * 空间 Tab：已删除空间列表 + 逐行恢复。
 * 数据面 GET /api/admin/namespaces/deleted，恢复 POST /api/admin/namespaces/restore（body: name）。
 */
function NamespaceDeletedTab() {
  const { t } = useTranslation()
  // 待确认恢复的空间：null = 弹窗关闭
  const [target, setTarget] = useState<NamespaceModel | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  // 恢复成功后的重取信号：+1 让列表回读当前页（被恢复的行自然从「已删除」列表消失）
  const [reloadTick, setReloadTick] = useState(0)

  const fetchList = useCallback(async (page: number, search: string): Promise<DeletedPage<NamespaceModel>> => {
    const { data, error: err } = await api.GET(API.adminNamespaceDeleted, {
      params: { query: { page, pageSize: PAGE_SIZE, search: search.trim() || undefined } },
    })
    if (err) throw new Error(err.message ?? String(err))
    return { count: data?.count ?? 0, items: data?.items ?? [] }
  }, [])

  /** 提交恢复：成功后关弹窗 + 提示 + 回读列表；失败保留弹窗就地展示原因（可改目标重试） */
  const submit = async () => {
    if (!target) return
    setSubmitting(true)
    setError('')
    try {
      const { error: err } = await api.POST(API.adminNamespaceRestore, { body: { name: target.name } })
      if (err) {
        setError(err.message ?? t('restore.failed'))
        return
      }
      toast.success(t('restore.successNs', { name: target.name }))
      setTarget(null)
      setReloadTick((k) => k + 1)
    } catch (e) {
      // 必须接住异常：fetch 本身失败（断网 / 后端重启）走的是「抛」而非返回 {error}
      //（openapi-fetch 只挂了 onRequest/onResponse，没有 onError 中间件）
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      // 成败都要复位 submitting：不复位则确认按钮永久 disabled、取消按钮 disabled、
      // onOpenChange 的 !submitting 守卫又挡住 Esc/点遮罩 → 弹窗彻底锁死，只能刷新页面
      setSubmitting(false)
    }
  }

  return (
    <>
      <DeletedList<NamespaceModel>
        hint={t('restore.nsHint')}
        searchPlaceholder={t('restore.nsSearchPlaceholder')}
        emptyText={t('restore.nsEmpty')}
        searchEmpty={(kw) => t('restore.nsSearchEmpty', { kw })}
        header={[
          t('restore.nsColName'),
          t('restore.nsColOwner'),
          t('restore.nsColProjects'),
          t('restore.colDeletedAt'),
          t('restore.colAction'),
        ]}
        headerGrid={NS_HEADER_GRID}
        fetch={fetchList}
        renderRow={(it, onRestore) => <NamespaceDeletedRow key={it.id} item={it} onRestore={onRestore} />}
        onRestore={setTarget}
        reloadTick={reloadTick}
      />
      <RestoreConfirmDialog
        open={target !== null}
        onOpenChange={(o) => {
          // 提交中不允许关闭（请求已发出，关掉会让人以为没生效）
          if (!o && !submitting) {
            setTarget(null)
            setError('')
          }
        }}
        title={t('restore.confirmNsTitle')}
        desc={t('restore.confirmNsDesc')}
        target={target?.name ?? ''}
        submitting={submitting}
        error={error}
        onConfirm={() => void submit()}
      />
    </>
  )
}

/**
 * 项目 Tab：已删除项目列表（只列单独删除的行）+ 逐行恢复。
 * 数据面 GET /api/admin/projects/deleted，恢复 POST /api/admin/projects/restore（body: namespace + name）。
 */
function ProjectDeletedTab() {
  const { t } = useTranslation()
  const [target, setTarget] = useState<ProjectModel | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [reloadTick, setReloadTick] = useState(0)

  const fetchList = useCallback(async (page: number, search: string): Promise<DeletedPage<ProjectModel>> => {
    const { data, error: err } = await api.GET(API.adminProjectDeleted, {
      params: { query: { page, pageSize: PAGE_SIZE, search: search.trim() || undefined } },
    })
    if (err) throw new Error(err.message ?? String(err))
    return { count: data?.count ?? 0, items: data?.items ?? [] }
  }, [])

  const submit = async () => {
    if (!target) return
    const namespace = target.namespace?.name ?? ''
    setSubmitting(true)
    setError('')
    try {
      const { error: err } = await api.POST(API.adminProjectRestore, {
        body: { namespace, name: target.name },
      })
      if (err) {
        setError(err.message ?? t('restore.failed'))
        return
      }
      toast.success(t('restore.successProj', { namespace, name: target.name }))
      setTarget(null)
      setReloadTick((k) => k + 1)
    } catch (e) {
      // 同空间 Tab：网络层失败是「抛」不是 {error}，不接住就会锁死弹窗（详见 NamespaceDeletedTab.submit）
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <>
      <DeletedList<ProjectModel>
        hint={t('restore.projHint')}
        searchPlaceholder={t('restore.projSearchPlaceholder')}
        emptyText={t('restore.projEmpty')}
        searchEmpty={(kw) => t('restore.projSearchEmpty', { kw })}
        header={[
          t('restore.projColName'),
          t('restore.projColNamespace'),
          t('restore.projColOperator'),
          t('restore.colDeletedAt'),
          t('restore.colAction'),
        ]}
        headerGrid={PROJECT_HEADER_GRID}
        fetch={fetchList}
        renderRow={(it, onRestore) => <ProjectDeletedRow key={it.id} item={it} onRestore={onRestore} />}
        onRestore={setTarget}
        reloadTick={reloadTick}
      />
      <RestoreConfirmDialog
        open={target !== null}
        onOpenChange={(o) => {
          if (!o && !submitting) {
            setTarget(null)
            setError('')
          }
        }}
        title={t('restore.confirmProjTitle')}
        desc={t('restore.confirmProjDesc')}
        target={target ? `${target.namespace?.name ?? ''}/${target.name}` : ''}
        submitting={submitting}
        error={error}
        onConfirm={() => void submit()}
      />
    </>
  )
}

/**
 * 误删恢复（管理员后台 · 仅内置超管）
 *
 * 背景：删除空间/项目都是**软删除**——集群侧资源会被物理删除，但 DB 记录只打 `deleted_at`
 * 标记，因此可以把「最近删除」列出来逐条还原，而无需记忆删除前的名字。
 * - 两个 Tab：空间 / 项目，各自搜索 + 分页 + 行内恢复（只挂载当前 Tab，非活动 Tab 不白拉一次接口）。
 * - 项目 Tab 只列**单独删除**的项目（deleted_with_namespace=false）：随空间级联删除的那批由恢复空间
 *   连带还原，后端对它们单独恢复会硬拒 400（所属空间已删），列出来只是不可点的死行——列表口径与
 *   projectRepo.RestoreDeleted 的前置校验（空间须存活）严格一致。
 * - 权限：后端 RequireSuperAdmin 强制名单（普通 admin 亦 403），前端 RequireSuperAdmin 路由守卫
 *   + 侧栏 superOnly 隐藏做可见性/可访问性双保险。
 * - 恢复只重建 k8s 命名空间 + docker secret 并清软删标记，**不重建 helm release**，故恢复后的项目
 *   仍需重新部署才能运行（各 Tab 提示条显式说明）。
 */
export function RestoreDeleted() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<TabKey>('namespaces')

  return (
    <div className="flex h-full flex-col gap-4">
      {/* 页头：只有标题——「仅内置超管」标识已由侧栏「超管」角标常驻承载，
          这里再放一个同义标签属于重复表达（进得来这个页面的人必然是超管） */}
      <h2 className="shrink-0 text-[16px] font-semibold text-ink">{t('restore.title')}</h2>

      {/* 实体 Tab：空间 / 项目 */}
      <div className="border-b border-line">
        <Tabs value={tab} onValueChange={(v) => setTab(v as TabKey)}>
          <TabsList variant="line">
            <TabsTrigger value="namespaces" className="group">
              <Icon name="namespace" className="size-4 group-data-[state=active]:text-primary" />
              {t('restore.tabNamespaces')}
            </TabsTrigger>
            <TabsTrigger value="projects" className="group">
              <Icon name="project" className="size-4 group-data-[state=active]:text-primary" />
              {t('restore.tabProjects')}
            </TabsTrigger>
          </TabsList>
        </Tabs>
      </div>

      {tab === 'namespaces' ? <NamespaceDeletedTab /> : <ProjectDeletedTab />}
    </div>
  )
}
