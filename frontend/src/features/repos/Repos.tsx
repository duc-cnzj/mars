import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSearchParams } from 'react-router-dom'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { Icon } from '@/components/Icons'
import { Empty, SkeletonList, Tag } from '@/components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { SearchInput } from '@/components/SearchInput'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/shadcn/alert-dialog'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { copyText } from '@/lib/copy'
import { RepoFormModal } from './RepoFormModal'

type RepoModel = components['schemas']['types.RepoModel']

const PAGE_SIZE = 15

/**
 * Repo 管理：列表（搜索 + 分页）+ 添加/编辑 + 克隆 + 启用/禁用 + 删除。
 * git 关联信息展示在行内；启用状态影响后续部署链路。
 */
export function Repos() {
  const { t } = useTranslation()

  const [items, setItems] = useState<RepoModel[]>([])
  const [count, setCount] = useState(0)
  const [page, setPage] = useState(1)
  // 搜索关键词从 URL query 恢复：刷新/分享链接/前进后退都保留上次搜索
  const [searchParams, setSearchParams] = useSearchParams()
  const [keyword, setKeyword] = useState(() => searchParams.get('name') ?? '')
  /** 搜索防抖 400ms（对齐 events 页）：避免每敲一个键就触发一次「整列表变骨架」的重拉 */
  const [debouncedKeyword, setDebouncedKeyword] = useState(() => searchParams.get('name') ?? '')
  const [initialLoading, setInitialLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [togglingId, setTogglingId] = useState(0)
  const sentinelRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const busyRef = useRef(false)
  const refreshingRef = useRef(false)
  /** 上次实际发出的 (debouncedKeyword, page)：刷新完成后的 effect 重跑据此去重，不重复拉第 1 页 */
  const lastKwRef = useRef('')
  const lastPageRef = useRef(0)
  const hasMore = items.length < count

  const [formOpen, setFormOpen] = useState(false)
  const [editItem, setEditItem] = useState<RepoModel | undefined>()
  const [cloneOpen, setCloneOpen] = useState(false)
  const [cloneId, setCloneId] = useState(0)
  const [cloneName, setCloneName] = useState('')

  const fetchList = useCallback(
    async (p: number, kw?: string, append = false) => {
      busyRef.current = true
      if (append) setLoadingMore(true)
      else setInitialLoading(true)
      try {
        const { data, error } = await api.GET('/api/repos', {
          params: { query: { page: p, pageSize: PAGE_SIZE, name: kw?.trim() || undefined } },
        })
        if (error) throw new Error(error.message ?? String(error))
        if (!data) return
        setItems((prev) => (append ? [...prev, ...data.items] : data.items))
        setCount(data.count)
      } catch (e) {
        toast.error(e instanceof Error ? e.message : String(e))
      } finally {
        busyRef.current = false
        setInitialLoading(false)
        setLoadingMore(false)
      }
    },
    [],
  )

  // 搜索防抖 400ms：keyword 同步 URL（立即），实际拉取走 debouncedKeyword
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedKeyword(keyword), 400)
    return () => clearTimeout(timer)
  }, [keyword])

  useEffect(() => {
    // 刷新已直接拉取第 1 页，跳过 setPage(1) 触发的重复请求
    if (page === 1 && refreshingRef.current) return
    // 防抖未落定（keyword 已变、debouncedKeyword 尚未跟上）不拉：否则 onChange 里 setPage(1)
    // 会让「深翻页后首次输入」以旧关键词抢拉一次第 1 页（旧数据闪现 + 一次浪费请求）
    if (keyword !== debouncedKeyword) return
    // 同 (debouncedKeyword, page) 去重：刷新完成后的 effect 重跑（refreshing 翻转触发）不重复拉第 1 页
    if (lastKwRef.current === debouncedKeyword && lastPageRef.current === page) return
    lastKwRef.current = debouncedKeyword
    lastPageRef.current = page
    void fetchList(page, debouncedKeyword, page > 1)
  }, [page, debouncedKeyword, keyword, fetchList, refreshing])

  // 搜索关键词变化 → 同步 URL query。replace 避免每次输入堆一条历史；
  // 空值直接删参数，URL 保持干净；其余参数原样保留
  useEffect(() => {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        if (keyword) next.set('name', keyword)
        else next.delete('name')
        return next
      },
      { replace: true },
    )
  }, [keyword, setSearchParams])

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

  const refresh = async () => {
    if (refreshingRef.current) return
    refreshingRef.current = true
    setRefreshing(true)
    // 重置期间同步锁 busyRef，阻止触底加载在刷新时连环翻页
    busyRef.current = true
    // 回到顶部：让哨兵离开视口，刷新后不会自动加载后续页
    if (scrollRef.current) scrollRef.current.scrollTop = 0
    setPage(1)
    // 预标记本次刷新要拉的 (debouncedKeyword, 1)，刷新完成后的 effect 重跑据此去重
    lastKwRef.current = debouncedKeyword
    lastPageRef.current = 1
    await fetchList(1, debouncedKeyword, false)
    busyRef.current = false
    refreshingRef.current = false
    setRefreshing(false)
  }

  const openCreate = () => {
    setEditItem(undefined)
    setFormOpen(true)
  }

  const openEdit = (item: RepoModel) => {
    // 直接用列表项数据打开弹窗，不再 await GET /api/repos/{id}：
    // 列表项 schema 已是完整 RepoModel（含 marsConfig），detail 响应纯属冗余，
    // 且它内嵌 valuesYaml，仓库配置大时响应体积大、慢请求会把
    // 「点击编辑 → 弹窗出现」整个卡住（无反馈直到请求完毕）。即时打开体验更好。
    setEditItem(item)
    setFormOpen(true)
  }

  const toggle = async (item: RepoModel) => {
    setTogglingId(item.id)
    try {
      const { error } = await api.POST('/api/repos/toggle_enabled', {
        body: { id: item.id, enabled: !item.enabled },
      })
      if (error) throw new Error(error.message ?? String(error))
      setItems((prev) =>
        prev.map((it) => (it.id === item.id ? { ...it, enabled: !item.enabled } : it)),
      )
      // toggle 后新状态是 !item.enabled：旧为启用 → 已禁用；旧为禁用 → 已启用
      toast.success(
        item.enabled
          ? t('repos.disableSuccess', { name: item.name })
          : t('repos.enableSuccess', { name: item.name }),
      )
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setTogglingId(0)
    }
  }

  const remove = async (item: RepoModel) => {
    try {
      const { error } = await api.DELETE('/api/repos/{id}', {
        params: { path: { id: item.id } },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('repos.deleteSuccess', { name: item.name }))
      refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    }
  }

  const clone = async () => {
    if (cloneId <= 0 || !cloneName.trim()) {
      toast.error(t('repos.cloneEmptyError'))
      return
    }
    try {
      const { error } = await api.POST('/api/repos/clone', {
        body: { id: cloneId, name: cloneName.trim() },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('repos.cloneSuccess', { name: cloneName.trim() }))
      setCloneOpen(false)
      setCloneName('')
      refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    }
  }

  const copyId = async (id: number) => {
    const ok = await copyText(String(id))
    if (ok) toast.success(t('repos.copyId'))
    else toast.error(t('common.retry'))
  }

  return (
    // 父级 main 是 min-h-screen 下的 flex-1（高度 auto），flex 链没有有界高度，
    // 用 calc(100dvh - 100px) 让内部滚动容器 flex-1 min-h-0 overflow-y-auto 真正可滚动。
    <div className="flex h-[calc(100dvh_-_100px)] flex-col gap-3">
      {/* 工具栏 */}
      <div className="flex shrink-0 flex-col gap-3">
        {/* 标题 + 刷新 + 添加 */}
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-[16px] font-semibold text-ink">{t('repos.title')}</h2>
          <div className="flex flex-wrap items-center gap-2">
            <Button size="sm" variant="outline" disabled={refreshing} onClick={refresh}>
              <Icon name="refresh" className="size-4" />
              {t('common.refresh')}
            </Button>
            <Button size="sm" variant="default" onClick={openCreate}>
              <Icon name="plus" className="size-4" />
              {t('repos.add')}
            </Button>
          </div>
        </div>

        {/* 搜索 */}
        <SearchInput
          value={keyword}
          onChange={(v) => {
            setKeyword(v)
            setPage(1)
          }}
          placeholder={t('repos.searchPlaceholder')}
          className="max-w-xs"
        />
      </div>

      {/* 列表：flex 撑满剩余高度，容器内滚动（对齐旧版 div 内无限加载） */}
      <div
        ref={scrollRef}
        className="min-h-0 flex-1 overflow-y-auto rounded-lg border border-line bg-surface"
      >
        {initialLoading ? (
          <SkeletonList count={8} bare />
        ) : items.length === 0 ? (
          <div className="flex h-full items-center justify-center p-8">
            <Empty text={t('common.empty')} icon="repo" />
          </div>
        ) : (
          <div className="divide-y divide-line">
            {items.map((item, index) => (
            <div
              key={item.id}
              // 进入动画：淡入 + 上浮，按行错峰（封顶 10 行×30ms，深滚加载的行不滞后）
              className="animate-list-in flex items-start justify-between gap-4 px-5 py-4 transition-colors hover:bg-raised"
              style={{ animationDelay: `${Math.min(index, 10) * 30}ms` }}
            >
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-[13px] font-medium text-ink">{item.name}</span>
                  <Tag tone={item.enabled ? 'ok' : 'mute'}>
                    {item.enabled ? t('repos.enable') : t('repos.disable')}
                  </Tag>
                  <span className="text-[11px] text-faint">
                    (id:{' '}
                    <button
                      type="button"
                      onClick={() => copyId(item.id)}
                      title={t('repos.copyId')}
                      className="inline-flex items-center gap-0.5 font-mono text-primary hover:text-primary-strong"
                    >
                      {item.id}
                      <Icon name="copy" className="text-[10px]" />
                    </button>
                    )
                  </span>
                </div>
                <div className="mt-1 text-[12px] text-mute">
                  {item.gitProjectId ? (
                    <span>
                      {t('repos.gitLinked', { name: item.gitProjectName })},{' '}
                      {t('repos.gitProjectId', { id: item.gitProjectId })}
                    </span>
                  ) : (
                    <span>{t('repos.noGit')}</span>
                  )}
                </div>
              </div>
              <div className="flex shrink-0 flex-wrap items-center gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => {
                    setCloneId(item.id)
                    setCloneName('')
                    setCloneOpen(true)
                  }}
                >
                  {t('repos.clone')}
                </Button>
                <Button size="sm" variant="outline" onClick={() => openEdit(item)}>
                  {t('common.edit')}
                </Button>
                {item.enabled ? (
                  // 禁用需确认：避免误点直接停掉线上的 repo（启用不弹）
                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <Button
                        size="sm"
                        variant="destructive"
                        disabled={togglingId === item.id}
                      >
                        {togglingId === item.id && <Icon name="loader" className="size-4 animate-spin" />}
                        {t('repos.disable')}
                      </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>
                          {t('repos.disableConfirm', { name: item.name })}
                        </AlertDialogTitle>
                        <AlertDialogDescription>
                          {t('repos.disableTip')}
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
                        <AlertDialogAction
                          className="bg-destructive text-white hover:bg-destructive/90"
                          onClick={() => toggle(item)}
                        >
                          {t('repos.disable')}
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                ) : (
                  <Button
                    size="sm"
                    variant="default"
                    disabled={togglingId === item.id}
                    onClick={() => toggle(item)}
                  >
                    {togglingId === item.id && <Icon name="loader" className="size-4 animate-spin" />}
                    {t('repos.enable')}
                  </Button>
                )}
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button size="sm" variant="destructive">
                      {t('repos.delete')}
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>{t('repos.deleteConfirm', { name: item.name })}</AlertDialogTitle>
                      <AlertDialogDescription>{t('repos.deleteTip')}</AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
                      <AlertDialogAction
                        className="bg-destructive text-white hover:bg-destructive/90"
                        onClick={() => remove(item)}
                      >
                        {t('common.delete')}
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </div>
          ))}
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

      {/* 添加 / 编辑 */}
      <RepoFormModal
        open={formOpen}
        editItem={editItem}
        onClose={() => setFormOpen(false)}
        onSaved={refresh}
      />

      {/* 克隆 */}
      <Dialog open={cloneOpen} onOpenChange={(o) => !o && setCloneOpen(false)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('repos.cloneTitle')}</DialogTitle>
          </DialogHeader>
          <Input
            aria-label={t('repos.clonePlaceholder')}
            value={cloneName}
            onChange={(e) => setCloneName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && clone()}
            placeholder={t('repos.clonePlaceholder')}
            autoFocus
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setCloneOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="default" onClick={clone}>
              {t('common.confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
