import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import dayjs from 'dayjs'
import { formatDateTime } from '@/lib/format'
import { toast } from '@/lib/toast'
import type { components } from '@/api/schema'
import { api } from '@/api/client'
import { Icon } from '@/components/Icons'
import { SearchInput } from '@/components/SearchInput'
import { Empty, SkeletonList, Tag } from '@/components/ui'
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
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/shadcn/alert-dialog'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/shadcn/dialog'
import { copyText } from '@/lib/copy'
import { useAuth } from '@/features/auth/AuthProvider'

type TokenModel = components['schemas']['types.AccessTokenModel']
type Unit = 'month' | 'day' | 'hour' | 'minute' | 'second'

/** 状态过滤取值（与后端 ListRequest.status 对齐）：'all' 不过滤，其余三态走服务端 query */
type TokenStatus = 'all' | 'valid' | 'expired' | 'revoked'

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
  // 管理员可查全平台令牌，普通用户服务端已收敛到本人 → 「我创建的」收敛按钮仅管理员可见
  const { user } = useAuth()
  const isAdmin = user?.roles.includes('mars_admin') ?? false

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
  // 按创建人搜索（admin 可查全平台令牌，普通用户仅本人；搜索走服务端 query）
  const [search, setSearch] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  // 「我创建的」快捷筛选：admin 全量视图收敛到本人令牌（与服务端 all 反极性映射，见 fetchList）；
  // 仅管理员可见——普通用户服务端本就只返回本人令牌，无需收敛入口
  const [mineOnly, setMineOnly] = useState(false)
  // 状态过滤：全部/有效/已过期/已撤销（走服务端 query，与搜索/「我创建的」正交，全用户可用）
  const [statusFilter, setStatusFilter] = useState<TokenStatus>('all')
  // 请求去重基准：记录上次真正发起的 (过滤条件, 页码)，跳过「新条件×旧页码」过期请求与重复请求。
  // 初始值必须「不可能等于」首屏状态——null ≠ '' 的 filterKey、0 ≠ 1 的 page，
  // 否则首屏挂载被去重判定吞掉，列表永远不请求（骨架屏常驻）
  const lastFilterKeyRef = useRef<string | null>(null)
  const lastPageRef = useRef(0)
  /** 最新过滤条件快照（渲染期同步）：fetchList 落地时校验响应是否仍属当前条件，
   *  丢弃「旧条件追加页晚到」的跨条件污染（对齐 NamespaceManager filterKeyRef） */
  const filterKeyRef = useRef('')
  const hasMore = items.length < count

  // 创建/续租弹窗
  const [modalOpen, setModalOpen] = useState(false)
  const [mode, setMode] = useState<'create' | 'renew'>('create')
  const [currToken, setCurrToken] = useState('')
  const [num, setNum] = useState('7')
  const [unit, setUnit] = useState<Unit>('day')
  const [usage, setUsage] = useState('')
  const [saving, setSaving] = useState(false)
  // 撤销确认弹窗：null=关闭；非空=确认撤销该令牌（受控弹窗，撤销成功前不关闭）
  const [revokeTarget, setRevokeTarget] = useState<TokenModel | null>(null)
  // 撤销进行中（防重复点击；进行中禁止关闭弹窗，失败保持打开可重试）
  const [revoking, setRevoking] = useState(false)

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

  const fetchList = useCallback(
    async (p: number, append = false) => {
      busyRef.current = true
      if (append) setLoadingMore(true)
      else setInitialLoading(true)
      try {
        const { data, error } = await api.GET('/api/access_tokens', {
          params: {
            query: {
              page: p,
              pageSize: PAGE_SIZE,
              search: debouncedSearch.trim() || undefined,
              // all=true = 管理员看全平台令牌；点「我创建的」收敛本人（all=false）。极性反转修复：
              // 后端字段是 all（true=展开全部）而非 mineOnly（true=只看本人），旧代码发 mineOnly
              // 被 grpc-gateway 静默忽略 → admin 全平台视图不可达且「我创建的」是 no-op。普通用户
              // 服务端本就收敛本人，不发 all（undefined）。
              all: isAdmin ? !mineOnly : undefined,
              status: statusFilter === 'all' ? undefined : statusFilter,
            },
          },
        })
        if (error) throw new Error(error.message ?? String(error))
        if (!data) return
        // 过期响应（搜索/「我创建的」/状态已切换、此响应属于旧条件）：丢弃不落地——触底加载的追加页
        // 晚到时不能把旧条件的下一页 append 进新条件结果，否则跨条件污染持续到下次取数
        if (filterKeyRef.current !== filterKey) return
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
    [debouncedSearch, mineOnly, statusFilter, isAdmin],
  )

  // 搜索防抖 300ms：停顿后才把关键词交给 fetchList
  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedSearch(search), 300)
    return () => window.clearTimeout(timer)
  }, [search])

  /** 搜索指纹：关键词/「我创建的」/状态过滤/是否管理员任一变化都视为新的过滤条件
   *  （配合页码去重 refs；isAdmin 影响 all 参数——未受 RequireAuth 门控的挂载（如冒烟）
   *   在会话解析后 isAdmin 翻转，须按新 all 值重拉第 1 页） */
  const filterKey = `${debouncedSearch}|${mineOnly}|${statusFilter}|${isAdmin}`
  // 渲染期同步最新条件（供 fetchList 落地校验丢弃旧条件响应，见 filterKeyRef 声明注释）
  filterKeyRef.current = filterKey

  // 搜索条件变化 → 回到顶部 + 重置第 1 页（由 page 变化驱动下面的 fetch）
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: 0 })
    setPage(1)
  }, [filterKey])

  useEffect(() => {
    // 刷新已直接拉取第 1 页，跳过 setPage(1) 触发的重复请求
    if (page === 1 && refreshingRef.current) return
    // 过滤条件刚变、page 尚未重置到 1（filterKey effect 已 setPage(1)，本 commit 里拿到的还是旧 page）：
    // 跳过这次「新条件×旧页码」的过期请求，真正的第 1 页由 setPage(1) 触发的下一次 effect 承担
    if (lastFilterKeyRef.current !== filterKey && page !== 1) return
    // 同 (filterKey, page) 去重：刷新完成后的 effect 重跑不会重复拉取
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

  const refresh = async () => {
    if (refreshingRef.current) return
    refreshingRef.current = true
    setRefreshing(true)
    // 重置期间同步锁 busyRef，阻止触底加载在刷新时连环翻页
    busyRef.current = true
    // 回到顶部：让哨兵离开视口，刷新后不会自动加载后续页
    if (scrollRef.current) scrollRef.current.scrollTop = 0
    setPage(1)
    // 预标记本次刷新要拉的 (filterKey, 1)：刷新完成后的 effect 重跑据此去重、不重复拉第 1 页，
    // 且后续触底翻页从第 2 页继续（对齐 Events/Repos 同款——缺此标记则 page>1 刷新后触底
    // setPage 被「同 (filterKey,page) 去重」guard 吞掉，列表永久停在第 1 页）
    lastFilterKeyRef.current = filterKey
    lastPageRef.current = 1
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

  /** 撤销令牌：受控弹窗，DELETE 成功才关闭；失败保持打开（toast 报错）供重试，杜绝「还没撤销成功弹窗先消失」 */
  const revoke = async () => {
    if (!revokeTarget || revoking) return
    setRevoking(true)
    try {
      const { error } = await api.DELETE('/api/access_tokens/{token}', {
        params: { path: { token: revokeTarget.token } },
      })
      if (error) throw new Error(error.message ?? String(error))
      toast.success(t('tokens.revokeSuccess', { name: revokeTarget.usage || revokeTarget.token }))
      setRevokeTarget(null)
      void refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setRevoking(false)
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
    // 双布局共享：/admin/tokens（AdminLayout）父级有界（h-full 生效），/tokens（AppLayout）
    // 父级 main 高度 auto、flex 链无有界高度（h-full 退化）——max-h 兜底视口高度上限，
    // 两种上下文下内部滚动容器 flex-1 min-h-0 overflow-y-auto 都能真正滚动且不产生双滚动条。
    <div className="flex h-full max-h-[calc(100dvh_-_100px)] flex-col gap-3">
      {/* 工具栏 */}
      <div className="flex shrink-0 flex-col gap-3">
        {/* 标题 + 刷新 + 创建 */}
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-[16px] font-semibold text-ink">{t('tokens.list')}</h2>
          <div className="flex flex-wrap items-center gap-2">
            <Button size="sm" variant="outline" disabled={refreshing} onClick={refresh}>
              {refreshing ? (
                <Icon name="loader" className="size-4 animate-spin" />
              ) : (
                <Icon name="refresh" className="size-4" />
              )}
              {t('common.refresh')}
            </Button>
            <Button size="sm" variant="default" onClick={openCreate}>
              <Icon name="plus" className="size-4" />
              {t('tokens.create')}
            </Button>
          </div>
        </div>
        {/* 按创建人搜索（admin 看全平台 / 普通用户仅本人，搜索走服务端 query） +
            我创建的快捷筛选（仅管理员可见：全平台视图收敛到本人令牌，与服务端 mine_only 对齐） */}
        <div className="flex flex-wrap items-center gap-2">
          <SearchInput
            value={search}
            onChange={setSearch}
            placeholder={t('tokens.searchPlaceholder')}
            className="w-72"
            size="sm"
          />
          {/* 状态过滤：走服务端 query（分页列表前端过滤只会滤当前已加载页、count/hasMore 全错），
              与搜索/「我创建的」正交；全用户可用（普通用户过滤本人的令牌） */}
          <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v as TokenStatus)}>
            <SelectTrigger size="sm" className="w-32" aria-label={t('tokens.statusFilter')}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t('tokens.statusAll')}</SelectItem>
              <SelectItem value="valid">{t('tokens.statusValid')}</SelectItem>
              <SelectItem value="expired">{t('tokens.statusExpired')}</SelectItem>
              <SelectItem value="revoked">{t('tokens.statusRevoked')}</SelectItem>
            </SelectContent>
          </Select>
          {isAdmin && (
            <Button
              size="sm"
              variant={mineOnly ? 'default' : 'outline'}
              aria-pressed={mineOnly}
              title={t('tokens.mineOnlyTip')}
              onClick={() => setMineOnly((v) => !v)}
            >
              <Icon name="user" className="size-3.5" />
              {t('tokens.mineOnly')}
            </Button>
          )}
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
                    title={t('tokens.copyTip')}
                    className="inline-flex max-w-full items-center gap-1.5 rounded bg-raised px-1.5 py-0.5 font-mono text-[12px] text-primary transition-colors hover:bg-primary-soft"
                  >
                    <Icon name="copy" className="shrink-0 text-[11px]" />
                    {/* 展示后端返回的完整令牌：复制/撤销/续租都直接消费它，掩码会让三个功能全废 */}
                    <span className="truncate" translate="no">{item.token}</span>
                  </button>
                  <span>
                    {t('tokens.expiresAt')} {formatDateTime(item.expiredAt)}
                  </span>
                  {/* 创建人仅管理员可见：普通用户服务端只返回本人令牌，创建人恒为本人，展示无意义还暴露邮箱 */}
                  {isAdmin && (
                    <span className="inline-flex items-center gap-1">
                      {t('tokens.creator')}
                      <span className="font-mono text-mute">{item.email}</span>
                    </span>
                  )}
                  <span>{item.lastUsedAt ? `${item.lastUsedAt} ${t('tokens.used')}` : t('tokens.neverUsed')}</span>
                </div>
              </div>
              {!item.isDeleted && !item.isExpired && (
                <div className="flex shrink-0 items-center gap-2">
                  <Button size="sm" variant="outline" onClick={() => openRenew(item.token)}>
                    {t('tokens.renew')}
                  </Button>
                  <AlertDialog
                    open={revokeTarget?.token === item.token}
                    onOpenChange={(o) => !o && !revoking && setRevokeTarget(null)}
                  >
                    <Button size="sm" variant="destructive" onClick={() => setRevokeTarget(item)}>
                      {t('tokens.revoke')}
                    </Button>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>
                          {t('tokens.revokeConfirm', { name: item.usage || item.token })}
                        </AlertDialogTitle>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel disabled={revoking}>{t('common.cancel')}</AlertDialogCancel>
                        {/* 普通 Button 而非 AlertDialogAction：后者点击默认关闭弹窗，会「还没撤销成功就消失」 */}
                        <Button variant="destructive" disabled={revoking} onClick={revoke}>
                          {revoking && <Icon name="loader" className="size-4 animate-spin" />}
                          {t('tokens.revoke')}
                        </Button>
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
              {saving && <Icon name="loader" className="size-4 animate-spin" />}
              {t('tokens.submit')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
