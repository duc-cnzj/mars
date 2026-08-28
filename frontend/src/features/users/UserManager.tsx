import { memo, useCallback, useEffect, useMemo, useRef, useState, type CSSProperties } from 'react'
import { useTranslation } from 'react-i18next'
import { Icon } from '@/components/Icons'
import { SearchInput } from '@/components/SearchInput'
import { StatCard } from '@/components/StatCard'
import { Empty, RefreshFade, SkeletonList, Tag } from '@/components/ui'
import { Avatar, AvatarFallback } from '@/components/ui/shadcn/avatar'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/shadcn/alert-dialog'
import { Button } from '@/components/ui/shadcn/button'
import { copyText } from '@/lib/copy'
import { formatDateTime } from '@/lib/format'
import { humanizeDateTime } from '@/lib/humanizeDateTime'
import { toast } from '@/lib/toast'
import { api } from '@/api/client'
import type { components } from '@/api/schema'
import type { TKey } from '@/i18n/keys'

type UserModel = components['schemas']['user.UserModel']
type UserStats = components['schemas']['user.UserStats']

/** 单次拉取上限：百级用户全量回填后本地排序（对齐 governance 服务端内存分页语义，
 *  搜索/角色过滤走服务端，最近登录排序无服务端支持故本地做） */
const FETCH_LIMIT = 100

/** 无限下拉滚动揭示块大小：一次拉全量后本地排序，滚动到底逐块揭示 */
const CHUNK = 20

/** 超级管理员邮箱（对齐真实系统 SuperAdminEmail，恒为管理员、不可降级） */
const SUPER_ADMIN_EMAIL = '1025434218@qq.com'

/** 角色 → 词条键（服务端已归一化为 admin/user 两种取值） */
const roleKey = (role: string): TKey => (role === 'admin' ? 'users.roleAdmin' : 'users.roleUser')

/** 用户行（React.memo）：行级 props 全部稳定（user 引用 + useCallback handler），
 *  列表状态变化（关键词输入/只看管理员/排序切换/滚动揭示）时行不重渲染，杜绝 O(n) 整表刷新。 */
const UserRow = memo(function UserRow({
  user,
  isOpen,
  toggling,
  onToggle,
  onClose,
  onConfirm,
  onCopyEmail,
  className,
  style,
}: {
  user: UserModel
  isOpen: boolean
  toggling: boolean
  onToggle: (u: UserModel) => void
  onClose: () => void
  onConfirm: () => void
  onCopyEmail: (email: string) => void
  /** RefreshFade 经 cloneElement 注入的渐入 class/延迟——须转发到根元素才生效 */
  className?: string
  style?: CSSProperties
}) {
  const { t } = useTranslation()
  const isAdmin = user.roles.includes('admin')
  const isSuper = user.email === SUPER_ADMIN_EMAIL
  return (
    <div
      className={`grid grid-cols-1 gap-2 border-b border-line px-4 py-2.5 last:border-b-0 sm:grid-cols-2 lg:grid-cols-[minmax(0,2fr)_minmax(0,1.5fr)_7rem_8rem] lg:items-center ${className ?? ''}`}
      style={style}
    >
      {/* 用户：头像（系统无头像字段，统一图标占位）+ 姓名 + 邮箱 */}
      <div className="flex min-w-0 items-center gap-2.5">
        <Avatar className="size-8 shrink-0">
          <AvatarFallback className="bg-primary-soft text-primary">
            <Icon name="user" className="size-4" />
          </AvatarFallback>
        </Avatar>
        <div className="min-w-0">
          <div className="truncate text-[13px] font-medium text-ink">{user.name}</div>
          <div className="flex min-w-0 items-center gap-1">
            <span className="truncate font-mono text-[11px] text-faint">{user.email}</span>
            <button
              type="button"
              onClick={() => onCopyEmail(user.email)}
              aria-label={t('common.copy')}
              title={t('common.copy')}
              className="shrink-0 text-faint opacity-60 transition-opacity hover:text-ink hover:opacity-100"
            >
              <Icon name="copy" className="size-3" />
            </button>
          </div>
        </div>
      </div>

      {/* 角色：超级管理员单独打标，其余按角色 tag */}
      <div className="flex flex-wrap items-center gap-1">
        {isSuper && (
          <Tag tone="accent" dot={false}>
            {t('users.superAdmin')}
          </Tag>
        )}
        {user.roles.map((r) => (
          <Tag key={r} tone={r === 'admin' ? 'accent' : 'mute'} dot={false}>
            {t(roleKey(r))}
          </Tag>
        ))}
      </div>

      {/* 最近登录：从未登录显示占位；有登录记录则相对时间为主（humanizeDateTime 跟随
          locale），精确时间放原生 tooltip 可溯源。lastLogin 可能缺失（后端省略字段），
          formatDateTime/humanizeDateTime 对非法输入均返回空串，不再抛 RangeError */}
      <span className="text-[11px] text-ink">
        {user.lastLogin ? (
          <time dateTime={user.lastLogin} title={formatDateTime(user.lastLogin)}>
            {humanizeDateTime(user.lastLogin)}
          </time>
        ) : (
          <span className="text-faint">{t('users.neverLoggedIn')}</span>
        )}
      </span>

      {/* 操作：添加/删除管理员（二次确认；超级管理员不可降级） */}
      {isSuper ? (
        <Button size="sm" variant="ghost" disabled title={t('users.superAdminCannotDemote')}>
          <Icon name="shield" className="size-3.5" />
          {t('users.demote')}
        </Button>
      ) : isAdmin ? (
        <AlertDialog open={isOpen} onOpenChange={(o) => !o && !toggling && onClose()}>
          <Button size="sm" variant="outline" onClick={() => onToggle(user)}>
            <Icon name="minus" className="size-3.5" />
            {t('users.demote')}
          </Button>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{t('users.demoteConfirmTitle', { name: user.name })}</AlertDialogTitle>
              <AlertDialogDescription>{t('users.demoteConfirmDesc')}</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={toggling}>{t('common.cancel')}</AlertDialogCancel>
              {/* 普通 Button 而非 AlertDialogAction：后者点击默认关闭弹窗，会「还没生效就消失」 */}
              <Button variant="destructive" disabled={toggling} onClick={onConfirm}>
                {toggling && <Icon name="loader" className="size-4 animate-spin" />}
                {t('users.demote')}
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      ) : (
        <AlertDialog open={isOpen} onOpenChange={(o) => !o && !toggling && onClose()}>
          <Button size="sm" variant="default" onClick={() => onToggle(user)}>
            <Icon name="shield" className="size-3.5" />
            {t('users.promote')}
          </Button>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{t('users.promoteConfirmTitle', { name: user.name })}</AlertDialogTitle>
              <AlertDialogDescription>{t('users.promoteConfirmDesc')}</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={toggling}>{t('common.cancel')}</AlertDialogCancel>
              {/* 普通 Button 而非 AlertDialogAction：后者点击默认关闭弹窗，会「还没生效就消失」 */}
              <Button disabled={toggling} onClick={onConfirm}>
                {toggling && <Icon name="loader" className="size-4 animate-spin" />}
                {t('users.promote')}
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
    </div>
  )
})

/**
 * 用户管理（管理员后台）
 *
 * 数据来自 /api/admin/users（服务端搜索 + 只看管理员过滤 + 全量口径统计）：
 * - 顶部三卡统计来自服务端 stats（全量口径，不受搜索/角色过滤影响）
 * - 关键词搜索（300ms 防抖）与「只看管理员」均走服务端 query
 * - 最近登录列点击切换升/降序（本地排序，服务端无此能力）；行内一键复制邮箱
 * - 行内「设为管理员 / 移除管理员」走角色管理接口（PUT /api/admin/users/{email}/role），
 *   均带 AlertDialog 二次确认（超级管理员不可降级）
 * - 「同步用户」走 POST /api/admin/users/sync：把内置管理员 + 命名空间成员对账为 users
 *   投影（幂等可重复调用），带 AlertDialog 二次确认，成功后自动刷新列表拉最新投影
 * - 无限下拉（客户端滚动揭示）：一次拉全量 + 本地排序后逐块揭示，搜索/筛选/排序变化自动回顶部
 * - ⚠️ 系统无封号/禁用状态，本页不提供该能力
 */
export function UserManager() {
  const { t } = useTranslation()
  const [users, setUsers] = useState<UserModel[]>([])
  const [count, setCount] = useState(0)
  const [stats, setStats] = useState<UserStats>({ total: 0, admins: 0, regular: 0 })
  const [keyword, setKeyword] = useState('')
  // 防抖后的关键词：避免每次击键都打后端
  const [debouncedKeyword, setDebouncedKeyword] = useState('')
  const [adminOnly, setAdminOnly] = useState(false)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  // 同步用户进行中（防重复点击；同步完成后自动 refresh 拉最新投影）
  const [syncing, setSyncing] = useState(false)
  // 同步确认弹窗开关：点「同步用户」先二次确认再执行（受控弹窗，请求成功才关闭）
  const [syncOpen, setSyncOpen] = useState(false)
  const [error, setError] = useState('')
  // 手动刷新计数：递增触发 fetch effect 重跑
  const [reloadKey, setReloadKey] = useState(0)
  // 渐入版本号：取数成功 +1，RefreshFade 依 key 重挂载列表重播渐入
  const [version, setVersion] = useState(0)
  // 最近登录列排序方向（默认降序 = 最近登录在前，点击表头切换升/降）
  const [loginSort, setLoginSort] = useState<'asc' | 'desc'>('desc')
  // 提权/降权确认弹窗：null=关闭；非空=确认变更该用户管理员角色（受控弹窗，请求成功前不关闭）
  const [toggleTarget, setToggleTarget] = useState<UserModel | null>(null)
  // 角色变更进行中（防重复点击；进行中禁止关闭弹窗，失败保持打开可重试）
  const [toggling, setToggling] = useState(false)
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

  // 拉取用户清单：搜索/只看管理员过滤交给服务端，stats 为全量口径统计
  useEffect(() => {
    let ignore = false
    setLoading(true)
    void api
      .GET('/api/admin/users', {
        params: {
          query: {
            page: 1,
            pageSize: FETCH_LIMIT,
            search: debouncedKeyword.trim() || undefined,
            role: adminOnly ? 'admin' : undefined,
          },
        },
      })
      .then(({ data, error: err }) => {
        if (ignore) return
        if (err) {
          setError(err.message ?? String(err))
          setUsers([])
          return
        }
        setError('')
        if (!data) return
        setUsers(data.items)
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
  }, [debouncedKeyword, adminOnly, reloadKey])

  /** 手动刷新：拉取最新一版用户清单（useCallback 稳定引用，供行 memo 与 toggleAdmin 复用） */
  const refresh = useCallback(() => {
    setRefreshing(true)
    setReloadKey((k) => k + 1)
  }, [])

  /** 同步用户：把真实身份源（内置管理员/命名空间成员）同步为 users 投影，幂等可重复调用。
   *  受控弹窗：请求成功才关闭，失败保持打开（toast 报错）供重试，杜绝「没同步完弹窗先消失」；
   *  完成后自动刷新列表展示最新投影。 */
  const syncUsers = useCallback(async () => {
    if (syncing) return
    setSyncing(true)
    try {
      const { error: err } = await api.POST('/api/admin/users/sync', {})
      if (err) throw new Error(err.message ?? String(err))
      toast.success(t('users.syncSuccess'))
      setSyncOpen(false)
      refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setSyncing(false)
    }
  }, [refresh, syncing, t])

  // 搜索 / 只看管理员 / 手动刷新 / 排序方向变化 → 回顶部并重置揭示量
  //（排序切换走本地重排，无需重新拉取，故不进入 fetch effect 依赖）
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: 0 })
    setVisible(CHUNK)
  }, [debouncedKeyword, adminOnly, reloadKey, loginSort])

  // 按最近登录排序（默认最近登录在前）。lastLogin 缺失（从未登录）→ 排序基准 NaN，
  // 排序时 desc 映射 0 / asc 映射 +Infinity：升、降序两个方向从未登录者都沉底
  //（asc = 最旧登录在前，从未登录应垫底而非置顶；diff 为 NaN 时 Array.sort 视为相等保持原序）。
  // 辅助逻辑内联进 useMemo（闭包只依赖 loginSort，依赖数组齐整，不悬挂组件体内 helper）
  const sorted = useMemo(() => {
    const base = (s: string | undefined): number => {
      if (!s) return Number.NaN
      const ts = new Date(s).getTime()
      return Number.isNaN(ts) ? Number.NaN : ts
    }
    const toSort = (s: string | undefined): number => {
      const ts = base(s)
      if (Number.isNaN(ts)) return loginSort === 'asc' ? Number.POSITIVE_INFINITY : 0
      return ts
    }
    return [...users].sort((a, b) => {
      const diff = toSort(a.lastLogin) - toSort(b.lastLogin)
      return loginSort === 'desc' ? -diff : diff
    })
  }, [users, loginSort])

  // 滚动揭示切片 + 是否还有未揭示行（揭示完毕哨兵显示「没有更多了」）
  const revealed = sorted.slice(0, visible)
  const hasMore = visible < sorted.length

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

  // 弹窗目标/进行中状态同步 ref：供 useCallback 稳定读取最新值，避免 handler 每渲染重建令行 memo 失效
  const toggleTargetRef = useRef<UserModel | null>(null)
  const togglingRef = useRef(false)
  toggleTargetRef.current = toggleTarget
  togglingRef.current = toggling

  /** 添加/删除管理员：调角色管理接口（超级管理员不可降级由服务端保证）。
   *  受控弹窗：请求成功才关闭，失败保持打开（toast 报错）供重试，杜绝「还没生效弹窗先消失」。
   *  useCallback 稳定引用：行 memo 的 onConfirm prop 不因父级状态抖动而重建。 */
  const toggleAdmin = useCallback(async () => {
    const target = toggleTargetRef.current
    if (!target || togglingRef.current) return
    const makeAdmin = !target.roles.includes('admin')
    setToggling(true)
    try {
      const { error: err } = await api.PUT('/api/admin/users/{email}/role', {
        params: { path: { email: target.email } },
        body: { email: target.email, admin: makeAdmin },
      })
      if (err) throw new Error(err.message ?? String(err))
      toast.success(t(makeAdmin ? 'users.promoteSuccess' : 'users.demoteSuccess'))
      setToggleTarget(null)
      void refresh()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : String(e))
    } finally {
      setToggling(false)
    }
  }, [refresh, t])

  /** 复制用户邮箱（邀请成员/联系用户），成功 toast 反馈（useCallback 稳定引用） */
  const copyEmail = useCallback(async (email: string) => {
    const ok = await copyText(email)
    if (ok) toast.success(t('users.copyEmail'))
    else toast.error(t('common.retry'))
  }, [t])

  /** 打开提权/降权确认弹窗（setToggleTarget 恒稳定，useCallback 仅为行 memo 的 prop 引用一致） */
  const handleToggle = useCallback((u: UserModel) => setToggleTarget(u), [])

  /** 关闭确认弹窗（受控弹窗 onOpenChange(false) → 关） */
  const handleClose = useCallback(() => setToggleTarget(null), [])

  return (
    <div className="flex h-full flex-col gap-4">
      {/* 页头：标题 + 刷新 */}
      <div className="flex shrink-0 flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <h2 className="text-[16px] font-semibold text-ink">{t('users.title')}</h2>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={refresh} disabled={refreshing}>
            {refreshing ? (
              <Icon name="loader" className="size-4 animate-spin" />
            ) : (
              <Icon name="refresh" className="size-4" />
            )}
            {t('common.refresh')}
          </Button>
          <Button variant="outline" size="sm" onClick={() => setSyncOpen(true)} disabled={syncing}>
            {syncing ? (
              <Icon name="loader" className="size-4 animate-spin" />
            ) : (
              <Icon name="refresh-cw" className="size-4" />
            )}
            {t('users.sync')}
          </Button>
        </div>
      </div>

      {/* 顶部三卡统计（服务端全量口径） */}
      <div className="grid shrink-0 grid-cols-1 gap-3 sm:grid-cols-3">
        <StatCard label={t('users.total')} value={stats.total} icon="users" tone="mute" />
        <StatCard label={t('users.admins')} value={stats.admins} icon="shield" tone="accent" />
        <StatCard label={t('users.regular')} value={stats.regular} icon="user" tone="ok" />
      </div>

      {/* 工具栏：搜索 + 一键只看管理员 + 结果计数（SearchInput 内置 ⌘K 聚焦快捷键） */}
      <div className="flex shrink-0 flex-wrap items-center gap-3">
        <SearchInput
          value={keyword}
          onChange={setKeyword}
          placeholder={t('users.searchPlaceholder')}
          className="w-full sm:w-72"
        />
        <Button
          size="sm"
          variant={adminOnly ? 'default' : 'outline'}
          aria-pressed={adminOnly}
          onClick={() => setAdminOnly((v) => !v)}
        >
          <Icon name="shield" className="size-3.5" />
          {t('users.filterAdmin')}
        </Button>
        <span className="text-[12px] text-faint">{t('users.resultCount', { count })}</span>
      </div>

      {/* 用户列表：固定表头 + 内部滚动容器（无限下拉 root） */}
      <section className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-line bg-surface">
        <div className="hidden grid-cols-[minmax(0,2fr)_minmax(0,1.5fr)_7rem_8rem] items-center gap-2 border-b border-line px-4 py-2 text-[11px] font-medium text-faint lg:grid">
          <span>{t('users.user')}</span>
          <span>{t('users.role')}</span>
          <button
            type="button"
            onClick={() => setLoginSort((v) => (v === 'desc' ? 'asc' : 'desc'))}
            className="flex w-fit items-center gap-0.5 hover:text-ink"
            title={
              loginSort === 'desc' ? t('users.sortDesc') : t('users.sortAsc')
            }
          >
            {t('users.lastLogin')}
            <Icon
              name={loginSort === 'desc' ? 'chevron-down' : 'chevron-up'}
              className="size-3 opacity-70"
            />
          </button>
          <span>{t('users.action')}</span>
        </div>

        <div ref={scrollRef} className="min-h-0 flex-1 overflow-y-auto">
        {loading && sorted.length === 0 ? (
          <SkeletonList count={8} bare />
        ) : error ? (
          <Empty icon="alert" text={error} />
        ) : sorted.length === 0 ? (
          <Empty
            icon="users"
            text={keyword ? t('users.searchEmpty', { kw: keyword.trim() }) : t('common.empty')}
          />
        ) : (
          <RefreshFade version={version}>
          {revealed.map((u) => (
            <UserRow
              key={u.email}
              user={u}
              isOpen={toggleTarget?.email === u.email}
              toggling={toggling}
              onToggle={handleToggle}
              onClose={handleClose}
              onConfirm={toggleAdmin}
              onCopyEmail={copyEmail}
            />
          ))}
          </RefreshFade>
        )}
        {/* 无限下拉哨兵：进入视口揭示下一块；揭示完毕显示「没有更多了」 */}
        <div ref={sentinelRef} className="flex h-10 items-center justify-center gap-2">
          {hasMore ? (
            <Icon name="loader" className="size-3.5 animate-spin text-faint" />
          ) : (
            <span className="text-[11px] text-faint">{t('common.noMore')}</span>
          )}
        </div>
        </div>
      </section>

      {/* 同步用户二次确认弹窗：点「同步用户」→ 确认 → 执行（幂等，成功才关闭，失败保持打开可重试） */}
      <AlertDialog open={syncOpen} onOpenChange={(o) => !o && !syncing && setSyncOpen(false)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('users.syncConfirmTitle')}</AlertDialogTitle>
            <AlertDialogDescription>{t('users.syncConfirmDesc')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={syncing}>{t('common.cancel')}</AlertDialogCancel>
            {/* 普通 Button 而非 AlertDialogAction：后者点击默认关闭弹窗，会「没同步完就消失」 */}
            <Button disabled={syncing} onClick={syncUsers}>
              {syncing ? (
                <Icon name="loader" className="size-4 animate-spin" />
              ) : (
                <Icon name="refresh-cw" className="size-4" />
              )}
              {t('users.sync')}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
