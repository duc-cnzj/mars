import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSearchParams } from 'react-router-dom'
import { formatDateTime } from '@/lib/format'
import { toHumanizeDateTime } from '@/lib/humanizeDateTime'
import type {
  components,
  PathsApiEventsGetParametersQueryActionTypes as SchemaActionTypes,
} from '../../api/schema'
import { api } from '../../api/client'
import { getToken } from '../../api/token'
import { toast } from '@/lib/toast'
import { Icon } from '../../components/icons'
import { Empty, SkeletonList, Tag, type Tone } from '../../components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { SearchInput } from '../../components/SearchInput'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/shadcn/alert-dialog'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { DiffModal } from './DiffModal'
import { EventTitle } from './EventTitle'
import { AsciinemaPlayer } from './AsciinemaPlayer'
import { useAuth } from '../auth/AuthContext'

type EventModel = components['schemas']['types.EventModel']

const PAGE_SIZE = 15

/** action 枚举取值（与后端 enum 字符串一致，schema 仅声明文件，运行时不能 import 枚举） */
const ACTION_VALUES = [
  'Unknown',
  'Create',
  'Update',
  'Delete',
  'Upload',
  'Download',
  'DryRun',
  'Shell',
  'Login',
  'CancelDeploy',
  'Exec',
  'ForceDeletePod',
] as const
type ActionType = (typeof ACTION_VALUES)[number]

/** action 枚举 → 词条键（用于标签与筛选），as const 保字面量类型以过 i18next 校验 */
const ACTION_KEY = {
  Unknown: 'events.actionUnknown',
  Create: 'events.actionCreate',
  Update: 'events.actionUpdate',
  Delete: 'events.actionDelete',
  Upload: 'events.actionUpload',
  Download: 'events.actionDownload',
  DryRun: 'events.actionDryRun',
  Shell: 'events.actionShell',
  Login: 'events.actionLogin',
  CancelDeploy: 'events.actionCancelDeploy',
  Exec: 'events.actionExec',
  ForceDeletePod: 'events.actionForceDeletePod',
} as const

/**
 * 筛选标签展示顺序：纯前端主观按使用频率排（高频在前），与 ACTION_VALUES 枚举顺序解耦。
 * 想调整顺序直接改这个数组即可，无需动数据流。
 */
const FILTER_ORDER: readonly ActionType[] = [
  'Create',
  'Update',
  'Login',
  'Download',
  'Upload',
  'Shell',
  'Delete',
  'Exec',
  'CancelDeploy',
  'DryRun',
  'ForceDeletePod',
]

const ACTION_TONE: Record<ActionType, Tone> = {
  Unknown: 'mute',
  Create: 'info',
  Update: 'ok',
  Delete: 'err',
  Upload: 'accent',
  Download: 'accent',
  DryRun: 'accent',
  Shell: 'warn',
  Login: 'info',
  CancelDeploy: 'warn',
  Exec: 'info',
  ForceDeletePod: 'err',
}

/**
 * 事件日志：动作类型多选标签 + 内容搜索 + 分页列表。
 * 行内操作：查看改动（diff）/ 查看操作记录（终端回放）/ 下载文件 / 删除文件。
 * 列表滚动限制在固定高度容器内（对齐旧版 div 内无限加载），筛选变化回到第 1 页。
 */
export function Events() {
  const { t } = useTranslation()
  // 事件列表对普通用户开放，但删除文件仅 admin（后端强制，前端不展示必然 403 的入口）
  const { user } = useAuth()
  const isAdmin = user?.roles.includes('mars_admin') ?? false

  const [items, setItems] = useState<EventModel[]>([])
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  // 筛选/搜索从 URL query 恢复：刷新/分享链接/前进后退都保留上次条件。
  // 首帧即带上恢复值（debouncedSearch 同步），避免先空载再按条件重拉的闪烁
  const [searchParams, setSearchParams] = useSearchParams()
  const initSearch = searchParams.get('search') ?? ''
  const initActionTypes = (searchParams.get('actionTypes') ?? '')
    .split(',')
    .filter((v): v is ActionType => v !== '' && ACTION_VALUES.includes(v as ActionType))
  /** 多选动作类型：空数组 = 全部（不传 actionType） */
  const [actionType, setActionType] = useState<ActionType[]>(initActionTypes)
  const [search, setSearch] = useState(initSearch)
  const [debouncedSearch, setDebouncedSearch] = useState(initSearch)
  const [initialLoading, setInitialLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [diskUsage, setDiskUsage] = useState('')
  const sentinelRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const busyRef = useRef(false)
  const reqIdRef = useRef(0)
  const refreshingRef = useRef(false)

  const [diff, setDiff] = useState<{
    open: boolean
    username: string
    message: string
    old: string
    new: string
  }>({ open: false, username: '', message: '', old: '', new: '' })
  const [record, setRecord] = useState<{ open: boolean; username: string; message: string }>({
    open: false,
    username: '',
    message: '',
  })
  const [records, setRecords] = useState<string[]>([])
  const [recordKey, setRecordKey] = useState(0)

  const fetchList = useCallback(
    async (p: number, act: ActionType[], kw: string, append: boolean) => {
      const id = ++reqIdRef.current
      busyRef.current = true
      if (append) setLoadingMore(true)
      else setInitialLoading(true)
      try {
        const { data, error } = await api.GET('/api/events', {
          params: {
            query: {
              page: p,
              pageSize: PAGE_SIZE,
              actionTypes: act.length ? (act as unknown as SchemaActionTypes[]) : undefined,
              search: kw.trim() || undefined,
            },
          },
        })
        if (error) throw new Error(error.message ?? String(error))
        if (!data) return
        if (id !== reqIdRef.current) return // 已有更新的请求，丢弃过期响应（防止筛选切换竞态）
        setItems((prev) => (append ? [...prev, ...data.items] : data.items))
        // 事件接口无 total：回满一页即认为可能还有下一页
        setHasMore(data.items.length === PAGE_SIZE)
      } catch (e) {
        if (id !== reqIdRef.current) return
        toast.error(e instanceof Error ? e.message : String(e))
      } finally {
        if (id === reqIdRef.current) {
          busyRef.current = false
          setInitialLoading(false)
          setLoadingMore(false)
        }
      }
    },
    [],
  )

  // 文件占用仅 admin 可查（后端 fileSvc.Authorize 除 MaxUploadSize 外全 admin），
  // 普通用户跳过请求，避免必 403 的调用与未捕获的 rejection
  useEffect(() => {
    if (!isAdmin) return
    void api.GET('/api/files/disk_info').then(({ data }) => {
      if (data) setDiskUsage(data.humanizeUsage)
    })
  }, [isAdmin])

  // 搜索防抖 400ms
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 400)
    return () => clearTimeout(timer)
  }, [search])

  // 筛选/搜索任一变化 → 同步 URL query。replace 避免每次输入堆一条历史；
  // 空值直接删参数，URL 保持干净；其余参数原样保留
  useEffect(() => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        if (search) next.set('search', search)
        else next.delete('search')
        if (actionType.length) next.set('actionTypes', actionType.join(','))
        else next.delete('actionTypes')
        return next
      },
      { replace: true },
    )
  }, [search, actionType, setSearchParams])

  /** 筛选/搜索指纹：任一变化都视为新的过滤条件 */
  const filterKey = useMemo(
    () => `${debouncedSearch}|${actionType.join(',')}`,
    [debouncedSearch, actionType],
  )

  // 筛选条件变化回到第 1 页（由 page 变化驱动下面的 fetch）。
  // 同时回滚滚动到顶部：否则深滚到底部后切换筛选，page-1 短列表的哨兵仍在视口内，
  // 会连续翻页请求直至 hasMore 为 false（用户不动滚动也会一直拉取）
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: 0 })
    setPage(1)
  }, [filterKey])

  useEffect(() => {
    // 刷新已直接拉取第 1 页，跳过 setPage(1) 触发的重复请求
    if (page === 1 && refreshingRef.current) return
    void fetchList(page, actionType, debouncedSearch, page > 1)
  }, [page, filterKey, fetchList]) // eslint-disable-line react-hooks/exhaustive-deps

  // 触底加载：sentinel 进入列表容器视口（提前 200px）且非忙碌/非刷新时翻下一页
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
      { root, rootMargin: '200px' },
    )
    io.observe(el)
    return () => io.disconnect()
  }, [hasMore, initialLoading, loadingMore, refreshing])

  const refresh = async () => {
    if (refreshingRef.current) return
    refreshingRef.current = true
    setRefreshing(true)
    // 重置期间同步锁 busyRef，阻止触底加载在刷新时连环翻页
    busyRef.current = true
    // 回到顶部：让哨兵离开视口，刷新后不会自动加载后续页
    if (scrollRef.current) scrollRef.current.scrollTop = 0
    setPage(1)
    await fetchList(1, actionType, debouncedSearch, false)
    busyRef.current = false
    refreshingRef.current = false
    setRefreshing(false)
  }

  const viewDiff = async (id: number) => {
    try {
      const { data, error } = await api.GET('/api/events/{id}', { params: { path: { id } } })
      if (error) throw new Error(error.message ?? String(error))
      if (!data?.item) return
      setDiff({
        open: true,
        username: data.item.username,
        message: data.item.message,
        old: data.item.old ?? '',
        new: data.item.new ?? '',
      })
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    }
  }

  const openRecord = async (item: EventModel) => {
    setRecord({ open: true, username: item.username, message: item.message })
    try {
      const { data, error } = await api.GET('/api/record_files/{id}', {
        params: { path: { id: item.fileId } },
      })
      if (error) throw new Error(error.message ?? String(error))
      if (data) {
        setRecords(data.items)
        setRecordKey(0)
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    }
  }

  const deleteFile = async (item: EventModel) => {
    try {
      const { error } = await api.DELETE('/api/files/{id}', {
        params: { path: { id: item.fileId } },
      })
      if (error) throw new Error(error.message ?? String(error))
      setItems((prev) => prev.map((v) => (v.id === item.id ? { ...v, fileId: 0 } : v)))
      toast.success(t('events.deleteSuccess', { name: item.file?.path ?? String(item.fileId) }))
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    }
  }

  const downloadFile = async (fileId: number) => {
    try {
      const res = await fetch(`/api/download_file/${fileId}`, {
        headers: { Authorization: getToken(), 'X-Requested-With': 'XMLHttpRequest' },
      })
      if (!res.ok) throw new Error('download failed')
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `file-${fileId}`
      a.click()
      URL.revokeObjectURL(url)
      toast.success(t('events.downloadSuccess'))
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    }
  }

  /** 筛选标签：平铺的可点标签（多选），"全部"= 空选择 */
  const toggleAction = (v: ActionType) => {
    setActionType((prev) => (prev.includes(v) ? prev.filter((x) => x !== v) : [...prev, v]))
  }
  const chipCls = (active: boolean) =>
    `cursor-pointer select-none rounded-full border px-3 py-1 text-[12px] transition-colors ${
      active
        ? 'border-primary bg-primary-soft font-medium text-primary'
        : 'border-line bg-surface text-mute hover:border-primary hover:text-primary'
    }`

  const renderActions = (item: EventModel) => {
    const act = item.action as unknown as ActionType
    const hasFile = !!item.file && item.fileId > 0
    const isShell = act === 'Shell' || act === 'Exec'
    const isUpload = act === 'Upload' || act === 'Download'
    return (
      <div className="flex shrink-0 flex-wrap items-center gap-2">
        {/* 终端回放对所有者开放（后端 ShowRecords 先 allowlist 过 admin 门禁、
            再 RequireFileAccess 判定所有者/admin，普通用户只能回放自己的会话）；
            删除文件仍仅 admin（后端 Delete 为 admin 门禁）。 */}
        {hasFile && isShell && (
          <>
            {/* 时长+大小 内嵌按钮，对齐旧版「查看操作记录 (时长: X, 大小: Y)」 */}
            <Button size="sm" variant="outline" onClick={() => openRecord(item)}>
              {t('events.viewRecord')}
              {item.duration && (
                <span className="ml-1 text-[10px] text-mute">
                  ({t('events.duration')}: {item.duration}, {t('events.size')}:{' '}
                  {item.file?.humanizeSize})
                </span>
              )}
            </Button>
            {isAdmin && (
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button size="sm" variant="destructive">
                  {t('events.deleteFile')}
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>
                    {t('events.deleteFileConfirm', { name: item.file?.path ?? String(item.fileId) })}
                  </AlertDialogTitle>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
                  <AlertDialogAction
                    className="bg-destructive text-white hover:bg-destructive/90"
                    onClick={() => deleteFile(item)}
                  >
                    {t('common.delete')}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
            )}
          </>
        )}
        {hasFile && isUpload && (
          <>
            <Button size="sm" variant="outline" onClick={() => downloadFile(item.fileId)}>
              {t('events.download')}
            </Button>
            {isAdmin && (
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button size="sm" variant="destructive">
                    {t('events.deleteFile')}
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>
                      {t('events.deleteFileConfirm', { name: item.file?.path ?? String(item.fileId) })}
                    </AlertDialogTitle>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
                    <AlertDialogAction
                      className="bg-destructive text-white hover:bg-destructive/90"
                      onClick={() => deleteFile(item)}
                    >
                      {t('common.delete')}
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            )}
          </>
        )}
        {item.hasDiff && (
          <Button size="sm" variant="outline" onClick={() => viewDiff(item.id)}>
            {t('events.viewDiff')}
          </Button>
        )}
      </div>
    )
  }

  return (
    // 父级 main 是 min-h-screen 下的 flex-1（高度 auto），flex 链没有有界高度，
    // 用 calc(100dvh - 100px) 让内部滚动容器 flex-1 min-h-0 overflow-y-auto 真正可滚动。
    <div className="flex h-[calc(100dvh_-_100px)] flex-col gap-3">
      {/* 工具栏 */}
      <div className="flex shrink-0 flex-col gap-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-[16px] font-semibold text-ink">{t('events.title')}</h2>
          <div className="flex flex-wrap items-center gap-2">
            <SearchInput
              value={search}
              onChange={setSearch}
              placeholder={t('events.searchPlaceholder')}
              className="w-56"
            />
            {diskUsage && (
              <span className="text-[12px] text-mute">
                {t('events.diskUsage')}: <span className="font-mono text-primary">{diskUsage}</span>
              </span>
            )}
            <Button size="sm" variant="outline" disabled={refreshing} onClick={refresh}>
              <Icon name="refresh" className="size-4" />
              {t('events.refresh')}
            </Button>
          </div>
        </div>
        {/* 动作类型多选标签 */}
        <div className="flex flex-wrap items-center gap-1.5">
          <span className="text-[12px] text-faint">{t('events.actionFilter')}</span>
          <button
            type="button"
            onClick={() => setActionType([])}
            className={chipCls(actionType.length === 0)}
          >
            {t('events.actionAll')}
          </button>
          {FILTER_ORDER.map((v) => (
            <button
              key={v}
              type="button"
              onClick={() => toggleAction(v)}
              className={chipCls(actionType.includes(v))}
            >
              {t(ACTION_KEY[v])}
            </button>
          ))}
        </div>
      </div>

      {/* 列表：flex 撑满剩余高度，容器内滚动（对齐旧版 div 内无限加载） */}
      <div
        ref={scrollRef}
        className="min-h-0 flex-1 overflow-y-auto overscroll-contain rounded-lg border border-line bg-surface"
      >
        {initialLoading ? (
          <SkeletonList count={8} bare />
        ) : items.length === 0 ? (
          <div className="flex h-full items-center justify-center p-8">
            <Empty text={t('common.empty')} icon="pulse" />
          </div>
        ) : (
          <div className="divide-y divide-line">
            {items.map((item, index) => {
              const act = item.action as unknown as ActionType
              return (
                <div
                  key={item.id}
                  // 进入动画：淡入 + 上浮，按行错峰（封顶 10 行×30ms，深滚加载的行不滞后）
                  className="animate-list-in flex items-start justify-between gap-4 px-5 py-3.5 transition-colors hover:bg-raised"
                  style={{ animationDelay: `${Math.min(index, 10) * 30}ms` }}
                >
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="text-[13px] font-semibold text-primary">{item.username}</span>
                      <Tag tone={ACTION_TONE[act]}>{t(ACTION_KEY[act])}</Tag>
                      {/* 相对时间由前端基于 createdAt 本地化计算（后端 event_at 写死中文，
                          新 UI 不再消费），与下方精确时间并存，对齐旧版布局 */}
                      <span className="text-[11px] text-ink">
                        {toHumanizeDateTime(item.createdAt)}
                      </span>
                      <span className="font-mono text-[11px] text-faint">
                        {formatDateTime(item.createdAt)}
                      </span>
                    </div>
                    <div className="mt-1 break-words text-[13px] text-mute">{item.message}</div>
                  </div>
                  {renderActions(item)}
                </div>
              )
            })}
          </div>
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

      {/* 改动 diff */}
      <DiffModal
        open={diff.open}
        onClose={() => setDiff((d) => ({ ...d, open: false }))}
        username={diff.username}
        message={diff.message}
        oldText={diff.old}
        newText={diff.new}
      />

      {/* 操作记录（终端回放） */}
      <Dialog
        open={record.open}
        onOpenChange={(o) => {
          if (!o) {
            setRecord({ open: false, username: '', message: '' })
            setRecords([])
            setRecordKey(0)
          }
        }}
      >
        <DialogContent className="flex h-[calc(100vh-2rem)] w-[calc(100vw-2rem)] max-w-[calc(100vw-2rem)] flex-col sm:max-w-[calc(100vw-2rem)]">
          <DialogHeader>
            <DialogTitle className="flex items-center">
              <EventTitle username={record.username} message={record.message} />
            </DialogTitle>
          </DialogHeader>
          {/* 内容区：撑满剩余高度，容器内滚动（对齐旧版全宽 Drawer 行为） */}
          <div className="min-h-0 flex-1 overflow-auto overscroll-contain rounded-md border border-line bg-surface">
            <div className="flex flex-col gap-3 p-4">
              {records.length > 1 && (
                <div className="flex flex-wrap gap-1.5">
                  {records.map((_, i) => (
                    <button
                      key={i}
                      type="button"
                      onClick={() => setRecordKey(i)}
                      className={`rounded-md border px-2.5 py-1 text-[12px] transition-colors ${
                        recordKey === i
                          ? 'border-primary bg-primary-soft text-primary'
                          : 'border-line text-mute hover:border-primary hover:text-primary'
                      }`}
                    >
                      {t('events.segment')} {i + 1}
                    </button>
                  ))}
                </div>
              )}
              {/* 终端超宽时横向滚动，不撑破弹窗 */}
              <div className="overflow-x-auto">
                {records[recordKey] ? (
                  <AsciinemaPlayer key={recordKey} src={records[recordKey]} />
                ) : (
                  <div className="px-3 py-4 text-[12px] text-mute">{t('common.empty')}</div>
                )}
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
