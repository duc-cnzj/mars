import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'
import { formatDateTime } from '@/lib/format'
import { toast } from '@/lib/toast'
import { Loader2 } from 'lucide-react'
import type { components } from '../../api/schema'
import { api } from '../../api/client'
import { Icon } from '../../components/icons'
import { Empty, SkeletonList, Tag } from '../../components/ui'
import { Button } from '@/components/ui/shadcn/button'
import { Input } from '@/components/ui/shadcn/input'
import { Textarea } from '@/components/ui/shadcn/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/shadcn/select'
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { copyText } from '../../utils/copy'

type TokenModel = components['schemas']['types.AccessTokenModel']
type Unit = 'month' | 'day' | 'hour' | 'minute' | 'second'

const PAGE_SIZE = 15

/** 单位 → 词条键，避免动态 key 破坏 i18next 类型校验 */
const UNIT_KEY = {
  month: 'tokens.unitMonth',
  day: 'tokens.unitDay',
  hour: 'tokens.unitHour',
  minute: 'tokens.unitMinute',
  second: 'tokens.unitSecond',
} as const

/** 按单位把数值换算成秒（month 用日历月差，其余按固定进制） */
function getSeconds(num: number, unit: Unit): number {
  switch (unit) {
    case 'month':
      return dayjs().add(num, 'months').diff(dayjs(), 'seconds')
    case 'day':
      return 86400 * num
    case 'hour':
      return 3600 * num
    case 'minute':
      return 60 * num
    case 'second':
      return num
  }
}

/**
 * 访问令牌管理：列表（分页）+ 创建 + 续租 + 撤销。
 * token 点击即复制；撤销前二次确认；状态标签区分 已撤销/已过期/即将过期。
 */
export function AccessTokenManager() {
  const { t } = useTranslation()

  const [items, setItems] = useState<TokenModel[]>([])
  const [count, setCount] = useState(0)
  const [page, setPage] = useState(1)
  const [initialLoading, setInitialLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const sentinelRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const busyRef = useRef(false)
  const refreshingRef = useRef(false)
  const hasMore = items.length < count

  // 创建/续租弹窗
  const [modalOpen, setModalOpen] = useState(false)
  const [mode, setMode] = useState<'create' | 'renew'>('create')
  const [currToken, setCurrToken] = useState('')
  const [num, setNum] = useState('7')
  const [unit, setUnit] = useState<Unit>('day')
  const [usage, setUsage] = useState('')
  const [saving, setSaving] = useState(false)

  const unitOptions = useMemo(
    () => [
      { value: 'month', label: t('tokens.unitMonth') },
      { value: 'day', label: t('tokens.unitDay') },
      { value: 'hour', label: t('tokens.unitHour') },
      { value: 'minute', label: t('tokens.unitMinute') },
      { value: 'second', label: t('tokens.unitSecond') },
    ],
    [t],
  )

  const fetchList = useCallback(async (p: number, append = false) => {
    busyRef.current = true
    if (append) setLoadingMore(true)
    else setInitialLoading(true)
    try {
      const { data, error } = await api.GET('/api/access_tokens', {
        params: { query: { page: p, pageSize: PAGE_SIZE } },
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
  }, [])

  useEffect(() => {
    // 刷新已直接拉取第 1 页，跳过 setPage(1) 触发的重复请求
    if (page === 1 && refreshingRef.current) return
    void fetchList(page, page > 1)
  }, [page, fetchList])

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
    await fetchList(1, false)
    busyRef.current = false
    refreshingRef.current = false
    setRefreshing(false)
  }

  const copy = async (token: string) => {
    const ok = await copyText(token)
    if (ok) toast.success(t('tokens.copySuccess'))
    else toast.error(t('common.retry'))
  }

  const revoke = async (item: TokenModel) => {
    try {
      const { error } = await api.DELETE('/api/access_tokens/{token}', {
        params: { path: { token: item.token } },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('tokens.revokeSuccess', { name: item.usage || item.token }))
      void refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    }
  }

  const openCreate = () => {
    setMode('create')
    setCurrToken('')
    setNum('7')
    setUnit('day')
    setUsage('')
    setModalOpen(true)
  }

  const openRenew = (token: string) => {
    setMode('renew')
    setCurrToken(token)
    setNum('7')
    setUnit('day')
    setUsage('')
    setModalOpen(true)
  }

  const submit = async () => {
    const n = Number(num)
    if (!Number.isInteger(n) || n < 1) {
      toast.error(t('tokens.validityMin', { unit: t(UNIT_KEY[unit]) }))
      return
    }
    if (mode === 'create' && !usage.trim()) {
      toast.error(t('tokens.usageRequired'))
      return
    }
    const expireSeconds = getSeconds(n, unit)
    setSaving(true)
    try {
      if (mode === 'create') {
        const { error } = await api.POST('/api/access_tokens', {
          body: { usage: usage.trim(), expireSeconds },
        })
        if (error) throw new Error(error.message ?? String(error))
        toast.success(t('tokens.createSuccess'))
      } else {
        const { error } = await api.PUT('/api/access_tokens/{token}', {
          body: { token: currToken, expireSeconds },
          params: { path: { token: currToken } },
        })
        if (error) throw new Error(error.message ?? String(error))
        toast.success(t('tokens.renewSuccess'))
      }
      setModalOpen(false)
      void refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setSaving(false)
    }
  }

  const renderTags = (item: TokenModel) => {
    if (item.isDeleted) return <Tag tone="err">{t('tokens.statusRevoked')}</Tag>
    if (item.isExpired) return <Tag tone="mute">{t('tokens.statusExpired')}</Tag>
    const exp = dayjs(item.expiredAt)
    if (exp.isBefore(dayjs().add(1, 'day')) && exp.isAfter(dayjs())) {
      return <Tag tone="warn">{t('tokens.statusExpiring')}</Tag>
    }
    return null
  }

  return (
    // 父级 main 是 min-h-screen 下的 flex-1（高度 auto），flex 链没有有界高度，
    // 用 calc(100dvh - 100px) 让内部滚动容器 flex-1 min-h-0 overflow-y-auto 真正可滚动。
    <div className="flex h-[calc(100dvh_-_100px)] flex-col gap-3">
      {/* 工具栏 */}
      <div className="flex shrink-0 flex-col gap-3">
        {/* 标题 + 刷新 + 创建 */}
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-[16px] font-semibold text-ink">{t('tokens.list')}</h2>
          <div className="flex flex-wrap items-center gap-2">
            <Button size="sm" variant="outline" disabled={refreshing} onClick={refresh}>
              <Icon name="refresh" className="size-4" />
              {t('common.refresh')}
            </Button>
            <Button size="sm" variant="default" onClick={openCreate}>
              <Icon name="plus" className="size-4" />
              {t('tokens.create')}
            </Button>
          </div>
        </div>
      </div>

      {/* 列表：flex 撑满剩余高度，容器内滚动（对齐 repos/events 无限加载） */}
      <div
        ref={scrollRef}
        className="min-h-0 flex-1 overflow-y-auto rounded-lg border border-line bg-surface"
      >
        {initialLoading ? (
          <SkeletonList count={8} bare />
        ) : items.length === 0 ? (
          <div className="flex h-full items-center justify-center p-8">
            <Empty text={t('common.empty')} icon="key" />
          </div>
        ) : (
          <div className="divide-y divide-line">
            {items.map((item, index) => (
            <div
              key={item.token}
              // 进入动画：淡入 + 上浮，按行错峰（封顶 10 行×30ms，深滚加载的行不滞后）
              className="animate-list-in flex items-start justify-between gap-4 px-5 py-4 transition-colors hover:bg-raised"
              style={{ animationDelay: `${Math.min(index, 10) * 30}ms` }}
            >
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-[13px] font-medium text-ink">{item.usage}</span>
                  {renderTags(item)}
                </div>
                <div
                  className={`mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-[12px] text-mute ${
                    item.isDeleted || item.isExpired ? 'opacity-60 line-through' : ''
                  }`}
                >
                  <button
                    type="button"
                    onClick={() => copy(item.token)}
                    title={item.token}
                    className="inline-flex max-w-full items-center gap-1.5 rounded bg-raised px-1.5 py-0.5 font-mono text-[12px] text-primary transition-colors hover:bg-primary-soft"
                  >
                    <Icon name="copy" className="shrink-0 text-[11px]" />
                    <span className="truncate" translate="no">{item.token}</span>
                  </button>
                  <span>
                    {t('tokens.expiresAt')} {formatDateTime(item.expiredAt)}
                  </span>
                  <span>{item.lastUsedAt ? `${item.lastUsedAt} ${t('tokens.used')}` : t('tokens.neverUsed')}</span>
                </div>
              </div>
              {!item.isDeleted && !item.isExpired && (
                <div className="flex shrink-0 items-center gap-2">
                  <Button size="sm" variant="outline" onClick={() => openRenew(item.token)}>
                    {t('tokens.renew')}
                  </Button>
                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <Button size="sm" variant="destructive">
                        {t('tokens.revoke')}
                      </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>
                          {t('tokens.revokeConfirm', { name: item.usage || item.token })}
                        </AlertDialogTitle>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
                        <AlertDialogAction
                          className="bg-destructive text-white hover:bg-destructive/90"
                          onClick={() => revoke(item)}
                        >
                          {t('tokens.revoke')}
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </div>
              )}
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

      {/* 创建 / 续租弹窗 */}
      <Dialog open={modalOpen} onOpenChange={(o) => !o && setModalOpen(false)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{mode === 'create' ? t('tokens.create') : t('tokens.renew')}</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4">
            <label className="flex flex-col gap-1.5">
              <span className="text-[12px] font-medium text-mute">{t('tokens.validity')}</span>
              <div className="flex items-stretch gap-2">
                <Input
                  type="number"
                  min={1}
                  value={num}
                  onChange={(e) => setNum(e.target.value)}
                />
                <div className="w-28">
                  <Select value={unit} onValueChange={(v) => setUnit(v as Unit)}>
                    <SelectTrigger className="w-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {unitOptions.map((o) => (
                        <SelectItem key={o.value} value={o.value}>
                          {o.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </label>
            {mode === 'create' ? (
              <label className="flex flex-col gap-1.5">
                <span className="text-[12px] font-medium text-mute">{t('tokens.usage')}</span>
                <Textarea
                  value={usage}
                  onChange={(e) => setUsage(e.target.value)}
                  maxLength={30}
                  rows={2}
                  className="min-h-16"
                />
              </label>
            ) : (
              <div className="break-all rounded-md bg-raised px-3 py-2 font-mono text-[12px] text-primary">
                {currToken}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setModalOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button variant="default" disabled={saving} onClick={submit}>
              {saving && <Loader2 className="size-4 animate-spin" />}
              {t('tokens.submit')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
